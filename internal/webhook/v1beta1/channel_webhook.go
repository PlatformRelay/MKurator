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
// +kubebuilder:webhook:path=/validate-messaging-mkurator-dev-v1beta1-channel,mutating=false,failurePolicy=fail,sideEffects=None,groups=messaging.mkurator.dev,resources=channels,verbs=create;update,versions=v1beta1,name=vchannel-v1beta1.kb.io,admissionReviewVersions=v1

type channelCustomValidator struct {
	Client client.Reader
}

var _ admission.Validator[*messagingv1beta1.Channel] = &channelCustomValidator{}

func setupChannelWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &messagingv1beta1.Channel{}).
		WithValidator(&channelCustomValidator{Client: mgr.GetClient()}).
		Complete()
}

func (v *channelCustomValidator) ValidateCreate(
	ctx context.Context,
	channel *messagingv1beta1.Channel,
) (admission.Warnings, error) {
	return validateCreateUpdate(ctx, v.Client, channel, v.validateChannel, validation.ChannelInvalidV1Beta1)
}

func (v *channelCustomValidator) ValidateUpdate(
	ctx context.Context,
	_ *messagingv1beta1.Channel,
	newChannel *messagingv1beta1.Channel,
) (admission.Warnings, error) {
	return validateCreateUpdate(ctx, v.Client, newChannel, v.validateChannel, validation.ChannelInvalidV1Beta1)
}

func (v *channelCustomValidator) ValidateDelete(
	_ context.Context,
	_ *messagingv1beta1.Channel,
) (admission.Warnings, error) {
	return nil, nil
}

func (v *channelCustomValidator) validateChannel(
	ctx context.Context,
	reader client.Reader,
	channel *messagingv1beta1.Channel,
) ([]string, field.ErrorList) {
	return validation.ValidateChannelSpecV1Beta1(ctx, reader, channel.Namespace, channel.Name, &channel.Spec)
}
