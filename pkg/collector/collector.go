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

package collector

import (
	"github.com/openebs/lvm-localpv/pkg/lvm"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

/*
lvmCollector collects vg total size, vg free size and lv size metrics
*/
type lvmCollector struct {
	vgFreeMetric *prometheus.Desc
	vgSizeMetric *prometheus.Desc
	lvSizeMetric *prometheus.Desc
}

func NewLvmCollector() prometheus.Collector {
	return &lvmCollector{
		vgFreeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "free_size_bytes"),
			"LVM VG free size in bytes",
			[]string{"name"}, nil,
		),
		vgSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "total_size_bytes"),
			"LVM VG total size in bytes",
			[]string{"name"}, nil,
		),
		// Metric name is openebs_size_of_volume which stores the size of lv
		lvSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("openebs", "size_of", "volume"),
			"LVM LV total size in bytes",
			[]string{"volumename", "dm_path", "device"}, nil,
		),
	}
}

func (c *lvmCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.vgFreeMetric
	ch <- c.vgSizeMetric
	ch <- c.lvSizeMetric
}

func (c *lvmCollector) Collect(ch chan<- prometheus.Metric) {
	vgList, err := lvm.ListLVMVolumeGroup(false)
	if err != nil {
		klog.Errorf("error in getting the list of lvm volume groups: %v", err)
	} else {
		for _, vg := range vgList {
			ch <- prometheus.MustNewConstMetric(c.vgFreeMetric, prometheus.GaugeValue, vg.Free.AsApproximateFloat64(), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgSizeMetric, prometheus.GaugeValue, vg.Size.AsApproximateFloat64(), vg.Name)
		}
	}

	lvList, err := lvm.ListLVMLogicalVolume()
	if err != nil {
		klog.Errorf("error in getting the list of lvm logical volumes: %v", err)
	} else {
		for _, lv := range lvList {
			ch <- prometheus.MustNewConstMetric(c.lvSizeMetric, prometheus.GaugeValue, float64(lv.Size), lv.Name, lv.DMPath, lv.Device)

		}
	}
}
