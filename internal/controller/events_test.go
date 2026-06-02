package controller

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	messagingv1alpha1 "github.com/konradheimel/kurator/api/v1alpha1"
	"github.com/konradheimel/kurator/internal/mqadmin"
)

func TestRecordTerminalEvent_NilRecorder(t *testing.T) {
	t.Parallel()
	recordTerminalEvent(nil, &messagingv1alpha1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "default"},
	}, &mqadmin.TerminalError{Message: "bad mqsc"})
}

func TestRecordTerminalEvent_SkipsTransient(t *testing.T) {
	t.Parallel()
	recordTerminalEvent(nil, &messagingv1alpha1.Queue{}, &mqadmin.TransientError{Message: "timeout"})
}

func TestRecordTerminalEvent_EmitsWarning(t *testing.T) {
	t.Parallel()
	recorder := record.NewFakeRecorder(1)
	q := &messagingv1alpha1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "orders", Namespace: "default"},
	}
	recordTerminalEvent(recorder, q, &mqadmin.TerminalError{Message: "bad mqsc"})
	select {
	case ev := <-recorder.Events:
		if !strings.Contains(ev, corev1.EventTypeWarning) || !strings.Contains(ev, "TerminalError") {
			t.Fatalf("event = %q", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("expected event")
	}

	recorder = record.NewFakeRecorder(1)
	recordTerminalEvent(recorder, q, mqadmin.ErrNotFound)
	select {
	case ev := <-recorder.Events:
		if !strings.Contains(ev, "ReconcileError") {
			t.Fatalf("event = %q", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("expected event for non-terminal error")
	}
}
