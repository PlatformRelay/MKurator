package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ChannelAuthRuleType is the CHLAUTH rule TYPE.
// +kubebuilder:validation:Enum=ADDRESSMAP;BLOCKUSER;USERMAP;SSLPEERMAP;QMGRMAP;BLOCKADDR
type ChannelAuthRuleType string

const (
	ChannelAuthRuleTypeAddressMap ChannelAuthRuleType = "ADDRESSMAP"
	ChannelAuthRuleTypeBlockUser  ChannelAuthRuleType = "BLOCKUSER"
	ChannelAuthRuleTypeUserMap    ChannelAuthRuleType = "USERMAP"
	ChannelAuthRuleTypeSSLPeerMap ChannelAuthRuleType = "SSLPEERMAP"
	ChannelAuthRuleTypeQMGRMap    ChannelAuthRuleType = "QMGRMAP"
	ChannelAuthRuleTypeBlockAddr  ChannelAuthRuleType = "BLOCKADDR"
)

// ChannelAuthRuleFinalizer ensures CHLAUTH is removed before the CR is deleted.
const ChannelAuthRuleFinalizer = "messaging.kurator.dev/channelauthrule"

// ChannelAuthRuleSpec defines a SET CHLAUTH rule on a referenced queue manager.
type ChannelAuthRuleSpec struct {
	// ConnectionRef names a QueueManagerConnection in the same namespace.
	// +kubebuilder:validation:Required
	ConnectionRef LocalObjectReference `json:"connectionRef"`

	// ChannelName is the IBM MQ channel name in SET CHLAUTH('…').
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ChannelName string `json:"channelName"`

	// RuleType maps to CHLAUTH TYPE(...).
	// +kubebuilder:validation:Required
	RuleType ChannelAuthRuleType `json:"ruleType"`

	// Address maps to ADDRESS(...) for ADDRESSMAP and BLOCKADDR rules.
	// +optional
	Address string `json:"address,omitempty"`

	// UserSource maps to USERSRC(...) for ADDRESSMAP rules.
	// +optional
	UserSource string `json:"userSource,omitempty"`

	// CheckClient maps to CHCKCLNT(...) for ADDRESSMAP rules.
	// +optional
	CheckClient string `json:"checkClient,omitempty"`

	// Description maps to DESCR(...).
	// +optional
	Description string `json:"description,omitempty"`
}

// ChannelAuthRuleStatus defines the observed state of ChannelAuthRule.
type ChannelAuthRuleStatus struct {
	// Conditions represent the current state of the rule.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation last successfully synced.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	MQObjectStatusFields `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=car
// +kubebuilder:printcolumn:name="Synced",type=string,JSONPath=`.status.conditions[?(@.type=="Synced")].status`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="Synced")].reason`
// +kubebuilder:printcolumn:name="Channel",type=string,JSONPath=`.spec.channelName`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ChannelAuthRule maintains an IBM MQ CHLAUTH rule on a referenced queue manager.
type ChannelAuthRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ChannelAuthRuleSpec   `json:"spec,omitempty"`
	Status ChannelAuthRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ChannelAuthRuleList contains a list of ChannelAuthRule.
type ChannelAuthRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ChannelAuthRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ChannelAuthRule{}, &ChannelAuthRuleList{})
}
