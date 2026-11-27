/*
Copyright 2026 The OpenEBS Authors

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

	mount "k8s.io/mount-utils"
)

func TestIsBlockTargetMounted(t *testing.T) {
	dir := t.TempDir()
	mountedTarget := filepath.Join(dir, "mounted")
	plainTarget := filepath.Join(dir, "plain")
	missingTarget := filepath.Join(dir, "missing")
	for _, p := range []string{mountedTarget, plainTarget} {
		if err := os.WriteFile(p, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}

	fake := mount.NewFakeMounter([]mount.MountPoint{{Device: "/dev/dm-1", Path: mountedTarget}})

	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{"already mounted target is reported as mounted", mountedTarget, true},
		{"existing file that is not a mount point", plainTarget, false},
		{"target that does not exist yet", missingTarget, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := isBlockTargetMounted(fake, tc.target)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("isBlockTargetMounted(%q) = %v, want %v", tc.target, got, tc.want)
			}
		})
	}
}
