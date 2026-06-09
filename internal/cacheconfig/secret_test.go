package cacheconfig

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStripSecretCredentialTransform(t *testing.T) {
	t.Parallel()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"},
		Data:       map[string][]byte{"username": []byte("admin"), "password": []byte("s3cret")},
		StringData: map[string]string{"note": "x"},
	}
	out, err := stripSecretCredentialTransform(secret)
	if err != nil {
		t.Fatalf("transform: %v", err)
	}
	stripped, ok := out.(*corev1.Secret)
	if !ok || stripped.Data != nil || stripped.StringData != nil {
		t.Fatalf("stripped = %+v", stripped)
	}
}

func TestManagerOptionsConfiguresSecretScoping(t *testing.T) {
	t.Parallel()
	cacheOpts, clientOpts := ManagerOptions()
	found := false
	for obj, cfg := range cacheOpts.ByObject {
		if _, ok := obj.(*corev1.Secret); ok {
			found = cfg.Transform != nil
			break
		}
	}
	if !found || clientOpts.Cache == nil {
		t.Fatal("expected secret scoping config")
	}
}
