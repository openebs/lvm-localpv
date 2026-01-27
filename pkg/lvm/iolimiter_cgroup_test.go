package lvm

import (
	"os"
	"path/filepath"
	"testing"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
)

func TestGetPodUIDFromPath(t *testing.T) {
	kDir = "/var/lib/kubelet/"
	tests := []struct {
		name string
		in   string
		uid  string
		ok   bool
	}{
		{
			name: "valid kubelet csi mount path",
			in:   "/var/lib/kubelet/pods/eb9d54fe-07d5-4428-9c46-f937f4e12043/volumes/kubernetes.io~csi/pvc-aaa/mount",
			uid:  "eb9d54fe-07d5-4428-9c46-f937f4e12043",
			ok:   true,
		},
		{
			name: "path without kubelet prefix",
			in:   "/something/else/pods/eb9d54fe-07d5-4428-9c46-f937f4e12043/mount",
			uid:  "",
			ok:   false,
		},
		{
			name: "prefix but no slash after uid",
			in:   "/var/lib/kubelet/pods/eb9d54fe-07d5-4428-9c46-f937f4e12043",
			uid:  "",
			ok:   false,
		},
		{
			name: "prefix but empty uid",
			in:   "/var/lib/kubelet/pods//volumes/kubernetes.io~csi/x/mount",
			uid:  "",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, ok := getPodUIDFromPath(tt.in)
			if ok != tt.ok {
				t.Fatalf("ok mismatch: got %v want %v (uid=%q)", ok, tt.ok, uid)
			}
			if uid != tt.uid {
				t.Fatalf("uid mismatch: got %q want %q", uid, tt.uid)
			}
		})
	}
}

func TestQoSValuesConvert(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "max", want: "max", wantErr: false},
		{in: "1", want: "1", wantErr: false},
		{in: "123456", want: "123456", wantErr: false},
		{in: "", wantErr: true},
		{in: "0", wantErr: true},
		{in: "nope", wantErr: true},
		{in: " 1", wantErr: true},
	}

	for _, tt := range tests {
		got, err := qosValuesConvert(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("QoSValuesConvert(%q) expected error, got %q", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("QoSValuesConvert(%q) unexpected error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("QoSValuesConvert(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestSetQoSValuesLine(t *testing.T) {
	maj := uint32(253)
	min := uint32(7)

	t.Run("nil qos means all max", func(t *testing.T) {
		got, err := setQoSValuesLine(maj, min, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "253:7 rbps=max wbps=max riops=max wiops=max"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("mixed values", func(t *testing.T) {
		qos := &apis.VolumeQoS{
			ReadBPS:   "100",
			WriteBPS:  "max",
			ReadIOPS:  "5",
			WriteIOPS: "max",
		}
		got, err := setQoSValuesLine(maj, min, qos)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "253:7 rbps=100 wbps=max riops=5 wiops=max"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("invalid value triggers error", func(t *testing.T) {
		qos := &apis.VolumeQoS{
			ReadBPS:   "0",
			WriteBPS:  "max",
			ReadIOPS:  "5",
			WriteIOPS: "max",
		}
		_, err := setQoSValuesLine(maj, min, qos)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestQoSValuesWriteFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "io.max")

	if err := os.WriteFile(p, []byte(""), 0o644); err != nil {
		t.Fatalf("prepare file: %v", err)
	}

	line := "253:7 rbps=100 wbps=max riops=5 wiops=max"
	if err := QoSValuesWriteFile(p, line); err != nil {
		t.Fatalf("QoSValuesWriteFile error: %v", err)
	}

	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	want := line + "\n"
	if string(b) != want {
		t.Fatalf("file content mismatch:\n got: %q\nwant: %q", string(b), want)
	}
}
