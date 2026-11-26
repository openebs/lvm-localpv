package driver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	lvmapi "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/lvm"
)

const testSnapshotReadyTimeout = 500 * time.Millisecond

func fakeSnapshotGetter(t *testing.T, states func(call int) (string, error)) *int {
	t.Helper()
	calls := 0
	origGet, origInitial, origMax := getLVMSnapshot, snapshotReadyPollInitialInterval, snapshotReadyPollMaxInterval
	t.Cleanup(func() {
		getLVMSnapshot, snapshotReadyPollInitialInterval, snapshotReadyPollMaxInterval = origGet, origInitial, origMax
	})
	snapshotReadyPollInitialInterval, snapshotReadyPollMaxInterval = 5*time.Millisecond, 20*time.Millisecond
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

	assert.ErrorContains(t, err, `is still "Pending"`)
	assert.Equal(t, lvm.LVMStatusPending, snap.Status.State)
	assert.GreaterOrEqual(t, time.Since(start), testSnapshotReadyTimeout)
}

func TestWaitForLVMSnapshotReady_TimesOutBeforeContextDeadline(t *testing.T) {
	fakeSnapshotGetter(t, func(int) (string, error) {
		return lvm.LVMStatusPending, nil
	})
	origMargin := snapshotReadyReplyMargin
	t.Cleanup(func() { snapshotReadyReplyMargin = origMargin })
	snapshotReadyReplyMargin = 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	_, err := waitForLVMSnapshotReady(ctx, "snap-1", testSnapshotReadyTimeout)

	assert.ErrorContains(t, err, `is still "Pending"`)
	assert.NoError(t, ctx.Err())
}

func TestWaitForLVMSnapshotReady_ReturnsContextError(t *testing.T) {
	fakeSnapshotGetter(t, func(int) (string, error) {
		return lvm.LVMStatusPending, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	time.AfterFunc(50*time.Millisecond, cancel)

	start := time.Now()
	_, err := waitForLVMSnapshotReady(ctx, "snap-1", testSnapshotReadyTimeout)

	assert.ErrorIs(t, err, context.Canceled)
	assert.ErrorContains(t, err, `is still "Pending"`)
	assert.Equal(t, codes.Canceled, snapshotWaitErrorCode(err))
	assert.Less(t, time.Since(start), testSnapshotReadyTimeout)
}

func TestSnapshotWaitErrorCode(t *testing.T) {
	assert.Equal(t, codes.DeadlineExceeded, snapshotWaitErrorCode(context.DeadlineExceeded))
	assert.Equal(t, codes.Canceled, snapshotWaitErrorCode(context.Canceled))
	assert.Equal(t, codes.ResourceExhausted, snapshotWaitErrorCode(k8serror.NewServiceUnavailable("apiserver unavailable")))
}
