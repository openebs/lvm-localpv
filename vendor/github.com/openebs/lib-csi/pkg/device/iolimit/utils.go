/*
Copyright 2020 The OpenEBS Authors

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

package iolimit

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	"github.com/openebs/lib-csi/pkg/common/helpers"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
)

const (
	baseCgroupPath = "/sys/fs/cgroup"
)

// SetIOLimits sets iops, bps limits for a pod with uid podUid for accessing a device named deviceName
// provided that the underlying cgroup used for pod namespacing is cgroup2 (cgroup v2)
func SetIOLimits(request *Request) error {
	if !helpers.DirExists(baseCgroupPath) {
		return errors.New(baseCgroupPath + " does not exist")
	}
	if err := checkCgroupV2(); err != nil {
		return err
	}
	validRequest, err := validate(request)
	if err != nil {
		return err
	}
	err = setIOLimits(validRequest)
	return err
}

func validate(request *Request) (*ValidRequest, error) {
	if !helpers.IsValidUUID(request.PodUid) {
		return nil, errors.New("Expected PodUid in UUID format, Got " + request.PodUid)
	}
	podCGPath, err := getPodCGroupPath(request.PodUid, request.ContainerRuntime)
	if err != nil {
		return nil, err
	}
	ioMaxFile := podCGPath + "/io.max"
	if !helpers.FileExists(ioMaxFile) {
		return nil, errors.New("io.max file is not present in pod CGroup")
	}
	deviceNumber, err := getDeviceNumber(request.DeviceName)
	if err != nil {
		return nil, errors.New("Device Major:Minor numbers could not be obtained")
	}
	return &ValidRequest{
		FilePath:     ioMaxFile,
		DeviceNumber: deviceNumber,
		IOMax:      request.IOLimit,
	}, nil
}

func getPodCGroupPath(podUid string, cruntime string) (string, error) {
	switch cruntime {
	case "containerd":
		path, err := getContainerdCGPath(podUid)
		if err != nil {
			return "", err
		}
		return path, nil
	default:
		return "", errors.New(cruntime + " runtime support is not present")
	}

}

func checkCgroupV2() error {
	if !helpers.FileExists(baseCgroupPath + "/cgroup.controllers") {
		return errors.New("CGroupV2 not enabled")
	}
	return nil
}

func getContainerdPodCGSuffix(podUid string) string {
	return "pod" + strings.ReplaceAll(podUid, "-", "_")
}

func getContainerdCGPath(podUid string) (string, error) {
	kubepodsCGPath := baseCgroupPath + "/kubepods.slice"
	podSuffix := getContainerdPodCGSuffix(podUid)
	podCGPath := kubepodsCGPath + "/kubepods-besteffort.slice/kubepods-besteffort-" + podSuffix + ".slice"
	if helpers.DirExists(podCGPath) {
		return podCGPath, nil
	}
	podCGPath = kubepodsCGPath + "/kubepods-burstable.slice/kubepods-burstable-" + podSuffix + ".slice"
	if helpers.DirExists(podCGPath) {
		return podCGPath, nil
	}
	return "", errors.New("CGroup Path not found for pod with Uid: " + podUid)
}

func getDeviceNumber(deviceName string) (*DeviceNumber, error) {
	stat := syscall.Stat_t{}
	if err := syscall.Stat(deviceName, &stat); err != nil {
		return nil, err
	}
	return &DeviceNumber{
		Major: uint64(stat.Rdev/256),
		Minor: uint64(stat.Rdev%256),
	}, nil
}

func getIOLimitsStr(deviceNumber *DeviceNumber, ioMax *IOMax) string {
	line := strconv.FormatUint(deviceNumber.Major, 10) + ":" + strconv.FormatUint(deviceNumber.Minor, 10)
	if ioMax.Riops != 0 {
		line += " riops=" + strconv.FormatUint(ioMax.Riops, 10)
	}
	if ioMax.Wiops != 0 {
		line += " wiops=" + strconv.FormatUint(ioMax.Wiops, 10)
	}
	if ioMax.Rbps != 0 {
		line += " rbps=" + strconv.FormatUint(ioMax.Rbps, 10)
	}
	if ioMax.Wbps != 0 {
		line += " wbps=" + strconv.FormatUint(ioMax.Wbps, 10)
	}
	return line
}

func setIOLimits(request *ValidRequest) error {
	line := getIOLimitsStr(request.DeviceNumber, request.IOMax)
	err := ioutil.WriteFile(request.FilePath, []byte(line), 0700)
	return err
}
