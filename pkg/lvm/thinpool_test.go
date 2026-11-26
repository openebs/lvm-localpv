package lvm

import "testing"

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

func TestParseThinPools(t *testing.T) {
	pools, err := parseThinPools([]byte(lvsThinPoolsReport), "lvmvg")
	if err != nil {
		t.Fatalf("parseThinPools() error = %v, want nil", err)
	}
	if len(pools) != 2 {
		t.Fatalf("parseThinPools() returned %d pools, want 2: %+v", len(pools), pools)
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
