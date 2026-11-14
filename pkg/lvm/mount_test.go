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

func TestMakeFile(t *testing.T) {
	dir := t.TempDir()

	target := filepath.Join(dir, "block")
	if err := makeFile(target); err != nil {
		t.Fatalf("makeFile(%q) = %v, want nil", target, err)
	}
	if fi, err := os.Stat(target); err != nil || !fi.Mode().IsRegular() {
		t.Fatalf("makeFile(%q) did not create a regular file: %v", target, err)
	}

	// an existing target is reused
	if err := makeFile(target); err != nil {
		t.Fatalf("makeFile(%q) on an existing file = %v, want nil", target, err)
	}

	// a missing parent directory is reported as an error
	missing := filepath.Join(dir, "missing", "block")
	if err := makeFile(missing); err == nil {
		t.Fatalf("makeFile(%q) = nil, want an error", missing)
	}
}
