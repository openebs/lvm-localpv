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
pvCollector collects all the necessary metrics related to physical volume
*/
type pvCollector struct {
	pvSizeMetric         *prometheus.Desc
	pvFreeMetric         *prometheus.Desc
	pvUsedMetric         *prometheus.Desc
	pvDeviceSizeMetric   *prometheus.Desc
	pvMetadataSizeMetric *prometheus.Desc
	pvMetadataFreeMetric *prometheus.Desc
}

func NewPvCollector() prometheus.Collector {
	return &pvCollector{
		pvSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "pv", "total_size_bytes"),
			"LVM PV total size in bytes",
			[]string{"name", "allocatable", "vg", "missing", "in_use"}, nil,
		),
		pvFreeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "pv", "free_size_bytes"),
			"LVM PV free size in bytes",
			[]string{"name", "allocatable", "vg", "missing", "in_use"}, nil,
		),
		pvUsedMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "pv", "used_size_bytes"),
			"LVM PV used size in bytes",
			[]string{"name", "allocatable", "vg", "missing", "in_use"}, nil,
		),
		pvDeviceSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "pv", "device_size_bytes"),
			"LVM PV underlying device size in bytes",
			[]string{"name", "allocatable", "vg", "missing", "in_use"}, nil,
		),
		pvMetadataSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "pv", "mda_total_size_bytes"),
			"LVM PV device smallest metadata area size in bytes",
			[]string{"name", "allocatable", "vg", "missing", "in_use"}, nil,
		),
		pvMetadataFreeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "pv", "mda_free_size_bytes"),
			"LVM PV device free metadata area space in bytes",
			[]string{"name", "allocatable", "vg", "missing", "in_use"}, nil,
		),
	}
}

func (c *pvCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.pvSizeMetric
	ch <- c.pvFreeMetric
	ch <- c.pvUsedMetric
	ch <- c.pvDeviceSizeMetric
	ch <- c.pvMetadataSizeMetric
	ch <- c.pvMetadataFreeMetric
}

func (c *pvCollector) Collect(ch chan<- prometheus.Metric) {
	pvList, err := lvm.ListLVMPhysicalVolume()
	if err != nil {
		klog.Errorf("error in getting the list of lvm physical volumes: %v", err)
	} else {
		for _, pv := range pvList {
			ch <- prometheus.MustNewConstMetric(c.pvSizeMetric, prometheus.GaugeValue, pv.Size.AsApproximateFloat64(), pv.Name, pv.Allocatable, pv.VGName, pv.Missing, pv.InUse)
			ch <- prometheus.MustNewConstMetric(c.pvFreeMetric, prometheus.GaugeValue, pv.Free.AsApproximateFloat64(), pv.Name, pv.Allocatable, pv.VGName, pv.Missing, pv.InUse)
			ch <- prometheus.MustNewConstMetric(c.pvUsedMetric, prometheus.GaugeValue, pv.Used.AsApproximateFloat64(), pv.Name, pv.Allocatable, pv.VGName, pv.Missing, pv.InUse)
			ch <- prometheus.MustNewConstMetric(c.pvDeviceSizeMetric, prometheus.GaugeValue, pv.DeviceSize.AsApproximateFloat64(), pv.Name, pv.Allocatable, pv.VGName, pv.Missing, pv.InUse)
			ch <- prometheus.MustNewConstMetric(c.pvMetadataSizeMetric, prometheus.GaugeValue, pv.MetadataSize.AsApproximateFloat64(), pv.Name, pv.Allocatable, pv.VGName, pv.Missing, pv.InUse)
			ch <- prometheus.MustNewConstMetric(c.pvMetadataFreeMetric, prometheus.GaugeValue, pv.MetadataFree.AsApproximateFloat64(), pv.Name, pv.Allocatable, pv.VGName, pv.Missing, pv.InUse)
		}
	}
}
