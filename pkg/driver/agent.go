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

package driver

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/openebs/lvm-localpv/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/openebs/lib-csi/pkg/btrfs"
	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
	"github.com/openebs/lib-csi/pkg/mount"
	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/builder/volbuilder"
	"github.com/openebs/lvm-localpv/pkg/lvm"
	"github.com/openebs/lvm-localpv/pkg/mgmt/lvmnode"
	"github.com/openebs/lvm-localpv/pkg/mgmt/snapshot"
	"github.com/openebs/lvm-localpv/pkg/mgmt/volume"
	"golang.org/x/net/context"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// node is the server implementation
// for CSI NodeServer
type node struct {
	driver *CSIDriver
}

// NewNode returns a new instance
// of CSI NodeServer
func NewNode(d *CSIDriver) csi.NodeServer {
	var ControllerMutex = sync.RWMutex{}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// start the lvm node resource watcher
	go func() {
		err := lvmnode.Start(&ControllerMutex, stopCh)
		if err != nil {
			klog.Fatalf("Failed to start LVM node controller: %s", err.Error())
		}
	}()

	// start the lvmvolume watcher
	go func() {
		err := volume.Start(&ControllerMutex, stopCh)
		if err != nil {
			klog.Fatalf("Failed to start LVM volume management controller: %s", err.Error())
		}
	}()

	// start the lvm snapshot watcher
	go func() {
		err := snapshot.Start(&ControllerMutex, stopCh)
		if err != nil {
			klog.Fatalf("Failed to start LVM volume snapshot management controller: %s", err.Error())
		}
	}()

	if d.config.ListenAddress != "" {
		exposeMetrics(d.config.ListenAddress, d.config.MetricsPath, d.config.DisableExporterMetrics)
	}

	return &node{
		driver: d,
	}
}

// Function to register collectors to collect LVM related metrics and exporter metrics.
//
// If disableExporterMetrics is set to false, exporter will include metrics about itself i.e (process_*, go_*).
func registerCollectors(disableExporterMetrics bool) (*prometheus.Registry, error) {
	registry := prometheus.NewRegistry()

	if !disableExporterMetrics {
		processCollector := collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})
		err := registry.Register(processCollector)
		if err != nil {
			klog.Errorf("failed to register process collector for exporter metrics collection: %s", err.Error())
			return nil, err
		}
		goProcessCollector := collectors.NewGoCollector()
		err = registry.Register(goProcessCollector)
		if err != nil {
			klog.Errorf("failed to register go process collector for exporter metrics collection: %s", err.Error())
			return nil, err
		}
	}

	lvmVgCollector := collector.NewVgCollector()
	err := registry.Register(lvmVgCollector)
	if err != nil {
		klog.Errorf("failed to register LVM VG collector for LVM metrics collection: %s", err.Error())
		return nil, err
	}

	lvmLvCollector := collector.NewLvCollector()
	err = registry.Register(lvmLvCollector)
	if err != nil {
		klog.Errorf("failed to register LVM LV collector for LVM metrics collection: %s", err.Error())
		return nil, err
	}

	lvmPvCollector := collector.NewPvCollector()
	err = registry.Register(lvmPvCollector)
	if err != nil {
		klog.Errorf("failed to register LVM PV collector for LVM metrics collection: %s", err.Error())
		return nil, err
	}

	return registry, nil
}

type promLog struct{}

// Implementation of Println(...) method of Logger interface of prometheus client_go.
func (p *promLog) Println(v ...interface{}) {
	klog.Error(v...)
}

func promLogger() *promLog {
	return &promLog{}
}

// Function to start HTTP server to expose LVM metrics.
//
// Parameters:
//
// listenAddr: TCP network address where the prometheus metrics endpoint will listen.
//
// metricsPath: The HTTP path where prometheus metrics will be exposed.
//
// disableExporterMetrics: Exclude metrics about the exporter itself (process_*, go_*).
func exposeMetrics(listenAddr string, metricsPath string, disableExporterMetrics bool) {

	// Registry with all the collectors registered
	registry, err := registerCollectors(disableExporterMetrics)
	if err != nil {
		klog.Fatalf("Failed to register collectors for LVM metrics collection: %s", err.Error())
	}

	http.Handle(metricsPath, promhttp.InstrumentMetricHandler(registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog: promLogger(),
	})))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>LVM Exporter</title></head>
			<body>
			<h1>LVM Exporter</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	go func() {
		if err := http.ListenAndServe(listenAddr, nil); err != nil {
			klog.Fatalf("Failed to start HTTP server at specified address (%q) and metrics path (%q) to expose LVM metrics: %s", listenAddr, metricsPath, err.Error())
		}
	}()
}

// GetVolAndMountInfo get volume and mount info from node csi volume request
func GetVolAndMountInfo(
	req *csi.NodePublishVolumeRequest,
) (*apis.LVMVolume, *lvm.MountInfo, error) {
	var mountinfo lvm.MountInfo

	mountinfo.FSType = req.GetVolumeCapability().GetMount().GetFsType()
	mountinfo.MountPath = req.GetTargetPath()
	mountinfo.MountOptions = append(mountinfo.MountOptions, req.GetVolumeCapability().GetMount().GetMountFlags()...)

	if req.GetReadonly() {
		mountinfo.MountOptions = append(mountinfo.MountOptions, "ro")
	}

	volName := strings.ToLower(req.GetVolumeId())

	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
		WithNamespace(lvm.LvmNamespace).
		Get(volName, getOptions)

	if err != nil {
		return nil, nil, err
	}

	return vol, &mountinfo, nil
}

func getPodLVInfo(req *csi.NodePublishVolumeRequest) (*lvm.PodLVInfo, error) {
	var podLVInfo lvm.PodLVInfo
	var ok bool
	if podLVInfo.UID, ok = req.VolumeContext["csi.storage.k8s.io/pod.uid"]; !ok {
		return nil, errors.New("csi.storage.k8s.io/pod.uid key missing in VolumeContext")
	}
	if podLVInfo.LVGroup, ok = req.VolumeContext["openebs.io/volgroup"]; !ok {
		return nil, errors.New("openebs.io/volgroup key missing in VolumeContext")
	}
	return &podLVInfo, nil
}

// NodePublishVolume publishes (mounts) the volume
// at the corresponding node at a given path
//
// This implements csi.NodeServer
func (ns *node) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {

	var (
		err error
	)

	if err = ns.validateNodePublishReq(req); err != nil {
		return nil, err
	}

	vol, mountInfo, err := GetVolAndMountInfo(req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	podLVinfo, err := getPodLVInfo(req)
	if err != nil {
		klog.Warningf("PodLVInfo could not be obtained for volume_id: %s, err = %v", req.VolumeId, err)
	}
	switch req.GetVolumeCapability().GetAccessType().(type) {
	case *csi.VolumeCapability_Block:
		// attempt block mount operation on the requested path
		err = lvm.MountBlock(vol, mountInfo, podLVinfo)
	case *csi.VolumeCapability_Mount:
		// attempt filesystem mount operation on the requested path
		err = lvm.MountFilesystem(vol, mountInfo, podLVinfo)
	}

	if err != nil {
		return nil, err
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unpublishes (unmounts) the volume
// from the corresponding node from the given path
//
// This implements csi.NodeServer
func (ns *node) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {

	var (
		err error
		vol *apis.LVMVolume
	)

	if err = ns.validateNodeUnpublishReq(req); err != nil {
		return nil, err
	}

	targetPath := req.GetTargetPath()
	volumeID := req.GetVolumeId()

	if vol, err = lvm.GetLVMVolume(volumeID); err != nil {
		return nil, status.Errorf(codes.Internal,
			"not able to get the LVMVolume %s err : %s",
			volumeID, err.Error())
	}

	err = lvm.UmountVolume(vol, targetPath)

	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"unable to umount the volume %s err : %s",
			volumeID, err.Error())
	}
	klog.Infof("hostpath: volume %s path: %s has been unmounted.",
		volumeID, targetPath)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetInfo returns node details
//
// This implements csi.NodeServer
func (ns *node) NodeGetInfo(
	ctx context.Context,
	req *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {

	node, err := k8sapi.GetNode(ns.driver.config.NodeID)
	if err != nil {
		klog.Errorf("failed to get the node %s", ns.driver.config.NodeID)
		return nil, err
	}
	/*
	 * The driver will support all the keys and values defined in the node's label.
	 * if nodes are labeled with the below keys and values
	 * map[beta.kubernetes.io/arch:amd64 beta.kubernetes.io/os:linux kubernetes.io/arch:amd64 kubernetes.io/hostname:pawan-node-1 kubernetes.io/os:linux node-role.kubernetes.io/worker:true openebs.io/zone:zone1 openebs.io/zpool:ssd]
	 * The driver will support below key and values
	 * {
	 *	beta.kubernetes.io/arch:amd64
	 *	beta.kubernetes.io/os:linux
	 *	kubernetes.io/arch:amd64
	 *	kubernetes.io/hostname:pawan-node-1
	 *	kubernetes.io/os:linux
	 *	node-role.kubernetes.io/worker:true
	 *	openebs.io/zone:zone1
	 *	openebs.io/zpool:ssd
	 * }
	 */

	// add driver's topology key
	topology := map[string]string{
		lvm.LVMTopologyKey: ns.driver.config.NodeID,
	}

	// support topologykeys from env ALLOWED_TOPOLOGIES
	allowedTopologies := os.Getenv("ALLOWED_TOPOLOGIES")
	allowedKeys := strings.Split(allowedTopologies, ",")
	for _, key := range allowedKeys {
		if key != "" {
			v, ok := node.Labels[key]
			if ok {
				topology[key] = v
			}
		}
	}

	return &csi.NodeGetInfoResponse{
		NodeId: ns.driver.config.NodeID,
		AccessibleTopology: &csi.Topology{
			Segments: topology,
		},
	}, nil
}

// NodeGetCapabilities returns capabilities supported
// by this node service
//
// This implements csi.NodeServer
func (ns *node) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
		},
	}, nil
}

// TODO
// This needs to be implemented
//
// NodeStageVolume mounts the volume on the staging
// path
//
// This implements csi.NodeServer
func (ns *node) NodeStageVolume(
	ctx context.Context,
	req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// NodeUnstageVolume unmounts the volume from
// the staging path
//
// This implements csi.NodeServer
func (ns *node) NodeUnstageVolume(
	ctx context.Context,
	req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

// TODO
// Verify if this needs to be implemented
//
// # NodeExpandVolume resizes the filesystem if required
//
// If ControllerExpandVolumeResponse returns true in
// node_expansion_required then FileSystemResizePending
// condition will be added to PVC and NodeExpandVolume
// operation will be queued on kubelet
//
// This implements csi.NodeServer
func (ns *node) NodeExpandVolume(
	ctx context.Context,
	req *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if req.GetVolumePath() == "" || volumeID == "" {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"path not provided for NodeExpandVolume Request %s",
			volumeID,
		)
	}

	vol, err := lvm.GetLVMVolume(volumeID)

	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"failed to handle NodeExpandVolume Request for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}

	isBlockMode := req.GetVolumeCapability().GetBlock() != nil
	fsType := req.GetVolumeCapability().GetMount().GetFsType()

	resizeFS := true
	if isBlockMode || fsType == "btrfs" {
		// In case of volume block mode (or) btrfs filesystem mode
		// lvm doesn't expand the fs natively
		resizeFS = false
	}

	err = lvm.ResizeLVMVolume(vol, resizeFS)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"failed to handle NodeExpandVolume Request for %s, {%s}",
			req.VolumeId,
			err.Error(),
		)
	}

	// Expand btrfs filesystem
	if fsType == "btrfs" {
		err = btrfs.ResizeBTRFS(req.GetVolumePath())
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"failed to handle NodeExpandVolume Request for %s, {%s}",
				req.VolumeId,
				err.Error(),
			)
		}
	}

	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
	}, nil
}

// NodeGetVolumeStats returns statistics for the
// given volume
func (ns *node) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {

	volID := req.GetVolumeId()
	path := req.GetVolumePath()

	if len(volID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume id is not provided")
	}
	if len(path) == 0 {
		return nil, status.Error(codes.InvalidArgument, "path is not provided")
	}

	if !mount.IsMountPath(path) {
		return nil, status.Error(codes.NotFound, "path is not a mount path")
	}

	var sfs unix.Statfs_t
	if err := unix.Statfs(path, &sfs); err != nil {
		return nil, status.Errorf(codes.Internal, "statfs on %s failed: %v", path, err)
	}

	var usage []*csi.VolumeUsage
	usage = append(usage, &csi.VolumeUsage{
		Unit:      csi.VolumeUsage_BYTES,
		Total:     int64(sfs.Blocks) * int64(sfs.Bsize),
		Used:      int64(sfs.Blocks-sfs.Bfree) * int64(sfs.Bsize),
		Available: int64(sfs.Bavail) * int64(sfs.Bsize),
	})
	usage = append(usage, &csi.VolumeUsage{
		Unit:      csi.VolumeUsage_INODES,
		Total:     int64(sfs.Files),
		Used:      int64(sfs.Files - sfs.Ffree),
		Available: int64(sfs.Ffree),
	})

	return &csi.NodeGetVolumeStatsResponse{Usage: usage}, nil
}

func (ns *node) validateNodePublishReq(
	req *csi.NodePublishVolumeRequest,
) error {
	if req.GetVolumeCapability() == nil {
		return status.Error(codes.InvalidArgument,
			"Volume capability missing in request")
	}

	if len(req.GetVolumeId()) == 0 {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}
	return nil
}

func (ns *node) validateNodeUnpublishReq(
	req *csi.NodeUnpublishVolumeRequest,
) error {
	if req.GetVolumeId() == "" {
		return status.Error(codes.InvalidArgument,
			"Volume ID missing in request")
	}

	if req.GetTargetPath() == "" {
		return status.Error(codes.InvalidArgument,
			"Target path missing in request")
	}
	return nil
}
