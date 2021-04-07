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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"strings"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	VGList   = "vgs"

	LVCreate = "lvcreate"
	LVRemove = "lvremove"
	LVExtend = "lvextend"

	PVScan = "pvscan"
)

// ExecError holds the process output along with underlying
// error returned by exec.CombinedOutput function.
type ExecError struct {
	Output []byte
	Err    error
}

// Error implements the error interface.
func (e *ExecError) Error() string {
	return fmt.Sprintf("%v - %v", string(e.Output), e.Err)
}

func newExecError(output []byte, err error) error {
	if err == nil {
		return nil
	}
	return &ExecError{
		Output: output,
		Err:    err,
	}
}

// builldLVMCreateArgs returns lvcreate command for the volume
func buildLVMCreateArgs(vol *apis.LVMVolume) []string {
	var LVMVolArg []string

	volume := vol.Name
	size := vol.Spec.Capacity + "b"
	// thinpool name required for thinProvision volumes
	pool := vol.Spec.VolGroup + "_thinpool"

	if len(vol.Spec.Capacity) != 0 {
		// check if thin pool exists for given volumegroup requested thin volume
		if strings.TrimSpace(vol.Spec.ThinProvision) != "yes" || !lvThinExists(vol.Spec.VolGroup, pool) {
			LVMVolArg = append(LVMVolArg, "-L", size)
		}
	}

	// command to create thinpool and thin volume if thinProvision is enabled
	// `lvcreate -L 1G -T lvmvg/mythinpool -V 1G -n thinvol`
	if strings.TrimSpace(vol.Spec.ThinProvision) == "yes" {
		LVMVolArg = append(LVMVolArg, "-T", vol.Spec.VolGroup+"/"+pool, "-V", size)
	}

	if len(vol.Spec.VolGroup) != 0 {
		LVMVolArg = append(LVMVolArg, "-n", volume)
	}

	if strings.TrimSpace(vol.Spec.ThinProvision) != "yes" {
		LVMVolArg = append(LVMVolArg, vol.Spec.VolGroup)
	}
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

	volExists, err := CheckVolumeExists(vol)
	if err != nil {
		return err
	}
	if volExists {
		klog.Infof("lvm: volume (%s) already exists, skipping its creation", volume)
		return nil
	}

	args := buildLVMCreateArgs(vol)
	cmd := exec.Command(LVCreate, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		err = newExecError(out, err)
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

	volExists, err := CheckVolumeExists(vol)
	if err != nil {
		return err
	}
	if !volExists {
		klog.Infof("lvm: volume (%s) doesn't exists, skipping its deletion", volume)
		return nil
	}

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

// CheckVolumeExists validates if lvm volume exists
func CheckVolumeExists(vol *apis.LVMVolume) (bool, error) {
	devPath, err := GetVolumeDevPath(vol)
	if err != nil {
		return false, err
	}
	if _, err = os.Stat(devPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

// builldVolumeResizeArgs returns resize command for the lvm volume
func buildVolumeResizeArgs(vol *apis.LVMVolume, resizefs bool) []string {
	var LVMVolArg []string

	dev := DevPath + vol.Spec.VolGroup + "/" + vol.Name
	size := vol.Spec.Capacity + "b"

	LVMVolArg = append(LVMVolArg, dev, "-L", size)

	if resizefs == true {
		LVMVolArg = append(LVMVolArg, "-r")
	}

	return LVMVolArg
}

// ResizeLVMVolume resizes the volume
func ResizeLVMVolume(vol *apis.LVMVolume, resizefs bool) error {
	volume := vol.Spec.VolGroup + "/" + vol.Name

	args := buildVolumeResizeArgs(vol, resizefs)
	cmd := exec.Command(LVExtend, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		klog.Errorf(
			"lvm: could not resize the volume %v cmd %v error: %s", volume, args, string(out),
		)
	}

	return err
}

func buildLVMSnapCreateArgs(snap *apis.LVMSnapshot) []string {
	var LVMSnapArg []string

	volName := snap.Labels[LVMVolKey]
	volPath := DevPath + snap.Spec.VolGroup + "/" + volName
	size := snap.Spec.Capacity + "b"

	LVMSnapArg = append(LVMSnapArg,
		// snapshot argument
		"--snapshot",
		// name of snapshot
		"--name", getLVMSnapName(snap.Name),
		// size of the snapshot, will be same as source volume
		"--size", size,
		// set the permission to make the snapshot read-only. By default LVM snapshots are RW
		"--permission", "r",
		// volume to snapshot
		volPath,
	)

	return LVMSnapArg
}

func buildLVMSnapDestroyArgs(snap *apis.LVMSnapshot) []string {
	var LVMSnapArg []string

	dev := DevPath + snap.Spec.VolGroup + "/" + getLVMSnapName(snap.Name)

	LVMSnapArg = append(LVMSnapArg, "-y", dev)

	return LVMSnapArg
}

// CreateSnapshot creates the lvm volume snapshot
func CreateSnapshot(snap *apis.LVMSnapshot) error {

	volume := snap.Labels[LVMVolKey]

	snapVolume := snap.Spec.VolGroup + "/" + getLVMSnapName(snap.Name)

	args := buildLVMSnapCreateArgs(snap)
	cmd := exec.Command(LVCreate, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		klog.Errorf("lvm: could not create snapshot %s cmd %v error: %s", snapVolume, args, string(out))
		return err
	}

	klog.Infof("created snapshot %s from %s", snapVolume, volume)
	return nil

}

// DestroySnapshot deletes the lvm volume snapshot
func DestroySnapshot(snap *apis.LVMSnapshot) error {
	snapVolume := snap.Spec.VolGroup + "/" + getLVMSnapName(snap.Name)

	args := buildLVMSnapDestroyArgs(snap)
	cmd := exec.Command(LVRemove, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		klog.Errorf("lvm: could not remove snapshot %s cmd %v error: %s", snapVolume, args, string(out))
		return err
	}

	klog.Infof("removed snapshot %s", snapVolume)
	return nil

}

// getSnapName is used to remove the snapshot prefix from the snapname. since names starting
// with "snapshot" are reserved in lvm2
func getLVMSnapName(snapName string) string {
	return strings.TrimPrefix(snapName, "snapshot-")
}

func decodeVgsJSON(raw []byte) ([]apis.VolumeGroup, error) {
	output := &struct {
		Report []struct {
			VolumeGroups []map[string]string `json:"vg"`
		} `json:"report"`
	}{}
	var err error
	if err = json.Unmarshal(raw, output); err != nil {
		return nil, err
	}

	if len(output.Report) != 1 {
		return nil, fmt.Errorf("expected exactly one lvm report")
	}

	items := output.Report[0].VolumeGroups
	vgs := make([]apis.VolumeGroup, 0, len(items))
	for _, item := range items {
		var vg apis.VolumeGroup
		if vg, err = parseVolumeGroup(item); err != nil {
			return vgs, err
		}
		vgs = append(vgs, vg)
	}
	return vgs, nil
}

func parseVolumeGroup(m map[string]string) (apis.VolumeGroup, error) {
	var vg apis.VolumeGroup
	vg.Name = m["vg_name"]
	vg.UUID = m["vg_uuid"]

	int32Map := map[string]*int32{
		"pv_count": &vg.PVCount,
		"lv_count": &vg.LVCount,
	}
	for key, value := range int32Map {
		count, err := strconv.Atoi(m[key])
		if err != nil {
			err = fmt.Errorf("invalid format of %v=%v for vg %v: %v", key, m[key], vg.Name, err)
		}
		*value = int32(count)
	}

	resQuantityMap := map[string]*resource.Quantity{
		"vg_size": &vg.Size,
		"vg_free": &vg.Free,
	}

	for key, value := range resQuantityMap {
		sizeBytes, err := strconv.ParseInt(
			strings.TrimSuffix(strings.ToLower(m[key]), "b"),
			10, 64)
		if err != nil {
			err = fmt.Errorf("invalid format of %v=%v for vg %v: %v", key, m[key], vg.Name, err)
		}
		quantity := resource.NewQuantity(sizeBytes, resource.BinarySI)
		*value = *quantity //
	}
	return vg, nil
}

// ReloadLVMMetadataCache refreshes lvmetad daemon cache used for
// serving vgs or other lvm utility.
func ReloadLVMMetadataCache() error {
	args := []string{"--cache"}
	cmd := exec.Command(PVScan, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("lvm: reload lvm metadata cache: %v - %v", string(output), err)
		return err
	}
	return nil
}

// ListLVMVolumeGroup invokes `vgs` to list all the available volume
// groups in the node.
func ListLVMVolumeGroup() ([]apis.VolumeGroup, error) {
	if err := ReloadLVMMetadataCache(); err != nil {
		return nil, err
	}

	args := []string{
		"--options", "vg_all",
		"--reportformat", "json",
		"--units", "b",
	}
	cmd := exec.Command(VGList, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("lvm: list volume group cmd %v: %v", args, err)
		return nil, err
	}
	return decodeVgsJSON(output)
}

// lvThinExists verifies if thin pool/volume already exists for given volumegroup
func lvThinExists(vg string, name string) bool {
	cmd := exec.Command("lvs", vg+"/"+name, "--noheadings", "-o", "lv_name")
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Infof("unable to list existing volumes:%v", err)
		return false
	}
	return name == strings.TrimSpace(string(out))
}
