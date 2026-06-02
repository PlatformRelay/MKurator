package mqadmin

import (
	"context"
	"errors"
	"fmt"

	messagingv1alpha1 "github.com/konradheimel/kurator/api/v1alpha1"
)

// Factory builds Admin clients for a QueueManagerConnection.
type Factory interface {
	ForConnection(ctx context.Context, conn *messagingv1alpha1.QueueManagerConnection) (Admin, error)
}

// Admin is the seam between reconcilers and IBM MQ.
type Admin interface {
	Ping(ctx context.Context) error
	GetQueue(ctx context.Context, name string) (*QueueState, error)
	DefineQueue(ctx context.Context, spec QueueSpec) error
	DeleteQueue(ctx context.Context, name string) error
}

// QueueSpec is the domain shape for defining a local queue via MQSC.
type QueueSpec struct {
	Name       string
	Type       QueueType
	Attributes map[string]string
}

// QueueType mirrors the CRD queue type (local only in v1alpha1).
type QueueType string

const (
	QueueTypeLocal QueueType = "local"
)

// QueueState is the observed MQSC attributes of a queue.
type QueueState struct {
	Name       string
	Attributes map[string]string
}

// Sentinel errors for controller branching.
var (
	ErrNotFound  = errors.New("mq object not found")
	ErrTerminal  = errors.New("mq terminal error")
	ErrTransient = errors.New("mq transient error")
)

// TerminalError wraps a non-retryable MQ or REST failure.
type TerminalError struct {
	Reason  string
	Message string
	Cause   error
}

func (e *TerminalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *TerminalError) Unwrap() error { return e.Cause }

func (e *TerminalError) Is(target error) bool { return target == ErrTerminal }

// TransientError wraps a retryable failure (5xx, timeout, QM unavailable).
type TransientError struct {
	Message string
	Cause   error
}

func (e *TransientError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *TransientError) Unwrap() error { return e.Cause }

func (e *TransientError) Is(target error) bool { return target == ErrTransient }

// NotFoundError indicates the MQ object does not exist.
type NotFoundError struct {
	Object string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("mq object %q not found", e.Object)
}

func (e *NotFoundError) Is(target error) bool { return target == ErrNotFound }
