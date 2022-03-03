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
	"math"
	"regexp"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openebs/lvm-localpv/pkg/builder/nodebuilder"
	"github.com/openebs/lvm-localpv/pkg/builder/volbuilder"
	"github.com/openebs/lvm-localpv/pkg/lvm"
)

// scheduling algorithm constants
const (
	// pick the node where less volumes are provisioned for the given volume group
	VolumeWeighted = "VolumeWeighted"

	// pick the node where total provisioned volumes have occupied less capacity from the given volume group
	CapacityWeighted = "CapacityWeighted"

	// pick the node which is less loaded space wise
	// this will be the default scheduler when none provided
	SpaceWeightedMap = "SpaceWeighted"
)

// getVolumeWeightedMap goes through all the volumegroup on all the nodes
// and creates the node mapping of the volume for all the nodes.
// It returns a map which has nodes as key and volumes present
// on the nodes as corresponding value.
func getVolumeWeightedMap(re *regexp.Regexp) (map[string]int64, error) {
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
		if re.MatchString(vol.Spec.VolGroup) {
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
func getCapacityWeightedMap(re *regexp.Regexp) (map[string]int64, error) {
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
		if re.MatchString(vol.Spec.VolGroup) {
			volSize, err := strconv.ParseInt(vol.Spec.Capacity, 10, 64)
			if err == nil {
				nmap[vol.Spec.OwnerNodeID] += volSize
			}
		}
	}

	return nmap, nil
}

// getSpaceWeightedMap returns how weighted a node is space wise.
// The node which has max free space available is less loaded and
// can accumulate more volumes.
func getSpaceWeightedMap(re *regexp.Regexp) (map[string]int64, error) {
	nmap := map[string]int64{}

	nodeList, err := nodebuilder.NewKubeclient().
		WithNamespace(lvm.LvmNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	for _, node := range nodeList.Items {
		var maxFree int64 = 0
		for _, vg := range node.VolumeGroups {
			if re.MatchString(vg.Name) {
				freeCapacity := vg.Free.Value()
				if maxFree < freeCapacity {
					maxFree = freeCapacity
				}
			}
		}
		if maxFree > 0 {
			// converting to SpaceWeighted by subtracting it with MaxInt64
			// as the node which has max free space available is less loaded.
			nmap[node.Name] = math.MaxInt64 - maxFree
		}
	}

	return nmap, nil
}

// getNodeMap returns the node mapping for the given scheduling algorithm
func getNodeMap(schd string, vgPattern *regexp.Regexp) (map[string]int64, error) {
	switch schd {
	case VolumeWeighted:
		return getVolumeWeightedMap(vgPattern)
	case CapacityWeighted:
		return getCapacityWeightedMap(vgPattern)
	case SpaceWeightedMap:
		return getSpaceWeightedMap(vgPattern)
	}
	// return getSpaceWeightedMap(default) if not specified
	return getSpaceWeightedMap(vgPattern)
}
