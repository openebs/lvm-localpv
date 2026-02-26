package lvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	mnt "github.com/openebs/lib-csi/pkg/mount"
	"golang.org/x/sys/unix"
	klog "k8s.io/klog/v2"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/driver/config"
)

const qosValueMax = "max"

var (
	kDir        string
	qosValuesMu sync.Mutex
)

// SetQoSValuesConf sets config parametrs
func SetQoSValuesConf(config *config.Config) {
	kDir = config.KubeletDir
}

// ReconcileSetQoSValues update vol.Spec.QoS(VAC) on all pods currently using this volume on
// this node. Logic of the operation is based on use of information about mounting points.
// Get the path to the device, all mount points for this device and set VAC parametrs.
func ReconcileSetQoSValues(vol *apis.LVMVolume) error {
	if vol == nil || vol.Spec.QoS == nil {
		return nil
	}

	devPath := GetVolumeDevPath(vol)

	mounts, err := mnt.GetMounts(devPath)
	if err != nil {
		return fmt.Errorf("get mounts for %s: %w", devPath, err)
	}

	podUIDs := make(map[string]struct{})

	// when pvc volumeMode: filesystem
	for _, mp := range mounts {
		if uid, ok := getPodUIDFromPath(mp); ok {
			podUIDs[uid] = struct{}{}
		}
	}
	// when pvc volumeMode: Block
	if len(podUIDs) == 0 {
		uids, err := getPodUIDFromVolDevice(vol.Name)
		if err != nil {
			return err
		}
		for _, uid := range uids {
			podUIDs[uid] = struct{}{}
		}
	}

	if len(podUIDs) == 0 {
		klog.Infof("Not found pod uuid from mount on node")
		return nil
	}

	for uid := range podUIDs {
		if err := SetQoSValues(vol, uid); err != nil {
			klog.Warningf("io.max apply failed vol=%s podUID=%s: %v", vol.Name, uid, err)
		}
	}
	return nil
}

// getPodUIDFromVolDevice helper function tries to discover podUIDs for CSI block volumes.
// Entries are typically named by UUID.
func getPodUIDFromVolDevice(volumeID string) ([]string, error) {
	publishDir := filepath.Join(
		kDir,
		"plugins/kubernetes.io/csi/volumeDevices/publish",
		volumeID,
	)

	ents, err := os.ReadDir(publishDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("readdir %s: %w", publishDir, err)
	}

	out := make([]string, 0, len(ents))
	seen := make(map[string]struct{}, len(ents))

	for _, e := range ents {
		name := e.Name()
		if len(name) != 36 || strings.Count(name, "-") != 4 {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}

	return out, nil
}

// getPodUIDFromPath helper function to calculate the <podUID> from the kubelet mount points:
// /var/lib/kubelet/pods/<podUID>/volumes/.../mount
func getPodUIDFromPath(mountPath string) (string, bool) {
	prefix := filepath.Join(kDir, "pods")

	rest, found := strings.CutPrefix(mountPath, prefix)
	if !found {
		return "", false
	}

	// cut first slash
	rest = strings.TrimPrefix(rest, string(filepath.Separator))

	// get uid
	uid, _, ok := strings.Cut(rest, string(filepath.Separator))
	if !ok || uid == "" {
		return "", false
	}
	return uid, true
}

// SetQoSValues applies vol.Spec.QoS for the given podUID. After getting the path to the
// device we can find out its id. With the path to the device, we will find out its id in
// order to form the parameters for writing to io.max and apply the limits.
func SetQoSValues(vol *apis.LVMVolume, podUID string) error {
	if vol == nil || vol.Spec.QoS == nil {
		return nil
	}

	devPath := GetVolumeDevPath(vol)

	maj, min, err := getDevMajorMinor(devPath)
	if err != nil {
		return fmt.Errorf("stat device %s: %w", devPath, err)
	}

	line, err := setQoSValuesLine(maj, min, vol.Spec.QoS)
	if err != nil {
		return fmt.Errorf("build io.max line: %w", err)
	}

	cgDirs, err := getPodCgroupPath(podUID)
	if err != nil {
		return err
	}
	if len(cgDirs) == 0 {
		return nil
	}

	qosValuesMu.Lock()
	defer qosValuesMu.Unlock()

	for _, dir := range cgDirs {
		p := filepath.Join(dir, "io.max")
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if err := QoSValuesWriteFile(p, line); err != nil {
			klog.Warningf("failed writing %s: %v", p, err)
			continue
		}
		klog.Infof("Applied io.max vol=%s podUID=%s file=%s line=%q", vol.Name, podUID, p, line)
	}

	return nil
}

// getDevMajorMinor helper function to calculate device id number, for example:
// major minor => 253 0
func getDevMajorMinor(devPath string) (uint32, uint32, error) {
	var st unix.Stat_t
	if err := unix.Stat(devPath, &st); err != nil {
		return 0, 0, err
	}
	return unix.Major(uint64(st.Rdev)), unix.Minor(uint64(st.Rdev)), nil
}

// setQoSValuesLine helper function which transforms all keys to desired state.
// cgroup v2 io.max format => "<maj>:<min> rbps=<v> wbps=<v> riops=<v> wiops=<v>"
func setQoSValuesLine(maj, min uint32, qos *apis.VolumeQoS) (string, error) {
	rbps := qosValueMax
	wbps := qosValueMax
	riops := qosValueMax
	wiops := qosValueMax

	if qos != nil {
		var err error

		rbps, err = qosValuesConvert(qos.ReadBPS)
		if err != nil {
			return "", fmt.Errorf("readBPS: %w", err)
		}
		wbps, err = qosValuesConvert(qos.WriteBPS)
		if err != nil {
			return "", fmt.Errorf("writeBPS: %w", err)
		}
		riops, err = qosValuesConvert(qos.ReadIOPS)
		if err != nil {
			return "", fmt.Errorf("readIOPS: %w", err)
		}
		wiops, err = qosValuesConvert(qos.WriteIOPS)
		if err != nil {
			return "", fmt.Errorf("writeIOPS: %w", err)
		}
	}

	return fmt.Sprintf("%d:%d rbps=%s wbps=%s riops=%s wiops=%s", maj, min, rbps, wbps, riops, wiops), nil
}

// getPodCgroupPath helper function which find pod cgroup directory and supports both
// systemd and cgroupfs paths.
func getPodCgroupPath(podUID string) ([]string, error) {
	kDriver := "/sys/fs/cgroup/kubepods.slice"
	cDriver := "/sys/fs/cgroup/kubepods"

	uidDash := podUID
	uidUnderscore := strings.ReplaceAll(podUID, "-", "_")

	candidates := []string{
		// systemd
		filepath.Join(kDriver, "kubepods-burstable.slice", fmt.Sprintf("kubepods-burstable-pod%s.slice", uidUnderscore)),
		filepath.Join(kDriver, "kubepods-besteffort.slice", fmt.Sprintf("kubepods-besteffort-pod%s.slice", uidUnderscore)),
		filepath.Join(kDriver, "kubepods-guaranteed.slice", fmt.Sprintf("kubepods-guaranteed-pod%s.slice", uidUnderscore)),
		filepath.Join(kDriver, fmt.Sprintf("kubepods-pod%s.slice", uidUnderscore)),

		// cgroupfs
		filepath.Join(cDriver, "burstable", fmt.Sprintf("pod%s", uidDash)),
		filepath.Join(cDriver, "besteffort", fmt.Sprintf("pod%s", uidDash)),
		filepath.Join(cDriver, fmt.Sprintf("pod%s", uidDash)),
	}

	out := make([]string, 0, 2)
	for _, dir := range candidates {
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, "io.max")); err == nil {
				out = append(out, dir)
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("pod cgroup dir not found for podUID=%s", podUID)
	}
	return out, nil
}

// QoSValuesWriteFile helper function which write a VAC value to the file.
func QoSValuesWriteFile(path, line string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}

// QoSValuesConvert helper function which converts a VAC value to io.max value
// with checking for invalid values ​​in case of parsing error.
func qosValuesConvert(v string) (string, error) {
	if v == qosValueMax {
		return qosValueMax, nil
	}
	if v == "" {
		return "", fmt.Errorf("qos value is empty")
	}
	u, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return "", fmt.Errorf("qos value %q is not a valid uint64", v)
	}
	if u == 0 {
		return "", fmt.Errorf("qos value must be > 0 or 'max', got %q", v)
	}

	return strconv.FormatUint(u, 10), nil
}
