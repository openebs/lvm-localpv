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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"

	"github.com/openebs/lvm-localpv/pkg/builder/volbuilder"
	"github.com/openebs/lvm-localpv/pkg/lvm"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given volume group
	VolumeWeighted = "VolumeWeighted"

	// pick the node where total provisioned volumes have occupied less capacity from the given volume group
	// this will be the default scheduler when none provided
	CapacityWeighted = "CapacityWeighted"
)

// getVolumeWeightedMap goes through all the volumegroup on all the nodes
// and creates the node mapping of the volume for all the nodes.
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

// getCapacityWeightedMap goes through all the volume groups on all the nodes
// and creates the node mapping of the capacity for all the nodes.
// It returns a map which has nodes as key and capacity provisioned
// on the nodes as corresponding value. The scheduler will use this map
// and picks the node which is less weighted.
func getCapacityWeightedMap(vg string) (map[string]int64, error) {
	nmap := map[string]int64{}

	volList, err := volbuilder.NewKubeclient().
		WithNamespace(lvm.LvmNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	// create the map of the volume capacity
	// for the given volume group
	for _, vol := range volList.Items {
		if vol.Spec.VolGroup == vg {
			volSize, err := strconv.ParseInt(vol.Spec.Capacity, 10, 64)
			if err == nil {
				nmap[vol.Spec.OwnerNodeID] += volSize
			}
		}
	}

	return nmap, nil
}

// getNodeMap returns the node mapping for the given scheduling algorithm
func getNodeMap(schd string, vg string) (map[string]int64, error) {
	switch schd {
	case VolumeWeighted:
		return getVolumeWeightedMap(vg)
	case CapacityWeighted:
		return getCapacityWeightedMap(vg)
	}
	// return CapacityWeighted(default) if not specified
	return getCapacityWeightedMap(vg)
}
