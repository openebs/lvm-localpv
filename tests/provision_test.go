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
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = Describe("[lvmpv] TEST VOLUME PROVISIONING", func() {
	Context("App is deployed with lvm driver", func() {
		It("Running volume Creation Tests", volumeCreationTest)
		It("Running scheduling Tests", schedulingTest)
		It("Running volume/snapshot Capacity Tests", capacityTest)
		It("Running thin snapshot restore Tests", thinSnapshotRestoreToNewThinVolume)
		It("Running volume Deletion Tests", volumeDeletionTest)
	})
})

func deleteAppAndPvc(appnames []string, pvcname string) {
	for _, appName := range appnames {
		By("Deleting the application deployment " + appName)
		deleteAppDeployment(appName)
	}
	By("Deleting the PVC")
	deleteAndVerifyPVC(pvcname)
}

func setupVg(size int, name string) string {
	device := createPV(size)
	createVg(name, device)
	return device
}

func cleanupVg(device string, name string) {
	removeVg(name)
	removePV(device)
}

func fsVolCreationTest() {
	fstypes := []string{"ext4", "xfs", "btrfs"}
	for _, fstype := range fstypes {
		By("####### Creating the storage class : " + fstype + " #######")
		createFstypeStorageClass(fstype)
		By("Creating and verifying PVC bound status")
		createAndVerifyPVC(true)
		By("Creating and deploying app pod")
		createDeployVerifyApp(appNames, pvcObj)
		By("Verifying LVMVolume object to be Ready")
		VerifyLVMVolume(true, "", pvcObj)

		resizeAndVerifyPVC(true, "8Gi")
		// do not resize after creating the snapshot(not supported)
		createSnapshot(pvcName, snapName, snapYAML)
		verifySnapshotCreated(snapName)

		if fstype != "btrfs" {
			// if snapshot is there, resize should fail
			resizeAndVerifyPVC(false, "10Gi")
		}

		deleteAppAndPvc(appNames, pvcName)

		// PV should be present after PVC deletion since snapshot is present
		By("Verifying that PV exists after PVC deletion")
		verifyPVForPVC(true, pvcName)

		deleteSnapshot(pvcName, snapName, snapYAML)

		By("Verifying that PV is deleted after snapshot deletion")
		verifyPVForPVC(false, pvcName)
		By("Deleting storage class", deleteStorageClass)
	}
}

func formatOptionsTest() {
	formatOptions := "-b 4096 -N 5000000"
	By("####### Creating the storage class with formatOptions : " + formatOptions + " #######")
	createFormatOptionsStorageClass(formatOptions)
	By("Creating and verifying PVC bound status")
	createAndVerifyPVC(true)
	By("Creating and deploying app pod", createDeployVerifyFormatOptions)
	By("Verifying LVMVolume object to be Ready")
	VerifyLVMVolume(true, "", pvcObj)
	By("Deleting verifier pvc/pod")
	deleteAppAndPvc([]string{"format-options-verifier"}, pvcName)
	By("Verifying that PV is deleted after deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func blockVolCreationTest() {
	By("Creating default storage class", createStorageClass)
	By("Creating and verifying PVC bound status")
	createAndVerifyBlockPVC(true)
	By("Creating and deploying app pod", createDeployVerifyBlockApp)
	By("Verifying LVMVolume object to be Ready")
	VerifyLVMVolume(true, "", pvcObj)
	By("Online resizing the block volume")
	resizeAndVerifyPVC(true, "8Gi")
	By("Creating snapshot")
	createSnapshot(pvcName, snapName, snapYAML)
	By("Verifying snapshot")
	verifySnapshotCreated(snapName)
	deleteAppAndPvc(appNames, pvcName)
	By("Verifying that PV exists after PVC deletion")
	verifyPVForPVC(true, pvcName)
	By("Deleting snapshot")
	deleteSnapshot(pvcName, snapName, snapYAML)
	By("Verifying that PV is deleted after snapshot deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func vgExtendNeededForProvsioningTest() {
	device_0 := setupVg(7, "lvmvgdiff")
	device := setupVg(3, "lvmvg")
	device_1 := createPV(4)
	defer removePV(device_1)
	defer cleanupVg(device_0, "lvmvgdiff")
	defer cleanupVg(device, "lvmvg")
	By("Creating default storage class", createStorageClass)
	By("Creating and verifying PVC Not Bound status")
	createAndVerifyPVC(false)
	By("Verifying LVMVolume object to be not Ready")
	VerifyLVMVolume(false, "", pvcObj)
	extendVg("lvmvg", device_1)
	By("Verifying PVC bound status after vg extend")
	VerifyBlockPVC()
	By("Verifying LVMVolume object to be Ready after vg extend")
	VerifyLVMVolume(true, "", pvcObj)
	By("Deleting pvc")
	deleteAndVerifyPVC(pvcName)
	By("Verifying that PV doesnt exists after PVC deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func vgPatternMatchPresentTest() {
	device := setupVg(20, "lvmvg112")
	device_1 := setupVg(20, "lvmvg")
	defer cleanupVg(device_1, "lvmvg")
	defer cleanupVg(device, "lvmvg112")
	By("Creating custom storage class with non existing vg parameter", createVgPatternStorageClass)
	By("Creating and verifying PVC Bound status")
	createAndVerifyPVC(true)
	By("Verifying LVMVolume object to be Ready")
	VerifyLVMVolume(true, "lvmvg112", pvcObj)
	deleteAndVerifyPVC(pvcName)
	By("Verifying that PV doesnt exists after PVC deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func scheduleOnCordonedNodeTest() {
	device := setupVg(20, "lvmvg")
	defer cleanupVg(device, "lvmvg")
	By("Creating storage class", createStorageClass)
	By("Cordoning the node", cordonk8sNode)
	By("Creating and verifying PVC Bound status. It should not be Bound")
	createAndVerifyPVC(false)
	By("Uncordon the node", uncordonk8sNode)
	By("Verify the PVC gets Bound")
	verifyPVCStatus(pvcName, true)
	deleteAndVerifyPVC(pvcName)
	By("Verifying that PV doesnt exists after PVC deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func vgPatternNoMatchPresentTest() {
	device := setupVg(20, "lvmvg212")
	device_1 := setupVg(20, "lvmvg")
	defer cleanupVg(device_1, "lvmvg")
	defer cleanupVg(device, "lvmvg212")
	By("Creating custom storage class with non existing vg parameter", createVgPatternStorageClass)
	By("Creating and verifying PVC Not Bound status")
	createAndVerifyPVC(false)
	By("Verifying LVMVolume object to be Not Ready")
	VerifyLVMVolume(false, "", pvcObj)
	deleteAndVerifyPVC(pvcName)
	By("Verifying that PV doesnt exists after PVC deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func vgSpecifiedNotPresentTest() {
	device := setupVg(40, "lvmvg")
	defer cleanupVg(device, "lvmvg")
	By("Creating custom storage class with non existing vg parameter", createStorageClassWithNonExistingVg)
	By("creating and verifying PVC Not Bound status")
	createAndVerifyPVC(false)
	By("Verifying LVMVolume object to be Not Ready")
	VerifyLVMVolume(false, "", pvcObj)
	By("Deleting pvc")
	deleteAndVerifyPVC(pvcName)
	By("Verifying that PV doesnt exists after PVC deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func sharedVolumeTest() {
	By("Creating shared LV storage class", createSharedVolStorageClass)
	By("creating and verifying PVC bound status")
	createAndVerifyPVC(true)
	//we use two fio app pods for this test.
	appNames = append(appNames, "fio-ci-1")
	By("Creating and deploying app pod")
	createDeployVerifyApp(appNames, pvcObj)
	By("Verifying LVMVolume object to be Not Ready")
	VerifyLVMVolume(true, "", pvcObj)
	By("Online resizing the shared volume")
	resizeAndVerifyPVC(true, "8Gi")
	deleteAppAndPvc(appNames, pvcName)
	By("Verifying that PV doesnt exists after PVC deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
	// Reset the app list back to original
	appNames = appNames[:len(appNames)-1]
}

func thinVolCreationTest() {
	By("Creating thinProvision storage class", createThinStorageClass)
	By("creating and verifying PVC bound status")
	createAndVerifyPVC(true)
	By("Creating and deploying app pod")
	createDeployVerifyApp(appNames, pvcObj)
	By("verifying LVMVolume object")
	VerifyLVMVolume(true, "", pvcObj)
	By("Online resizing the block volume")
	resizeAndVerifyPVC(true, "8Gi")
	By("create snapshot")
	createSnapshot(pvcName, snapName, snapYAML)
	By("verify snapshot")
	verifySnapshotCreated(snapName)
	deleteAppAndPvc(appNames, pvcName)
	By("Verifying that PV exists after PVC deletion")
	verifyPVForPVC(true, pvcName)
	By("Deleting snapshot")
	deleteSnapshot(pvcName, snapName, snapYAML)
	By("Verifying that PV is deleted after snapshot deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting thinProvision storage class", deleteStorageClass)
}

func thinVolCapacityTest() {
	By("Creating thinProvision storage class", createThinStorageClass)
	By("creating and verifying PVC bound status")
	createAndVerifyPVC(true)
	By("enabling monitoring on thinpool", enableThinpoolMonitoring)
	By("Creating and deploying app pod")
	createDeployVerifyApp(appNames, pvcObj)
	By("verifying thinpool auto-extended", VerifyThinpoolExtend)
	By("verifying LVMVolume object")
	VerifyLVMVolume(true, "", pvcObj)
	deleteAppAndPvc(appNames, pvcName)
	By("Verifying that PV doesnt exists after PVC deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting thinProvision storage class", deleteStorageClass)
}

func sizedSnapFSTest() {
	createFstypeStorageClass("ext4")
	By("creating and verifying PVC bound status")
	createAndVerifyPVC(true)
	By("Creating and deploying app pod")
	createDeployVerifyApp(appNames, pvcObj)
	By("verifying LVMVolume object")
	VerifyLVMVolume(true, "", pvcObj)
	createSnapshot(pvcName, snapName, sizedsnapYAML)
	verifySnapshotCreated(snapName)
	deleteAppAndPvc(appNames, pvcName)
	By("Verifying that PV exists before Snapshot deletion")
	verifyPVForPVC(true, pvcName)
	deleteSnapshot(pvcName, snapName, sizedsnapYAML)
	By("Verifying that PV doesnt exists after Snapshot deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func sizedSnapBlockTest() {
	By("Creating default storage class", createStorageClass)
	By("creating and verifying PVC bound status")
	createAndVerifyPVC(true)
	By("Creating and deploying app pod")
	createDeployVerifyApp(appNames, pvcObj)
	By("verifying LVMVolume object")
	VerifyLVMVolume(true, "", pvcObj)
	createSnapshot(pvcName, snapName, sizedsnapYAML)
	verifySnapshotCreated(snapName)
	deleteAppAndPvc(appNames, pvcName)
	By("Verifying that PV exists before Snapshot deletion")
	verifyPVForPVC(true, pvcName)
	deleteSnapshot(pvcName, snapName, sizedsnapYAML)
	By("Verifying that PV doesnt exists after Snapshot deletion")
	verifyPVForPVC(false, pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func sizedSnapshotTest() {
	By("Sized snapshot for filesystem volume", sizedSnapFSTest)
	By("Sized snapshot for block volume", sizedSnapBlockTest)
}

func leakProtectionTest() {
	By("Creating default storage class", createStorageClass)
	ds := deleteNodeDaemonSet() // ensure that provisioning remains in pending state.

	time.Sleep(30 * time.Second)

	By("Creating PVC", createPVC)
	time.Sleep(30 * time.Second) // wait for external provisioner to pick up new pvc
	By("Verify pending lvm volume resource")
	verifyPendingLVMVolume(getGeneratedVolName(pvcObj))

	existingSize := scaleControllerPlugin(0) // remove the external provisioner
	createNodeDaemonSet(ds)                  // provision the volume now by restoring node plugin
	By("Wait for lvm volume resource to become ready", WaitForLVMVolumeReady)

	deleteAndVerifyLeakedPVC(pvcName)
	scaleControllerPlugin(existingSize)

	gomega.Expect(IsPVCDeletedEventually(pvcName)).To(gomega.Equal(true),
		"failed to garbage collect leaked pvc")
	By("Deleting storage class", deleteStorageClass)
}

func volumeCreationTest() {
	device := setupVg(40, "lvmvg")
	defer cleanupVg(device, "lvmvg")
	By("###Running filesystem volume creation test###", fsVolCreationTest)
	By("###Running filesystem with formatOptions creation test###", formatOptionsTest)
	By("###Running block volume creation test###", blockVolCreationTest)
	By("###Running thin volume creation test###", thinVolCreationTest)
	By("###Running leak protection test###", leakProtectionTest)
	By("###Running shared volume for two app pods on same node test###", sharedVolumeTest)
}

func volumeDeletionTest() {
	device := setupVg(40, "lvmvg")
	defer cleanupVg(device, "lvmvg")
	By("Creating thinProvision storage class", createThinStorageClass)
	By("Creating and verifying a PVC bound status")
	createAndVerifyBlockPVCId(true, "1")
	By("Verifying LVMVolume object to be Ready")
	VerifyLVMVolume(true, "", pvcObj)
	By("Creating and verifying another PVC bound status")
	createAndVerifyBlockPVCId(true, "2")
	By("Verifying LVMVolume object to be Ready")
	VerifyLVMVolume(true, "", pvcObj)
	deleteAndVerifyPVC(pvcName + "-1")
	VerifyThinPoolDeletion("lvmvg", "lvmvg_thinpool", false)
	deleteAndVerifyPVC(pvcName + "-2")
	VerifyThinPoolDeletion("lvmvg", "lvmvg_thinpool", true)
}

func schedulingTest() {
	By("###Running vg extend needed to provision test###", vgExtendNeededForProvsioningTest)
	By("###Running vg specified in sc not present test###", vgSpecifiedNotPresentTest)
	By("###Running lvmnode has vg matching vgpattern test###", vgPatternMatchPresentTest)
	By("###Running lvmnode doesnt have vg matching vgpattern test###", vgPatternNoMatchPresentTest)
	By("###Running volume schedule on Cordoned node test###", scheduleOnCordonedNodeTest)
}

func capacityTest() {
	device := setupVg(40, "lvmvg")
	defer cleanupVg(device, "lvmvg")
	By("###Running thin volume capacity test###", thinVolCapacityTest)
	By("###Running sized snapshot test###", sizedSnapshotTest)
}

func thinSnapshotRestoreToNewThinVolume() {
	device := setupVg(40, "lvmvg")
	defer cleanupVg(device, "lvmvg")
	By("Creating thinProvision storage class", createThinStorageClass)
	By("creating and verifying PVC bound status")
	createAndVerifyPVC(true)
	By("Creating and deploying app pod")
	createDeployVerifyApp(appNames, pvcObj)
	By("verifying LVMVolume object")
	VerifyLVMVolume(true, "", pvcObj)
	By("create snapshot")
	createSnapshot(pvcName, snapName, snapYAML)
	By("verify snapshot")
	verifySnapshotCreated(snapName)
	By("creating and verifying PVC with thin snapshot as source")
	restorePvcObj := createAndVerifySnapshotRestorePVC(true)
	By("verifying restored LVMVolume object")
	VerifyLVMVolume(true, "", restorePvcObj)
	By("Creating and deploying restore app pod")
	createDeployVerifyApp(restoreAppNames, restorePvcObj)
	By("Deleting source thin snapshot")
	deleteSnapshot(pvcName, snapName, snapYAML)
	By("Delete origin app and pvc")
	deleteAppAndPvc(appNames, pvcName)
	By("Delete restored app and pvc")
	deleteAppAndPvc(restoreAppNames, restorePvcObj.Name)
	By("Deleting thinProvision storage class", deleteStorageClass)
}
