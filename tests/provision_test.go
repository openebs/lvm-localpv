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
	. "github.com/onsi/ginkgo"
)

var _ = Describe("[lvmpv] TEST VOLUME PROVISIONING", func() {
	Context("App is deployed with lvm driver", func() {
		It("Running volume Creation Test", volumeCreationTest)
	})
})

func fsVolCreationTest() {
	fstypes := []string{"ext4", "xfs", "btrfs"}
	for _, fstype := range fstypes {
		By("####### Creating the storage class : " + fstype + " #######")
		createFstypeStorageClass(fstype)
		By("creating and verifying PVC bound status", createAndVerifyPVC)
		By("Creating and deploying app pod", createDeployVerifyApp)
		By("verifying LVMVolume object", VerifyLVMVolume)

		// btrfs does not support online resize
		if fstype != "btrfs" {
			resizeAndVerifyPVC(true, "8Gi")
		}
		// do not resize after creating the snapshot(not supported)
		createSnapshot(pvcName, snapName)
		verifySnapshotCreated(snapName)

		if fstype != "btrfs" {
			// if snapshot is there, resize should fail
			resizeAndVerifyPVC(false, "10Gi")
		}

		By("Deleting the application deployment")
		deleteAppDeployment(appName)

		// delete PVC should succeed
		By("Deleting the PVC")
		deleteAndVerifyPVC(pvcName)

		// PV should be present after PVC deletion since snapshot is present
		By("Verifying that PV exists after PVC deletion")
		verifyPVForPVC(true, pvcName)

		By("Deleting snapshot")
		deleteSnapshot(pvcName, snapName)

		By("Verifying that PV is deleted after snapshot deletion")
		verifyPVForPVC(false, pvcName)
		By("Deleting storage class", deleteStorageClass)
	}
}

func blockVolCreationTest() {
	By("Creating default storage class", createStorageClass)
	By("creating and verifying PVC bound status", createAndVerifyBlockPVC)

	By("Creating and deploying app pod", createDeployVerifyBlockApp)
	By("verifying LVMVolume object", VerifyLVMVolume)
	By("Deleting application deployment")
	deleteAppDeployment(appName)
	By("Deleting pvc")
	deleteAndVerifyPVC(pvcName)
	By("Deleting storage class", deleteStorageClass)
}

func thinVolCreationTest() {
	By("Creating thinProvision storage class", createThinStorageClass)
	By("creating and verifying PVC bound status", createAndVerifyPVC)

	By("Creating and deploying app pod", createDeployVerifyApp)
	By("verifying LVMVolume object", VerifyLVMVolume)
	By("Deleting application deployment")
	deleteAppDeployment(appName)
	By("Deleting pvc")
	deletePVC(pvcName)
	By("Deleting thinProvision storage class", deleteStorageClass)
}

func volumeCreationTest() {
	By("Running volume creation test", fsVolCreationTest)
	By("Running block volume creation test", blockVolCreationTest)
	By("Running thin volume creation test", thinVolCreationTest)
}
