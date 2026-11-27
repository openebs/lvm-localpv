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
	"testing"

	"github.com/openebs/lvm-localpv/pkg/driver/config"
	"github.com/stretchr/testify/assert"
)

func TestExtractingIoLimits(t *testing.T) {
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
		rbps  uint64
		wbps  uint64
	}
	testSuite := map[string]struct {
		config   *config.Config
		vgNames  *[]string
		expected map[string]expectedIoLimitRate
	}{
		"Riops and Wiops": {
			config: &config.Config{
				SetIOLimits:     true,
				RIopsLimitPerGB: &[]string{"lvmvg1:50", "lvmvg2:100"},
				WIopsLimitPerGB: &[]string{"lvmvg1:70", "lvmvg2:120"},
				RBpsLimitPerGB:  nil,
				WBpsLimitPerGB:  nil,
			}, vgNames: &[]string{"lvmvg1-id1", "lvmvg2", "lvmvg3", "prfx-lvmvg2-sffx"},
			expected: map[string]expectedIoLimitRate{
				"lvmvg1-id1":       {riops: 50, wiops: 70, rbps: 0, wbps: 0},
				"lvmvg2":           {riops: 100, wiops: 120, rbps: 0, wbps: 0},
				"lvmvg3":           {riops: 0, wiops: 0, rbps: 0, wbps: 0},
				"prfx-lvmvg2-sffx": {riops: 0, wiops: 0, rbps: 0, wbps: 0},
			},
		},
	}
	for _, testData := range testSuite {
		SetIORateLimits(testData.config)
		for _, vgName := range *testData.vgNames {
			assert.Equal(t, testData.expected[vgName].riops, GetRIopsPerGB(vgName))
			assert.Equal(t, testData.expected[vgName].wiops, GetWIopsPerGB(vgName))
			assert.Equal(t, testData.expected[vgName].rbps, GetRBpsPerGB(vgName))
			assert.Equal(t, testData.expected[vgName].wbps, GetWBpsPerGB(vgName))
		}
	}
}

// A malformed rate entry must not take the node plugin down. run() in
// cmd/main.go calls SetIORateLimits right after flag parsing, and setValues
// already treats an unusable rate list as non-fatal: it logs a warning and
// falls back to an empty map. pflag's StringSlice splits on commas, so
// --riops-per-gb="lvmvg1:50,lvmvg2" yields an entry with no colon in it and
// a trailing comma yields an empty one.
func TestSetIORateLimitsSurvivesMalformedRate(t *testing.T) {
	// SetIORateLimits applies a config once per process, so put the package
	// back the way a freshly started plugin would find it.
	reset := func() {
		rwlock.Lock()
		set = false
		rwlock.Unlock()
	}
	reset()
	t.Cleanup(reset)

	cfg := &config.Config{
		SetIOLimits:     true,
		RIopsLimitPerGB: &[]string{"lvmvg1:50", "lvmvg2"},
		WIopsLimitPerGB: &[]string{"lvmvg1:70", ""},
		RBpsLimitPerGB:  nil,
		WBpsLimitPerGB:  nil,
	}

	assert.NotPanics(t, func() { SetIORateLimits(cfg) })

	// A bad entry discards that whole list, which is what setValues already
	// does when strconv.ParseUint fails, so no rate is applied.
	assert.Equal(t, uint64(0), GetRIopsPerGB("lvmvg1"))
	assert.Equal(t, uint64(0), GetWIopsPerGB("lvmvg1"))
}
