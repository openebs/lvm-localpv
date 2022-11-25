/*
Copyright 2020 The OpenEBS Authors

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

package volume

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/lvm"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtimenew "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// isDeletionCandidate checks if a lvm volume is a deletion candidate.
func (c *VolController) isDeletionCandidate(Vol *apis.LVMVolume) bool {
	return Vol.ObjectMeta.DeletionTimestamp != nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *VolController) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Vol resource with this namespace/name
	klog.Infof("Getting lvmvol object name:%s, ns:%s from cache\n", name, namespace)
	unstructuredVol, err := c.VolLister.Namespace(namespace).Get(name)
	if k8serror.IsNotFound(err) {
		runtime.HandleError(fmt.Errorf("lvmvolume '%s' has been deleted", key))
		return nil
	}
	if err != nil {
		return err
	}
	vol := &apis.LVMVolume{}
	err = runtimenew.DefaultUnstructuredConverter.FromUnstructured(unstructuredVol.UnstructuredContent(), &vol)
	if err != nil {
		klog.Infof("err %s, While converting unstructured obj to typed object\n", err.Error())
	}
	VolCopy := vol.DeepCopy()
	err = c.syncVol(VolCopy)
	return err
}

// addVol is the add event handler for LVMVolume
func (c *VolController) addVol(obj interface{}) {
	Vol, ok := c.getStructuredObject(obj)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get Vol object %#v", obj))
		return
	}

	if lvm.NodeID != Vol.Spec.OwnerNodeID {
		return
	}
	klog.Infof("Got add event for Vol %s", Vol.Name)
	klog.Infof("lvmvolume object to be enqueued by Add handler: %v", Vol)
	c.enqueueVol(Vol)
}

// updateVol is the update event handler for LVMVolume
func (c *VolController) updateVol(oldObj, newObj interface{}) {
	newVol, ok := c.getStructuredObject(newObj)
	if !ok {
		runtime.HandleError(fmt.Errorf("Couldn't get Vol object %#v", newVol))
		return
	}

	if lvm.NodeID != newVol.Spec.OwnerNodeID {
		return
	}

	if c.isDeletionCandidate(newVol) {
		klog.Infof("Got update event for deleted Vol %s, Deletion timestamp %s", newVol.Name, newVol.ObjectMeta.DeletionTimestamp)
		c.enqueueVol(newVol)
	}
}

// deleteVol is the delete event handler for LVMVolume
func (c *VolController) deleteVol(obj interface{}) {
	Vol, ok := c.getStructuredObject(obj)
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
		Vol, ok = tombStone.Obj.(*apis.LVMVolume)
		if !ok {
			runtime.HandleError(fmt.Errorf("tombstone contained object that is not a lvmvolume %#v", obj))
			return
		}
	}

	if lvm.NodeID != Vol.Spec.OwnerNodeID {
		return
	}

	klog.Infof("Got delete event for Vol %s", Vol.Name)
	c.enqueueVol(Vol)
}

// enqueueVol takes a LVMVolume resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than LVMVolume.
func (c *VolController) enqueueVol(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// Obj from queue is not readily in lvmvol type. This function would convert obj into lvmvolume type.
func (c *VolController) getStructuredObject(obj interface{}) (*apis.LVMVolume, bool) {
	unstructuredInterface, ok := obj.(*unstructured.Unstructured)
	if !ok {
		runtime.HandleError(fmt.Errorf("couldnt type assert obj: %#v to unstructured obj", obj))
		return nil, false
	}
	vol := &apis.LVMVolume{}
	err := runtimenew.DefaultUnstructuredConverter.FromUnstructured(unstructuredInterface.UnstructuredContent(), &vol)
	if err != nil {
		runtime.HandleError(fmt.Errorf("err %s, While converting unstructured obj to typed object\n", err.Error()))
		return nil, false
	}
	return vol, true
}

// synVol is the function which tries to converge to a desired state for the
// LVMVolume
func (c *VolController) syncVol(vol *apis.LVMVolume) error {
	var err error
	// LVM Volume should be deleted. Check if deletion timestamp is set
	if c.isDeletionCandidate(vol) {
		err = lvm.DestroyVolume(vol)
		if err == nil {
			err = lvm.RemoveVolFinalizer(vol)
		}
		return err
	}

	// if status is Pending then it means we are creating the volume.
	// Otherwise, we are just ignoring the event.
	switch vol.Status.State {
	case lvm.LVMStatusFailed:
		klog.Warningf("Skipping retrying lvm volume provisioning as its already in failed state: %+v", vol.Status.Error)
		return nil
	case lvm.LVMStatusReady:
		klog.Info("lvm volume already provisioned")
		return nil
	}

	// if there is already a volGroup field set for lvmvolume resource,
	// we'll first try to create a volume in that volume group.
	if vol.Spec.VolGroup != "" {
		err = lvm.CreateVolume(vol)
		if err == nil {
			return lvm.UpdateVolInfo(vol, lvm.LVMStatusReady)
		}
	}

	vgs, err := c.getVgPriorityList(vol)
	if err != nil {
		return err
	}

	if len(vgs) == 0 {
		err = fmt.Errorf("no vg available to serve volume request having regex=%q & capacity=%q",
			vol.Spec.VgPattern, vol.Spec.Capacity)
		klog.Errorf("lvm volume %v - %v", vol.Name, err)
	} else {
		for _, vg := range vgs {
			// first update volGroup field in lvm volume resource for ensuring
			// idempotency and avoiding volume leaks during crash.
			if vol, err = lvm.UpdateVolGroup(vol, vg.Name); err != nil {
				klog.Errorf("failed to update volGroup to %v: %v", vg.Name, err)
				return err
			}
			if err = lvm.CreateVolume(vol); err == nil {
				return lvm.UpdateVolInfo(vol, lvm.LVMStatusReady)
			}
		}
	}

	// In case no vg available or lvm.CreateVolume fails for all vgs, mark
	// the volume provisioning failed so that controller can reschedule it.
	vol.Status.Error = c.transformLVMError(err)
	return lvm.UpdateVolInfo(vol, lvm.LVMStatusFailed)
}

// getVgPriorityList returns ordered list of volume groups from higher to lower
// priority to use for provisioning a lvm volume. As of now, we are prioritizing
// the vg having least amount free space available to fit the volume.
func (c *VolController) getVgPriorityList(vol *apis.LVMVolume) ([]apis.VolumeGroup, error) {
	re, err := regexp.Compile(vol.Spec.VgPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression %v for lvm volume %s: %v",
			vol.Spec.VgPattern, vol.Name, err)
	}
	capacity, err := strconv.Atoi(vol.Spec.Capacity)
	if err != nil {
		return nil, fmt.Errorf("invalid requested capacity %v for lvm volume %s: %v",
			vol.Spec.Capacity, vol.Name, err)
	}

	vgs, err := lvm.ListLVMVolumeGroup(true)
	if err != nil {
		return nil, fmt.Errorf("failed to list vgs available on node: %v", err)
	}
	filteredVgs := make([]apis.VolumeGroup, 0)
	for _, vg := range vgs {
		if !re.MatchString(vg.Name) {
			continue
		}
		// skip the vgs capacity comparison in case of thin provision enable volume
		if vol.Spec.ThinProvision != "yes" {
			// filter vgs having insufficient capacity.
			if vg.Free.Value() < int64(capacity) {
				continue
			}
		}
		filteredVgs = append(filteredVgs, vg)
	}

	// prioritize the volume group having less free space available.
	sort.SliceStable(filteredVgs, func(i, j int) bool {
		return filteredVgs[i].Free.Cmp(filteredVgs[j].Free) < 0
	})
	return filteredVgs, nil
}

func (c *VolController) transformLVMError(err error) *apis.VolumeError {
	volErr := &apis.VolumeError{
		Code:    apis.Internal,
		Message: err.Error(),
	}
	execErr, ok := err.(*lvm.ExecError)
	if !ok {
		return volErr
	}

	if strings.Contains(strings.ToLower(string(execErr.Output)),
		"insufficient free space") {
		volErr.Code = apis.InsufficientCapacity
	}
	return volErr
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *VolController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Vol controller")

	// Wait for the k8s caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.VolSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	klog.Info("Starting Vol workers")
	// Launch worker to process Vol resources
	// Threadiness will decide the number of workers you want to launch to process work items from queue
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started Vol workers")
	<-stopCh
	klog.Info("Shutting down Vol workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *VolController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *VolController) processNextWorkItem() bool {
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
		// Vol resource to be synced.
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
