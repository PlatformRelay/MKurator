package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
