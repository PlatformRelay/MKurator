package controller

import (
	"testing"
	"time"
)

func TestSetMQRequestTimeout(t *testing.T) {
	t.Parallel()
	prev := MQRequestTimeout()
	t.Cleanup(func() { SetMQRequestTimeout(prev) })
	SetMQRequestTimeout(0)
	if MQRequestTimeout() != prev {
		t.Fatal("zero ignored")
	}
	SetMQRequestTimeout(12 * time.Second)
	if MQRequestTimeout() != 12*time.Second {
		t.Fatal("timeout not set")
	}
}

func TestMQRequestContextDeadline(t *testing.T) {
	t.Parallel()
	prev := MQRequestTimeout()
	t.Cleanup(func() { SetMQRequestTimeout(prev) })
	SetMQRequestTimeout(50 * time.Millisecond)
	ctx, cancel := MQRequestContext(t.Context())
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) <= 0 {
		t.Fatal("expected future deadline")
	}
}
