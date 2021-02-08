package lvm

import (
	"github.com/openebs/lvm-localpv/pkg/config"
	"k8s.io/klog"

	"strconv"
	"strings"
	"sync"
)

var(
	set = false
	ioLimitsEnabled = false
	iopsRate map[string]*uint64
	bpsRate map[string]*uint64
	rwlock sync.RWMutex
)

func isSet() bool {
	rwlock.RLock()
	defer rwlock.RUnlock()
	if set {
		return true
	}
	return false
}

func extractRateValues(rateVals *[]string) (map[string]*uint64, error) {
	rate := map[string]*uint64{}
	for _, kv := range *rateVals {
		parts := strings.Split(kv, "=")
		key := parts[0]
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}
		rate[key] = &value
	}
	return rate, nil
}

func setValues(config *config.Config) {
	var err error
	iopsVals := config.VGIopsLimitPerKB
	bpsVals := config.VGBpsLimitPerKB

	iopsRate, err = extractRateValues(iopsVals)
	if err != nil {
		klog.Warningf("IOPS limit rates could not be extracted from config", err)
		iopsRate = map[string]*uint64{}
	}

	bpsRate, err = extractRateValues(bpsVals)
	if err != nil {
		klog.Warningf("BPS limit rates could not be extracted from config", err)
		bpsRate = map[string]*uint64{}
	}
}

// SetIORateLimits sets io limit rates for the volume group (prefixes) provided in config
func SetIORateLimits(config *config.Config) {
	if isSet() {
		return
	}
	rwlock.Lock()
	defer rwlock.Unlock()

	if !config.SetIOLimits {
		set = true
		return
	}

	ioLimitsEnabled = true
	setValues(config)
	set = true
}

func getRatePerKB(vgName string, rateMap map[string]*uint64) *uint64 {
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
	return nil
}

// getIopsPerKB returns the iops per KB limit for a volume group name
func getIopsPerKB(vgName string) *uint64 {
	return getRatePerKB(vgName, iopsRate)
}

// getBpsPerKB returns the bps per KB limit for a volume group name
func getBpsPerKB(vgName string) *uint64 {
	return getRatePerKB(vgName, bpsRate)
}
