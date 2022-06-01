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
	"context"
	"sync"
	"time"

	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	"github.com/openebs/lvm-localpv/pkg/lvm"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Start starts the lvmnode controller.
func Start(controllerMtx *sync.RWMutex, stopCh <-chan struct{}) error {

	// Get in cluster config
	cfg, err := k8sapi.Config().Get()
	if err != nil {
		return errors.Wrap(err, "error building kubeconfig")
	}

	// Building Kubernetes Clientset
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building kubernetes clientset")
	}

	openebsClientNew, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error building dynamic client for lvmnode cr")
	}

	// setup watch only on node we are interested in.
	nodeInformerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(openebsClientNew, 5*time.Minute,
		lvm.LvmNamespace, func(options *metav1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", lvm.NodeID).String()
		})

	k8sNode, err := kubeClient.CoreV1().Nodes().Get(context.TODO(), lvm.NodeID, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "fetch k8s node %s", lvm.NodeID)
	}
	isTrue := true
	// as object returned by client go clears all TypeMeta from it.
	nodeGVK := &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Node",
	}
	ownerRef := metav1.OwnerReference{
		APIVersion: nodeGVK.GroupVersion().String(),
		Kind:       nodeGVK.Kind,
		Name:       k8sNode.Name,
		UID:        k8sNode.GetUID(),
		Controller: &isTrue,
	}

	// Build() fn of all controllers calls AddToScheme to adds all types of this
	// clientset into the given scheme.
	// If multiple controllers happen to call this AddToScheme same time,
	// it causes panic with error saying concurrent map access.
	// This lock is used to serialize the AddToScheme call of all controllers.
	controllerMtx.Lock()

	controller := newNodeController(kubeClient, openebsClientNew, nodeInformerFactory, ownerRef)

	// blocking call, can't use defer to release the lock
	controllerMtx.Unlock()

	if err != nil {
		return errors.Wrapf(err, "error building controller instance")
	}

	klog.Info("Starting informer for lvm node controller")
	nodeInformerFactory.Start(stopCh)

	klog.Info("Starting Lvm node controller")
	// Threadiness defines the number of workers to be launched in Run function
	return controller.Run(1, stopCh)
}
