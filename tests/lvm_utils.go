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
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

// This creates loopdevice using the size passed as arg,
// Uses the new loop device to create PV. returns loopdevice name to the caller.
func createPV(size int) string {
	ginkgo.By("Creating Pv")

	back_file_args := []string{
		"mktemp",
		"-t",
		"openebs_lvm_localpv_disk_XXXXX",
		"--dry-run",
	}

	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		back_file_args = append(back_file_args, "-p", dataDir)
	}

	file, _, _ := execAtLocal("sudo", nil, back_file_args...)

	file_str := strings.TrimSpace(string(file[:]))
	size_str := strconv.Itoa(size) + "G"
	device_args := []string{
		"truncate",
		"-s",
		size_str, file_str,
	}
	_, _, err := execAtLocal("sudo", nil, device_args...)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "create device failed")

	args_loop := []string{
		"losetup",
		"-f",
		file_str, "--show",
	}
	stdout_loop, _, err := execAtLocal("sudo", nil, args_loop...)
	gomega.Expect(err).To(gomega.BeNil(), "loop device create failed")

	stdout_loop_str := strings.TrimSpace(string(stdout_loop[:]))
	args_pv := []string{
		"pvcreate",
		stdout_loop_str,
	}
	_, _, err_pv := execAtLocal("sudo", nil, args_pv...)
	gomega.Expect(err_pv).To(gomega.BeNil(), "pv create failed")
	return stdout_loop_str
}

// Gets lv_count of a specified vg, returns false if its not empty at the end of poll.
func vgEmpty(name string) bool {
	args_lvs := []string{
		"vgs",
		name,
		"--options",
		"lv_count",
		"--noheadings",
	}
	lvs, _, _ := execAtLocal("sudo", nil, args_lvs...)
	lvs_str := strings.TrimSpace(string(lvs))
	lv_cnt, _ := strconv.Atoi(lvs_str)
	if lv_cnt != 0 {
		return false
	} else {
		return true
	}
}

// Does pvremove on specified device. Deletes loop device and the file backing loop device.
func removePV(device string) {
	ginkgo.By("Removing pv")
	args_pv := []string{
		"pvremove",
		device,
		"-y",
	}
	_, _, err_pv := execAtLocal("sudo", nil, args_pv...)
	gomega.Expect(err_pv).To(gomega.BeNil(), "pv remove failed")

	args_lo := []string{
		"losetup",
		device,
		"-O",
		"BACK-FILE",
		"--noheadings",
	}
	dev, _, _ := execAtLocal("sudo", nil, args_lo...)
	dev_str := strings.TrimSpace(string(dev))

	args_loop := []string{
		"losetup",
		"-d",
		device,
	}
	_, _, err_loop := execAtLocal("sudo", nil, args_loop...)
	gomega.Expect(err_loop).To(gomega.BeNil(), "loop device remove failed")

	args_file := []string{
		"rm",
		"-f",
		dev_str,
	}
	_, _, err_file := execAtLocal("sudo", nil, args_file...)
	gomega.Expect(err_file).To(gomega.BeNil(), "file remove failed")
}

// Creates vg on the specified device, Device passed should be a pv.
func createVg(name string, device string) {
	ginkgo.By("Creating vg")
	args_vg := []string{
		"vgcreate", name,
		device,
	}
	_, _, err_vg := execAtLocal("sudo", nil, args_vg...)
	gomega.Expect(err_vg).To(gomega.BeNil(), "vg create failed")
}

// Takes vg name and pv device, extends vg using the supplied pv.
func extendVg(name string, device string) {
	ginkgo.By("Extending vg")
	args_vg := []string{
		"vgextend", name,
		device,
	}
	_, _, err_vg := execAtLocal("sudo", nil, args_vg...)
	gomega.Expect(err_vg).To(gomega.BeNil(), "vg extend failed")
}

// Does vgremove on specified vg with -y flag if vg isnt empty after few retires.
func removeVg(name string) {
	ginkgo.By("Removing vg")
	retries := 3
	current_retry := 0
	args_vg := []string{
		"vgremove",
		name,
	}
	for {
		if current_retry < retries {
			vg_empty := vgEmpty(name)
			if vg_empty {
				break
			} else {
				fmt.Printf("lv in vg during retry %d\n", current_retry)
			}
		} else {
			fmt.Printf("vg still not empty after 6 seconds, moving on with force delete\n")
			args_vg = append(args_vg, "-f")
			break
		}
		current_retry += 1
		time.Sleep(2 * time.Second)
	}
	_, _, err_vg := execAtLocal("sudo", nil, args_vg...)
	gomega.Expect(err_vg).To(gomega.BeNil(), "vg remove failed")
}

// enable the monitoring on thinpool created for test, on local node which
// is part of single node cluster.
func enableThinpoolMonitoring() {
	ginkgo.By("Enable thinpool monitoring")
	lv := VOLGROUP + "/" + pvcObj.Spec.VolumeName

	args := []string{
		"lvdisplay", "--columns",
		"--options", "pool_lv",
		"--noheadings",
		lv,
	}
	stdout, _, err := execAtLocal("sudo", nil, args...)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "display LV")
	gomega.Expect(strings.TrimSpace(string(stdout))).To(gomega.Not(gomega.Equal("")), "get thinpool LV")

	thinpool := VOLGROUP + "/" + strings.TrimSpace(string(stdout))

	args = []string{
		"lvchange",
		"--monitor", "y",
		thinpool,
	}

	_, _, err = execAtLocal("sudo", nil, args...)
	gomega.Expect(err).To(gomega.BeNil(), "run lvchange command")
}

// verify that the thinpool has extended in capacity to an expected size.
func VerifyThinpoolExtend() {
	ginkgo.By("Verify thinpool extend")
	expect_size, _ := strconv.ParseInt(expanded_capacity, 10, 64)
	lv := VOLGROUP + "/" + pvcObj.Spec.VolumeName

	args := []string{
		"lvdisplay", "--columns",
		"--options", "pool_lv",
		"--noheadings",
		lv,
	}

	//stdout will contain the pool name
	stdout, _, err := execAtLocal("sudo", nil, args...)
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "display LV")
	gomega.Expect(strings.TrimSpace(string(stdout))).To(gomega.Not(gomega.Equal("")), "get thinpool LV")

	thinpool := VOLGROUP + "/" + strings.TrimSpace(string(stdout))

	args = []string{
		"lvdisplay", "--columns",
		"--options", "lv_size",
		"--units", "b",
		"--noheadings",
		thinpool,
	}

	gomega.Eventually(func() bool {
		// stdout will contain the size
		stdout, _, err = execAtLocal("sudo", nil, args...)
		gomega.Expect(err).To(gomega.BeNil(), "display thinpool LV")

		// Remove unit suffix from the size.
		size_str := strings.TrimSuffix(strings.TrimSpace(string(stdout)), "B")
		// This expectation is a factor of the lvm.conf settings we do from ci-test.sh
		// and the original volume size.
		size_int64, _ := strconv.ParseInt(size_str, 10, 64)
		return size_int64 == expect_size
	},
		45*time.Second, 5*time.Second).
		Should(gomega.BeTrue())
}

// Verify if thinpool is deleted or not as per input expectation.
func VerifyThinPoolDeletion(vg string, thinpool string, expect_deleted bool) {
	thinpool_lv := vg + "/" + thinpool

	args := []string{
		"lvs",
		"--noheadings",
		thinpool_lv,
	}

	stdout, _, err := execAtLocal("sudo", nil, args...)
	fmt.Printf("VerifyThinPoolDeletion: expect_deleted: %v, err: %v, output: \n%s\n", expect_deleted, err, stdout)
	// When expect_deleted is true it means we are expecting `lvs` command to list the thinpool, which
	// means err will be nil. Otherwise, err is non-nil (error code 5) when thinpool is not found.
	if expect_deleted {
		gomega.Expect(err).Should(gomega.HaveOccurred(), "shouldn't find the thinpool")
	} else {
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred(), "should've found the thinpool")
	}
}
