package driver

import (
	"testing"
	"time"
)

func TestLockVolumeWithSnapshot_WaitsForSameSnapshot(t *testing.T) {
	l := newVolumeLock()
	unlock := l.LockVolumeWithSnapshot("vol-1", "snap-1")

	acquired := make(chan struct{})
	go func() {
		l.LockVolumeWithSnapshot("vol-1", "snap-1")()
		close(acquired)
	}()

	select {
	case <-acquired:
		t.Fatal("second lock acquired while the first is held")
	case <-time.After(50 * time.Millisecond):
	}

	unlock()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second lock not acquired after unlock")
	}
}

func TestLockVolume_IndependentNames(t *testing.T) {
	l := newVolumeLock()
	unlock := l.LockVolumeWithSnapshot("vol-1", "snap-1")
	defer unlock()

	done := make(chan struct{})
	go func() {
		l.LockVolumeWithSnapshot("vol-2", "snap-2")()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("lock on an unrelated volume and snapshot blocked")
	}
}
