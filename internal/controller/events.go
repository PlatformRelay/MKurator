package controller

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	"github.com/konradheimel/kurator/internal/mqadmin"
)

func recordTerminalEvent(recorder record.EventRecorder, obj runtime.Object, err error) {
	if recorder == nil || err == nil || errors.Is(err, mqadmin.ErrTransient) {
		return
	}
	var reason string
	switch {
	case errors.Is(err, mqadmin.ErrTerminal):
		reason = "TerminalError"
	default:
		reason = "ReconcileError"
	}
	recorder.Event(obj, corev1.EventTypeWarning, reason, err.Error())
}
