//nolint:dupl // workload webhook validators share the same controller-runtime shape
package webhookv1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
	"github.com/platformrelay/mkurator/internal/validation"
)

//nolint:lll // kubebuilder webhook marker is a single line
// +kubebuilder:webhook:path=/validate-messaging-mkurator-dev-v1beta1-queue,mutating=false,failurePolicy=fail,sideEffects=None,groups=messaging.mkurator.dev,resources=queues,verbs=create;update,versions=v1beta1,name=vqueue-v1beta1.kb.io,admissionReviewVersions=v1

type queueCustomValidator struct {
	Client client.Reader
}

var _ admission.Validator[*messagingv1beta1.Queue] = &queueCustomValidator{}

func setupQueueWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &messagingv1beta1.Queue{}).
		WithValidator(&queueCustomValidator{Client: mgr.GetClient()}).
		Complete()
}

func (v *queueCustomValidator) ValidateCreate(
	ctx context.Context,
	queue *messagingv1beta1.Queue,
) (admission.Warnings, error) {
	return validateCreateUpdate(ctx, v.Client, queue, v.validateQueue, validation.QueueInvalidV1Beta1)
}

func (v *queueCustomValidator) ValidateUpdate(
	ctx context.Context,
	_ *messagingv1beta1.Queue,
	newQueue *messagingv1beta1.Queue,
) (admission.Warnings, error) {
	return validateCreateUpdate(ctx, v.Client, newQueue, v.validateQueue, validation.QueueInvalidV1Beta1)
}

func (v *queueCustomValidator) ValidateDelete(
	_ context.Context,
	_ *messagingv1beta1.Queue,
) (admission.Warnings, error) {
	return nil, nil
}

func (v *queueCustomValidator) validateQueue(
	ctx context.Context,
	reader client.Reader,
	queue *messagingv1beta1.Queue,
) ([]string, field.ErrorList) {
	return validation.ValidateQueueSpecV1Beta1(ctx, reader, queue.Namespace, queue.Name, &queue.Spec)
}
