package validation

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

func TestValidateChannelAuthRuleSpecAddressMapRequiresAddress(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "dev-app-addressmap",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
		})
	if len(errs) == 0 {
		t.Fatal("expected address required error")
	}
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeRequired && err.Field == "spec.address" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.address required, got %v", errs)
	}
}

func TestValidateAuthorityRecordSpecRequiresPrincipalOrGroup(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateAuthorityRecordSpec(context.Background(), cl, "default", "auth1",
		&messagingv1alpha1.AuthorityRecordSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1alpha1.AuthorityObjectTypeQueue,
			Authorities:   []string{"GET"},
		})
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeRequired && err.Field == "spec.principal" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.principal required, got %v", errs)
	}
}

func TestValidateAuthorityRecordSpecValid(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateAuthorityRecordSpec(context.Background(), cl, "default", "auth1",
		&messagingv1alpha1.AuthorityRecordSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1alpha1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   []string{"GET", "PUT"},
		})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecValid(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "dev-app-addressmap",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
			UserSource:    "CHANNEL",
		})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateAuthorityRecordSpecBothPrincipalAndGroup(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateAuthorityRecordSpec(context.Background(), cl, "default", "auth1",
		&messagingv1alpha1.AuthorityRecordSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1alpha1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Group:         "apps",
			Authorities:   []string{"GET"},
		})
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeInvalid && err.Field == "spec.group" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.group invalid, got %v", errs)
	}
}

func TestValidateAuthorityRecordSpecEmptyAuthority(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateAuthorityRecordSpec(context.Background(), cl, "default", "auth1",
		&messagingv1alpha1.AuthorityRecordSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1alpha1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   []string{" "},
		})
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeInvalid && err.Field == "spec.authorities[0]" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected empty authority error, got %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecMissingRuleType(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "default"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "default"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "car1",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
		})
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeRequired && err.Field == "spec.ruleType" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.ruleType required, got %v", errs)
	}
}
