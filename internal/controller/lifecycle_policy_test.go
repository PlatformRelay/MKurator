package controller

import (
	"testing"

	messagingv1beta1 "github.com/platformrelay/mkurator/api/v1beta1"
)

func TestOrphanDeletionRequested(t *testing.T) {
	t.Parallel()
	q := &messagingv1beta1.Queue{}
	if orphanDeletionRequested(q) {
		t.Fatal("expected false")
	}
	q.Annotations = map[string]string{messagingv1beta1.ForceOrphanAnnotation: "true"}
	if !orphanDeletionRequested(q) {
		t.Fatal("expected force-orphan")
	}
	q.Annotations = nil
	q.Spec.DeletionPolicy = messagingv1beta1.DeletionPolicyOrphan
	if !orphanDeletionRequested(q) {
		t.Fatal("expected Orphan policy")
	}
}
