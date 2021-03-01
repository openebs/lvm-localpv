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
	clientset "github.com/openebs/lvm-localpv/pkg/generated/clientset/internalclientset"
	openebsScheme "github.com/openebs/lvm-localpv/pkg/generated/clientset/internalclientset/scheme"
	informers "github.com/openebs/lvm-localpv/pkg/generated/informer/externalversions"
	listers "github.com/openebs/lvm-localpv/pkg/generated/lister/lvm/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

const controllerAgentName = "lvmvolume-controller"

// VolController is the controller implementation for volume resources
type VolController struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// clientset is a openebs custom resource package generated for custom API group.
	clientset clientset.Interface

	VolLister listers.LVMVolumeLister

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

// VolControllerBuilder is the builder object for controller.
type VolControllerBuilder struct {
	VolController *VolController
}

// NewVolControllerBuilder returns an empty instance of controller builder.
func NewVolControllerBuilder() *VolControllerBuilder {
	return &VolControllerBuilder{
		VolController: &VolController{},
	}
}

// withKubeClient fills kube client to controller object.
func (cb *VolControllerBuilder) withKubeClient(ks kubernetes.Interface) *VolControllerBuilder {
	cb.VolController.kubeclientset = ks
	return cb
}

// withOpenEBSClient fills openebs client to controller object.
func (cb *VolControllerBuilder) withOpenEBSClient(cs clientset.Interface) *VolControllerBuilder {
	cb.VolController.clientset = cs
	return cb
}

// withVolLister fills Vol lister to controller object.
func (cb *VolControllerBuilder) withVolLister(sl informers.SharedInformerFactory) *VolControllerBuilder {
	VolInformer := sl.Local().V1alpha1().LVMVolumes()
	cb.VolController.VolLister = VolInformer.Lister()
	return cb
}

// withVolSynced adds object sync information in cache to controller object.
func (cb *VolControllerBuilder) withVolSynced(sl informers.SharedInformerFactory) *VolControllerBuilder {
	VolInformer := sl.Local().V1alpha1().LVMVolumes()
	cb.VolController.VolSynced = VolInformer.Informer().HasSynced
	return cb
}

// withWorkqueue adds workqueue to controller object.
func (cb *VolControllerBuilder) withWorkqueueRateLimiting() *VolControllerBuilder {
	cb.VolController.workqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Vol")
	return cb
}

// withRecorder adds recorder to controller object.
func (cb *VolControllerBuilder) withRecorder(ks kubernetes.Interface) *VolControllerBuilder {
	klog.Infof("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: ks.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})
	cb.VolController.recorder = recorder
	return cb
}

// withEventHandler adds event handlers controller object.
func (cb *VolControllerBuilder) withEventHandler(cvcInformerFactory informers.SharedInformerFactory) *VolControllerBuilder {
	cvcInformer := cvcInformerFactory.Local().V1alpha1().LVMVolumes()
	// Set up an event handler for when Vol resources change
	cvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    cb.VolController.addVol,
		UpdateFunc: cb.VolController.updateVol,
		DeleteFunc: cb.VolController.deleteVol,
	})
	return cb
}

// Build returns a controller instance.
func (cb *VolControllerBuilder) Build() (*VolController, error) {
	err := openebsScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	return cb.VolController, nil
}
