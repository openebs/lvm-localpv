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

package driver

import (
	"github.com/openebs/lib-csi/pkg/common/helpers"
)

// VolumeParams holds collection of supported settings that can
// be configured in storage class.
type VolumeParams struct {
	// VolumeGroup specifies vg name to use for
	// provisioning logical volumes.
	VolumeGroup string

	Scheduler     string
	Shared        string
	ThinProvision string
	// extra optional metadata passed by external provisioner
	// if enabled. See --extra-create-metadata flag for more details.
	// https://github.com/kubernetes-csi/external-provisioner#recommended-optional-arguments
	PVCName      string
	PVCNamespace string
	PVName       string
}

// NewVolumeParams parses the input params and instantiates new VolumeParams.
func NewVolumeParams(m map[string]string) (*VolumeParams, error) {
	params := &VolumeParams{ // set up defaults, if any.
		Scheduler:     CapacityWeighted,
		Shared:        "no",
		ThinProvision: "no",
	}
	// parameter keys may be mistyped from the CRD specification when declaring
	// the storageclass, which kubectl validation will not catch. Because
	// parameter keys (not values!) are all lowercase, keys may safely be forced
	// to the lower case.
	m = helpers.GetCaseInsensitiveMap(&m)
	params.VolumeGroup = m["volgroup"]

	// parse string params
	stringParams := map[string]*string{
		"scheduler":     &params.Scheduler,
		"shared":        &params.Shared,
		"thinprovision": &params.ThinProvision,
	}
	for key, param := range stringParams {
		value, ok := m[key]
		if !ok {
			continue
		}
		*param = value
	}

	params.PVCName = m["csi.storage.k8s.io/pvc/name"]
	params.PVCNamespace = m["csi.storage.k8s.io/pvc/namespace"]
	params.PVName = m["csi.storage.k8s.io/pv/name"]

	return params, nil
}
