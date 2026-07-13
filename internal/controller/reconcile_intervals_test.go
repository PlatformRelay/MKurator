package controller

import (
	"testing"
	"time"
)

// TestSetConnectionWaitInterval and TestSetTransientRequeueInterval mutate the
// package-global reconcile intervals, which parallel reconciler tests read via
// ConnectionWaitInterval()/TransientRequeueInterval() (e.g.
// queue_controller_unit_test.go, queuemanagerconnection_reconciler_test.go
// assert result.RequeueAfter against the current global). They must NOT call
// t.Parallel(): running them in the serial phase keeps the global at its
// default (restored by t.Cleanup) throughout the parallel phase, so the
// readers never observe a mid-test mutation. See TESTQ-1 / audit TQ-Q6,TQ-Q7.
func TestSetConnectionWaitInterval(t *testing.T) {
	prev := ConnectionWaitInterval()
	t.Cleanup(func() { SetConnectionWaitInterval(prev) })

	SetConnectionWaitInterval(0)
	if ConnectionWaitInterval() != prev {
		t.Fatalf("zero ignored: got %v want %v", ConnectionWaitInterval(), prev)
	}

	SetConnectionWaitInterval(12 * time.Second)
	if ConnectionWaitInterval() != 12*time.Second {
		t.Fatalf("got %v", ConnectionWaitInterval())
	}
}

func TestSetTransientRequeueInterval(t *testing.T) {
	prev := TransientRequeueInterval()
	t.Cleanup(func() { SetTransientRequeueInterval(prev) })

	SetTransientRequeueInterval(-1 * time.Second)
	if TransientRequeueInterval() != prev {
		t.Fatalf("negative ignored: got %v want %v", TransientRequeueInterval(), prev)
	}

	SetTransientRequeueInterval(45 * time.Second)
	if TransientRequeueInterval() != 45*time.Second {
		t.Fatalf("got %v", TransientRequeueInterval())
	}
}
