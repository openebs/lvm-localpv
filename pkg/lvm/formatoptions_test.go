package lvm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetFormatOptions puts the node defaults back to the empty state a fresh
// agent starts with, so one test does not decide the outcome of the next.
func resetFormatOptions(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		defaultFormatOptions = map[string][]string{}
	})
}

func TestSetDefaultFormatOptions(t *testing.T) {
	testSuite := map[string]struct {
		entries   []string
		expectErr bool
		expected  map[string][]string
	}{
		"no entries": {
			entries:  nil,
			expected: map[string][]string{},
		},
		"one entry": {
			entries:  []string{"xfs=-i nrext64=0"},
			expected: map[string][]string{"xfs": {"-i", "nrext64=0"}},
		},
		"one entry per filesystem": {
			entries: []string{"xfs=-i nrext64=0", "ext4=-O ^orphan_file"},
			expected: map[string][]string{
				"xfs":  {"-i", "nrext64=0"},
				"ext4": {"-O", "^orphan_file"},
			},
		},
		"options holding a comma stay one entry": {
			entries:  []string{"ext4=-O ^has_journal,^orphan_file"},
			expected: map[string][]string{"ext4": {"-O", "^has_journal,^orphan_file"}},
		},
		"extra spaces are dropped": {
			entries:  []string{"xfs=  -i   nrext64=0  "},
			expected: map[string][]string{"xfs": {"-i", "nrext64=0"}},
		},
		"the fstype is case insensitive": {
			entries:  []string{" XFS =-i nrext64=0"},
			expected: map[string][]string{"xfs": {"-i", "nrext64=0"}},
		},
		"an empty value leaves the filesystem without a default": {
			entries:  []string{"xfs="},
			expected: map[string][]string{},
		},
		"a later entry replaces an earlier one": {
			entries:  []string{"xfs=-i nrext64=0", "xfs=-b size=2048"},
			expected: map[string][]string{"xfs": {"-b", "size=2048"}},
		},
		"an entry without the fstype is an error": {
			entries:   []string{"-i nrext64=0"},
			expectErr: true,
		},
		"an entry for a filesystem the driver does not format is an error": {
			entries:   []string{"zfs=-o compression=on"},
			expectErr: true,
		},
		"a mistyped fstype is an error": {
			entries:   []string{"xsf=-i nrext64=0"},
			expectErr: true,
		},
	}

	for name, test := range testSuite {
		t.Run(name, func(t *testing.T) {
			resetFormatOptions(t)

			err := SetDefaultFormatOptions(test.entries)
			if test.expectErr {
				assert.Error(t, err, "expected %v to be rejected", test.entries)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expected, defaultFormatOptions)
		})
	}
}

// TestSetDefaultFormatOptionsKeepsOldOnError checks that a bad entry leaves the
// defaults of the node untouched, it does not half apply the new ones.
func TestSetDefaultFormatOptionsKeepsOldOnError(t *testing.T) {
	resetFormatOptions(t)

	assert.NoError(t, SetDefaultFormatOptions([]string{"xfs=-i nrext64=0"}))

	assert.Error(t, SetDefaultFormatOptions(
		[]string{"ext4=-O ^orphan_file", "zfs=-o compression=on"},
	))

	assert.Equal(t, []string{"-i", "nrext64=0"}, FormatOptions("xfs", ""))
	assert.Nil(t, FormatOptions("ext4", ""))
}

func TestFormatOptions(t *testing.T) {
	testSuite := map[string]struct {
		fstype    string
		scOptions string
		expected  []string
	}{
		"the node default is used when the storage class is silent": {
			fstype:   "xfs",
			expected: []string{"-i", "nrext64=0"},
		},
		"the storage class replaces the default, the two are not merged": {
			fstype:    "xfs",
			scOptions: "-b size=2048",
			expected:  []string{"-b", "size=2048"},
		},
		"a storage class can ask the feature back on": {
			fstype:    "xfs",
			scOptions: "-i nrext64=1",
			expected:  []string{"-i", "nrext64=1"},
		},
		"a filesystem with no default and no storage class value gets nothing": {
			fstype:   "btrfs",
			expected: nil,
		},
		"a filesystem with no default still takes its storage class value": {
			fstype:    "btrfs",
			scOptions: "-M",
			expected:  []string{"-M"},
		},
		"an empty fstype falls back on the ext4 default": {
			fstype:   "",
			expected: []string{"-O", "^orphan_file"},
		},
		"the fstype is case insensitive": {
			fstype:   "XFS",
			expected: []string{"-i", "nrext64=0"},
		},
		"an all blank storage class value is no value": {
			fstype:    "xfs",
			scOptions: "   ",
			expected:  []string{"-i", "nrext64=0"},
		},
		"extra spaces in the storage class value are dropped": {
			fstype:    "ext4",
			scOptions: "-b  4096   -N 5000000",
			expected:  []string{"-b", "4096", "-N", "5000000"},
		},
	}

	for name, test := range testSuite {
		t.Run(name, func(t *testing.T) {
			resetFormatOptions(t)

			assert.NoError(t, SetDefaultFormatOptions(
				[]string{"xfs=-i nrext64=0", "ext4=-O ^orphan_file"},
			))

			assert.Equal(t, test.expected, FormatOptions(test.fstype, test.scOptions))
		})
	}
}

// TestFormatOptionsBeforeSet covers the controller, and any node whose agent
// was given no defaults, where nothing must be added to the mkfs command line.
func TestFormatOptionsBeforeSet(t *testing.T) {
	resetFormatOptions(t)

	for _, fstype := range []string{"ext2", "ext3", "ext4", "xfs", "btrfs", ""} {
		assert.Nil(t, FormatOptions(fstype, ""), "expected no options for %q", fstype)
	}

	assert.Equal(t, []string{"-b", "4096"}, FormatOptions("xfs", "-b 4096"))
}
