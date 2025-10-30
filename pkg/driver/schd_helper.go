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
	"k8s.io/utils/strings/slices"

	"github.com/openebs/lvm-localpv/pkg/builder/nodebuilder"
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
	SpaceWeighted = "SpaceWeighted"
)

// Goes through all the volumegroup on all the nodes
// and returns a map which has nodes as key and number of volumes present
// on the nodes as corresponding value. In case of thick volume, We discard node
// in case it has no Volume group that can accommodate it.
func getVolumeWeightedMap(params *VolumeParams, capacity string) (map[string]int64, error) {
	nmap := map[string]int64{}

	nodeList, err := nodebuilder.NewKubeclient().
		WithNamespace(lvm.LvmNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	suitableNodes, err := getSuitableNodes(params.VgPattern, capacity)

	if err != nil {
		return nmap, err
	}

	for _, node := range nodeList.Items {
		if !(params.ThinProvision == "yes") && !slices.Contains(suitableNodes, node.Name) {
			continue
		}
		for _, vg := range node.VolumeGroups {
			if params.VgPattern.MatchString(vg.Name) {
				nmap[node.Name] += int64(vg.LVCount)
			}
		}
	}

	return nmap, nil
}

// Goes through all the volume groups on all the nodes
// and returns a map which has nodes as key and capacity provisioned
// on the nodes as corresponding value. The scheduler will use this map
// and picks the node which is less weighted. In case of thick volume, We discard node
// in case it has no Volume group that can accommodate it.
func getCapacityWeightedMap(params *VolumeParams, capacity string) (map[string]int64, error) {
	nmap := map[string]int64{}

	nodeList, err := nodebuilder.NewKubeclient().
		WithNamespace(lvm.LvmNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nmap, err
	}

	suitableNodes, err := getSuitableNodes(params.VgPattern, capacity)

	if err != nil {
		return nmap, err
	}

	for _, node := range nodeList.Items {
		if !(params.ThinProvision == "yes") && !slices.Contains(suitableNodes, node.Name) {
			continue
		}
		for _, vg := range node.VolumeGroups {
			if params.VgPattern.MatchString(vg.Name) {
				nmap[node.Name] += vg.Size.Value() - vg.Free.Value()
			}
		}
	}

	return nmap, nil
}

// Goes through all volume groups in the node to determine vg with higest
// free space. In case of thick volume, We discard node in case it has
// no Volume group that can accommodate it.
func getSpaceWeightedMap(params *VolumeParams, capacity string) (map[string]int64, error) {
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
			if params.VgPattern.MatchString(vg.Name) {
				freeCapacity := vg.Free.Value()
				if maxFree < freeCapacity {
					maxFree = freeCapacity
				}
			}
		}
		if maxFree > 0 {
			pvcSize, err := strconv.ParseInt(capacity, 10, 64)
			if err != nil {
				return nmap, err
			}
			if !(params.ThinProvision == "yes") && pvcSize > maxFree {
				continue
			}
			// To implement SpaceWeighted scheduling, we invert the free space metric by subtracting
			// the node's free space from math.MaxInt64. This way, nodes with more free space will have
			// lower weights, making them preferred by the scheduler as less loaded nodes.
			nmap[node.Name] = math.MaxInt64 - maxFree
		}
	}

	return nmap, nil
}

// Returns nodes which can accommodate a requested capacity.
func getSuitableNodes(reg *regexp.Regexp, capacity string) ([]string, error) {
	var nodes []string
	nodeList, err := nodebuilder.NewKubeclient().
		WithNamespace(lvm.LvmNamespace).
		List(metav1.ListOptions{})

	if err != nil {
		return nodes, err
	}
	for _, node := range nodeList.Items {
		var maxFree int64 = 0
		for _, vg := range node.VolumeGroups {
			if reg.MatchString(vg.Name) && vg.Free.Value() > maxFree {
				maxFree = vg.Free.Value()
			}
		}
		pvcSize, err := strconv.ParseInt(capacity, 10, 64)
		if err != nil {
			return nodes, err
		}
		if maxFree > pvcSize {
			nodes = append(nodes, node.Name)
		}
	}
	return nodes, nil
}

// getNodeMap returns the node mapping for the given scheduling algorithm
func getNodeMap(params *VolumeParams, capacity string) (map[string]int64, error) {
	switch params.Scheduler {
	case VolumeWeighted:
		return getVolumeWeightedMap(params, capacity)
	case CapacityWeighted:
		return getCapacityWeightedMap(params, capacity)
	case SpaceWeighted:
		return getSpaceWeightedMap(params, capacity)
	}
	// return getSpaceWeightedMap(default) if not specified
	return getSpaceWeightedMap(params, capacity)
}
