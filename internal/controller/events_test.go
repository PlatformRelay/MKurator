package controller

import (
	"fmt"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/events"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
	"github.com/platformrelay/mkurator/internal/mqadmin"
)

func TestClassifyReconcileError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantReason string
		wantMsg    string
	}{
		{
			name:       "terminal with reason",
			err:        &mqadmin.TerminalError{Reason: "MQSCError", Message: "define failed: AMQ8405E"},
			wantReason: "MQSCError",
			wantMsg:    "define failed: AMQ8405E",
		},
		{
			// REQ-REL-2026-08: terminal failures get no retry, so their condition reason
			// must be distinguishable from a retryable "Error".
			name:       "terminal without reason",
			err:        &mqadmin.TerminalError{Message: "bad mqsc"},
			wantReason: messagingv1beta1.ReasonTerminalError,
			wantMsg:    "bad mqsc",
		},
		{
			name: "connection not found",
			err: &mqadmin.ConnectionNotFoundError{
				Name: "qm1",
				Cause: apierrors.NewNotFound(
					schema.GroupResource{Resource: "queuemanagerconnections"},
					"qm1",
				),
			},
			wantReason: EventReasonConnectionNotFound,
		},
		{
			name: "credentials secret not found",
			err: &mqadmin.SecretNotFoundError{
				Name: "mq-creds",
				Role: "credentials",
				Cause: apierrors.NewNotFound(
					schema.GroupResource{Resource: "secrets"},
					"mq-creds",
				),
			},
			wantReason: EventReasonSecretNotFound,
		},
		{
			name: "ca secret not found",
			err: &mqadmin.SecretNotFoundError{
				Name: "mq-ca",
				Role: "CA",
				Cause: apierrors.NewNotFound(
					schema.GroupResource{Resource: "secrets"},
					"mq-ca",
				),
			},
			wantReason: EventReasonSecretNotFound,
		},
		{
			name: "substring-only connection error is generic",
			err: fmt.Errorf(
				`get connection "qm1": %w`,
				apierrors.NewNotFound(schema.GroupResource{Resource: "queuemanagerconnections"}, "qm1"),
			),
			wantReason: messagingv1beta1.ReasonError,
		},
		{
			name:       "generic error",
			err:        mqadmin.ErrNotFound,
			wantReason: messagingv1beta1.ReasonError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reason, msg := classifyReconcileError(tt.err)
			if reason != tt.wantReason {
				t.Fatalf("reason = %q, want %q", reason, tt.wantReason)
			}
			if tt.wantMsg != "" && msg != tt.wantMsg {
				t.Fatalf("message = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

func TestConditionChanged(t *testing.T) {
	t.Parallel()

	conditions := []metav1.Condition{{
		Type:   messagingv1beta1.ConditionSynced,
		Status: metav1.ConditionFalse,
		Reason: messagingv1beta1.ReasonProgressing,
	}}

	if !conditionChanged(
		conditions,
		messagingv1beta1.ConditionSynced,
		metav1.ConditionTrue,
		messagingv1beta1.ReasonAvailable,
	) {
		t.Fatal("expected changed on status transition")
	}
	if conditionChanged(
		conditions,
		messagingv1beta1.ConditionSynced,
		metav1.ConditionFalse,
		messagingv1beta1.ReasonProgressing,
	) {
		t.Fatal("expected unchanged when status and reason match")
	}
	if !conditionChanged(
		nil,
		messagingv1beta1.ConditionSynced,
		metav1.ConditionTrue,
		messagingv1beta1.ReasonAvailable,
	) {
		t.Fatal("expected changed when condition missing")
	}
}

// Smoke test: nil recorder must not panic (optional dependency in unit tests).
func TestRecordReconcileWarning_NilRecorder(t *testing.T) {
	t.Parallel()
	recordReconcileWarning(nil, &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "default"},
	}, &mqadmin.TerminalError{Message: "bad mqsc"})
}

func TestRecordReconcileWarning_SkipsTransient(t *testing.T) {
	t.Parallel()
	recorder := events.NewFakeRecorder(1)
	recordReconcileWarning(recorder, &messagingv1beta1.Queue{}, &mqadmin.TransientError{Message: "timeout"})
	select {
	case ev := <-recorder.Events:
		t.Fatalf("unexpected event: %q", ev)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestRecordReconcileWarning_EmitsWarning(t *testing.T) {
	t.Parallel()
	recorder := events.NewFakeRecorder(2)
	q := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "default"},
	}
	recordReconcileWarning(recorder, q, &mqadmin.TerminalError{Reason: "MQSCError", Message: "bad mqsc"})
	select {
	case ev := <-recorder.Events:
		if !strings.Contains(ev, corev1.EventTypeWarning) || !strings.Contains(ev, "MQSCError") {
			t.Fatalf("event = %q", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("expected event")
	}

	recordReconcileWarning(recorder, q, mqadmin.ErrNotFound)
	select {
	case ev := <-recorder.Events:
		if !strings.Contains(ev, messagingv1beta1.ReasonError) {
			t.Fatalf("event = %q", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("expected event for non-terminal error")
	}
}

func TestRecordNormalEvent(t *testing.T) {
	t.Parallel()
	recorder := events.NewFakeRecorder(1)
	q := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "default"},
	}
	recordNormalEvent(recorder, q, messagingv1beta1.ReasonAvailable, "Queue matches spec")
	select {
	case ev := <-recorder.Events:
		if !strings.Contains(ev, corev1.EventTypeNormal) || !strings.Contains(ev, messagingv1beta1.ReasonAvailable) {
			t.Fatalf("event = %q", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("expected normal event")
	}
}
