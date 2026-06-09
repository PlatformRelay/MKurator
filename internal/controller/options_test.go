package controller

import "testing"

func TestControllerOptionsRecoverPanic(t *testing.T) {
	t.Parallel()
	opts := controllerOptions()
	if opts.RecoverPanic == nil || !*opts.RecoverPanic {
		t.Fatal("expected RecoverPanic enabled on controller options")
	}
}
