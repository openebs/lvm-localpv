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
}

// LVMNodeList is a collection of LVMNode resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=lvmnodes
type LVMNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LVMNode `json:"items"`
}
