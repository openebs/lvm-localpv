/*
Copyright 2020 The OpenEBS Authors.

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
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lvmapi "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/generated/clientset/internalclientset/fake"
	informers "github.com/openebs/lvm-localpv/pkg/generated/informer/externalversions"
	"github.com/openebs/lvm-localpv/pkg/lvm"
)

func TestRoundOff(t *testing.T) {

	tests := map[string]struct {
		input    int64
		expected int64
	}{
		"Minimum allocatable is 1Mi": {input: 1, expected: Mi},
		"roundOff to same Mi size":   {input: Mi, expected: Mi},
		"roundOff to nearest Mi":     {input: Mi + 1, expected: Mi * 2},
		"roundOff to same Gi size":   {input: Gi, expected: Gi},
		"roundOff to nearest Gi":     {input: Gi + 1, expected: Gi * 2},
		"roundOff MB size":           {input: 5 * MB, expected: 5 * Mi},
		"roundOff GB size":           {input: 5 * GB, expected: 5 * Gi},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, getRoundedCapacity(test.input))
		})
	}
}

func Test_getSnapSize(t *testing.T) {
	type args struct {
		params   *SnapshotParams
		capacity int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "snapSize percent",
			args: args{
				params: &SnapshotParams{
					SnapSize:    50,
					AbsSnapSize: false,
				},
				capacity: 4 * Gi,
			},
			want: 2 * Gi,
		},
		{
			name: "snapSize absolute less than capacity",
			args: args{
				params: &SnapshotParams{
					SnapSize:    3 * GB,
					AbsSnapSize: true,
				},
				capacity: 4 * Gi,
			},
			want: 3 * Gi,
		},
		{
			name: "snapSize absolute more than capacity",
			args: args{
				params: &SnapshotParams{
					SnapSize:    5 * Gi,
					AbsSnapSize: true,
				},
				capacity: 4 * Gi,
			},
			want: 4 * Gi,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSnapSize(tt.args.params, tt.args.capacity); got != tt.want {
				t.Errorf("getSnapSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVolumeGroupFreeSpace(t *testing.T) {
	namespace := lvm.LvmNamespace
	lvm.LvmNamespace = "openebs"
	t.Cleanup(func() { lvm.LvmNamespace = namespace })

	lvmNode := &lvmapi.LVMNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node1",
			Namespace: lvm.LvmNamespace,
		},
		VolumeGroups: []lvmapi.VolumeGroup{
			{
				Name: "lvmvg",
				Size: *resource.NewQuantity(50*Gi, resource.BinarySI),
				Free: *resource.NewQuantity(10*Gi, resource.BinarySI),
			},
		},
	}

	lvmNodeInformer := informers.NewSharedInformerFactory(fake.NewSimpleClientset(), 0).
		Local().V1alpha1().LVMNodes().Informer()
	if err := lvmNodeInformer.GetIndexer().Add(lvmNode); err != nil {
		t.Fatalf("failed to add LVMNode to the informer cache: %v", err)
	}
	cs := &controller{lvmNodeInformer: lvmNodeInformer}

	tests := map[string]struct {
		ownerNode     string
		volGroup      string
		thinProvision string
		growth        int64
		code          codes.Code
	}{
		"growth less than the free space": {
			ownerNode: "node1", volGroup: "lvmvg", growth: 2 * Gi, code: codes.OK,
		},
		"growth equal to the free space": {
			ownerNode: "node1", volGroup: "lvmvg", growth: 10 * Gi, code: codes.OK,
		},
		"growth more than the free space": {
			ownerNode: "node1", volGroup: "lvmvg", growth: 10*Gi + Mi, code: codes.OutOfRange,
		},
		"thin volume growth more than the free space": {
			ownerNode: "node1", volGroup: "lvmvg", thinProvision: lvm.YES, growth: 20 * Gi, code: codes.OK,
		},
		"volume group not on the owner node": {
			ownerNode: "node1", volGroup: "othervg", growth: Mi, code: codes.Unavailable,
		},
		"owner node without LVMNode": {
			ownerNode: "node2", volGroup: "lvmvg", growth: Mi, code: codes.Unavailable,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			vol := &lvmapi.LVMVolume{
				ObjectMeta: metav1.ObjectMeta{Name: "pvc-1"},
				Spec: lvmapi.VolumeInfo{
					OwnerNodeID:   test.ownerNode,
					VolGroup:      test.volGroup,
					ThinProvision: test.thinProvision,
				},
			}
			err := cs.validateVolumeGroupFreeSpace(vol, test.growth)
			assert.Equal(t, test.code, status.Code(err), "error: %v", err)
		})
	}
}
