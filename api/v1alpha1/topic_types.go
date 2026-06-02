package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TopicFinalizer ensures the MQ topic is deleted before the CR is removed.
const TopicFinalizer = "messaging.kurator.dev/topic"

// TopicSpec defines an administrative topic object on a referenced queue manager.
type TopicSpec struct {
	// ConnectionRef names a QueueManagerConnection in the same namespace.
	// +kubebuilder:validation:Required
	ConnectionRef LocalObjectReference `json:"connectionRef"`

	// TopicName is the IBM MQ topic object name (e.g. RETAIL.ORDERS).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	TopicName string `json:"topicName"`

	// Attributes map to MQSC parameters (lowercase keys in mqweb runCommandJSON).
	// Common keys: topstr, descr, pub, sub, defpsist, pubscope, subscope.
	// +optional
	Attributes map[string]string `json:"attributes,omitempty"`
}

// TopicStatus defines the observed state of Topic.
type TopicStatus struct {
	// Conditions represent the current state of the topic.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation last successfully synced.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=tp
// +kubebuilder:printcolumn:name="Synced",type=string,JSONPath=`.status.conditions[?(@.type=="Synced")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Synced")].reason`
// +kubebuilder:printcolumn:name="Topic",type=string,JSONPath=`.spec.topicName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Topic maintains an IBM MQ topic object on a referenced queue manager.
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec,omitempty"`
	Status TopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicList contains a list of Topic.
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Topic{}, &TopicList{})
}
