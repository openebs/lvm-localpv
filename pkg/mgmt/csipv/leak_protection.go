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

package csipv

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	VolumeAnnotation        = "csi-volume-name"
	LeakProtectionFinalizer = "csi-leak-protection"
)

// LeakProtectionController gracefully cleans up any orphan volume created
// by csi plugin before external provisioner creates pv for given pvc.
// See https://github.com/kubernetes-csi/external-provisioner/issues/486 for
// more details.
// Note: As a storage vendor, you should be able to lookup your volumes
// uniquely based on csi CreateVolume request name parameter.
type LeakProtectionController struct {
	driverName  string
	onPVCDelete func(pvc *corev1.PersistentVolumeClaim, createVolumeName string) error

	client clientset.Interface

	pvcLister       corelisters.PersistentVolumeClaimLister
	pvcListerSynced cache.InformerSynced

	queue workqueue.RateLimitingInterface

	// track set of pending volumes creation (stores pvc namespaced name string).
	// It is used in synchronizing BeginCreateVolume (invoked by csi.CreateVolume)
	// and onPVCDelete which deletes the created volume if any. Since CSI spec
	// doesn't expect create and delete volume rpcs per volume to be concurrent safe,
	// the controller loop here needs to ensure that it doesn't call onPVCDelete
	// method if there is any in-flight create volume rpcs running.
	claimsInProgress *syncSet
}

func NewLeakProtectionController(
	client clientset.Interface,
	pvcInformer coreinformers.PersistentVolumeClaimInformer,
	driverName string,
	onPVCDelete func(pvc *corev1.PersistentVolumeClaim, createVolumeName string) error,
) (*LeakProtectionController, error) {
	if driverName == "" {
		return nil, fmt.Errorf("empty csi driver name")
	}

	if onPVCDelete == nil {
		return nil, fmt.Errorf("invalid pvc onDelete callback")
	}

	c := &LeakProtectionController{
		driverName:  driverName,
		onPVCDelete: onPVCDelete,
		client:      client,

		pvcLister:       pvcInformer.Lister(),
		pvcListerSynced: pvcInformer.Informer().HasSynced,

		queue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(), "leak-protection"),
		claimsInProgress: newSyncSet(),
	}

	pvcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.onAddUpdate,
		UpdateFunc: func(old, new interface{}) {
			c.onAddUpdate(new)
		},
	})

	return c, nil
}

// onAddUpdate reacts to pvc added/updated events
func (c *LeakProtectionController) onAddUpdate(obj interface{}) {
	pvc, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("pvc informer returned non-pvc object: %#v", obj))
		return
	}
	key, err := cache.MetaNamespaceKeyFunc(pvc)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for persistent volume claim %#v: %v", pvc, err))
		return
	}
	klog.V(4).InfoS("received informer event on pvc", "key", key)
	c.queue.Add(key)
}

// Run runs the controller goroutines.
func (c *LeakProtectionController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	klog.InfoS("starting up csi pvc controller")
	defer klog.InfoS("shutting down csi pvc provisioning controller")

	if !cache.WaitForNamedCacheSync("CSI Provisioner", stopCh, c.pvcListerSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *LeakProtectionController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem deals with one pvcKey off the queue.  It returns false when it's time to quit.
func (c *LeakProtectionController) processNextWorkItem() bool {
	pvcKey, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(pvcKey)

	pvcNamespace, pvcName, err := cache.SplitMetaNamespaceKey(pvcKey.(string))
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error parsing pvc key %q: %v", pvcKey, err))
		return true
	}

	err = c.processPVC(pvcNamespace, pvcName)
	if err == nil {
		c.queue.Forget(pvcKey)
		return true
	}

	utilruntime.HandleError(fmt.Errorf("failed to process pvc %v: %v", pvcKey, err))
	c.queue.AddRateLimited(pvcKey)

	return true
}

func (c *LeakProtectionController) processPVC(pvcNamespace, pvcName string) error {
	pvc, err := c.pvcLister.PersistentVolumeClaims(pvcNamespace).Get(pvcName)
	if apierrors.IsNotFound(err) {
		klog.V(4).InfoS("pvc not found, ignoring...", "pvc", klog.KRef(pvcNamespace, pvcName))
		return nil
	}
	// if relevant finalizer doesn't exists, skip processing that pvc.
	if !c.finalizerExists(pvc) {
		return nil
	}

	klog.InfoS("leak controller processing pvc", "pvc", klog.KRef(pvcNamespace, pvcName))
	// if pvc gets bound to a persistent volume, we can safely remove the finalizer
	// since csi external-provisioner guarantees to call csi spec DeleteVolume method.
	if pvc.Status.Phase == corev1.ClaimBound {
		return c.removeFinalizer(pvc)
	}

	// process pvc in case it's marked for deletion.
	if pvc.GetDeletionTimestamp() != nil {
		volumeName, exists := pvc.GetAnnotations()[c.GetAnnotationKey()]
		if !exists {
			return fmt.Errorf("failed to find volume name used by csi create volume request")
		}

		if err := func() error {
			if alreadyExists := c.claimsInProgress.Add(c.claimsInProgressKey(pvc)); alreadyExists {
				return fmt.Errorf("csi driver already has volume creation in progress, will retry after sometime")
			}
			defer c.claimsInProgress.Remove(c.claimsInProgressKey(pvc))
			return c.onPVCDelete(pvc, volumeName)
		}(); err != nil {
			return fmt.Errorf("failed to finalize pvc deletion: %v", err)
		}
		klog.InfoS("deleted volume via csi driver if exists", "volume", volumeName,
			"driver", c.driverName, "pvc", klog.KRef(pvcNamespace, pvcName))
		return c.removeFinalizer(pvc)
	}
	return nil
}

func (c *LeakProtectionController) claimsInProgressKey(pvc *corev1.PersistentVolumeClaim) string {
	return pvc.Namespace + "/" + pvc.Name
}

func (c *LeakProtectionController) finalizerExists(pvc *corev1.PersistentVolumeClaim) bool {
	finalizers := pvc.GetFinalizers()
	for _, finalizer := range finalizers {
		if finalizer == c.GetFinalizer() {
			return true
		}
	}
	return false
}

func (c *LeakProtectionController) addFinalizer(pvc *corev1.PersistentVolumeClaim, volumeName string) error {
	finalizer := c.GetFinalizer()
	if c.finalizerExists(pvc) {
		klog.V(4).InfoS("finalizer already exists, ignoring...",
			"finalizer", finalizer, "pvc", klog.KObj(pvc))
		return nil
	}

	claimClone := pvc.DeepCopy()
	claimClone.ObjectMeta.Annotations[c.GetAnnotationKey()] = volumeName
	claimClone.ObjectMeta.Finalizers = append(claimClone.ObjectMeta.Finalizers, finalizer)
	_, err := c.client.CoreV1().PersistentVolumeClaims(claimClone.Namespace).Update(context.TODO(), claimClone, metav1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to add finalizer to pvc", "pvc", klog.KObj(pvc))
		return err
	}
	klog.V(3).InfoS("added finalizer to pvc",
		"finalizer", finalizer, "pvc", klog.KObj(pvc))
	return nil
}

func (c *LeakProtectionController) removeFinalizer(pvc *corev1.PersistentVolumeClaim) error {
	finalizer := c.GetFinalizer()
	claimClone := pvc.DeepCopy()

	// remove the annotation added previously.
	delete(claimClone.ObjectMeta.Annotations, c.GetAnnotationKey())

	currFinalizerList := claimClone.ObjectMeta.Finalizers
	newFinalizerList := make([]string, 0, len(currFinalizerList))
	for _, v := range currFinalizerList {
		if v == finalizer {
			continue
		}
		newFinalizerList = append(newFinalizerList, v)
	}
	claimClone.ObjectMeta.Finalizers = newFinalizerList

	_, err := c.client.CoreV1().PersistentVolumeClaims(claimClone.Namespace).Update(context.TODO(), claimClone, metav1.UpdateOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to remove finalizer from PVC",
			"finalizer", finalizer,
			"pvc", klog.KObj(pvc))
		return err
	}
	klog.V(3).InfoS("removed finalizer from PVC",
		"finalizer", finalizer, "pvc", klog.KObj(pvc))
	return nil
}

// BeginCreateVolume add relevant finalizer to the given pvc to avoid potential
// csi volume leak. It must be called from the create volume csi method
// implementation just before actual volume provisioning.
// volumeName param should be same as csi.CreateVolumeRequest Name parameter.
// In case of error, the csi driver should return non-retryable grpc error codes
// to external provisioner.
// Returned finishCreateVolume function must be called (preferably under defer)
// after attempting to provision volume.
// e.g
// {
//		finishCreateVolume, err := c.BeginCreateVolume("volumeId", "namespace", "name")
//		if err != nil {
//			return nil, status.Errorf(codes.FailedPrecondition, err.Error())
//		}
//		defer finishCreateVolume()
//		..... start provisioning volume here .....
// }
func (c *LeakProtectionController) BeginCreateVolume(volumeName,
	pvcNamespace, pvcName string) (func(), error) {
	pvc, err := c.client.CoreV1().PersistentVolumeClaims(pvcNamespace).
		Get(context.TODO(), pvcName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "failed to fetch pvc", "pvc", klog.KRef(pvcNamespace, pvcName))
		return nil, status.Errorf(codes.FailedPrecondition, "failed to fetch pvc: %v", err)
	} else if pvc.GetDeletionTimestamp() != nil {
		// if pvc is already marked for deletion, return err.
		err = fmt.Errorf("pvc already marked for deletion")
		klog.ErrorS(err, "", "pvc", klog.KRef(pvcNamespace, pvcName))
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	key := c.claimsInProgressKey(pvc)
	finishCreateVolume := func() {
		c.claimsInProgress.Remove(key)
	}
	alreadyExists := c.claimsInProgress.Add(key)
	if alreadyExists {
		return nil, status.Errorf(codes.Aborted,
			"csi driver already has volume creation in progress")
	}

	if err = c.addFinalizer(pvc, volumeName); err != nil {
		finishCreateVolume() // make sure we clean up on error.
		return nil, err
	}
	return finishCreateVolume, nil
}

func (c *LeakProtectionController) GetFinalizer() string {
	return c.driverName + "/" + LeakProtectionFinalizer
}

func (c *LeakProtectionController) GetAnnotationKey() string {
	return c.driverName + "/" + VolumeAnnotation
}

// syncSet is synchronised set of strings
type syncSet struct {
	sync.Mutex
	m map[string]struct{}
}

func newSyncSet() *syncSet {
	return &syncSet{
		m: make(map[string]struct{}),
	}
}

func (s *syncSet) Add(k string) bool {
	s.Lock()
	_, ok := s.m[k]
	s.m[k] = struct{}{}
	s.Unlock()
	return ok
}

func (s *syncSet) Remove(k string) {
	s.Lock()
	delete(s.m, k)
	s.Unlock()
}
