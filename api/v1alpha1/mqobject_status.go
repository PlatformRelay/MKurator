package v1alpha1

import messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"

// MQObjectStatusFields are shared status fields for Queue, Topic, Channel, and auth CRs.
//
// During the Phase-8e v1beta1 cut-over (8e-3) this is a type alias to the v1beta1 hub struct so
// a single controller-side MQObject interface is satisfied by BOTH the not-yet-migrated v1alpha1
// workloads and the migrated v1beta1 Queue. The alias disappears when v1alpha1 is dropped (8e-8).
type MQObjectStatusFields = messagingv1beta1.MQObjectStatusFields
