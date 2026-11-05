/*
Copyright 2026 The OpenEBS Authors.

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

package driver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	lvmapi "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/openebs/lvm-localpv/pkg/lvm"
)

func fakeSnapshotGetter(t *testing.T, states func(call int) (string, error)) *int {
	t.Helper()
	calls := 0
	origGet, origTimeout, origInterval := getLVMSnapshot, snapshotReadyTimeout, snapshotReadyPollInterval
	t.Cleanup(func() {
		getLVMSnapshot, snapshotReadyTimeout, snapshotReadyPollInterval = origGet, origTimeout, origInterval
	})
	snapshotReadyTimeout = 500 * time.Millisecond
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

	snap, err := waitForLVMSnapshotReady(context.Background(), "snap-1")

	assert.NoError(t, err)
	assert.Equal(t, lvm.LVMStatusReady, snap.Status.State)
	assert.Equal(t, 5, *calls)
}

func TestWaitForLVMSnapshotReady_NotFound(t *testing.T) {
	calls := fakeSnapshotGetter(t, func(int) (string, error) {
		return "", k8serror.NewNotFound(schema.GroupResource{Group: "local.openebs.io", Resource: "lvmsnapshots"}, "snap-1")
	})

	_, err := waitForLVMSnapshotReady(context.Background(), "snap-1")

	assert.True(t, k8serror.IsNotFound(err))
	assert.Equal(t, 1, *calls)
}

func TestWaitForLVMSnapshotReady_TimesOutWithLastState(t *testing.T) {
	fakeSnapshotGetter(t, func(int) (string, error) {
		return lvm.LVMStatusPending, nil
	})

	start := time.Now()
	snap, err := waitForLVMSnapshotReady(context.Background(), "snap-1")

	assert.NoError(t, err)
	assert.Equal(t, lvm.LVMStatusPending, snap.Status.State)
	assert.GreaterOrEqual(t, time.Since(start), snapshotReadyTimeout)
}

func TestWaitForLVMSnapshotReady_StopsWhenContextIsDone(t *testing.T) {
	fakeSnapshotGetter(t, func(int) (string, error) {
		return lvm.LVMStatusPending, nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	snap, err := waitForLVMSnapshotReady(ctx, "snap-1")

	assert.NoError(t, err)
	assert.Equal(t, lvm.LVMStatusPending, snap.Status.State)
	assert.Less(t, time.Since(start), snapshotReadyTimeout)
}
