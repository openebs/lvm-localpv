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
	"path/filepath"
	"strconv"

	"strings"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
)

// lvm related constants
const (
	DevPath       = "/dev/"
	DevMapperPath = "/dev/mapper/"
	// MinExtentRoundOffSize represents minimum size (256Mi) to roundoff the volume
	// group size in case of thin pool provisioning
	MinExtentRoundOffSize = 268435456

	// BlockCleanerCommand is the command used to clean filesystem on the device
	BlockCleanerCommand = "wipefs"
)

// lvm command related constants
const (
	VGCreate = "vgcreate"
	VGList   = "vgs"

	LVCreate = "lvcreate"
	LVRemove = "lvremove"
	LVExtend = "lvextend"
	LVList   = "lvs"

	PVScan = "pvscan"

	YES = "yes"
)

// LogicalVolume specifies attributes of a given lv that exists on the node.
type LogicalVolume struct {

	// Name of the lvm logical volume(name: pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d)
	Name string

	// TODO
	// FullName is not getting populated currently. We will have to populate it for using it as a label in lvm metrics.
	// Full name of the lvm logical volume (fullName: linuxlvmvg/pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d)
	FullName string

	// TODO
	// UUID is not getting populated currently. We will have to populate it for using it as a label in lvm metrics.
	// UUID denotes a unique identity of a lvm logical volume.
	UUID string

	// Size specifies the total size of logical volume in Bytes
	Size int64

	// Path specifies LVM logical volume path
	Path string

	// TODO
	// DMPath is not getting populated currently. We will have to populate it for using it as a label in lvm metrics.
	// DMPath specifies device mapper path
	DMPath string

	// LVM logical volume device
	Device string

	// TODO
	// VGName is not getting populated currently. We will have to populate it for using it as a label in lvm metrics.
	// Name of the VG in which LVM logical volume is created
	VGName string
}

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
		if strings.TrimSpace(vol.Spec.ThinProvision) != YES {
			LVMVolArg = append(LVMVolArg, "-L", size)
		} else if !lvThinExists(vol.Spec.VolGroup, pool) {
			// thinpool size can't be equal or greater than actual volumegroup size
			LVMVolArg = append(LVMVolArg, "-L", getThinPoolSize(vol.Spec.VolGroup, vol.Spec.Capacity))
		}
	}

	// command to create thinpool and thin volume if thinProvision is enabled
	// `lvcreate -L 1G -T lvmvg/mythinpool -V 1G -n thinvol`
	if strings.TrimSpace(vol.Spec.ThinProvision) == YES {
		LVMVolArg = append(LVMVolArg, "-T", vol.Spec.VolGroup+"/"+pool, "-V", size)
	}

	if len(vol.Spec.VolGroup) != 0 {
		LVMVolArg = append(LVMVolArg, "-n", volume)
	}

	if strings.TrimSpace(vol.Spec.ThinProvision) != YES {
		LVMVolArg = append(LVMVolArg, vol.Spec.VolGroup)
	}

	// -y is used to wipe the signatures before creating LVM volume
	LVMVolArg = append(LVMVolArg, "-y")
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
	if vol.Spec.VolGroup == "" {
		klog.Infof("volGroup not set for lvm volume %v, skipping its deletion", vol.Name)
		return nil
	}

	volume := vol.Spec.VolGroup + "/" + vol.Name

	volExists, err := CheckVolumeExists(vol)
	if err != nil {
		return err
	}
	if !volExists {
		klog.Infof("lvm: volume (%s) doesn't exists, skipping its deletion", volume)
		return nil
	}

	err = removeVolumeFilesystem(vol)
	if err != nil {
		return err
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

	if resizefs {
		LVMVolArg = append(LVMVolArg, "-r")
	}

	return LVMVolArg
}

// ResizeLVMVolume resizes the underlying LVM volume and FS if resizefs
// is set to true
// Note:
//	1. Triggering `lvextend <dev_path> -L <size> -r` multiple times with
//     same size will not return any errors
//  2. Triggering `lvextend <dev_path> -L <size>` more than one time will
//     cause errors
func ResizeLVMVolume(vol *apis.LVMVolume, resizefs bool) error {

	// In case if resizefs is not enabled then check current size
	// before exapnding LVM volume(If volume is already expanded then
	// it might be error prone). This also makes ResizeLVMVolume func
	// idempotent
	if !resizefs {
		desiredVolSize, err := strconv.ParseUint(vol.Spec.Capacity, 10, 64)
		if err != nil {
			return err
		}

		curVolSize, err := getLVSize(vol)
		if err != nil {
			return err
		}

		// Trigger resize only when desired volume size is greater than
		// current volume size else return
		if desiredVolSize <= curVolSize {
			return nil
		}
	}

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

// getLVSize will return current LVM volume size in bytes
func getLVSize(vol *apis.LVMVolume) (uint64, error) {
	lvmVolumeName := vol.Spec.VolGroup + "/" + vol.Name

	args := []string{
		lvmVolumeName,
		"--noheadings",
		"-o", "lv_size",
		"--units", "b",
		"--nosuffix",
	}

	cmd := exec.Command(LVList, args...)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return 0, errors.Wrapf(
			err,
			"could not get size of volume %v output: %s",
			lvmVolumeName,
			string(raw),
		)
	}

	volSize, err := strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil {
		return 0, err
	}

	return volSize, nil
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

	ok, err := snapshotExists(snapVolume)
	if !ok {
		klog.Infof("lvm: snapshot %s does not exist, skipping deletion", snapVolume)
		return nil
	}

	if err != nil {
		klog.Errorf("lvm: error checking for snapshot %s, error: %v", snapVolume, err)
		return err
	}

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
	var count int
	var sizeBytes int64
	var err error

	vg.Name = m["vg_name"]
	vg.UUID = m["vg_uuid"]

	int32Map := map[string]*int32{
		"pv_count": &vg.PVCount,
		"lv_count": &vg.LVCount,
	}
	for key, value := range int32Map {
		count, err = strconv.Atoi(m[key])
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
		sizeBytes, err = strconv.ParseInt(
			strings.TrimSuffix(strings.ToLower(m[key]), "b"),
			10, 64)
		if err != nil {
			err = fmt.Errorf("invalid format of %v=%v for vg %v: %v", key, m[key], vg.Name, err)
		}
		quantity := resource.NewQuantity(sizeBytes, resource.BinarySI)
		*value = *quantity //
	}
	return vg, err
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
//
//In case reloadCache is false, we skip refreshing lvm metadata cache.
func ListLVMVolumeGroup(reloadCache bool) ([]apis.VolumeGroup, error) {
	if reloadCache {
		if err := ReloadLVMMetadataCache(); err != nil {
			return nil, err
		}
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

//Function to get LVM Logical volume device
//It returns LVM logical volume device(dm-*).
//This is used as a label in metrics(lvm_lv_total_size) which helps us to map lv_name to device.
//
//Example: pvc-f147582c-adbd-4015-8ca9-fe3e0a4c2452(lv_name) -> dm-0(device)
func getLvDeviceName(path string) (string, error) {
	dmPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		klog.Errorf("failed to resolve device mapper from lv path %v: %v", path, err)
		return "", err
	}
	deviceName := strings.Split(dmPath, "/")
	return deviceName[len(deviceName)-1], nil
}

//To parse the output of lvs command and store it in LogicalVolume
//It returns LogicalVolume.
//
//Example: LogicalVolume{
//		Name:     "pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
//		Size:     3221225472,
//		Path:     "/dev/linuxlvmvg/pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
//	}
func parseLogicalVolume(m map[string]string) (LogicalVolume, error) {
	var lv LogicalVolume
	var err error

	lv.Name = m["lv_name"]
	lv.Path = m["lv_path"]
	sizeBytes, err := strconv.ParseInt(strings.TrimSuffix(strings.ToLower(m["lv_size"]), "b"), 10, 64)

	if err != nil {
		err = fmt.Errorf("invalid format of lv_size=%v for lv %v: %v", m["lv_size"], lv.Name, err)
		return LogicalVolume{}, err
	}

	lv.Size = sizeBytes
	return lv, err
}

//decodeLvsJSON([]bytes): Decode json format and pass the unmarshalled json to parseLogicalVolume to store logical volumes in LogicalVolume
//
//Output of lvs command will be in json format:
//
//{
//	"report": [
//		{
//			"lv": [
//					{
//						"lv_name":"pvc-ba7b648e-b08b-47bb-beef-60738a33fbd2",
//						...
//					}
//				]
//		}
//	]
//}
//
//This function is used to decode the output of lvs command.
//It returns []LogicalVolume.
//
//Example: []LogicalVolume{
//	{
//		Name:     "pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
//		Size:     3221225472,
//		Path:     "/dev/linuxlvmvg/pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
//		Device:	  "dm-0"
//	}
//}
func decodeLvsJSON(raw []byte) ([]LogicalVolume, error) {
	output := &struct {
		Report []struct {
			LogicalVolumes []map[string]string `json:"lv"`
		} `json:"report"`
	}{}
	var err error
	if err = json.Unmarshal(raw, output); err != nil {
		return nil, err
	}

	if len(output.Report) != 1 {
		return nil, fmt.Errorf("expected exactly one lvm report")
	}

	items := output.Report[0].LogicalVolumes
	lvs := make([]LogicalVolume, 0, len(items))
	for _, item := range items {
		var lv LogicalVolume
		if lv, err = parseLogicalVolume(item); err != nil {
			return lvs, err
		}
		deviceName, err := getLvDeviceName(lv.Path)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		lv.Device = deviceName
		lvs = append(lvs, lv)
	}
	return lvs, nil
}

func ListLVMLogicalVolume() ([]LogicalVolume, error) {
	args := []string{
		"--options", "lv_all,vg_name",
		"--reportformat", "json",
		"--units", "b",
	}
	cmd := exec.Command(LVList, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("lvm: error while running command %s %v: %v", LVList, args, err)
		return nil, err
	}
	return decodeLvsJSON(output)
}

// lvThinExists verifies if thin pool/volume already exists for given volumegroup
func lvThinExists(vg string, name string) bool {
	cmd := exec.Command("lvs", vg+"/"+name, "--noheadings", "-o", "lv_name")
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("failed to list existing volumes:%v", err)
		return false
	}
	return name == strings.TrimSpace(string(out))
}

// snapshotExists checks if a snapshot volume exists given the name of the volume.
// The name should be <vg-name>/<snapshot-name>
func snapshotExists(snapVolumeName string) (bool, error) {
	snapPath := DevPath + snapVolumeName
	if _, err := os.Stat(snapPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// getVGSize get the size in bytes for given volumegroup name
func getVGSize(vgname string) string {
	cmd := exec.Command("vgs", vgname, "--noheadings", "-o", "vg_free", "--units", "b", "--nosuffix")
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.Errorf("failed to list existing volumegroup:%v , %v", vgname, err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getThinPoolSize gets size for a given volumegroup, compares it with
// the requested volume size and returns the minimum size as a thin pool size
func getThinPoolSize(vgname, volsize string) string {
	outStr := getVGSize(vgname)
	vgFreeSize, err := strconv.ParseInt(strings.TrimSpace(string(outStr)), 10, 64)
	if err != nil {
		klog.Errorf("failed to convert vg_size to int, got size,:%v , %v", outStr, err)
		return ""
	}

	volSize, err := strconv.ParseInt(strings.TrimSpace(string(volsize)), 10, 64)
	if err != nil {
		klog.Errorf("failed to convert volsize to int, got size,:%v , %v", volSize, err)
		return ""
	}

	if vgFreeSize < volSize {
		// reducing 268435456 bytes (256Mi) from the total byte size to round off
		// blocks extent
		return fmt.Sprint(vgFreeSize-MinExtentRoundOffSize) + "b"
	}
	return volsize + "b"
}

// removeVolumeFilesystem will erases the filesystem signature from lvm volume
func removeVolumeFilesystem(lvmVolume *apis.LVMVolume) error {
	devicePath := filepath.Join(DevPath, lvmVolume.Spec.VolGroup, lvmVolume.Name)

	// wipefs erases the filesystem signature from the lvm volume
	// -a    wipe all magic strings
	// -f    force erasure
	// Command: wipefs -af /dev/lvmvg/volume1
	cleanCommand := exec.Command(BlockCleanerCommand, "-af", devicePath)
	output, err := cleanCommand.CombinedOutput()
	if err != nil {
		return errors.Wrapf(
			err,
			"failed to wipe filesystem on device path: %s resp: %s",
			devicePath,
			string(output),
		)
	}
	klog.V(4).Infof("Successfully wiped filesystem on device path: %s", devicePath)
	return nil
}
