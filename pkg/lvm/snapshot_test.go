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

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fakeLVMTools puts fake lvs and lvremove commands first in PATH. lvs prints
// lvsStdout and lvsStderr and exits with lvsExit; lvremove records its call in
// the returned file.
func fakeLVMTools(t *testing.T, lvsStdout, lvsStderr string, lvsExit string) string {
	t.Helper()
	dir := t.TempDir()
	removed := filepath.Join(dir, "lvremove.called")
	lvs := "#!/bin/sh\nprintf '%s' '" + lvsStdout + "'\nprintf '%s' '" + lvsStderr + "' >&2\nexit " + lvsExit + "\n"
	lvremove := "#!/bin/sh\necho \"$@\" > '" + removed + "'\n"
	for name, script := range map[string]string{LVList: lvs, LVRemove: lvremove} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return removed
}

func TestIsSnapshotExists(t *testing.T) {
	tests := map[string]struct {
		stdout, stderr, exit string
		want                 bool
		wantErr              bool
	}{
		"exists": {
			stdout: "  pvc-snap\n", exit: "0",
			want: true,
		},
		"logical volume not found": {
			stderr: "  Failed to find logical volume \"lvmvg/pvc-snap\"\n", exit: "5",
		},
		"volume group not found": {
			stderr: "  Volume group \"lvmvg\" not found\n  Cannot process volume group lvmvg\n", exit: "5",
		},
		"other lvs failure": {
			stderr: "  Global lock failed: check that global lockspace is started\n", exit: "5",
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fakeLVMTools(t, tt.stdout, tt.stderr, tt.exit)
			got, err := isSnapshotExists("lvmvg", "pvc-snap")
			if (err != nil) != tt.wantErr {
				t.Fatalf("isSnapshotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("isSnapshotExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDestroySnapshot(t *testing.T) {
	snap := &apis.LVMSnapshot{
		ObjectMeta: metav1.ObjectMeta{Name: "snapshot-pvc-snap"},
		Spec:       apis.LVMSnapshotSpec{VolGroup: "lvmvg"},
	}

	t.Run("lvs failure keeps the snapshot", func(t *testing.T) {
		removed := fakeLVMTools(t, "", "  Global lock failed\n", "5")
		if err := DestroySnapshot(snap); err == nil {
			t.Fatal("DestroySnapshot() = nil, want the lvs error so the finalizer is kept")
		}
		if _, err := os.Stat(removed); err == nil {
			t.Fatal("lvremove was called although the snapshot could not be checked")
		}
	})

	t.Run("missing snapshot is skipped", func(t *testing.T) {
		removed := fakeLVMTools(t, "", "  Failed to find logical volume \"lvmvg/pvc-snap\"\n", "5")
		if err := DestroySnapshot(snap); err != nil {
			t.Fatalf("DestroySnapshot() = %v, want nil", err)
		}
		if _, err := os.Stat(removed); err == nil {
			t.Fatal("lvremove was called for a missing snapshot")
		}
	})

	t.Run("existing snapshot is removed", func(t *testing.T) {
		removed := fakeLVMTools(t, "  pvc-snap\n", "", "0")
		if err := DestroySnapshot(snap); err != nil {
			t.Fatalf("DestroySnapshot() = %v, want nil", err)
		}
		if _, err := os.Stat(removed); err != nil {
			t.Fatal("lvremove was not called for an existing snapshot")
		}
	})
}
