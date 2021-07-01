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

package lvm

import (
	"reflect"
	"testing"
)

var (
	fakeLogicalVolume = LogicalVolume{
		Name:                "pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
		FullName:            "linuxlvmvg/pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
		UUID:                "UJp2Dh-Knfo-E0fO-KjPB-RSHO-X7JO-AI2FZW",
		Size:                3221225472,
		Path:                "/dev/linuxlvmvg/pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
		SegType:             "thin",
		Permission:          1,
		BehaviourWhenFull:   -1,
		HealthStatus:        0,
		RaidSyncAction:      -1,
		ActiveStatus:        "active",
		UsedSizePercent:     0,
		MetadataSize:        0,
		MetadataUsedPercent: 0,
		SnapshotUsedPercent: 0,
		Host:                "node1",
		PoolName:            "thin_pool",
		DMPath:              "/dev/mapper/linuxlvmvg-pvc--213ca1e6--e271--4ec8--875c--c7def3a4908d",
		VGName:              "linuxlvmvg",
	}
)

func Test_parseLogicalVolume(t *testing.T) {
	type args struct {
		m map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    LogicalVolume
		wantErr bool
	}{
		{
			name: "Test case for successful parsing",
			args: args{
				map[string]string{"lv_uuid": "UJp2Dh-Knfo-E0fO-KjPB-RSHO-X7JO-AI2FZW",
					"lv_name":             "pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
					"lv_full_name":        "linuxlvmvg/pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
					"segtype":             "thin",
					"lv_permissions":      "writeable",
					"lv_when_full":        "",
					"lv_health_status":    "",
					"lv_raid_sync_action": "",
					"lv_active":           "active",
					"lv_host":             "node1",
					"pool_lv":             "thin_pool",
					"data_percent":        "0.00",
					"lv_metadata_size":    "",
					"metadata_percent":    "",
					"snap_percent":        "",
					"lv_path":             "/dev/linuxlvmvg/pvc-213ca1e6-e271-4ec8-875c-c7def3a4908d",
					"lv_dm_path":          "/dev/mapper/linuxlvmvg-pvc--213ca1e6--e271--4ec8--875c--c7def3a4908d",
					"lv_size":             "3221225472",
					"vg_name":             "linuxlvmvg"},
			},
			want:    fakeLogicalVolume,
			wantErr: false,
		},
		{
			name: "Test case for failed parsing",
			args: args{
				map[string]string{"lv_uuid": "fake-uuid",
					"lv_name":      "fake-name",
					"lv_full_name": "fake-full_name",
					"lv_path":      "fakse_path",
					"lv_size":      "invalid-format", "vg_name": "fake-vg"},
			},
			want:    fakeLogicalVolume,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLogicalVolume(tt.args.m)
			if tt.wantErr && !(err != nil) {
				t.Errorf("parseLogicalVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLogicalVolume() got = %v, want %v", got, tt.want)
			}
		})
	}
}
