/*
 Copyright © 2021 The OpenEBS Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package lvmnode

import (
	"fmt"
	"reflect"
	"time"

	"github.com/openebs/lib-csi/pkg/common/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimenew "k8s.io/apimachinery/pkg/runtime"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/builder/nodebuilder"
	"github.com/openebs/lvm-localpv/pkg/equality"
	"github.com/openebs/lvm-localpv/pkg/lvm"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func (c *NodeController) listLVMVolumeGroup() ([]apis.VolumeGroup, error) {
	return lvm.ListLVMVolumeGroup(true)
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *NodeController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	return c.syncNode(namespace, name)
}

// syncNode is the function which tries to converge to a desired state for the
// LVMNode
func (c *NodeController) syncNode(namespace string, name string) error {
	// Get the node resource with this namespace/name
	cachedNode, err := c.NodeLister.Namespace(namespace).Get(name)
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}

	var node *apis.LVMNode
	if cachedNode != nil {
		nodeStruct, ok := c.getStructuredObject(cachedNode)
		if !ok {
			return errors.Errorf("couldn't get node object %#v", cachedNode)
		}
		node = nodeStruct.DeepCopy()
	}

	vgs, err := c.listLVMVolumeGroup()
	if err != nil {
		return err
	}

	if node == nil { // if it doesn't exists, create lvm node object
		if node, err = nodebuilder.NewBuilder().
			WithNamespace(namespace).WithName(name).
			WithVolumeGroups(vgs).
			WithOwnerReferences(c.ownerRef).
			Build(); err != nil {
			return err
		}

		klog.Infof("lvm node controller: creating new node object for %+v", node)
		if _, err = nodebuilder.NewKubeclient().WithNamespace(namespace).Create(node); err != nil {
			return errors.Errorf("create lvm node %s/%s: %v", namespace, name, err)
		}
		klog.Infof("lvm node controller: created node object %s/%s", namespace, name)
		return nil
	}

	// lvm node already exists check if we need to update it.
	var updateRequired bool
	// validate if owner reference updated.
	if ownerRefs, req := c.isOwnerRefsUpdateRequired(node.OwnerReferences); req {
		klog.Infof("lvm node controller: node owner references updated current=%+v, required=%+v",
			node.OwnerReferences, ownerRefs)
		node.OwnerReferences = ownerRefs
		updateRequired = true
	}

	// validate if node volume groups are upto date.
	if !equality.Semantic.DeepEqual(node.VolumeGroups, vgs) {
		klog.Infof("lvm node controller: node volume groups updated current=%+v, required=%+v",
			node.VolumeGroups, vgs)
		node.VolumeGroups = vgs
		updateRequired = true
	}

	if !updateRequired {
		return nil
	}

	klog.Infof("lvm node controller: updating node object with %+v", node)
	if _, err = nodebuilder.NewKubeclient().WithNamespace(namespace).Update(node); err != nil {
		return errors.Errorf("update lvm node %s/%s: %v", namespace, name, err)
	}
	klog.Infof("lvm node controller: updated node object %s/%s", namespace, name)

	return nil
}

// Obj from queue is not readily in lvmnode type. This function would convert obj into lvmnode type.
func (c *NodeController) getStructuredObject(obj interface{}) (*apis.LVMNode, bool) {
	unstructuredInterface, ok := obj.(*unstructured.Unstructured)
	if !ok {
		runtime.HandleError(errors.Errorf("couldnt type assert obj: %#v to unstructured obj", obj))
		return nil, false
	}
	node := &apis.LVMNode{}
	err := runtimenew.DefaultUnstructuredConverter.FromUnstructured(unstructuredInterface.UnstructuredContent(), &node)
	if err != nil {
		runtime.HandleError(fmt.Errorf("err %s, While converting unstructured obj to typed object\n", err.Error()))
		return nil, false
	}
	return node, true
}

// addNode is the add event handler for LVMNode
func (c *NodeController) addNode(obj interface{}) {
	node, ok := c.getStructuredObject(obj)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get node object %#v", obj))
		return
	}

	klog.Infof("Got add event for lvm node %s/%s", node.Namespace, node.Name)
	c.enqueueNode(node)
}

// updateNode is the update event handler for LVMNode
func (c *NodeController) updateNode(oldObj, newObj interface{}) {
	newNode, ok := c.getStructuredObject(newObj)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get node object %#v", newNode))
		return
	}

	klog.Infof("Got update event for lvm node %s/%s", newNode.Namespace, newNode.Name)
	c.enqueueNode(newNode)
}

// deleteNode is the delete event handler for LVMNode
func (c *NodeController) deleteNode(obj interface{}) {
	node, ok := c.getStructuredObject(obj)
	if !ok {
		unstructuredObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			runtime.HandleError(fmt.Errorf("couldnt type assert obj: %#v to unstructured obj", obj))
			return
		}
		tombStone := cache.DeletedFinalStateUnknown{}
		err := runtimenew.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(), &tombStone)
		if err != nil {
			runtime.HandleError(fmt.Errorf("couldn't get object from tombstone %#v", obj))
			return
		}
		node, ok = tombStone.Obj.(*apis.LVMNode)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a lvmnode %#v", obj))
			return
		}
	}

	klog.Infof("Got delete event for node %s/%s", node.Namespace, node.Name)
	c.enqueueNode(node)
}

// enqueueNode takes a LVMNode resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than LVMNode.
func (c *NodeController) enqueueNode(node *apis.LVMNode) {
	// node must exists in openebs namespace & must equal to the node id.
	if node.Namespace != lvm.LvmNamespace ||
		node.Name != lvm.NodeID {
		klog.Warningf("skipping lvm node object %s/%s", node.Namespace, node.Name)
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(node)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *NodeController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Node controller")

	// Wait for the k8s caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.NodeSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting Node workers")
	// Launch worker to process Node resources
	// Threadiness will decide the number of workers you want to launch to process work items from queue
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started Node workers")

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
		case <-stopCh:
			klog.Info("Shutting down Node controller")
			return nil
		}
		item := lvm.LvmNamespace + "/" + lvm.NodeID
		c.workqueue.Add(item) // add the item to worker queue.
		timer.Reset(c.pollInterval)
	}
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *NodeController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *NodeController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Node resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// isOwnerRefUpdateRequired validates if relevant owner references is being
// set for lvm node. If not, it returns the final owner references that needs
// to be set.
func (c *NodeController) isOwnerRefsUpdateRequired(ownerRefs []metav1.OwnerReference) ([]metav1.OwnerReference, bool) {
	updated := false
	reqOwnerRef := c.ownerRef
	for idx := range ownerRefs {
		if ownerRefs[idx].UID != reqOwnerRef.UID {
			continue
		}
		// in case owner reference exists, validate
		// if controller field is set correctly or not.
		if !reflect.DeepEqual(ownerRefs[idx].Controller, reqOwnerRef.Controller) {
			updated = true
			ownerRefs[idx].Controller = reqOwnerRef.Controller
		}
		return ownerRefs, updated
	}
	updated = true
	ownerRefs = append(ownerRefs, reqOwnerRef)
	return ownerRefs, updated
}
