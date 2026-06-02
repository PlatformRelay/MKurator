package controller

import "testing"

func TestSetMaxConcurrentReconciles(t *testing.T) {
	t.Parallel()
	SetMaxConcurrentReconciles(0)
	if maxConcurrentReconciles != 1 {
		t.Fatalf("min clamp: got %d", maxConcurrentReconciles)
	}
	SetMaxConcurrentReconciles(4)
	if controllerOptions().MaxConcurrentReconciles != 4 {
		t.Fatalf("options = %+v", controllerOptions())
	}
	SetMaxConcurrentReconciles(1)
}
