/*
Copyright © 2026 The OpenEBS Authors

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

package lvm

import (
	"os"
	"path/filepath"
	"testing"
)

// lvsThinPoolsReport is the output of
// `lvs --reportformat json --units b --no-suffix` for a volume group with an
// active thin pool (data_percent set) and an inactive one (data_percent empty).
const lvsThinPoolsReport = `{
  "report": [
    {
      "lv": [
        {"lv_name":"lvmvg_thinpool", "vg_name":"lvmvg", "lv_attr":"twi-aotz--", "lv_size":"10737418240", "pool_lv":"", "origin":"", "data_percent":"25.00", "metadata_percent":"10.50", "move_pv":"", "mirror_log":"", "copy_percent":"", "convert_lv":""},
        {"lv_name":"pvc-1", "vg_name":"lvmvg", "lv_attr":"Vwi-aotz--", "lv_size":"4294967296", "pool_lv":"lvmvg_thinpool", "origin":"", "data_percent":"62.50", "metadata_percent":"", "move_pv":"", "mirror_log":"", "copy_percent":"", "convert_lv":""},
        {"lv_name":"other_pool", "vg_name":"lvmvg", "lv_attr":"twi---tz--", "lv_size":"2147483648", "pool_lv":"", "origin":"", "data_percent":"", "metadata_percent":"", "move_pv":"", "mirror_log":"", "copy_percent":"", "convert_lv":""}
      ]
    }
  ]
}`

func TestGetThinPools(t *testing.T) {
	dir := t.TempDir()
	report := filepath.Join(dir, "lvs.json")
	if err := os.WriteFile(report, []byte(lvsThinPoolsReport), 0o644); err != nil {
		t.Fatal(err)
	}
	lvs := "#!/bin/sh\ncat '" + report + "'\n"
	if err := os.WriteFile(filepath.Join(dir, LVList), []byte(lvs), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	pools, err := getThinPools("lvmvg")
	if err != nil {
		t.Fatalf("getThinPools() error = %v, want nil", err)
	}
	if len(pools) != 2 {
		t.Fatalf("getThinPools() returned %d pools, want 2: %+v", len(pools), pools)
	}

	want := map[string]struct{ size, free int64 }{
		"lvmvg_thinpool": {size: 10737418240, free: 8053063680},
		"other_pool":     {size: 2147483648, free: 0},
	}
	for _, pool := range pools {
		w, ok := want[pool.Name]
		if !ok {
			t.Fatalf("unexpected thin pool %q", pool.Name)
		}
		if pool.Size.Value() != w.size || pool.Free.Value() != w.free {
			t.Errorf("thin pool %q: size=%d free=%d, want size=%d free=%d",
				pool.Name, pool.Size.Value(), pool.Free.Value(), w.size, w.free)
		}
	}
}
