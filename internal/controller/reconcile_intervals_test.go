package controller

import (
	"testing"
	"time"
)

func TestSetConnectionWaitInterval(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
