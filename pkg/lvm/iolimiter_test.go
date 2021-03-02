// Copyright 2020 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lvm

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/openebs/lvm-localpv/pkg/config"
)

func TestExtractingIOLimits(t *testing.T) {
	//beforeFunc := func(value string) {
	//	if err := os.Setenv(string(OpenEBSPingPeriod), value); err != nil {
	//		t.Logf("Unable to set environment variable")
	//	}
	//}
	//afterFunc := func() {
	//	if err := os.Unsetenv(string(OpenEBSPingPeriod)); err != nil {
	//		t.Logf("Unable to unset environment variable")
	//	}
	//}
	type expectedIoLimitRate struct {
		riops uint64
		wiops uint64
		rbps uint64
		wbps uint64
	}
	testSuite := map[string]struct {
		config *config.Config
		vgNames *[]string
		expected map[string]expectedIoLimitRate
	}{
		"Riops and Wiops": {
			config: &config.Config{
				SetIOLimits:true,
				RIopsLimitPerGB:  &[]string{"lvmvg1:50", "lvmvg2:100"},
				WIopsLimitPerGB:  &[]string{"lvmvg1:70", "lvmvg2:120"},
				RBpsLimitPerGB:   nil,
				WBpsLimitPerGB:   nil,
			}, vgNames: &[]string{"lvmvg1-id1", "lvmvg2", "lvmvg3"},
			expected: map[string]expectedIoLimitRate{
				"lvmvg1-id1": {riops: 50, wiops: 70, rbps: 0, wbps: 0},
				"lvmvg2": {riops: 100, wiops: 120, rbps: 0, wbps: 0},
				"lvmvg3": {riops: 0, wiops: 0, rbps: 0, wbps: 0},
			},
		},
	}
	for _, testData := range testSuite {
		SetIORateLimits(testData.config)
		for _, vgName := range *testData.vgNames {
			assert.Equal(t, testData.expected[vgName].riops, getRIopsPerGB(vgName))
			assert.Equal(t, testData.expected[vgName].wiops, getWIopsPerGB(vgName))
			assert.Equal(t, testData.expected[vgName].rbps, getRBpsPerGB(vgName))
			assert.Equal(t, testData.expected[vgName].wbps, getWBpsPerGB(vgName))
		}
	}
}

