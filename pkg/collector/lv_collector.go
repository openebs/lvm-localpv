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
lvCollector collects all the necessary metrics related to logical volume
*/
type lvCollector struct {
	lvSizeMetric                *prometheus.Desc
	lvTotalSizeMetric           *prometheus.Desc
	lvUsedSizePercentMetric     *prometheus.Desc
	lvPermissionMetric          *prometheus.Desc
	lvBehaviourWhenFullMetric   *prometheus.Desc
	lvHealthStatusMetric        *prometheus.Desc
	lvRaidSyncActionMetric      *prometheus.Desc
	lvMetadataSizeMetric        *prometheus.Desc
	lvMetadataUsedPercentMetric *prometheus.Desc
	lvSnapshotUsedPercentMetric *prometheus.Desc
	lvRiopsLimitMetric          *prometheus.Desc
	lvWiopsLimitMetric          *prometheus.Desc
	lvRbpsLimitMetric           *prometheus.Desc
	lvWbpsLimitMetric           *prometheus.Desc
}

func NewLvCollector() prometheus.Collector {
	return &lvCollector{
		// Metric name is openebs_size_of_volume which stores the size of lv
		lvSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("openebs", "size_of", "volume"),
			"LVM LV total size in bytes",
			[]string{"volumename", "device"}, nil,
		),
		lvTotalSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "total_size_bytes"),
			"LVM LV total size in bytes with the required labels",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvUsedSizePercentMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "used_percent"),
			"LVM LV used size in percentage",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvPermissionMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "permission"),
			"LV permissions: [-1: undefined], [0: unknown], [1: writeable], [2: read-only], [3: read-only-override]",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvBehaviourWhenFullMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "when_full"),
			"For thin pools, behavior when full: [-1: undefined], [0: error], [1: queue]",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvHealthStatusMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "health_status"),
			"LV health status: [-1: undefined], [0: \"\"], [1: partial], [2: refresh needed], [3: mismatches exist]",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvRaidSyncActionMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "raid_sync_action"),
			"For LV RAID, the current synchronization action being performed: [-1: undefined], [0: idle], [1: frozen], [2: resync], [3: recover], [4: check], [5: repair]",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvMetadataSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "mda_total_size_bytes"),
			"LVM LV metadata size in bytes",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvMetadataUsedPercentMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "mda_used_percent"),
			"LVM LV metadata used size in percentage",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvSnapshotUsedPercentMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "snap_percent"),
			"LVM LV snap used size in percentage",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvRiopsLimitMetric: prometheus.NewDesc(prometheus.BuildFQName("openebs", "lv", "riops_limit"),
			"LVM LV riops cgroup limit, 0 means without limit",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvWiopsLimitMetric: prometheus.NewDesc(prometheus.BuildFQName("openebs", "lv", "wiops_limit"),
			"LVM LV wiops cgroup limit, 0 means without limit",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvRbpsLimitMetric: prometheus.NewDesc(prometheus.BuildFQName("openebs", "lv", "rbps_limit"),
			"LVM LV rbps cgroup limit, 0 means without limit",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
		lvWbpsLimitMetric: prometheus.NewDesc(prometheus.BuildFQName("openebs", "lv", "wbps_limit"),
			"LVM LV wbps cgroup limit, 0 means without limit",
			[]string{"name", "path", "dm_path", "vg", "device", "host", "segtype", "pool", "active_status"}, nil,
		),
	}
}

func (c *lvCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lvSizeMetric
	ch <- c.lvTotalSizeMetric
	ch <- c.lvUsedSizePercentMetric
	ch <- c.lvPermissionMetric
	ch <- c.lvBehaviourWhenFullMetric
	ch <- c.lvHealthStatusMetric
	ch <- c.lvRaidSyncActionMetric
	ch <- c.lvMetadataSizeMetric
	ch <- c.lvMetadataUsedPercentMetric
	ch <- c.lvSnapshotUsedPercentMetric
}

func (c *lvCollector) Collect(ch chan<- prometheus.Metric) {
	lvList, err := lvm.ListLVMLogicalVolume()
	if err != nil {
		klog.Errorf("error in getting the list of lvm logical volumes: %v", err)
	} else {
		var lvIDS []string
		for _, lv := range lvList {
			if contains(lvIDS, lv.UUID) {
				klog.V(2).Infof("duplicate entry for LV: %s", lv.UUID)
				continue
			}
			lvIDS = append(lvIDS, lv.UUID)
			ch <- prometheus.MustNewConstMetric(c.lvSizeMetric, prometheus.GaugeValue, float64(lv.Size), lv.Name, lv.Device)
			ch <- prometheus.MustNewConstMetric(c.lvTotalSizeMetric, prometheus.GaugeValue, float64(lv.Size), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvUsedSizePercentMetric, prometheus.GaugeValue, lv.UsedSizePercent, lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvPermissionMetric, prometheus.GaugeValue, float64(lv.Permission), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvBehaviourWhenFullMetric, prometheus.GaugeValue, float64(lv.BehaviourWhenFull), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvHealthStatusMetric, prometheus.GaugeValue, float64(lv.HealthStatus), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvRaidSyncActionMetric, prometheus.GaugeValue, float64(lv.RaidSyncAction), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvMetadataSizeMetric, prometheus.GaugeValue, float64(lv.MetadataSize), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvMetadataUsedPercentMetric, prometheus.GaugeValue, lv.MetadataUsedPercent, lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvSnapshotUsedPercentMetric, prometheus.GaugeValue, lv.SnapshotUsedPercent, lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvRiopsLimitMetric, prometheus.GaugeValue, float64(lvm.GetRIopsPerGB(lv.VGName))*float64(lv.Size>>30), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvWiopsLimitMetric, prometheus.GaugeValue, float64(lvm.GetWIopsPerGB(lv.VGName))*float64(lv.Size>>30), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvRbpsLimitMetric, prometheus.GaugeValue, float64(lvm.GetRBpsPerGB(lv.VGName))*float64(lv.Size>>30), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
			ch <- prometheus.MustNewConstMetric(c.lvWbpsLimitMetric, prometheus.GaugeValue, float64(lvm.GetWBpsPerGB(lv.VGName))*float64(lv.Size>>30), lv.Name, lv.Path, lv.DMPath, lv.VGName, lv.Device, lv.Host, lv.SegType, lv.PoolName, lv.ActiveStatus)
		}
	}
}
