package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func copyObjectMeta(dst *metav1.ObjectMeta, src metav1.ObjectMeta) {
	src.DeepCopyInto(dst)
}

func copyWorkloadPolicies(dst *messagingv1beta1.WorkloadLifecyclePolicies, src WorkloadLifecyclePolicies) {
	dst.DeletionPolicy = messagingv1beta1.DeletionPolicy(src.DeletionPolicy)
	dst.AdoptionPolicy = messagingv1beta1.AdoptionPolicy(src.AdoptionPolicy)
}

func copyWorkloadPoliciesFromHub(dst *WorkloadLifecyclePolicies, src messagingv1beta1.WorkloadLifecyclePolicies) {
	dst.DeletionPolicy = DeletionPolicy(src.DeletionPolicy)
	dst.AdoptionPolicy = AdoptionPolicy(src.AdoptionPolicy)
}

func copyMQObjectStatusFields(dst *messagingv1beta1.MQObjectStatusFields, src MQObjectStatusFields) {
	dst.Message = src.Message
	dst.LastSyncTime = src.LastSyncTime
	dst.MQObjectExists = src.MQObjectExists
}

func copyMQObjectStatusFieldsFromHub(dst *MQObjectStatusFields, src messagingv1beta1.MQObjectStatusFields) {
	dst.Message = src.Message
	dst.LastSyncTime = src.LastSyncTime
	dst.MQObjectExists = src.MQObjectExists
}

func copyLocalObjectRef(dst *messagingv1beta1.LocalObjectReference, src LocalObjectReference) {
	dst.Name = src.Name
}

func copyLocalObjectRefFromHub(dst *LocalObjectReference, src messagingv1beta1.LocalObjectReference) {
	dst.Name = src.Name
}

func copyInt32Ptr(dst **int32, src *int32) {
	if src == nil {
		*dst = nil
		return
	}
	v := *src
	*dst = &v
}

func copyBoolPtr(dst **bool, src *bool) {
	if src == nil {
		*dst = nil
		return
	}
	v := *src
	*dst = &v
}

func copyStringSlice(dst *[]string, src []string) {
	if len(src) == 0 {
		*dst = nil
		return
	}
	out := make([]string, len(src))
	copy(out, src)
	*dst = out
}

func copyConditionsToHub(dst *[]metav1.Condition, src []metav1.Condition) {
	if len(src) == 0 {
		*dst = nil
		return
	}
	out := make([]metav1.Condition, len(src))
	copy(out, src)
	*dst = out
}

func copyConditionsFromHub(dst *[]metav1.Condition, src []metav1.Condition) {
	copyConditionsToHub(dst, src)
}
