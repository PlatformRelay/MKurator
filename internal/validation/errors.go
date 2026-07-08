package validation

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	messagingv1alpha1 "github.com/conduit-ops/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

func invalidObject(gvk schema.GroupVersionKind, name string, errs field.ErrorList) error {
	return apierrors.NewInvalid(gvk.GroupKind(), name, errs)
}

// QueueInvalid returns an Invalid status error for Queue admission failures.
func QueueInvalid(name string, errs field.ErrorList) error {
	gvk := messagingv1alpha1.GroupVersion.WithKind("Queue")
	return invalidObject(gvk, name, errs)
}

// QueueInvalidV1Beta1 returns an Invalid status error for Queue v1beta1 admission failures.
func QueueInvalidV1Beta1(name string, errs field.ErrorList) error {
	gvk := messagingv1beta1.GroupVersion.WithKind("Queue")
	return invalidObject(gvk, name, errs)
}

// TopicInvalid returns an Invalid status error for Topic admission failures.
func TopicInvalid(name string, errs field.ErrorList) error {
	gvk := messagingv1alpha1.GroupVersion.WithKind("Topic")
	return invalidObject(gvk, name, errs)
}

// TopicInvalidV1Beta1 returns an Invalid status error for Topic v1beta1 admission failures.
func TopicInvalidV1Beta1(name string, errs field.ErrorList) error {
	gvk := messagingv1beta1.GroupVersion.WithKind("Topic")
	return invalidObject(gvk, name, errs)
}

// ChannelInvalid returns an Invalid status error for Channel admission failures.
func ChannelInvalid(name string, errs field.ErrorList) error {
	gvk := messagingv1alpha1.GroupVersion.WithKind("Channel")
	return invalidObject(gvk, name, errs)
}

// ChannelInvalidV1Beta1 returns an Invalid status error for Channel v1beta1 admission failures.
func ChannelInvalidV1Beta1(name string, errs field.ErrorList) error {
	gvk := messagingv1beta1.GroupVersion.WithKind("Channel")
	return invalidObject(gvk, name, errs)
}

// ChannelAuthRuleInvalid returns an Invalid status error for ChannelAuthRule admission failures.
func ChannelAuthRuleInvalid(name string, errs field.ErrorList) error {
	gvk := messagingv1alpha1.GroupVersion.WithKind("ChannelAuthRule")
	return invalidObject(gvk, name, errs)
}

// ChannelAuthRuleInvalidV1Beta1 returns an Invalid status error for ChannelAuthRule v1beta1 admission failures.
func ChannelAuthRuleInvalidV1Beta1(name string, errs field.ErrorList) error {
	gvk := messagingv1beta1.GroupVersion.WithKind("ChannelAuthRule")
	return invalidObject(gvk, name, errs)
}

// AuthorityRecordInvalid returns an Invalid status error for AuthorityRecord admission failures.
func AuthorityRecordInvalid(name string, errs field.ErrorList) error {
	gvk := messagingv1alpha1.GroupVersion.WithKind("AuthorityRecord")
	return invalidObject(gvk, name, errs)
}

// AuthorityRecordInvalidV1Beta1 returns an Invalid status error for AuthorityRecord v1beta1 admission failures.
func AuthorityRecordInvalidV1Beta1(name string, errs field.ErrorList) error {
	gvk := messagingv1beta1.GroupVersion.WithKind("AuthorityRecord")
	return invalidObject(gvk, name, errs)
}

// QueueManagerConnectionInvalid returns an Invalid status error for QMC admission failures.
func QueueManagerConnectionInvalid(name string, errs field.ErrorList) error {
	gvk := messagingv1alpha1.GroupVersion.WithKind("QueueManagerConnection")
	return apierrors.NewInvalid(
		schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind},
		name,
		errs,
	)
}

// QueueManagerConnectionInvalidV1Beta1 returns an Invalid status error for QMC v1beta1 admission failures.
func QueueManagerConnectionInvalidV1Beta1(name string, errs field.ErrorList) error {
	gvk := messagingv1beta1.GroupVersion.WithKind("QueueManagerConnection")
	return apierrors.NewInvalid(
		schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind},
		name,
		errs,
	)
}

// ObjectNameFromMeta returns the resource name used in Invalid errors.
func ObjectNameFromMeta(meta metav1.Object) string {
	if meta.GetName() != "" {
		return meta.GetName()
	}
	return meta.GetGenerateName()
}
