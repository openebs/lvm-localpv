package driver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	lvmapi "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/lvm"
)

const testSnapshotReadyTimeout = 500 * time.Millisecond

func fakeSnapshotGetter(t *testing.T, states func(call int) (string, error)) *int {
	t.Helper()
	calls := 0
	origGet, origInterval := getLVMSnapshot, snapshotReadyPollInterval
	t.Cleanup(func() {
		getLVMSnapshot, snapshotReadyPollInterval = origGet, origInterval
	})
	snapshotReadyPollInterval = 10 * time.Millisecond
	getLVMSnapshot = func(name string) (*lvmapi.LVMSnapshot, error) {
		calls++
		state, err := states(calls)
		if err != nil {
			return nil, err
		}
		snap := &lvmapi.LVMSnapshot{}
		snap.Name = name
		snap.Status.State = state
		return snap, nil
	}
	return &calls
}

func TestWaitForLVMSnapshotReady_BecomesReady(t *testing.T) {
	calls := fakeSnapshotGetter(t, func(call int) (string, error) {
		if call < 5 {
			return lvm.LVMStatusPending, nil
		}
		return lvm.LVMStatusReady, nil
	})

	snap, err := waitForLVMSnapshotReady(context.Background(), "snap-1", testSnapshotReadyTimeout)

	assert.NoError(t, err)
	assert.Equal(t, lvm.LVMStatusReady, snap.Status.State)
	assert.Equal(t, 5, *calls)
}

func TestWaitForLVMSnapshotReady_NotFound(t *testing.T) {
	calls := fakeSnapshotGetter(t, func(int) (string, error) {
		return "", k8serror.NewNotFound(schema.GroupResource{Group: "local.openebs.io", Resource: "lvmsnapshots"}, "snap-1")
	})

	_, err := waitForLVMSnapshotReady(context.Background(), "snap-1", testSnapshotReadyTimeout)

	assert.True(t, k8serror.IsNotFound(err))
	assert.Equal(t, 1, *calls)
}

func TestWaitForLVMSnapshotReady_RetriesAfterGetError(t *testing.T) {
	calls := fakeSnapshotGetter(t, func(call int) (string, error) {
		if call < 3 {
			return "", k8serror.NewServiceUnavailable("apiserver unavailable")
		}
		return lvm.LVMStatusReady, nil
	})

	snap, err := waitForLVMSnapshotReady(context.Background(), "snap-1", testSnapshotReadyTimeout)

	assert.NoError(t, err)
	assert.Equal(t, lvm.LVMStatusReady, snap.Status.State)
	assert.Equal(t, 3, *calls)
}

func TestWaitForLVMSnapshotReady_TimesOutWithLastState(t *testing.T) {
	fakeSnapshotGetter(t, func(int) (string, error) {
		return lvm.LVMStatusPending, nil
	})

	start := time.Now()
	snap, err := waitForLVMSnapshotReady(context.Background(), "snap-1", testSnapshotReadyTimeout)

	assert.NoError(t, err)
	assert.Equal(t, lvm.LVMStatusPending, snap.Status.State)
	assert.GreaterOrEqual(t, time.Since(start), testSnapshotReadyTimeout)
}

func TestWaitForLVMSnapshotReady_ReturnsContextError(t *testing.T) {
	fakeSnapshotGetter(t, func(int) (string, error) {
		return lvm.LVMStatusPending, nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := waitForLVMSnapshotReady(ctx, "snap-1", testSnapshotReadyTimeout)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), testSnapshotReadyTimeout)
}

func TestSnapshotWaitErrorCode(t *testing.T) {
	assert.Equal(t, codes.DeadlineExceeded, snapshotWaitErrorCode(context.DeadlineExceeded))
	assert.Equal(t, codes.Canceled, snapshotWaitErrorCode(context.Canceled))
	assert.Equal(t, codes.ResourceExhausted, snapshotWaitErrorCode(k8serror.NewServiceUnavailable("apiserver unavailable")))
}

func TestLockSnapshot(t *testing.T) {
	cs := &controller{}

	unlock, err := cs.lockSnapshot("snap-1")
	assert.NoError(t, err)

	_, err = cs.lockSnapshot("snap-1")
	assert.Equal(t, codes.Aborted, status.Code(err))

	unlockOther, err := cs.lockSnapshot("snap-2")
	assert.NoError(t, err)
	unlockOther()

	unlock()
	unlock, err = cs.lockSnapshot("snap-1")
	assert.NoError(t, err)
	unlock()
}
