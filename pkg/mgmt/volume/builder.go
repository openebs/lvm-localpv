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
	controllerAgentName = "lvmvolume-controller"
	GroupOpenebsIO      = "local.openebs.io"
	VersionV1alpha1     = "v1alpha1"
	Resource            = "lvmvolumes"
)

var volresource = schema.GroupVersionResource{
	Group:    GroupOpenebsIO,
	Version:  VersionV1alpha1,
	Resource: Resource,
}

// VolController is the controller implementation for volume resources
type VolController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a interface which will be used to list lvmvolumes from Api server
	clientset dynamic.Interface

	//VolLister is used to list lvmvolumes from informer cache
	VolLister dynamiclister.Lister

	// VolSynced is used for caches sync to get populated
	VolSynced cache.InformerSynced

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

// This function returns controller object with all required keys set to watch over lvmvolume object
func newVolController(kubeClient kubernetes.Interface, client dynamic.Interface,
	dynInformer dynamicinformer.DynamicSharedInformerFactory) *VolController {
	//Creating informer for lvmvolume resource
	volInformer := dynInformer.ForResource(volresource).Informer()
	//This ratelimiter requeues failed items after 5 secs for first 12 attempts. Then objects are requeued after 30 secs.
	rateLimiter := workqueue.NewItemFastSlowRateLimiter(5*time.Second, 30*time.Second, 12)

	klog.Infof("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	klog.Infof("Creating lvm volume controller object")
	volCtrller := &VolController{
		kubeclientset: kubeClient,
		clientset:     client,
		VolLister:     dynamiclister.New(volInformer.GetIndexer(), volresource),
		VolSynced:     volInformer.HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(rateLimiter, "Vol"),
		recorder:      recorder,
	}

	klog.Infof("Adding Event handler functions for lvm volume controller")
	volInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    volCtrller.addVol,
		DeleteFunc: volCtrller.deleteVol,
		UpdateFunc: volCtrller.updateVol,
	})

	return volCtrller
}
