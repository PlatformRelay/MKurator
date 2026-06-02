package mqadmin

import (
	"errors"
	"fmt"
	"testing"
)

func TestTerminalError(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("root")
	err := &TerminalError{Reason: "Auth", Message: "denied", Cause: cause}
	if err.Error() != "denied: root" {
		t.Fatalf("Error() = %q", err.Error())
	}
	if !errors.Is(err, ErrTerminal) {
		t.Fatal("expected ErrTerminal")
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected unwrap to cause")
	}
}

func TestTransientError(t *testing.T) {
	t.Parallel()
	err := &TransientError{Message: "timeout"}
	if !errors.Is(err, ErrTransient) {
		t.Fatal("expected ErrTransient")
	}
}

func TestNotFoundError(t *testing.T) {
	t.Parallel()
	err := &NotFoundError{Object: "APP.X"}
	if err.Error() != `mq object "APP.X" not found` {
		t.Fatalf("Error() = %q", err.Error())
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
}
