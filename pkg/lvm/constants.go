/*
Copyright 2017 The Kubernetes Authors.

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

// lvm vg, lv & pv fields related constants
const (
	VGName              = "vg_name"
	VGUUID              = "vg_uuid"
	VGPVvount           = "pv_count"
	VGLvCount           = "lv_count"
	VGMaxLv             = "max_lv"
	VGMaxPv             = "max_pv"
	VGSnapCount         = "snap_count"
	VGMissingPvCount    = "vg_missing_pv_count"
	VGMetadataCount     = "vg_mda_count"
	VGMetadataUsedCount = "vg_mda_used_count"
	VGSize              = "vg_size"
	VGFreeSize          = "vg_free"
	VGMetadataSize      = "vg_mda_size"
	VGMetadataFreeSize  = "vg_mda_free"
	VGPermissions       = "vg_permissions"
	VGAllocationPolicy  = "vg_allocation_policy"

	LVName            = "lv_name"
	LVFullName        = "lv_full_name"
	LVUUID            = "lv_uuid"
	LVPath            = "lv_path"
	LVDmPath          = "lv_dm_path"
	LVActive          = "lv_active"
	LVSize            = "lv_size"
	LVMetadataSize    = "lv_metadata_size"
	LVSegtype         = "segtype"
	LVHost            = "lv_host"
	LVPool            = "pool_lv"
	LVPermissions     = "lv_permissions"
	LVWhenFull        = "lv_when_full"
	LVHealthStatus    = "lv_health_status"
	RaidSyncAction    = "raid_sync_action"
	LVDataPercent     = "data_percent"
	LVMetadataPercent = "metadata_percent"
	LVSnapPercent     = "snap_percent"

	PVName             = "pv_name"
	PVUUID             = "pv_uuid"
	PVInUse            = "pv_in_use"
	PVAllocatable      = "pv_allocatable"
	PVMissing          = "pv_missing"
	PVSize             = "pv_size"
	PVFreeSize         = "pv_free"
	PVUsedSize         = "pv_used"
	PVMetadataSize     = "pv_mda_size"
	PVMetadataFreeSize = "pv_mda_free"
	PVDeviceSize       = "dev_size"
)
