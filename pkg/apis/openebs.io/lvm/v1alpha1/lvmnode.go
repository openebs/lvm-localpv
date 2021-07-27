/*
 Copyright © 2021 The OpenEBS Authors

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=lvmnode

// LVMNode records information about all lvm volume groups available
// in a node. In general, the openebs node-agent creates the LVMNode
// object & periodically synchronizing the volume groups available in the node.
// LVMNode has an owner reference pointing to the corresponding node object.
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=lvmnode
type LVMNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	VolumeGroups []VolumeGroup `json:"volumeGroups"`
}

// VolumeGroup specifies attributes of a given vg exists on node.
type VolumeGroup struct {
	// Name of the lvm volume group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// UUID denotes a unique identity of a lvm volume group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	UUID string `json:"uuid"`

	// Size specifies the total size of volume group.
	// +kubebuilder:validation:Required
	Size resource.Quantity `json:"size"`
	// Free specifies the available capacity of volume group.
	// +kubebuilder:validation:Required
	Free resource.Quantity `json:"free"`

	// LVCount denotes total number of logical volumes in
	// volume group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	LVCount int32 `json:"lvCount"`
	// PVCount denotes total number of physical volumes
	// constituting the volume group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	PVCount int32 `json:"pvCount"`

	// MaxLV denotes maximum number of logical volumes allowed
	// in volume group or 0 if unlimited.
	MaxLV int32 `json:"maxLv"`

	// MaxPV denotes maximum number of physical volumes allowed
	// in volume group or 0 if unlimited.
	MaxPV int32 `json:"maxPv"`

	// SnapCount denotes number of snapshots in volume group.
	SnapCount int32 `json:"snapCount"`

	// MissingPVCount denotes number of physical volumes in
	// volume group which are missing.
	MissingPVCount int32 `json:"missingPvCount"`

	// MetadataCount denotes number of metadata areas on the
	// volume group.
	MetadataCount int32 `json:"metadataCount"`

	// MetadataUsedCount denotes number of used metadata areas in
	// volume group
	MetadataUsedCount int32 `json:"metadataUsedCount"`

	// MetadataFree specifies the available metadata area space
	// for the volume group
	MetadataFree resource.Quantity `json:"metadataFree"`

	// MetadataSize specifies size of smallest metadata area
	// for the volume group
	MetadataSize resource.Quantity `json:"metadataSize"`

	// Permission indicates the volume group permission
	// which can be writable or read-only.
	// Permission has the following mapping between
	// int and string for its value:
	// [-1: "", 0: "writeable", 1: "read-only"]
	Permission int `json:"permissions"`

	// AllocationPolicy indicates the volume group allocation
	// policy.
	// AllocationPolicy has the following mapping between
	// int and string for its value:
	// [-1: "", 0: "normal", 1: "contiguous", 2: "cling", 3: "anywhere", 4: "inherited"]
	AllocationPolicy int `json:"allocationPolicy"`
}

// LVMNodeList is a collection of LVMNode resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=lvmnodes
type LVMNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LVMNode `json:"items"`
}
