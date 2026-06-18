package conversion

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

// SetupWithManager registers CRD conversion webhooks for all v1beta1 hub kinds.
func SetupWithManager(mgr ctrl.Manager) error {
	hubTypes := []struct {
		name string
		reg  func(ctrl.Manager) error
	}{
		{"Queue", func(m ctrl.Manager) error {
			return ctrl.NewWebhookManagedBy(m, &messagingv1beta1.Queue{}).Complete()
		}},
		{"Topic", func(m ctrl.Manager) error {
			return ctrl.NewWebhookManagedBy(m, &messagingv1beta1.Topic{}).Complete()
		}},
		{"Channel", func(m ctrl.Manager) error {
			return ctrl.NewWebhookManagedBy(m, &messagingv1beta1.Channel{}).Complete()
		}},
		{"ChannelAuthRule", func(m ctrl.Manager) error {
			return ctrl.NewWebhookManagedBy(m, &messagingv1beta1.ChannelAuthRule{}).Complete()
		}},
		{"AuthorityRecord", func(m ctrl.Manager) error {
			return ctrl.NewWebhookManagedBy(m, &messagingv1beta1.AuthorityRecord{}).Complete()
		}},
		{"QueueManagerConnection", func(m ctrl.Manager) error {
			return ctrl.NewWebhookManagedBy(m, &messagingv1beta1.QueueManagerConnection{}).Complete()
		}},
	}

	for _, hub := range hubTypes {
		if err := hub.reg(mgr); err != nil {
			return fmt.Errorf("setup %s conversion webhook: %w", hub.name, err)
		}
	}
	return nil
}
