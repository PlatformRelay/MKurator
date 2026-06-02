package controller

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

type syncStatusOpts struct {
	mqObjectExists *bool
}

func connectionWaitMessage(conn *messagingv1alpha1.QueueManagerConnection) string {
	c, ok := findCondition(conn.Status.Conditions, messagingv1alpha1.ConditionReady)
	if !ok {
		return fmt.Sprintf("waiting for connection %q to become Ready", conn.Name)
	}
	if c.Status == metav1.ConditionTrue {
		return fmt.Sprintf("waiting for connection %q to become Ready", conn.Name)
	}
	switch {
	case c.Reason != "" && c.Message != "":
		return fmt.Sprintf("waiting for connection %q: %s — %s", conn.Name, c.Reason, c.Message)
	case c.Message != "":
		return fmt.Sprintf("waiting for connection %q: %s", conn.Name, c.Message)
	case c.Reason != "":
		return fmt.Sprintf("waiting for connection %q (%s)", conn.Name, c.Reason)
	default:
		return fmt.Sprintf("waiting for connection %q to become Ready", conn.Name)
	}
}

func findCondition(conditions []metav1.Condition, condType string) (metav1.Condition, bool) {
	for _, c := range conditions {
		if c.Type == condType {
			return c, true
		}
	}
	return metav1.Condition{}, false
}

func applyMQObjectStatusFields(obj client.Object, opts syncStatusOpts, message string, lastSync *metav1.Time) {
	exists := opts.mqObjectExists
	switch o := obj.(type) {
	case *messagingv1alpha1.Queue:
		o.Status.Message = message
		if lastSync != nil {
			o.Status.LastSyncTime = lastSync
		}
		if exists != nil {
			o.Status.MQObjectExists = exists
		}
	case *messagingv1alpha1.Topic:
		o.Status.Message = message
		if lastSync != nil {
			o.Status.LastSyncTime = lastSync
		}
		if exists != nil {
			o.Status.MQObjectExists = exists
		}
	case *messagingv1alpha1.Channel:
		o.Status.Message = message
		if lastSync != nil {
			o.Status.LastSyncTime = lastSync
		}
		if exists != nil {
			o.Status.MQObjectExists = exists
		}
	}
}

func boolPtr(b bool) *bool {
	return &b
}
