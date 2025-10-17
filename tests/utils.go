package tests

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/openebs/lib-csi/pkg/csipv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/openebs/lvm-localpv/pkg/lvm"
	"github.com/openebs/lvm-localpv/tests/container"
	"github.com/openebs/lvm-localpv/tests/deploy"
	"github.com/openebs/lvm-localpv/tests/k8svolume"
	"github.com/openebs/lvm-localpv/tests/pod"
	"github.com/openebs/lvm-localpv/tests/pts"
	"github.com/openebs/lvm-localpv/tests/pvc"
	"github.com/openebs/lvm-localpv/tests/sc"

	"k8s.io/apimachinery/pkg/api/resource"
)

// This checks if the pvc is bound eventually within the poll period.
func IsPVCBoundEventually(pvcName string) bool {
	ginkgo.By("Verifying pvc status to be bound eventually\n")
	return gomega.Eventually(func() bool {
		volume, err := PVCClient.
			Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return pvc.NewForAPIObject(volume).IsBound()
	},
		60, 5).
		Should(gomega.BeTrue())
}

// This checks if the pvc is Not bound consistently over polling period.
func IsPVCPendingConsistently(pvcName string) bool {
	ginkgo.By("Verifying pvc status to be Pending consistently\n")
	return gomega.Consistently(func() bool {
		volume, err := PVCClient.Get(pvcName, metav1.GetOptions{})
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		return !pvc.NewForAPIObject(volume).IsBound()
	}, 30, 5).Should(gomega.BeTrue())
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
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
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
		WithVolumeExpansion(true).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building default storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a default storageclass {%s}", scName)
}

func createVgPatternStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"vgpattern": vgPattern,
	}

	ginkgo.By("building a vgpattern based storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building default storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a default storageclass {%s}", scName)
}

func createStorageClassWithNonExistingVg() {
	var (
		err error
	)

	parameters := map[string]string{
		"volgroup": NONEXIST_VOLGROUP,
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

func createSharedVolStorageClass() {
	var (
		err error
	)

	parameters := map[string]string{
		"volgroup": VOLGROUP,
		"shared":   "yes",
	}

	ginkgo.By("building a shared volume storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithVolumeExpansion(true).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building shared volume storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a shared volume storageclass {%s}", scName)
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
		WithVolumeExpansion(true).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building thinProvision storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a thinProvision storageclass {%s}", scName)
}

func createFormatOptionsStorageClass(formatOptions string) {
	var (
		err error
	)

	parameters := map[string]string{
		"volgroup":      VOLGROUP,
		"formatOptions": formatOptions,
	}

	ginkgo.By("building a default storage class")
	scObj, err = sc.NewBuilder().
		WithGenerateName(scName).
		WithVolumeExpansion(true).
		WithParametersNew(parameters).
		WithProvisioner(LocalProvisioner).Build()
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(),
		"while building formatOptions storageclass obj with prefix {%s}", scName)

	scObj, err = SCClient.Create(scObj)
	gomega.Expect(err).To(gomega.BeNil(), "while creating a storageclass with formatOptions {%s}", scName)
}

// VerifyLVMVolume verify the properties of a lvm-volume
// expected_vg is supposed to be passed only when vgpatten was used for scheduling.
// If its volgroup in sc then we can just match volgroup with lvmvolume's vg field.
func VerifyLVMVolume(expect_ready bool, expected_vg string) {
	ginkgo.By("fetching lvm volume")
	vol_name := ""
	if !expect_ready {
		vol_name = pvcObj.ObjectMeta.Annotations["local.csi.openebs.io/csi-volume-name"]
	} else {
		vol_name = pvcObj.Spec.VolumeName
	}
	vol, err := LVMClient.WithNamespace(OpenEBSNamespace).
		Get(vol_name, metav1.GetOptions{})

	if !expect_ready {
		if vol != nil && vol.ObjectMeta.Name == vol_name {
			// Even if scheduler cant find the vg for scheduling it creates lvmvolume cr anyway,
			// It gets deleted, by the csi provisioner only when the owner node of cr marks is
			// as Failed. So incase, we do a get of cr when the cr was being handled then we expect
			// state to be either Pending or Failed.
			gomega.Expect(vol.Status.State).To(gomega.Or(gomega.Equal("Pending"), gomega.Equal("Failed")),
				"While checking if lvmvolume: %s is in Pending or Failed state", pvcObj.Spec.VolumeName)
		}
	} else {
		gomega.Expect(err).To(gomega.BeNil(), "while fetching the lvm volume {%s}", pvcObj.Spec.VolumeName)
		if expected_vg != "" {
			gomega.Expect(vol.Spec.VolGroup).To(gomega.Equal(expected_vg),
				"while checking volume group of lvm volume", pvcObj.Spec.VolumeName)
		} else {
			gomega.Expect(vol.Spec.VolGroup).To(gomega.Equal(scObj.Parameters["volgroup"]),
				"while checking volume group of lvm volume", pvcObj.Spec.VolumeName)
		}
		gomega.Expect(vol.Status.State).To(gomega.Equal("Ready"),
			"While checking if lvmvolume: %s is in Ready state", pvcObj.Spec.VolumeName)
		gomega.Expect(vol.Finalizers[0]).To(gomega.Equal(lvm.LVMFinalizer), "while checking finializer to be set {%s}", pvcObj.Spec.VolumeName)
	}
}

func deleteStorageClass() {
	err := SCClient.Delete(scObj.Name, &metav1.DeleteOptions{})
	gomega.Expect(err).To(gomega.BeNil(),
		"while deleting lvm storageclass {%s}", scObj.Name)
}

func createAndVerifyPVC(expect_bound bool) {
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
	verifyPVCStatus(pvcName, expect_bound)

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}

// Verifies state of already created pvc based on expect_bound.
func verifyPVCStatus(pvc_name string, expect_bound bool) {
	ok := false
	if !expect_bound {
		ok = IsPVCPendingConsistently(pvc_name)
	} else {
		ok = IsPVCBoundEventually(pvc_name)
	}
	gomega.Expect(ok).To(gomega.Equal(true),
		"while checking the pvc status")
}

func createAndVerifyBlockPVCId(expect_bound bool, id_suffix string) {
	var pvcName = "lvmpv-pvc" + "-" + id_suffix

	createBlockPVC(pvcName)
	verifyBlockPVCName(expect_bound, pvcName)
}
func createAndVerifyBlockPVC(expect_bound bool) {
	var pvcName = "lvmpv-pvc"

	createBlockPVC(pvcName)
	verifyBlockPVCName(expect_bound, pvcName)
}

func verifyBlockPVCName(expect_bound bool, pvcName string) {
	var err error
	ginkgo.By("verifying pvc status as bound\n")

	ok := false
	if !expect_bound {
		ok = IsPVCPendingConsistently(pvcName)
	} else {
		ok = IsPVCBoundEventually(pvcName)
	}
	gomega.Expect(ok).To(gomega.Equal(true),
		"while checking the pvc status")

	pvcObj, err = PVCClient.WithNamespace(OpenEBSNamespace).Get(pvcObj.Name, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"while retrieving pvc {%s} in namespace {%s}",
		pvcName,
		OpenEBSNamespace,
	)
}
func createBlockPVC(pvcName string) {
	var err error
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
}

func VerifyBlockPVC() {
	var (
		err     error
		pvcName = "lvmpv-pvc"
	)

	ok := IsPVCBoundEventually(pvcName)
	gomega.Expect(ok).To(gomega.Equal(true),
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
	createAndDeployAppPod(appNames)
	ginkgo.By("verifying app pods are running", verifyAppPodRunning)
}

func createAndDeployAppPod(appnames []string) {
	var err error
	rwmode := "write"
	for index, appname := range appnames {
		labels := map[string]string{
			"role": "test",
			"app":  appname,
		}
		// only one app should do the write to ensure data safety.
		if index > 0 {
			rwmode = "read"
		}
		ginkgo.By("building app " + appname + " " + rwmode + " mode pod deployment using above lvm volume")
		deployObj, err = deploy.NewBuilder().
			WithName(appname).
			WithNamespace(OpenEBSNamespace).
			WithLabelsNew(labels).
			WithSelectorMatchLabelsNew(labels).
			WithPodTemplateSpecBuilder(
				pts.NewBuilder().
					WithLabelsNew(labels).
					WithContainerBuilders(
						container.NewBuilder().
							WithImage("openebs/fio").
							WithName("fio").
							WithImagePullPolicy(corev1.PullIfNotPresent).
							WithCommandNew(
								[]string{
									"sh",
									"-c",
									"fio --filename=/mnt/datadir/fioFile --direct=1 --rw=" + rwmode + " --bs=4k --ioengine=linuxaio --iodepth=32 --size=3GiB --numjobs=1 --name=fio-ci",
								},
							).
							WithVolumeMountsNew(
								[]corev1.VolumeMount{
									corev1.VolumeMount{
										Name: "datavol1",
										// If this path changes, modify the above fio command line accordingly.
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

		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appname)

		deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
		gomega.Expect(err).ShouldNot(
			gomega.HaveOccurred(),
			"while creating pod {%s} in namespace {%s}",
			appname,
			OpenEBSNamespace,
		)
	}
}

func createAndDeployBlockAppPod() {
	var err error
	for _, appName := range appNames {
		labels := map[string]string{
			"role": "test",
			"app":  appName,
		}
		ginkgo.By("building app " + appName + " pod deployment using above lvm volume")
		deployObj, err = deploy.NewBuilder().
			WithName(appName).
			WithNamespace(OpenEBSNamespace).
			WithLabelsNew(labels).
			WithSelectorMatchLabelsNew(labels).
			WithPodTemplateSpecBuilder(
				pts.NewBuilder().
					WithLabelsNew(labels).
					WithContainerBuilders(
						container.NewBuilder().
							WithImage("openebs/fio").
							WithName("fio").
							WithImagePullPolicy(corev1.PullIfNotPresent).
							WithCommandNew(
								[]string{
									"sh",
									"-c",
									"fio --filename=/dev/xvda/fioFile --direct=1 --rw=write --bs=4k --ioengine=linuxaio --iodepth=32 --size=3GiB --numjobs=1 --name=fio-ci",
								},
							).
							WithVolumeDevicesNew(
								[]corev1.VolumeDevice{
									corev1.VolumeDevice{
										Name: "datavol1",
										// If this path changes, modify the above fio command line accordingly.
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
}

func createDeployVerifyFormatOptions() {
	ginkgo.By("creating and deploying verifier pod")
	createAndDeployVerifyFormatOptions()
	ginkgo.By("verifying verifier pod is running", verifyFormatOptionsVerifierPodRunning)
}

func createAndDeployVerifyFormatOptions() {
	var err error
	appname := "format-options-verifier"
	labels := map[string]string{
		"role": "test",
		"app":  appname,
	}
	ginkgo.By("building app " + appname + " using above lvm volume")
	deployObj, err = deploy.NewBuilder().
		WithName(appname).
		WithNamespace(OpenEBSNamespace).
		WithLabelsNew(labels).
		WithSelectorMatchLabelsNew(labels).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabelsNew(labels).
				WithContainerBuilders(
					container.NewBuilder().
						WithImage("debian:stable-slim").
						WithName("verifier").
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithEnvsNew(
							[]corev1.EnvVar{
								{
									Name:  "MIN_INODES",
									Value: "5000000",
								},
							},
						).
						WithCommandNew(
							[]string{
								"bash",
								"-c",
								`test "$(df -i '/mnt/datadir' | tail -n 1 | awk '{print $2}')" -ge "$MIN_INODES" && tail -f /dev/null`,
							},
						).
						WithLifeCycle(
							&corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"bash",
											"-c",
											"kill -9 $(pidof tail)",
										},
									},
								},
							},
						).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								{
									Name: "datavol1",
									// If this path changes, modify the above command line accordingly (df command).
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

	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while building app deployement {%s}", appname)

	deployObj, err = DeployClient.WithNamespace(OpenEBSNamespace).Create(deployObj)
	gomega.Expect(err).ShouldNot(
		gomega.HaveOccurred(),
		"while creating pod {%s} in namespace {%s}",
		appname,
		OpenEBSNamespace,
	)
}

func verifyFormatOptionsVerifierPodRunning() {
	var err error
	appName := "format-options-verifier"
	labelValue := fmt.Sprintf("role=test,app=%s", appName)
	gomega.Eventually(func() bool {
		appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
			List(metav1.ListOptions{
				LabelSelector: labelValue,
			})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while verifying application pod")
		return len(appPod.Items) == 1
	},
		60, 5).
		Should(gomega.BeTrue())

	status := IsPodRunningEventually(OpenEBSNamespace, appPod.Items[0].Name)
	gomega.Expect(status).To(gomega.Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
}

func createDeployVerifyBlockApp() {
	ginkgo.By("creating and deploying app pod", createAndDeployBlockAppPod)
	ginkgo.By("verifying app pod is running", verifyAppPodRunning)
}

func verifyAppPodRunning() {
	var err error
	for _, appName := range appNames {
		labelValue := fmt.Sprintf("role=test,app=%s", appName)
		gomega.Eventually(func() bool {
			appPod, err = PodClient.WithNamespace(OpenEBSNamespace).
				List(metav1.ListOptions{
					LabelSelector: labelValue,
				})
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "while verifying application pod")
			return len(appPod.Items) == 1
		},
			60, 5).
			Should(gomega.BeTrue())

		status := IsPodRunningEventually(OpenEBSNamespace, appPod.Items[0].Name)
		gomega.Expect(status).To(gomega.Equal(true), "while checking status of pod {%s}", appPod.Items[0].Name)
	}
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
	ginkgo.By(fmt.Sprintf("scaling controller plugin deployment %v to size %v", controllerDeployment, num))

	scale, err := K8sClient.AppsV1().Deployments(OpenEBSNamespace).
		GetScale(context.Background(), controllerDeployment, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"fetch current replica of deployment %v", controllerDeployment)
	existingReplicas := scale.Spec.Replicas

	if scale.Spec.Replicas == num {
		return existingReplicas
	}
	scale.Spec.Replicas = num
	scale, err = K8sClient.AppsV1().Deployments(OpenEBSNamespace).
		UpdateScale(context.Background(), controllerDeployment, scale, metav1.UpdateOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"update replicas of deployment %v to %v", controllerDeployment, num)

	scaled := gomega.Eventually(func() bool {
		scale, err = K8sClient.AppsV1().Deployments(OpenEBSNamespace).
			GetScale(context.Background(), controllerDeployment, metav1.GetOptions{})
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		return scale.Spec.Replicas == num
	}, 120, 10).
		Should(gomega.BeTrue())
	gomega.Expect(scaled).To(gomega.BeTrue(),
		"failed to scale up deployment %v to size %v", controllerDeployment, num)
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
		DaemonSets(OpenEBSNamespace).
		Get(context.Background(), nodeDaemonSet, metav1.GetOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"fetching node plugin daemonset %v", nodeDaemonSet)
	policy := metav1.DeletePropagationForeground
	err = K8sClient.AppsV1().
		DaemonSets(OpenEBSNamespace).
		Delete(context.Background(), nodeDaemonSet, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"deleting node plugin daemonset %v", nodeDaemonSet)

	ginkgo.By("waiting for deletion of node plugin pods")
	status := gomega.Eventually(func() bool {
		_, err = K8sClient.AppsV1().
			DaemonSets(OpenEBSNamespace).
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
		DaemonSets(OpenEBSNamespace).
		Create(context.Background(), ds, metav1.CreateOptions{})
	gomega.Expect(err).To(
		gomega.BeNil(),
		"creating node plugin daemonset %v", nodeDaemonSet)
}

// Creates a k8s client using kubeconfig path.
func getk8sClient() (client *kubernetes.Clientset) {
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	gomega.Expect(err).To(
		gomega.BeNil(),
		"Could not created a k8s client",
	)
	client, c_err := kubernetes.NewForConfig(config)
	gomega.Expect(c_err).To(
		gomega.BeNil(),
		"Could not created a k8s client",
	)
	return client
}

// Lists k8s nodes.
func listNodes(client *kubernetes.Clientset) (nodes *corev1.NodeList) {
	nodes, n_err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	gomega.Expect(n_err).To(
		gomega.BeNil(),
		"Could not list nodes",
	)
	return nodes
}

// Cordons all k8s nodes.
func cordonk8sNode() {
	client := getk8sClient()
	nodes := listNodes(client)
	for _, node := range nodes.Items {
		if !node.Spec.Unschedulable {
			c_err := cordonNode(client, &node)
			gomega.Expect(c_err).To(
				gomega.BeNil(),
				"Could not cordon node",
			)
		}
	}
}

// UnCordons all k8s nodes.
func uncordonk8sNode() {
	client := getk8sClient()
	nodes := listNodes(client)
	for _, node := range nodes.Items {
		if node.Spec.Unschedulable {
			c_err := uncordonNode(client, &node)
			gomega.Expect(c_err).To(
				gomega.BeNil(),
				"Could not uncordon node",
			)
		}
	}
}

// Adds cordon taint to a specific node.
func cordonNode(clientset *kubernetes.Clientset, node *corev1.Node) error {
	updatedNode := node.DeepCopy()
	updatedNode.Spec.Unschedulable = true

	_, err := clientset.CoreV1().Nodes().Update(context.TODO(), updatedNode, metav1.UpdateOptions{})
	return err

}

// Removes cordon taint from a specific node.
func uncordonNode(clientset *kubernetes.Clientset, node *corev1.Node) error {
	updatedNode := node.DeepCopy()
	updatedNode.Spec.Unschedulable = false

	_, err := clientset.CoreV1().Nodes().Update(context.TODO(), updatedNode, metav1.UpdateOptions{})
	return err

}
