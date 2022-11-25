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
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const (
	controllerAgentName = "lvmnode-controller"
	GroupOpenebsIO      = "local.openebs.io"
	VersionV1alpha1     = "v1alpha1"
	Resource            = "lvmnodes"
)

var noderesource = schema.GroupVersionResource{
	Group:    GroupOpenebsIO,
	Version:  VersionV1alpha1,
	Resource: Resource,
}

// NodeController is the controller implementation for lvm node resources

type NodeController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a interface which will be used to list lvmnode from Api server
	clientset dynamic.Interface

	//NodeLister is used to list lvmnode from informer cache
	NodeLister dynamiclister.Lister

	// NodeSynced is used for caches sync to get populated
	NodeSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// pollInterval controls the polling frequency of syncing up the vg metadata.
	pollInterval time.Duration

	// ownerRef is used to set the owner reference to lvmnode objects.
	ownerRef metav1.OwnerReference
}

// This function returns controller object with all required keys set to watch over lvmnode object
func newNodeController(kubeClient kubernetes.Interface, client dynamic.Interface,
	dynInformer dynamicinformer.DynamicSharedInformerFactory, ownerRef metav1.OwnerReference) *NodeController {
	//Creating informer for lvm node resource
	nodeInformer := dynInformer.ForResource(noderesource).Informer()
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	klog.Infof("Creating lvm node controller object")
	nodeContrller := &NodeController{
		kubeclientset: kubeClient,
		clientset:     client,
		NodeLister:    dynamiclister.New(nodeInformer.GetIndexer(), noderesource),
		NodeSynced:    nodeInformer.HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Node"),
		recorder:      recorder,
		pollInterval:  60 * time.Second,
		ownerRef:      ownerRef,
	}

	klog.Infof("Adding Event handler functions for lvm node controller")
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    nodeContrller.addNode,
		UpdateFunc: nodeContrller.updateNode,
		DeleteFunc: nodeContrller.deleteNode,
	})
	return nodeContrller
}
