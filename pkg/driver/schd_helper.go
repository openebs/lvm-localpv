/*
Copyright 2020 The OpenEBS Authors

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
	"github.com/pawanpraka1/dynamic-lvm/pkg/builder/volbuilder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pawanpraka1/dynamic-lvm/pkg/lvm"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given volume group
	// this will be the default scheduler when none provided
	VolumeWeighted = "VolumeWeighted"
)

// getVolumeWeightedMap goes through all the volumegroup on all the nodes
// and creats the node mapping of the volume for all the nodes.
// It returns a map which has nodes as key and volumes present
// on the nodes as corresponding value.
func getVolumeWeightedMap(vg string) (map[string]int64, error) {
	nmap := map[string]int64{}

	vollist, err := volbuilder.NewKubeclient().
		WithNamespace(lvm.LvmNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	// create the map of the volume count
	// for the given vg
	for _, vol := range vollist.Items {
		if vol.Spec.VolGroup == vg {
			nmap[vol.Spec.OwnerNodeID]++
		}
	}

	return nmap, nil
}

// getNodeMap returns the node mapping for the given scheduling algorithm
func getNodeMap(schd string, vg string) (map[string]int64, error) {
	switch schd {
	case VolumeWeighted:
		return getVolumeWeightedMap(vg)
	}
	// return VolumeWeighted(default) if not specified
	return getVolumeWeightedMap(vg)
}
