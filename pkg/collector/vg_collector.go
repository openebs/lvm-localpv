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
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"

	"github.com/openebs/lvm-localpv/pkg/lvm"
)

/*
vgCollector collects all the necessary metrics related to volume group
*/
type vgCollector struct {
	vgSizeMetric              *prometheus.Desc
	vgFreeMetric              *prometheus.Desc
	vgLvCountMetric           *prometheus.Desc
	vgPvCountMetric           *prometheus.Desc
	vgMaxLvMetric             *prometheus.Desc
	vgMaxPvMetric             *prometheus.Desc
	vgSnapCountMetric         *prometheus.Desc
	vgMissingPvCountMetric    *prometheus.Desc
	vgMetadataCountMetric     *prometheus.Desc
	vgMetadataUsedCountMetric *prometheus.Desc
	vgMetadataFreeMetric      *prometheus.Desc
	vgMetadataSizeMetric      *prometheus.Desc
	vgPermissionsMetric       *prometheus.Desc
	vgAllocationPolicyMetric  *prometheus.Desc
}

func NewVgCollector() prometheus.Collector {
	return &vgCollector{
		vgFreeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "free_size_bytes"),
			"LVM VG free size in bytes",
			[]string{"name"}, nil,
		),
		vgSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "total_size_bytes"),
			"LVM VG total size in bytes",
			[]string{"name"}, nil,
		),
		vgLvCountMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "lv_count"),
			"Number of LVs in VG",
			[]string{"name"}, nil,
		),
		vgPvCountMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "pv_count"),
			"Number of PVs in VG",
			[]string{"name"}, nil,
		),
		vgMaxLvMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "max_lv_count"),
			"LMaximum number of LVs allowed in VG or 0 if unlimited",
			[]string{"name"}, nil,
		),
		vgMaxPvMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "max_pv_count"),
			"Maximum number of PVs allowed in VG or 0 if unlimited",
			[]string{"name"}, nil,
		),
		vgSnapCountMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "snap_count"),
			"Number of snapshots in VG",
			[]string{"name"}, nil,
		),
		vgMissingPvCountMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "missing_pv_count"),
			"Number of PVs in VG which are missing",
			[]string{"name"}, nil,
		),
		vgMetadataCountMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "mda_count"),
			"Number of metadata areas on this VG",
			[]string{"name"}, nil,
		),
		vgMetadataUsedCountMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "mda_used_count"),
			"Number of metadata areas in use on this VG",
			[]string{"name"}, nil,
		),
		vgMetadataFreeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "mda_free_size_bytes"),
			"Free metadata area space for this VG in bytes",
			[]string{"name"}, nil,
		),
		vgMetadataSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "mda_total_size_bytes"),
			"Size of smallest metadata area for this VG in bytes",
			[]string{"name"}, nil,
		),
		vgPermissionsMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "permission"),
			"VG permissions: [-1: undefined], [0: writeable], [1: read-only]",
			[]string{"name"}, nil,
		),
		vgAllocationPolicyMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "allocation_policy"),
			"VG allocation policy: [-1: undefined], [0: normal], [1: contiguous], [2: cling], [3: anywhere], [4: inherited]",
			[]string{"name"}, nil,
		),
	}
}

func (c *vgCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.vgFreeMetric
	ch <- c.vgSizeMetric
	ch <- c.vgLvCountMetric
	ch <- c.vgPvCountMetric
	ch <- c.vgMaxLvMetric
	ch <- c.vgMaxPvMetric
	ch <- c.vgSnapCountMetric
	ch <- c.vgMissingPvCountMetric
	ch <- c.vgMetadataCountMetric
	ch <- c.vgMetadataUsedCountMetric
	ch <- c.vgMetadataFreeMetric
	ch <- c.vgMetadataSizeMetric
	ch <- c.vgPermissionsMetric
	ch <- c.vgAllocationPolicyMetric
}

func (c *vgCollector) Collect(ch chan<- prometheus.Metric) {
	vgList, err := lvm.ListLVMVolumeGroup(false)
	if err != nil {
		klog.Errorf("error in getting the list of lvm volume groups: %v", err)
	} else {
		for _, vg := range vgList {
			ch <- prometheus.MustNewConstMetric(c.vgFreeMetric, prometheus.GaugeValue, vg.Free.AsApproximateFloat64(), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgSizeMetric, prometheus.GaugeValue, vg.Size.AsApproximateFloat64(), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgLvCountMetric, prometheus.GaugeValue, float64(vg.LVCount), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgPvCountMetric, prometheus.GaugeValue, float64(vg.PVCount), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgMaxLvMetric, prometheus.GaugeValue, float64(vg.MaxLV), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgMaxPvMetric, prometheus.GaugeValue, float64(vg.MaxPV), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgSnapCountMetric, prometheus.GaugeValue, float64(vg.SnapCount), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgMissingPvCountMetric, prometheus.GaugeValue, float64(vg.MissingPVCount), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgMetadataCountMetric, prometheus.GaugeValue, float64(vg.MetadataCount), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgMetadataUsedCountMetric, prometheus.GaugeValue, float64(vg.MetadataUsedCount), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgMetadataFreeMetric, prometheus.GaugeValue, vg.MetadataFree.AsApproximateFloat64(), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgMetadataSizeMetric, prometheus.GaugeValue, vg.MetadataSize.AsApproximateFloat64(), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgPermissionsMetric, prometheus.GaugeValue, float64(vg.Permission), vg.Name)
			ch <- prometheus.MustNewConstMetric(c.vgAllocationPolicyMetric, prometheus.GaugeValue, float64(vg.AllocationPolicy), vg.Name)
		}
	}
}
