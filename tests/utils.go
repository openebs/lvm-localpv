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

package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/openebs/lib-csi/pkg/csipv"

	"github.com/openebs/lvm-localpv/pkg/lvm"
	"github.com/openebs/lvm-localpv/tests/container"
	"github.com/openebs/lvm-localpv/tests/deploy"
	"github.com/openebs/lvm-localpv/tests/k8svolume"
	"github.com/openebs/lvm-localpv/tests/pod"
	"github.com/openebs/lvm-localpv/tests/pts"
	"github.com/openebs/lvm-localpv/tests/pvc"
	"github.com/openebs/lvm-localpv/tests/sc"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

// IsPVCBoundEventually checks if the pvc is bound or not eventually
func IsPVCBoundEventually(pvcName string) bool {
	return gomega.Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		60, 5).
		Should(gomega.BeTrue())
}

// IsPVCResizedEventually checks if the pvc is bound or not eventually
func IsPVCResizedEventually(pvcName string, newCapacity string, shouldPass bool) bool {
	newStorage, err := resource.ParseQuantity(newCapacity)
	if err != nil {
		return false
	}
	status := gomega.BeFalse()
	if shouldPass {
		status = gomega.BeTrue()
	}

	return gomega.Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		pvcStorage := volume.Status.Capacity[corev1.ResourceName(corev1.ResourceStorage)]
		return pvcStorage == newStorage
	},
		120, 5).
		Should(status)
}

// IsPodRunningEventually return true if the pod comes to running state
func IsPodRunningEventually(namespace, podName string) bool {
	return gomega.Eventually(func() bool {
		p, err := PodClient.
			WithNamespace(namespace).
			Get(podName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pod.NewForAPIObject(p).
			IsRunning()
	},
		60, 5).
		Should(gomega.BeTrue())
}

// IsPVCDeletedEventually checks if the PVC is deleted or not eventually
func IsPVCDeletedEventually(pvcName string) bool {
	return gomega.Eventually(func() bool {
		_, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 10).
		Should(gomega.BeTrue())
}

func createFstypeStorageClass(ftype string) {
	var (
		err error
	)

	parameters := map[string]string{
		"volgroup": VOLGROUP,
		"fstype":   ftype,
	}

	ginkgo.By("building a " + ftype + " storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithVolumeExpansion(true).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building ext4 storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a ext4 storageclass {%s}", scName)
}

func createStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"volgroup": VOLGROUP,
	}

	ginkgo.By("building a default storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building default storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a default storageclass {%s}", scName)
}

func createThinStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"volgroup":      VOLGROUP,
		"thinProvision": "yes",
	}

	ginkgo.By("building a thinProvision storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building thinProvision storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a thinProvision storageclass {%s}", scName)
}

func createRaidStorageClass(raidArgs map[string]string) {
	var (
		err error
	)

	// Add in the volume group to the args
	raidArgs["volgroup"] = RAIDGROUP

	ginkgo.By("building a RAID storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(raidArgs).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building RAID storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a RAID storageclass {%s}", scName)
}

// VerifyLVMVolume verify the properties of a lvm-volume
func VerifyLVMVolume() {
	ginkgo.By("fetching lvm volume")
	vol, err := LVMClient.WithNamespace(OpenEBSNamespace).
		Get(pvcObj.Spec.VolumeName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the lvm volume {%s}", pvcObj.Spec.VolumeName)

	ginkgo.By("verifying lvm volume")
	gomega.Expect(vol.Spec.VolGroup).To(gomega.Equal(scObj.Parameters["volgroup"]),
		"while checking volume group of lvm volume", pvcObj.Spec.VolumeName)

	gomega.Expect(vol.Finalizers[0]).To(gomega.Equal(lvm.LVMFinalizer), "while checking finializer to be set {%s}", pvcObj.Spec.VolumeName)
}

func deleteStorageClass() {
	err := SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	gomega.Expect(err).To(gomega.BeNil(),
		"while deleting lvm storageclass {%s}", scObj.Name)
}

func createAndVerifyPVC() {
	var (
		err     error
		pvcName = "lvmpv-pvc"
	)
	ginkgo.By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

func createAndVerifyBlockPVC() {
	var (
		err     error
		pvcName = "lvmpv-pvc"
	)

	volmode := corev1.PersistentVolumeBlock

	ginkgo.By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithVolumeMode(&volmode).
		WithCapacity(capacity).Build()
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc status as bound")

	status := IsPVCBoundEventually(pvcName)
	gomega.Expect(status).To(gomega.Equal(true),
		"while checking status equal to bound")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

func resizeAndVerifyPVC(shouldPass bool, size string) {
	var (
		err     error
		pvcName = "lvmpv-pvc"
	)
	ginkgo.By("updating the pvc with new size")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	pvcObj, err = pvc.BuildFrom(pvcObj).
		WithCapacity(size).Build()
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Update(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while updating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("verifying pvc size to be updated")

	IsPVCResizedEventually(pvcName, size, shouldPass)

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}
func createDeployVerifyApp() {
	ginkgo.By("creating and deploying app pod")
	createAndDeployAppPod(appName)
	time.Sleep(30 * time.Second)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func createAndDeployAppPod(appname string) {
	var err error
	ginkgo.By("building a busybox app pod deployment using above lvm volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appname).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithSelectorMatchLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(
					map[string]string{
						"app": "busybox",
					},
				).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/datadir/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol1",
									MountPath: "/mnt/datadir",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcObj.Name),
				),
		).
		Build()

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createAndDeployBlockAppPod() {
	var err error
	ginkgo.By("building a busybox app pod deployment using above lvm volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appName).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithSelectorMatchLabelsNew(
			map[string]string{
				"app": "busybox",
			},
		).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(
					map[string]string{
						"app": "busybox",
					},
				).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("busybox").
						WithName("busybox").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithCommandNew(
							[]string{
								"sh",
								"-c",
								"date > /mnt/datadir/date.txt; sync; sleep 5; sync; tail -f /dev/null;",
							},
						).
						WithVolumeDevicesNew(
							[]corev1.VolumeDevice{
								corev1.VolumeDevice{
									Name:       "datavol1",
									DevicePath: "/dev/xvda",
								},
							},
						),
				).
				WithVolumeBuilders(
					k8svolume.NewBuilder().
						WithName("datavol1").
						WithPVCSource(pvcObj.Name),
				),
		).
		Build()

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appName)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appName,
		OpenEBSNamespace,
	)
}

func createDeployVerifyBlockApp() {
	ginkgo.By("creating and deploying app pod", createAndDeployBlockAppPod)
	time.Sleep(30 * time.Second)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func verifyAppPodRunning() {
	var err error
	appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
		List(metav1.ListOptions{
			LabelSelector: "app=busybox",
		},
		)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while verifying application pod")

	status := IsPodRunningEventually(OpenEBSNamespace, appPod.Items[0].Name)
	gomega.Expect(status).To(gomega.Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func deleteAppDeployment(appname string) {
	err := DeployClient.WithNamespace(OpenEBSNamespace).
		Delete(appname, &metav1.DeleteOptions{})
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while deleting application pod")
}

func deleteAndVerifyPVC(pvcname string) {
	err := PVCClient.WithNamespace(OpenEBSNamespace).Delete(pvcname, &metav1.DeleteOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcname,
		OpenEBSNamespace,
	)
	ginkgo.By("verifying deleted pvc")
	status := IsPVCDeletedEventually(pvcname)
	gomega.Expect(status).To(gomega.Equal(true), "while trying to get deleted pvc")
}

func verifyPVForPVC(shouldExist bool, pvcName string) {
	pvList, err := PVClient.List(metav1.ListOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while listing PV for PVC: %{s}",
		pvcName,
	)

	shouldPVExist := gomega.BeFalse()
	if shouldExist {
		shouldPVExist = gomega.BeTrue()
	}

	ginkgo.By("verifying PV for PVC exists or not")
	var matchingPVName string
	matchingPVExists := false
	for _, pv := range pvList.Items {
		if pv.Spec.ClaimRef != nil &&
			pv.Spec.ClaimRef.Name == pvcName &&
			pv.Spec.ClaimRef.Namespace == OpenEBSNamespace {
			matchingPVExists = true
			matchingPVName = pv.Name
			break
		}
	}

	if matchingPVExists && !shouldExist {
		if IsPVDeletedEventually(shouldExist, matchingPVName) {
			matchingPVExists = false
		}
	}

	gomega.Expect(matchingPVExists).To(shouldPVExist)
}

// IsPVDeletedEventually checks if the PV is deleted or not eventually
func IsPVDeletedEventually(shouldExist bool, pvName string) bool {
	shouldPVExist := gomega.BeFalse()
	if shouldExist {
		shouldPVExist = gomega.BeTrue()
	}
	return gomega.Eventually(func() bool {
		_, err := PVClient.Get(pvName, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	},
		120, 10).
		Should(shouldPVExist)
}

func getGeneratedVolName(pvc *corev1.PersistentVolumeClaim) string {
	return fmt.Sprintf("pvc-%v", pvcObj.GetUID())
}

func createPVC() {
	var (
		err     error
		pvcName = "lvmpv-pvc"
	)
	ginkgo.By("building a pvc")
	pvcObj, err = pvc.NewBuilder().
		WithName(pvcName).
		WithNamespace(OpenEBSNamespace).
		WithStorageClass(scObj.Name).
		WithAccessModes(accessModes).
		WithCapacity(capacity).Build()
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while building pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)

	ginkgo.By("creating above pvc")
	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Create(pvcObj)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while creating pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

func deleteAndVerifyLeakedPVC(pvcName string) {
	ginkgo.By("Deleting pending PVC")
	err := PVCClient.WithNamespace(OpenEBSNamespace).Delete(pvcName, &metav1.DeleteOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while deleting pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
	ginkgo.By("Verify leaked pvc finalizer")

	status := gomega.Eventually(func() bool {
		pvcRes, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).To(
			gomega.BeNil(),
			"fetch pvc %v", pvcName)
		return len(pvcRes.GetFinalizers()) == 1 &&
			pvcRes.GetFinalizers()[0] == LocalProvisioner+"/"+csipv.LeakProtectionFinalizer
	}, 120, 10).Should(gomega.BeTrue())
	gomega.Expect(status).To(gomega.Equal(true), "expecting a leak protection finalizer")
}

func verifyPendingLVMVolume(volName string) {
	ginkgo.By("fetching lvm volume")
	vol, err := LVMClient.WithNamespace(OpenEBSNamespace).
		Get(volName, metav1.GetOptions{})
	gomega.Expect(err).To(gomega.BeNil(), "while fetching the lvm volume {%s}", volName)

	ginkgo.By("verifying lvm volume")
	gomega.Expect(scObj.Parameters["volgroup"]).To(gomega.MatchRegexp(vol.Spec.VgPattern),
		"while checking volume group of lvm volume", volName)
}

// WaitForLVMVolumeReady verify the if lvm-volume is ready
func WaitForLVMVolumeReady() {
	volName := getGeneratedVolName(pvcObj)
	status := gomega.Eventually(func() bool {
		vol, err := LVMClient.WithNamespace(OpenEBSNamespace).
			Get(volName, metav1.GetOptions{})
		gomega.Expect(err).To(gomega.BeNil(), "while fetching the lvm volume {%s}", volName)
		return vol.Status.State == "Ready"
	}, 120, 10).
		Should(gomega.BeTrue())
	gomega.Expect(status).To(gomega.Equal(true), "expecting a lvmvol resource to be ready")
}

func scaleControllerPlugin(num int32) int32 {
	ginkgo.By(fmt.Sprintf("scaling controller plugin statefulset %v to size %v", controllerStatefulSet, num))

	scale, err := K8sClient.AppsV1().StatefulSets(metav1.NamespaceSystem).
		GetScale(context.Background(), controllerStatefulSet, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"fetch current replica of stateful set %v", controllerStatefulSet)
	existingReplicas := scale.Spec.Replicas

	if scale.Spec.Replicas == num {
		return existingReplicas
	}
	scale.Spec.Replicas = num
	scale, err = K8sClient.AppsV1().StatefulSets(metav1.NamespaceSystem).
		UpdateScale(context.Background(), controllerStatefulSet, scale, metav1.UpdateOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"update replicas of stateful set %v to %v", controllerStatefulSet, num)

	scaled := gomega.Eventually(func() bool {
		scale, err = K8sClient.AppsV1().StatefulSets(metav1.NamespaceSystem).
			GetScale(context.Background(), controllerStatefulSet, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return scale.Spec.Replicas == num
	}, 120, 10).
		Should(gomega.BeTrue())
	gomega.Expect(scaled).To(gomega.BeTrue(),
		"failed to scale up stateful set %v to size %v", controllerStatefulSet, num)
	return existingReplicas
}

func deleteNodeDaemonSet() *appsv1.DaemonSet {
	csiNodes, err := K8sClient.StorageV1().CSINodes().List(context.Background(), metav1.ListOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(), "fetching csi node")
	if len(csiNodes.Items) == 0 {
		err = fmt.Errorf("expecting non-zero csi nodes in the cluster")
		gomega.Expect(err).To(gomega.BeNil())
	}
	csiNode := csiNodes.Items[0]

	ginkgo.By("deleting node plugin daemonset " + nodeDaemonSet)
	ds, err := K8sClient.AppsV1().
		DaemonSets(metav1.NamespaceSystem).
		Get(context.Background(), nodeDaemonSet, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"fetching node plugin daemonset %v", nodeDaemonSet)
	policy := metav1.DeletePropagationForeground
	err = K8sClient.AppsV1().
		DaemonSets(metav1.NamespaceSystem).
		Delete(context.Background(), nodeDaemonSet, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"deleting node plugin daemonset %v", nodeDaemonSet)

	ginkgo.By("waiting for deletion of node plugin pods")
	status := gomega.Eventually(func() bool {
		_, err = K8sClient.AppsV1().
			DaemonSets(metav1.NamespaceSystem).
			Get(context.Background(), nodeDaemonSet, metav1.GetOptions{})
		return k8serrors.IsNotFound(err)
	}, 120, 10).Should(gomega.BeTrue())
	gomega.Expect(status).To(gomega.Equal(true),
		"waiting for deletion of node plugin daemonset")

	// update the underlying csi node resource to ensure pvc gets scheduled
	// by external provisioner.
	ginkgo.By("patching csinode resource")
	newCSINode, err := K8sClient.StorageV1().CSINodes().Get(context.Background(), csiNode.GetName(), metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(), "fetching updated csi node")
	newCSINode.Spec.Drivers = csiNode.Spec.Drivers
	_, err = K8sClient.StorageV1().CSINodes().Update(context.Background(), newCSINode, metav1.UpdateOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(), "updating csi node %v", csiNode.GetName())

	return ds
}

func createNodeDaemonSet(ds *appsv1.DaemonSet) {
	ds.SetResourceVersion("") // reset the resource version for creation.
	ginkgo.By("creating node plugin daemonset " + nodeDaemonSet)
	_, err := K8sClient.AppsV1().
		DaemonSets(metav1.NamespaceSystem).
		Create(context.Background(), ds, metav1.CreateOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"creating node plugin daemonset %v", nodeDaemonSet)
}
