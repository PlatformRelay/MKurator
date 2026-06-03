package validation

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/konih/mkurator/api/v1alpha1"
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

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

func TestValidateChannelAuthRuleSpecBlockUserRequiresUserList(t *testing.T) {
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "dev-app-blockuser",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeBlockUser,
		})
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeRequired && err.Field == "spec.userList" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.userList required, got %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecBlockAddrRequiresAddress(t *testing.T) {
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "car-blockaddr",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeBlockAddr,
		})
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

func TestValidateChannelAuthRuleSpecBlockAddrValid(t *testing.T) {
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "car-blockaddr",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeBlockAddr,
			Address:       "192.0.2.1",
			Description:   "block example",
		})
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateChannelAuthRuleSpecBlockUserValid(t *testing.T) {
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

	errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "dev-app-blockuser",
		&messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeBlockUser,
			UserList:      "nobody",
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

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

func TestValidateChannelAuthRuleSpecMissingManagedChannel(t *testing.T) {
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
			RuleType:      messagingv1alpha1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		})
	found := false
	for _, err := range errs {
		if err.Type == field.ErrorTypeNotFound && err.Field == "spec.channelName" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected spec.channelName not found, got %v", errs)
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

func channelAuthRuleValidationClient(t *testing.T) client.Client {
	t.Helper()
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()
}

func channelAuthRuleFieldError(errs field.ErrorList, typ field.ErrorType, fld string) bool {
	for _, err := range errs {
		if err.Type == typ && err.Field == fld {
			return true
		}
	}
	return false
}

func TestValidateChannelAuthRuleSpecRuleTypeTable(t *testing.T) {
	t.Parallel()
	cl := channelAuthRuleValidationClient(t)

	baseSpec := func(ruleType messagingv1alpha1.ChannelAuthRuleType) *messagingv1alpha1.ChannelAuthRuleSpec {
		return &messagingv1alpha1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      ruleType,
		}
	}

	tests := []struct {
		name      string
		spec      *messagingv1alpha1.ChannelAuthRuleSpec
		wantField string
		wantType  field.ErrorType
	}{
		{
			name:      "addressmap missing address",
			spec:      baseSpec(messagingv1alpha1.ChannelAuthRuleTypeAddressMap),
			wantField: "spec.address",
			wantType:  field.ErrorTypeRequired,
		},
		{
			name:      "blockuser missing userList",
			spec:      baseSpec(messagingv1alpha1.ChannelAuthRuleTypeBlockUser),
			wantField: "spec.userList",
			wantType:  field.ErrorTypeRequired,
		},
		{
			name:      "blockaddr missing address",
			spec:      baseSpec(messagingv1alpha1.ChannelAuthRuleTypeBlockAddr),
			wantField: "spec.address",
			wantType:  field.ErrorTypeRequired,
		},
		{
			name: "usermap passes without deferred clientUser field",
			spec: baseSpec(messagingv1alpha1.ChannelAuthRuleTypeUserMap),
		},
		{
			name: "sslpeermap passes without deferred sslPeer field",
			spec: baseSpec(messagingv1alpha1.ChannelAuthRuleTypeSSLPeerMap),
		},
		{
			name: "qmgrmap passes without deferred qmgrName field",
			spec: baseSpec(messagingv1alpha1.ChannelAuthRuleTypeQMGRMap),
		},
		{
			name: "usermap does not require address",
			spec: func() *messagingv1alpha1.ChannelAuthRuleSpec {
				s := baseSpec(messagingv1alpha1.ChannelAuthRuleTypeUserMap)
				s.UserList = "nobody"
				return s
			}(),
		},
		{
			name: "blockaddr whitespace-only address is not treated as empty",
			spec: func() *messagingv1alpha1.ChannelAuthRuleSpec {
				s := baseSpec(messagingv1alpha1.ChannelAuthRuleTypeBlockAddr)
				s.Address = "  \t  "
				return s
			}(),
		},
		{
			name: "blockaddr channelName asterisk invalid MQ name",
			spec: func() *messagingv1alpha1.ChannelAuthRuleSpec {
				return &messagingv1alpha1.ChannelAuthRuleSpec{
					ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
					ChannelName:   "*",
					RuleType:      messagingv1alpha1.ChannelAuthRuleTypeBlockAddr,
					Address:       "192.0.2.1",
				}
			}(),
			wantField: "spec.channelName",
			wantType:  field.ErrorTypeInvalid,
		},
		{
			name: "blockaddr wrong-field combo userList only still requires address",
			spec: func() *messagingv1alpha1.ChannelAuthRuleSpec {
				s := baseSpec(messagingv1alpha1.ChannelAuthRuleTypeBlockAddr)
				s.UserList = "nobody"
				return s
			}(),
			wantField: "spec.address",
			wantType:  field.ErrorTypeRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateChannelAuthRuleSpec(context.Background(), cl, "default", "car-table",
				tt.spec)
			if tt.wantField == "" {
				if len(errs) > 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}
			if !channelAuthRuleFieldError(errs, tt.wantType, tt.wantField) {
				t.Fatalf("expected %s %s, got %v", tt.wantType, tt.wantField, errs)
			}
		})
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
	ch := sampleManagedChannel("default", "orders-app", "qm1", "ORDERS.APP")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, secret, ch).Build()

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
