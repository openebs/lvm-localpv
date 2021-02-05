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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=lvmvolume

// LVMVolume represents a LVM based volume
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced,shortName=lvmvol
// +kubebuilder:printcolumn:name="VolGroup",type=string,JSONPath=`.spec.volGroup`,description="volume group where the volume is created"
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.ownerNodeID`,description="Node where the volume is created"
// +kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.spec.capacity`,description="Size of the volume"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.state`,description="Status of the volume"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Age of the volume"
type LVMVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolumeInfo `json:"spec"`
	Status VolStatus  `json:"status,omitempty"`
}

// LVMVolumeList is a list of LVMVolume resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=lvmvolumes
type LVMVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LVMVolume `json:"items"`
}

// VolumeInfo defines LVM info
type VolumeInfo struct {

	// OwnerNodeID is the Node ID where the volume group is present which is where
	// the volume has been provisioned.
	// OwnerNodeID can not be edited after the volume has been provisioned.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	OwnerNodeID string `json:"ownerNodeID"`

	// VolGroup specifies the name of the volume group where the volume has been created.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	VolGroup string `json:"volGroup"`

	// Capacity of the volume
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Capacity string `json:"capacity"`

	// Shared specifies whether the volume can be shared among multiple pods.
	// If it is not set to "yes", then the LVM LocalPV Driver will not allow
	// the volumes to be mounted by more than one pods.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=yes;no
	Shared string `json:"shared,omitempty"`
}

// VolStatus string that specifies the current state of the volume provisioning request.
type VolStatus struct {
	// State specifies the current state of the volume provisioning request.
	// The state "Pending" means that the volume creation request has not
	// processed yet. The state "Ready" means that the volume has been created
	// and it is ready for the use.
	// +kubebuilder:validation:Enum=Pending;Ready
	State string `json:"state,omitempty"`
}
