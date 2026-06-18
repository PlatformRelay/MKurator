package v1beta1

// DeletionPolicy controls MQ cleanup when the CR is deleted.
// +kubebuilder:validation:Enum=Delete;Orphan
type DeletionPolicy string

const (
	DeletionPolicyDelete DeletionPolicy = "Delete"
	DeletionPolicyOrphan DeletionPolicy = "Orphan"
)

// AdoptionPolicy controls behaviour when the MQ object already exists on first reconcile.
// +kubebuilder:validation:Enum=Adopt;AdoptIfMatching;FailIfExists
type AdoptionPolicy string

const (
	AdoptionPolicyAdopt           AdoptionPolicy = "Adopt"
	AdoptionPolicyAdoptIfMatching AdoptionPolicy = "AdoptIfMatching"
	AdoptionPolicyFailIfExists    AdoptionPolicy = "FailIfExists"
)

type WorkloadLifecyclePolicies struct {
	// +kubebuilder:default=Delete
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`
	// +kubebuilder:default=Adopt
	// +optional
	AdoptionPolicy AdoptionPolicy `json:"adoptionPolicy,omitempty"`
}
