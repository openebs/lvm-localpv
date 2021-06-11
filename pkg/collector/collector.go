package collector

import (
	"github.com/openebs/lvm-localpv/pkg/lvm"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

/*
To store vg total size, vg free size and lv size metrics
*/
type lvmCollector struct {
	vgFreeMetric *prometheus.Desc
	vgSizeMetric *prometheus.Desc
	lvSizeMetric *prometheus.Desc
}

func NewLvmCollector() *lvmCollector {
	return &lvmCollector{
		vgFreeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "free_size"),
			"Shows LVM VG free size",
			[]string{"vg_name"}, nil,
		),
		vgSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "vg", "total_size"),
			"Shows LVM VG total size",
			[]string{"vg_name"}, nil,
		),
		lvSizeMetric: prometheus.NewDesc(prometheus.BuildFQName("lvm", "lv", "total_size"),
			"Shows LVM LV total size",
			[]string{"lv_name", "lv_full_name", "lv_uuid", "lv_path", "lv_dm_path", "vg_name", "device"}, nil,
		),
	}
}

func (collector *lvmCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.vgFreeMetric
	ch <- collector.vgSizeMetric
	ch <- collector.lvSizeMetric
}

func (collector *lvmCollector) Collect(ch chan<- prometheus.Metric) {
	vgList, err := lvm.ListLVMVolumeGroup()
	if err != nil {
		klog.Errorf("Error in getting the list of LVM volume groups:%v", err)
	}

	for _, vg := range vgList {
		ch <- prometheus.MustNewConstMetric(collector.vgFreeMetric, prometheus.GaugeValue, vg.Free.AsApproximateFloat64(), vg.Name)
		ch <- prometheus.MustNewConstMetric(collector.vgSizeMetric, prometheus.GaugeValue, vg.Size.AsApproximateFloat64(), vg.Name)
	}

	lvList, err := lvm.ListLVMLogicalVolume()
	if err != nil {
		klog.Errorf("Error in getting the list of LVM logical volume:%v", err)
	}

	for _, lv := range lvList {
		ch <- prometheus.MustNewConstMetric(collector.lvSizeMetric, prometheus.GaugeValue, float64(lv.Size), lv.Name, lv.FullName, lv.UUID, lv.Path, lv.DMPath, lv.VGName, lv.Device)

	}
}
