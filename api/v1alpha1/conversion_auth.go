package v1alpha1

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

// authenticationUnionSnapshotAnnotation stashes the hub's spec.authentication union (as JSON)
// on the v1alpha1 spoke during down-conversion, and is consumed (then stripped) on up-conversion.
//
// v1alpha1 has no field to carry the union, so a bare down-then-up round trip through the spoke
// — which is exactly what happens when a v1alpha1-typed client (e.g. the reconciler's
// finalizer-add/remove Update, or any kubectl edit against the v1alpha1 endpoint) writes an
// object back without ever touching .spec.authentication itself — would otherwise silently wipe
// it (AUTH-14 post-merge e2e regression: the union never had a v1alpha1 code path, but the FIRST
// reconcile of any new object adds a finalizer via a full v1alpha1 Update, converting the
// just-read lossy spoke back to the hub and nil-ing Authentication in the process). The snapshot
// makes the round trip lossless for exactly this case while leaving true v1alpha1-origin writes
// (annotation absent) to keep going through the legacy CredentialsSecretRef path unchanged.
//
// Concurrent v1beta1 writers are still safe: the spoke and hub share one storage object and
// resourceVersion, so a stale spoke write (and its stale snapshot) is rejected as a conflict by
// the API server's optimistic concurrency check rather than silently clobbering a newer union.
const authenticationUnionSnapshotAnnotation = "messaging.mkurator.dev/authentication-union-snapshot"

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
	// onto the hub's now-optional pointer field. If this spoke carries an
	// authenticationUnionSnapshotAnnotation (stashed by ConvertFrom below), restore the union
	// from it — this is what makes a metadata-only round trip through the spoke (e.g. the
	// reconciler's finalizer Update, which never touches .spec) lossless instead of silently
	// wiping the union (AUTH-14). A true v1alpha1-origin write has no snapshot, so Authentication
	// stays nil as before.
	dst.Spec.CredentialsSecretRef = &messagingv1beta1.SecretReference{Name: src.Spec.CredentialsSecretRef.Name}
	dst.Spec.Authentication = nil
	if snapshot, ok := dst.Annotations[authenticationUnionSnapshotAnnotation]; ok {
		delete(dst.Annotations, authenticationUnionSnapshotAnnotation)
		var auth messagingv1beta1.MQWebAuthentication
		if err := json.Unmarshal([]byte(snapshot), &auth); err != nil {
			return fmt.Errorf("restore authentication union snapshot: %w", err)
		}
		dst.Spec.Authentication = &auth
	}

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
	// Down-convert to the v1alpha1 spoke. v1alpha1 has no authentication union field, so it is
	// dropped from the visible spec here — but first snapshotted as JSON onto
	// authenticationUnionSnapshotAnnotation so ConvertTo can restore it if this spoke is ever
	// written back without having touched .spec (AUTH-14; see ConvertTo for why that matters).
	// The union is NOT folded into CredentialsSecretRef: credential resolution for a union QMC
	// happens in the mqrest factory, which re-reads the v1beta1 hub (ADR-0027). A hub object
	// that carried only the union (nil CredentialsSecretRef) yields an empty spoke ref, which is
	// expected — the spoke is a lossy view, and the factory never reads the spoke field for such
	// objects.
	if src.Spec.CredentialsSecretRef != nil {
		dst.Spec.CredentialsSecretRef = SecretReference{Name: src.Spec.CredentialsSecretRef.Name}
	} else {
		dst.Spec.CredentialsSecretRef = SecretReference{}
	}
	if src.Spec.Authentication != nil {
		snapshot, err := json.Marshal(src.Spec.Authentication)
		if err != nil {
			return fmt.Errorf("snapshot authentication union: %w", err)
		}
		if dst.Annotations == nil {
			dst.Annotations = make(map[string]string, 1)
		}
		dst.Annotations[authenticationUnionSnapshotAnnotation] = string(snapshot)
	} else {
		delete(dst.Annotations, authenticationUnionSnapshotAnnotation)
	}

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	return nil
}
