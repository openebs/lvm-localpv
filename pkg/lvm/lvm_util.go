/*
Copyright 2017 The Kubernetes Authors.

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

package lvm

import (
	"os/exec"

	"strings"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"k8s.io/klog"
)

// lvm related constants
const (
	DevPath       = "/dev/"
	DevMapperPath = "/dev/mapper/"
)

// lvm command related constants
const (
	VGCreate = "vgcreate"
	LVCreate = "lvcreate"
	LVRemove = "lvremove"
)

// builldLVMCreateArgs returns lvcreate command for the volume
func buildLVMCreateArgs(vol *apis.LVMVolume) []string {
	var LVMVolArg []string

	volume := vol.Name
	size := vol.Spec.Capacity + "b"

	if len(vol.Spec.Capacity) != 0 {
		LVMVolArg = append(LVMVolArg, "-L", size)
	}

	if len(vol.Spec.VolGroup) != 0 {
		LVMVolArg = append(LVMVolArg, "-n", volume)
	}

	LVMVolArg = append(LVMVolArg, vol.Spec.VolGroup)

	return LVMVolArg
}

// builldLVMDestroyArgs returns lvmremove command for the volume
func buildLVMDestroyArgs(vol *apis.LVMVolume) []string {
	var LVMVolArg []string

	dev := DevPath + vol.Spec.VolGroup + "/" + vol.Name

	LVMVolArg = append(LVMVolArg, "-y", dev)

	return LVMVolArg
}

// CreateVolume creates the lvm volume
func CreateVolume(vol *apis.LVMVolume) error {
	volume := vol.Spec.VolGroup + "/" + vol.Name

	args := buildLVMCreateArgs(vol)
	cmd := exec.Command(LVCreate, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		klog.Errorf(
			"lvm: could not create volume %v cmd %v error: %s", volume, args, string(out),
		)
		return err
	}
	klog.Infof("lvm: created volume %s", volume)

	return nil
}

// DestroyVolume deletes the lvm volume
func DestroyVolume(vol *apis.LVMVolume) error {
	volume := vol.Spec.VolGroup + "/" + vol.Name

	args := buildLVMDestroyArgs(vol)
	cmd := exec.Command(LVRemove, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		klog.Errorf(
			"lvm: could not destroy volume %v cmd %v error: %s", volume, args, string(out),
		)
		return err
	}

	klog.Infof("lvm: destroyed volume %s", volume)

	return nil
}

// GetVolumeDevPath returns devpath for the given volume
func GetVolumeDevPath(vol *apis.LVMVolume) (string, error) {
	// LVM doubles the hiphen for the mapper device name
	// and uses single hiphen to separate volume group from volume
	vg := strings.Replace(vol.Spec.VolGroup, "-", "--", -1)

	lv := strings.Replace(vol.Name, "-", "--", -1)
	dev := DevMapperPath + vg + "-" + lv

	return dev, nil
}
