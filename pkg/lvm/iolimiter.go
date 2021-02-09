/*
Copyright © 2019 The OpenEBS Authors

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
	"github.com/openebs/lvm-localpv/pkg/config"
	"k8s.io/klog"

	"strconv"
	"strings"
	"sync"
)

var (
	set              = false
	ioLimitsEnabled  = false
	containerRuntime string
	riopsPerGB       map[string]uint64
	wiopsPerGB       map[string]uint64
	rbpsPerGB        map[string]uint64
	wbpsPerGB        map[string]uint64
	rwlock           sync.RWMutex
)

func isSet() bool {
	rwlock.RLock()
	defer rwlock.RUnlock()
	if set {
		return true
	}
	return false
}

func extractRateValues(rateVals *[]string) (map[string]uint64, error) {
	rate := map[string]uint64{}
	for _, kv := range *rateVals {
		parts := strings.Split(kv, ":")
		key := parts[0]
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}
		rate[key] = value
	}
	return rate, nil
}

func setValues(config *config.Config) {
	var err error
	riopsVals := config.RIopsLimitPerGB
	wiopsVals := config.WIopsLimitPerGB
	rbpsVals := config.RBpsLimitPerGB
	wbpsVals := config.WBpsLimitPerGB

	riopsPerGB, err = extractRateValues(riopsVals)
	if err != nil {
		klog.Warning("Read IOPS limit rates could not be extracted from config", err)
		riopsPerGB = map[string]uint64{}
	}

	wiopsPerGB, err = extractRateValues(wiopsVals)
	if err != nil {
		klog.Warning("Write IOPS limit rates could not be extracted from config", err)
		wiopsPerGB = map[string]uint64{}
	}

	rbpsPerGB, err = extractRateValues(rbpsVals)
	if err != nil {
		klog.Warning("Read BPS limit rates could not be extracted from config", err)
		rbpsPerGB = map[string]uint64{}
	}

	wbpsPerGB, err = extractRateValues(wbpsVals)
	if err != nil {
		klog.Warning("Write BPS limit rates could not be extracted from config", err)
		wbpsPerGB = map[string]uint64{}
	}
}

// SetIORateLimits sets io limit rates for the volume group (prefixes) provided in config
func SetIORateLimits(config *config.Config) {
	if isSet() {
		return
	}
	rwlock.Lock()
	defer rwlock.Unlock()

	ioLimitsEnabled = true
	containerRuntime = config.ContainerRuntime
	setValues(config)
	set = true
}

func getRatePerGB(vgName string, rateMap map[string]uint64) uint64 {
	rwlock.RLock()
	defer rwlock.RUnlock()
	if ptr, ok := rateMap[vgName]; ok {
		return ptr
	}
	for k, v := range rateMap {
		if strings.Contains(vgName, k) {
			return v
		}
	}
	return uint64(0)
}

func getRIopsPerGB(vgName string) uint64 {
	return getRatePerGB(vgName, riopsPerGB)
}

func getWIopsPerGB(vgName string) uint64 {
	return getRatePerGB(vgName, wiopsPerGB)
}

func getRBpsPerGB(vgName string) uint64 {
	return getRatePerGB(vgName, rbpsPerGB)
}

func getWBpsPerGB(vgName string) uint64 {
	return getRatePerGB(vgName, wbpsPerGB)
}

func getContainerRuntime() string {
	rwlock.RLock()
	defer rwlock.RUnlock()
	return containerRuntime
}
