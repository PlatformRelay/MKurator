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
// +kubebuilder:webhook:path=/validate-messaging-mkurator-dev-v1beta1-topic,mutating=false,failurePolicy=fail,sideEffects=None,groups=messaging.mkurator.dev,resources=topics,verbs=create;update,versions=v1beta1,name=vtopic-v1beta1.kb.io,admissionReviewVersions=v1

type topicCustomValidator struct {
	Client client.Reader
}

var _ admission.Validator[*messagingv1beta1.Topic] = &topicCustomValidator{}

func setupTopicWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &messagingv1beta1.Topic{}).
		WithValidator(&topicCustomValidator{Client: mgr.GetClient()}).
		Complete()
}

func (v *topicCustomValidator) ValidateCreate(
	ctx context.Context,
	topic *messagingv1beta1.Topic,
) (admission.Warnings, error) {
	return validateCreateUpdate(ctx, v.Client, topic, v.validateTopic, validation.TopicInvalidV1Beta1)
}

func (v *topicCustomValidator) ValidateUpdate(
	ctx context.Context,
	_ *messagingv1beta1.Topic,
	newTopic *messagingv1beta1.Topic,
) (admission.Warnings, error) {
	return validateCreateUpdate(ctx, v.Client, newTopic, v.validateTopic, validation.TopicInvalidV1Beta1)
}

func (v *topicCustomValidator) ValidateDelete(
	_ context.Context,
	_ *messagingv1beta1.Topic,
) (admission.Warnings, error) {
	return nil, nil
}

func (v *topicCustomValidator) validateTopic(
	ctx context.Context,
	reader client.Reader,
	topic *messagingv1beta1.Topic,
) ([]string, field.ErrorList) {
	return validation.ValidateTopicSpecV1Beta1(ctx, reader, topic.Namespace, topic.Name, &topic.Spec)
}
