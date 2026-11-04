/*
Copyright 2026 The OpenEBS Authors.

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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"

	"github.com/openebs/lib-csi/pkg/common/env"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
)

// serveLVMNodes starts a fake API server that returns the given LVMNodes
// for every request and points the nodebuilder client at it.
func serveLVMNodes(t *testing.T, nodes []apis.LVMNode) {
	t.Helper()
	list := apis.LVMNodeList{
		TypeMeta: metav1.TypeMeta{APIVersion: "local.openebs.io/v1alpha1", Kind: "LVMNodeList"},
		Items:    nodes,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}))
	t.Cleanup(srv.Close)
	t.Setenv(env.KubeMaster, srv.URL)
	t.Setenv(env.KubeConfig, "")
}

func lvmNode(name string, vgFree int64) apis.LVMNode {
	return apis.LVMNode{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		VolumeGroups: []apis.VolumeGroup{{
			Name: "lvmvg",
			Size: *resource.NewQuantity(100*Gi, resource.BinarySI),
			Free: *resource.NewQuantity(vgFree, resource.BinarySI),
		}},
	}
}

func TestGetNodeMapVgFreeEqualsCapacity(t *testing.T) {
	const capacity = 10 * Gi
	serveLVMNodes(t, []apis.LVMNode{
		lvmNode("node-less", capacity-Mi),
		lvmNode("node-equal", capacity),
		lvmNode("node-more", capacity+Mi),
	})

	for _, scheduler := range []string{VolumeWeighted, CapacityWeighted, SpaceWeighted} {
		t.Run(scheduler, func(t *testing.T) {
			params := &VolumeParams{
				VgPattern:     regexp.MustCompile("^lvmvg$"),
				Scheduler:     scheduler,
				ThinProvision: "no",
			}
			nmap, err := getNodeMap(params, strconv.FormatInt(capacity, 10))
			assert.NoError(t, err)
			assert.NotContains(t, nmap, "node-less")
			assert.Contains(t, nmap, "node-equal")
			assert.Contains(t, nmap, "node-more")
		})
	}
}
