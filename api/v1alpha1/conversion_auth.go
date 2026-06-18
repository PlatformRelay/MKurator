package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
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
	dst.Spec.CredentialsSecretRef = messagingv1beta1.SecretReference{Name: src.Spec.CredentialsSecretRef.Name}

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
	dst.Spec.CredentialsSecretRef = SecretReference{Name: src.Spec.CredentialsSecretRef.Name}

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	return nil
}
