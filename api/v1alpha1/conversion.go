package v1alpha1

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

// ConvertTo copies this Queue (v1alpha1 spoke) into the v1beta1 hub.
func (src *Queue) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*messagingv1beta1.Queue)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.Queue hub, got %T", dstRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	dst.Spec.ConnectionRef = messagingv1beta1.LocalObjectReference{Name: src.Spec.ConnectionRef.Name}
	dst.Spec.QueueName = src.Spec.QueueName
	dst.Spec.Type = messagingv1beta1.QueueType(src.Spec.Type)
	dst.Spec.MaxDepth = src.Spec.MaxDepth
	dst.Spec.Description = src.Spec.Description
	dst.Spec.DefPersistence = messagingv1beta1.QueueDefaultPersistence(src.Spec.DefPersistence)
	dst.Spec.Get = messagingv1beta1.QueueAccessEnabled(src.Spec.Get)
	dst.Spec.Put = messagingv1beta1.QueueAccessEnabled(src.Spec.Put)
	dst.Spec.TargetQueue = src.Spec.TargetQueue
	dst.Spec.XmitQueue = src.Spec.XmitQueue
	dst.Spec.RemoteQueueManager = src.Spec.RemoteQueueManager
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPolicies(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	attrs := messagingv1beta1.CloneStringMap(src.Spec.Attributes)
	messagingv1beta1.FoldQueueAttributesToTyped(&dst.Spec, attrs)
	dst.Spec.Attributes = attrs

	copyConditionsToHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFields(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertFrom copies the v1beta1 hub Queue into this v1alpha1 spoke.
//
//nolint:revive // kubebuilder convention: ConvertFrom receiver is dst.
func (dst *Queue) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*messagingv1beta1.Queue)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.Queue hub, got %T", srcRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	copyLocalObjectRefFromHub(&dst.Spec.ConnectionRef, src.Spec.ConnectionRef)
	dst.Spec.QueueName = src.Spec.QueueName
	dst.Spec.Type = QueueType(src.Spec.Type)
	dst.Spec.Attributes = messagingv1beta1.CloneStringMap(src.Spec.Attributes)
	dst.Spec.MaxDepth = src.Spec.MaxDepth
	dst.Spec.Description = src.Spec.Description
	dst.Spec.DefPersistence = QueueDefaultPersistence(src.Spec.DefPersistence)
	dst.Spec.Get = QueueAccessEnabled(src.Spec.Get)
	dst.Spec.Put = QueueAccessEnabled(src.Spec.Put)
	dst.Spec.TargetQueue = src.Spec.TargetQueue
	dst.Spec.XmitQueue = src.Spec.XmitQueue
	dst.Spec.RemoteQueueManager = src.Spec.RemoteQueueManager
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPoliciesFromHub(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFieldsFromHub(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertTo copies this Topic (v1alpha1 spoke) into the v1beta1 hub.
func (src *Topic) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*messagingv1beta1.Topic)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.Topic hub, got %T", dstRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	dst.Spec.ConnectionRef = messagingv1beta1.LocalObjectReference{Name: src.Spec.ConnectionRef.Name}
	dst.Spec.TopicName = src.Spec.TopicName
	dst.Spec.TopicString = src.Spec.TopicString
	dst.Spec.Description = src.Spec.Description
	dst.Spec.Publish = messagingv1beta1.TopicAccessEnabled(src.Spec.Publish)
	dst.Spec.Subscribe = messagingv1beta1.TopicAccessEnabled(src.Spec.Subscribe)
	dst.Spec.DefPersistence = messagingv1beta1.QueueDefaultPersistence(src.Spec.DefPersistence)
	dst.Spec.PublishScope = src.Spec.PublishScope
	dst.Spec.SubscribeScope = src.Spec.SubscribeScope
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPolicies(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	attrs := messagingv1beta1.CloneStringMap(src.Spec.Attributes)
	messagingv1beta1.FoldTopicAttributesToTyped(&dst.Spec, attrs)
	dst.Spec.Attributes = attrs

	copyConditionsToHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFields(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertFrom copies the v1beta1 hub Topic into this v1alpha1 spoke.
//
//nolint:revive // kubebuilder convention: ConvertFrom receiver is dst.
func (dst *Topic) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*messagingv1beta1.Topic)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.Topic hub, got %T", srcRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	copyLocalObjectRefFromHub(&dst.Spec.ConnectionRef, src.Spec.ConnectionRef)
	dst.Spec.TopicName = src.Spec.TopicName
	dst.Spec.Attributes = messagingv1beta1.CloneStringMap(src.Spec.Attributes)
	dst.Spec.TopicString = src.Spec.TopicString
	dst.Spec.Description = src.Spec.Description
	dst.Spec.Publish = TopicAccessEnabled(src.Spec.Publish)
	dst.Spec.Subscribe = TopicAccessEnabled(src.Spec.Subscribe)
	dst.Spec.DefPersistence = QueueDefaultPersistence(src.Spec.DefPersistence)
	dst.Spec.PublishScope = src.Spec.PublishScope
	dst.Spec.SubscribeScope = src.Spec.SubscribeScope
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPoliciesFromHub(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFieldsFromHub(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertTo copies this Channel (v1alpha1 spoke) into the v1beta1 hub.
func (src *Channel) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*messagingv1beta1.Channel)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.Channel hub, got %T", dstRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	dst.Spec.ConnectionRef = messagingv1beta1.LocalObjectReference{Name: src.Spec.ConnectionRef.Name}
	dst.Spec.ChannelName = src.Spec.ChannelName
	dst.Spec.Type = messagingv1beta1.ChannelType(src.Spec.Type)
	dst.Spec.Description = src.Spec.Description
	dst.Spec.MaxMsgLength = src.Spec.MaxMsgLength
	dst.Spec.TransportType = messagingv1beta1.ChannelTransportType(src.Spec.TransportType)
	dst.Spec.ShareConv = src.Spec.ShareConv
	dst.Spec.McaUser = src.Spec.McaUser
	dst.Spec.MaxInstances = src.Spec.MaxInstances
	dst.Spec.MaxInstancesClient = src.Spec.MaxInstancesClient
	dst.Spec.SslCipherSpec = src.Spec.SslCipherSpec
	dst.Spec.SslClientAuth = messagingv1beta1.ChannelSslClientAuth(src.Spec.SslClientAuth)
	dst.Spec.ConnName = src.Spec.ConnName
	dst.Spec.XmitQueue = src.Spec.XmitQueue
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPolicies(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	attrs := messagingv1beta1.CloneStringMap(src.Spec.Attributes)
	messagingv1beta1.FoldChannelAttributesToTyped(&dst.Spec, attrs)
	dst.Spec.Attributes = attrs

	copyConditionsToHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFields(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}

// ConvertFrom copies the v1beta1 hub Channel into this v1alpha1 spoke.
//
//nolint:revive // kubebuilder convention: ConvertFrom receiver is dst.
func (dst *Channel) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*messagingv1beta1.Channel)
	if !ok {
		return fmt.Errorf("expected messaging.mkurator.dev/v1beta1.Channel hub, got %T", srcRaw)
	}

	copyObjectMeta(&dst.ObjectMeta, src.ObjectMeta)

	copyLocalObjectRefFromHub(&dst.Spec.ConnectionRef, src.Spec.ConnectionRef)
	dst.Spec.ChannelName = src.Spec.ChannelName
	dst.Spec.Type = ChannelType(src.Spec.Type)
	dst.Spec.Attributes = messagingv1beta1.CloneStringMap(src.Spec.Attributes)
	dst.Spec.Description = src.Spec.Description
	dst.Spec.MaxMsgLength = src.Spec.MaxMsgLength
	dst.Spec.TransportType = ChannelTransportType(src.Spec.TransportType)
	dst.Spec.ShareConv = src.Spec.ShareConv
	dst.Spec.McaUser = src.Spec.McaUser
	dst.Spec.MaxInstances = src.Spec.MaxInstances
	dst.Spec.MaxInstancesClient = src.Spec.MaxInstancesClient
	dst.Spec.SslCipherSpec = src.Spec.SslCipherSpec
	dst.Spec.SslClientAuth = ChannelSslClientAuth(src.Spec.SslClientAuth)
	dst.Spec.ConnName = src.Spec.ConnName
	dst.Spec.XmitQueue = src.Spec.XmitQueue
	dst.Spec.Suspend = src.Spec.Suspend
	copyWorkloadPoliciesFromHub(&dst.Spec.WorkloadLifecyclePolicies, src.Spec.WorkloadLifecyclePolicies)

	copyConditionsFromHub(&dst.Status.Conditions, src.Status.Conditions)
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.DesiredMQSC = src.Status.DesiredMQSC
	copyMQObjectStatusFieldsFromHub(&dst.Status.MQObjectStatusFields, src.Status.MQObjectStatusFields)
	return nil
}
