/*
Copyright © 2020 The OpenEBS Authors
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
	"errors"
	"fmt"
	"github.com/openebs/lib-csi/pkg/device/iolimit"
	"math"
	"os"
	"strconv"

	mnt "github.com/openebs/lib-csi/pkg/mount"
	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
)

// MountInfo contains the volume related info
// for all types of volumes in LVMVolume
type MountInfo struct {
	// FSType of a volume will specify the
	// format type - ext4(default), xfs of PV
	FSType string `json:"fsType"`

	// AccessMode of a volume will hold the
	// access mode of the volume
	AccessModes []string `json:"accessModes"`

	// MountPath of the volume will hold the
	// path on which the volume is mounted
	// on that node
	MountPath string `json:"mountPath"`

	// MountOptions specifies the options with
	// which mount needs to be attempted
	MountOptions []string `json:"mountOptions"`
}

// PodLVInfo contains the pod, LVGroup related info
type PodLVInfo struct {
	// UID is the Uid of the pod
	UID string

	// LVGroup is the LVM vg name in which lv needs to be provisioned
	LVGroup string
}

// FormatAndMountVol formats and mounts the created volume to the desired mount path
func FormatAndMountVol(devicePath string, mountInfo *MountInfo) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

	err := mounter.FormatAndMount(devicePath, mountInfo.MountPath, mountInfo.FSType, mountInfo.MountOptions)
	if err != nil {
		klog.Errorf(
			"lvm: failed to mount volume %s [%s] to %s, error %v",
			devicePath, mountInfo.FSType, mountInfo.MountPath, err,
		)
		return err
	}

	return nil
}

// UmountVolume unmounts the volume and the corresponding mount path is removed
func UmountVolume(vol *apis.LVMVolume, targetPath string,
) error {
	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

	dev, ref, err := mount.GetDeviceNameFromMount(mounter, targetPath)
	if err != nil {
		klog.Errorf(
			"lvm: umount volume: failed to get device from mnt: %s\nError: %v",
			targetPath, err,
		)
		return err
	}

	// device has already been un-mounted, return successful
	if len(dev) == 0 || ref == 0 {
		klog.Warningf(
			"Warning: Unmount skipped because volume %s not mounted: %v",
			vol.Name, targetPath,
		)
		return nil
	}

	if pathExists, pathErr := mount.PathExists(targetPath); pathErr != nil {
		return fmt.Errorf("error checking if path exists: %v", pathErr)
	} else if !pathExists {
		klog.Warningf(
			"Warning: Unmount skipped because path does not exist: %v",
			targetPath,
		)
		return nil
	}

	if err = mounter.Unmount(targetPath); err != nil {
		klog.Errorf(
			"lvm: failed to unmount %s: path %s err: %v",
			vol.Name, targetPath, err,
		)
		return err
	}

	if err := os.Remove(targetPath); err != nil {
		klog.Errorf("lvm: failed to remove mount path vol %s err : %v", vol.Name, err)
	}

	klog.Infof("umount done %s path %v", vol.Name, targetPath)

	return nil
}

func verifyMountRequest(vol *apis.LVMVolume, mountpath string) (bool, error) {
	if len(mountpath) == 0 {
		return false, status.Error(codes.InvalidArgument, "verifyMount: mount path missing in request")
	}

	if len(vol.Spec.OwnerNodeID) > 0 &&
		vol.Spec.OwnerNodeID != NodeID {
		return false, status.Error(codes.Internal, "verifyMount: volume is owned by different node")
	}
	if vol.Finalizers == nil {
		return false, status.Error(codes.Internal, "verifyMount: volume is not ready to be mounted")
	}

	devicePath, err := GetVolumeDevPath(vol)
	if err != nil {
		klog.Errorf("can not get device for volume:%s dev %s err: %v",
			vol.Name, devicePath, err.Error())
		return false, status.Errorf(codes.Internal, "verifyMount: GetVolumePath failed %s", err.Error())
	}

	/*
	 * This check is the famous *Wall Of North*
	 * It will not let the volume to be mounted
	 * at more than two places. The volume should
	 * be unmounted before proceeding to the mount
	 * operation.
	 */
	currentMounts, err := mnt.GetMounts(devicePath)
	if err != nil {
		klog.Errorf("can not get mounts for volume:%s dev %s err: %v",
			vol.Name, devicePath, err.Error())
		return false, status.Errorf(codes.Internal, "verifyMount: Getmounts failed %s", err.Error())
	} else if len(currentMounts) >= 1 {
		// if device is already mounted at the mount point, return successful
		for _, mp := range currentMounts {
			if mp == mountpath {
				return true, nil
			}
		}

		// if it is not a shared volume, then it should not mounted to more than one path
		if vol.Spec.Shared != "yes" {
			klog.Errorf(
				"can not mount, volume:%s already mounted dev %s mounts: %v",
				vol.Name, devicePath, currentMounts,
			)
			return false, status.Errorf(codes.Internal, "verifyMount: device already mounted at %s", currentMounts)
		}
	}
	return false, nil
}

// MountVolume mounts the disk to the specified path
func MountVolume(vol *apis.LVMVolume, mount *MountInfo, podLVInfo *PodLVInfo) error {
	volume := vol.Spec.VolGroup + "/" + vol.Name
	mounted, err := verifyMountRequest(vol, mount.MountPath)
	if err != nil {
		return err
	}

	if mounted {
		klog.Infof("lvm : already mounted %s => %s", volume, mount.MountPath)
		return nil
	}

	devicePath := DevPath + volume

	err = FormatAndMountVol(devicePath, mount)
	if err != nil {
		return status.Error(codes.Internal, "not able to format and mount the volume")
	}

	klog.Infof("lvm: volume %v mounted %v fs %v", volume, mount.MountPath, mount.FSType)

	if ioLimitsEnabled && podLVInfo != nil {
		if err := setIOLimits(vol, podLVInfo, devicePath); err != nil {
			klog.Warningf("lvm: error setting io limits: podUid %s, device %s, err=%v", podLVInfo.UID, devicePath, err)
		} else {
			klog.Infof("lvm: io limits set for podUid %v, device %s", podLVInfo.UID, devicePath)
		}
	}

	return nil
}

// MountFilesystem mounts the disk to the specified path
func MountFilesystem(vol *apis.LVMVolume, mount *MountInfo, podinfo *PodLVInfo) error {
	if err := os.MkdirAll(mount.MountPath, 0755); err != nil {
		return status.Errorf(codes.Internal, "Could not create dir {%q}, err: %v", mount.MountPath, err)
	}

	return MountVolume(vol, mount, podinfo)
}

// MountBlock mounts the block disk to the specified path
func MountBlock(vol *apis.LVMVolume, mountinfo *MountInfo, podLVInfo *PodLVInfo) error {
	target := mountinfo.MountPath
	volume := vol.Spec.VolGroup + "/" + vol.Name
	devicePath := DevPath + volume

	mountopt := []string{"bind"}

	mounter := &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()}

	// Create the mount point as a file since bind mount device node requires it to be a file
	err := mounter.MakeFile(target)
	if err != nil {
		return status.Errorf(codes.Internal, "Could not create target file %q: %v", target, err)
	}

	// do the bind mount of the device at the target path
	if err := mounter.Mount(devicePath, target, "", mountopt); err != nil {
		if removeErr := os.Remove(target); removeErr != nil {
			return status.Errorf(codes.Internal, "Could not remove mount target %q: %v", target, removeErr)
		}
		return status.Errorf(codes.Internal, "mount failed at %v err : %v", target, err)
	}

	klog.Infof("NodePublishVolume mounted block device %s at %s", devicePath, target)

	if ioLimitsEnabled && podLVInfo != nil {
		if err := setIOLimits(vol, podLVInfo, devicePath); err != nil {
			klog.Warningf(": error setting io limits for podUid %s, device %s, err=%v", podLVInfo.UID, devicePath, err)
		} else {
			klog.Infof("lvm: io limits set for podUid %s, device %s", podLVInfo.UID, devicePath)
		}
	}
	return nil
}

func setIOLimits(vol *apis.LVMVolume, podLVInfo *PodLVInfo, devicePath string) error {
	if podLVInfo == nil {
		return errors.New("PodLVInfo is missing. Skipping setting IOLimits")
	}
	capacityBytes, err := strconv.ParseUint(vol.Spec.Capacity, 10, 64)
	if err != nil {
		klog.Warning("error parsing LVMVolume.Spec.Capacity. Skipping setting IOLimits", err)
		return err
	}
	capacityGB := uint64(math.Ceil(float64(capacityBytes) / (1024 * 1024 * 1024)))
	klog.Infof("Capacity of device in GB: %v", capacityGB)
	riops := getRIopsPerGB(podLVInfo.LVGroup) * capacityGB
	wiops := getWIopsPerGB(podLVInfo.LVGroup) * capacityGB
	rbps := getRBpsPerGB(podLVInfo.LVGroup) * capacityGB
	wbps := getWBpsPerGB(podLVInfo.LVGroup) * capacityGB
	klog.Infof("Setting iolimits for podUId %s, device %s: riops=%v, wiops=%v, rbps=%v, wbps=%v",
		podLVInfo.UID, devicePath, riops, wiops, rbps, wbps,
	)
	err = iolimit.SetIOLimits(&iolimit.Request{
		DeviceName:       devicePath,
		PodUid:           podLVInfo.UID,
		ContainerRuntime: getContainerRuntime(),
		IOLimit: &iolimit.IOMax{
			Riops: riops,
			Wiops: wiops,
			Rbps:  rbps,
			Wbps:  wbps,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
