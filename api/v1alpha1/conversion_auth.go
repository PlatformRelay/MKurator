package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// ConvertTo copies this ChannelAuthRule (v1alpha1 spoke) into the v1beta1 hub.
func (src *ChannelAuthRule) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*messagingv1beta1.ChannelAuthRule)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.ChannelAuthRule hub, got %T", dstRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	copyLocalObjectRef(&dst.Spec.ConnectionRef, src.Spec.ConnectionRef)
	dst.Spec.ChannelName = src.Spec.ChannelName
	dst.Spec.RuleType = messagingv1beta1.ChannelAuthRuleType(src.Spec.RuleType)
	dst.Spec.Address = src.Spec.Address
	dst.Spec.UserList = src.Spec.UserList
	dst.Spec.ClientUser = src.Spec.ClientUser
	dst.Spec.SslPeerName = src.Spec.SslPeerName
	dst.Spec.RemoteQueueManager = src.Spec.RemoteQueueManager
	dst.Spec.McaUser = src.Spec.McaUser
	dst.Spec.UserSource = messagingv1beta1.ChannelAuthUserSource(src.Spec.UserSource)
	dst.Spec.CheckClient = messagingv1beta1.ChannelAuthCheckClient(src.Spec.CheckClient)
	dst.Spec.Description = src.Spec.Description
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPolicies(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	copyConditionsToHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFields(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertFrom copies the v1beta1 hub ChannelAuthRule into this v1alpha1 spoke.
//
//nolint:revive // kubebuilder convention: ConvertFrom receiver is dst.
func (dst *ChannelAuthRule) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*messagingv1beta1.ChannelAuthRule)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.ChannelAuthRule hub, got %T", srcRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	copyLocalObjectRefFromHub(&dst.Spec.ConnectionRef, src.Spec.ConnectionRef)
	dst.Spec.ChannelName = src.Spec.ChannelName
	dst.Spec.RuleType = ChannelAuthRuleType(src.Spec.RuleType)
	dst.Spec.Address = src.Spec.Address
	dst.Spec.UserList = src.Spec.UserList
	dst.Spec.ClientUser = src.Spec.ClientUser
	dst.Spec.SslPeerName = src.Spec.SslPeerName
	dst.Spec.RemoteQueueManager = src.Spec.RemoteQueueManager
	dst.Spec.McaUser = src.Spec.McaUser
	dst.Spec.UserSource = ChannelAuthUserSource(src.Spec.UserSource)
	dst.Spec.CheckClient = ChannelAuthCheckClient(src.Spec.CheckClient)
	dst.Spec.Description = src.Spec.Description
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPoliciesFromHub(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFieldsFromHub(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertTo copies this AuthorityRecord (v1alpha1 spoke) into the v1beta1 hub.
func (src *AuthorityRecord) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*messagingv1beta1.AuthorityRecord)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.AuthorityRecord hub, got %T", dstRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	copyLocalObjectRef(&dst.Spec.ConnectionRef, src.Spec.ConnectionRef)
	dst.Spec.Profile = src.Spec.Profile
	dst.Spec.ObjectType = messagingv1beta1.AuthorityObjectType(src.Spec.ObjectType)
	dst.Spec.Principal = src.Spec.Principal
	dst.Spec.Group = src.Spec.Group
	copyStringSlice(&dst.Spec.Authorities, src.Spec.Authorities)
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPolicies(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	copyConditionsToHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFields(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertFrom copies the v1beta1 hub AuthorityRecord into this v1alpha1 spoke.
//
//nolint:revive // kubebuilder convention: ConvertFrom receiver is dst.
func (dst *AuthorityRecord) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*messagingv1beta1.AuthorityRecord)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.AuthorityRecord hub, got %T", srcRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	copyLocalObjectRefFromHub(&dst.Spec.ConnectionRef, src.Spec.ConnectionRef)
	dst.Spec.Profile = src.Spec.Profile
	dst.Spec.ObjectType = AuthorityObjectType(src.Spec.ObjectType)
	dst.Spec.Principal = src.Spec.Principal
	dst.Spec.Group = src.Spec.Group
	copyStringSlice(&dst.Spec.Authorities, src.Spec.Authorities)
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPoliciesFromHub(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFieldsFromHub(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertTo copies this QueueManagerConnection (v1alpha1 spoke) into the v1beta1 hub.
func (src *QueueManagerConnection) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*messagingv1beta1.QueueManagerConnection)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.QueueManagerConnection hub, got %T", dstRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	dst.Spec.QueueManager = src.Spec.QueueManager
	dst.Spec.Endpoint = src.Spec.Endpoint
	dst.Spec.RESTPrefix = src.Spec.RESTPrefix
	if src.Spec.TLS != nil {
		dst.Spec.TLS = &messagingv1beta1.TLSConfig{
			InsecureSkipVerify: src.Spec.TLS.InsecureSkipVerify,
		}
		if src.Spec.TLS.CASecretRef != nil {
			dst.Spec.TLS.CASecretRef = &messagingv1beta1.SecretReference{Name: src.Spec.TLS.CASecretRef.Name}
		}
	} else {
		dst.Spec.TLS = nil
	}
	// v1alpha1 has no authentication union: map its (always-present) CredentialsSecretRef
	// onto the hub's now-optional pointer field. The hub's authentication union is left nil
	// on the way up — it has no v1alpha1 source. Round-trip is exact for v1alpha1-origin
	// objects; a hub object that set authentication loses it on down-conversion (ADR-0027
	// AUTH-12 AC4 — v1beta1 is the storage hub, so nothing is lost at rest).
	dst.Spec.CredentialsSecretRef = &messagingv1beta1.SecretReference{Name: src.Spec.CredentialsSecretRef.Name}
	dst.Spec.Authentication = nil

	copyConditionsToHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	return nil
}

// ConvertFrom copies the v1beta1 hub QueueManagerConnection into this v1alpha1 spoke.
//
//nolint:revive // kubebuilder convention: ConvertFrom receiver is dst.
func (dst *QueueManagerConnection) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*messagingv1beta1.QueueManagerConnection)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.QueueManagerConnection hub, got %T", srcRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	dst.Spec.QueueManager = src.Spec.QueueManager
	dst.Spec.Endpoint = src.Spec.Endpoint
	dst.Spec.RESTPrefix = src.Spec.RESTPrefix
	if src.Spec.TLS != nil {
		dst.Spec.TLS = &TLSConfig{
			InsecureSkipVerify: src.Spec.TLS.InsecureSkipVerify,
		}
		if src.Spec.TLS.CASecretRef != nil {
			dst.Spec.TLS.CASecretRef = &SecretReference{Name: src.Spec.TLS.CASecretRef.Name}
		}
	} else {
		dst.Spec.TLS = nil
	}
	// Down-convert to the v1alpha1 spoke. v1alpha1 has no authentication union, so it is
	// simply DROPPED (AUTH-12 AC4) — v1beta1 is the storage hub, so nothing is lost at rest.
	// The union is NOT folded into CredentialsSecretRef: credential resolution for a union
	// QMC happens in the mqrest factory, which re-reads the v1beta1 hub (ADR-0027). A
	// hub object that carried only the union (nil CredentialsSecretRef) yields an empty
	// spoke ref, which is expected — the spoke is a lossy view, and the factory never reads
	// the spoke field for such objects.
	if src.Spec.CredentialsSecretRef != nil {
		dst.Spec.CredentialsSecretRef = SecretReference{Name: src.Spec.CredentialsSecretRef.Name}
	} else {
		dst.Spec.CredentialsSecretRef = SecretReference{}
	}

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	return nil
}
