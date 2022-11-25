/*
Copyright 2021 The OpenEBS Authors

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

package snapshot

import (
	"time"

	corev1 "k8s.io/api/core/v1"
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
	controllerAgentName = "lvmsnap-controller"
	GroupOpenebsIO      = "local.openebs.io"
	VersionV1alpha1     = "v1alpha1"
	Resource            = "lvmsnapshots"
)

var snapresource = schema.GroupVersionResource{
	Group:    GroupOpenebsIO,
	Version:  VersionV1alpha1,
	Resource: Resource,
}

// SnapController is the controller implementation for Snap resources
type SnapController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a interface which will be used to list lvmsnapshot from Api server
	clientset dynamic.Interface

	//VolLister is used to list lvmsnapshot from informer cache
	snapLister dynamiclister.Lister

	// snapSynced is used for caches sync to get populated
	snapSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// This function returns controller object with all required keys set to watch over lvmsnapshot object
func newSnapController(kubeClient kubernetes.Interface, client dynamic.Interface,
	dynInformer dynamicinformer.DynamicSharedInformerFactory) *SnapController {
	//Creating informer for lvmsnapshot resource
	snapInformer := dynInformer.ForResource(snapresource).Informer()
	//This ratelimiter requeues failed items after 5 secs for first 12 attempts. Then objects are requeued after 30 secs.
	rateLimiter := workqueue.NewItemFastSlowRateLimiter(5*time.Second, 30*time.Second, 12)

	klog.Infof("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	klog.Infof("Creating lvm snapshot controller object")
	snapCtrller := &SnapController{
		kubeclientset: kubeClient,
		clientset:     client,
		snapLister:    dynamiclister.New(snapInformer.GetIndexer(), snapresource),
		snapSynced:    snapInformer.HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(rateLimiter, "Snap"),
		recorder:      recorder,
	}
	klog.Infof("Adding Event handler functions for lvm snapshot controller")
	snapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    snapCtrller.addSnap,
		DeleteFunc: snapCtrller.deleteSnap,
		UpdateFunc: snapCtrller.updateSnap,
	})
	return snapCtrller
}
