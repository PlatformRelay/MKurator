package validation

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
)

func TestValidateMQObjectName(t *testing.T) {
	t.Parallel()
	path := fieldRoot("name")

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid", value: "APP.ORDERS", wantErr: false},
		{name: "max length", value: "ABCDEFGHIJ.KLMNOPQRSTUVWXYZ0123456789./%&$#@", wantErr: false},
		{name: "too long", value: "ABCDEFGHIJ.KLMNOPQRSTUVWXYZ0123456789./%&$#@EXTRA", wantErr: true},
		{name: "empty", value: "", wantErr: true},
		{name: "leading dot", value: ".APP", wantErr: true},
		{name: "trailing dot", value: "APP.", wantErr: true},
		{name: "system prefix", value: "SYSTEM.DEFAULT", wantErr: true},
		{name: "system prefix lowercase", value: "system.default", wantErr: true},
		{name: "amq prefix", value: "AMQ.TEST", wantErr: true},
		{name: "amq prefix lowercase", value: "amq9999", wantErr: true},
		{name: "lowercase", value: "app.orders", wantErr: true},
		{name: "invalid char", value: "APP-ORDERS", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs := ValidateMQObjectName(path, tt.value)
			if tt.wantErr && len(errs) == 0 {
				t.Fatalf("expected error for %q", tt.value)
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Fatalf("unexpected error for %q: %v", tt.value, errs)
			}
		})
	}
}

func TestValidateConnectionRef(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)

	conn := &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
	deleting := conn.DeepCopy()
	deleting.Name = "qm-deleting"
	now := metav1.Now()
	deleting.DeletionTimestamp = &now
	deleting.Finalizers = []string{messagingv1alpha1.QueueManagerConnectionFinalizer}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, deleting).Build()
	path := fieldRoot("connectionRef").Child("name")

	t.Run("missing ref name", func(t *testing.T) {
		t.Parallel()
		if errs := ValidateConnectionRef(context.Background(), cl, "ns", "", path); len(errs) == 0 {
			t.Fatal("expected required error")
		}
	})
	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		if errs := ValidateConnectionRef(context.Background(), cl, "ns", "missing", path); len(errs) == 0 {
			t.Fatal("expected not found error")
		}
	})
	t.Run("deleting", func(t *testing.T) {
		t.Parallel()
		if errs := ValidateConnectionRef(context.Background(), cl, "ns", "qm-deleting", path); len(errs) == 0 {
			t.Fatal("expected deleting error")
		}
	})
	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		if errs := ValidateConnectionRef(context.Background(), cl, "ns", "qm1", path); len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
	})
}

func TestValidateQueueSpec(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	conn := sampleConnection("ns", "qm1")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()

	t.Run("alias without targq", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			QueueName:     "ALIAS.Q",
			Type:          messagingv1alpha1.QueueTypeAlias,
		}
		_, errs := ValidateQueueSpec(context.Background(), cl, "ns", "alias-queue", spec)
		if len(errs) == 0 {
			t.Fatal("expected missing targq error")
		}
	})
	t.Run("alias with target alias", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			QueueName:     "ALIAS.Q",
			Type:          messagingv1alpha1.QueueTypeAlias,
			Attributes:    map[string]string{"target": "APP.BASE"},
		}
		warnings, errs := ValidateQueueSpec(context.Background(), cl, "ns", "alias-queue", spec)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if len(warnings) != 0 {
			t.Fatalf("unexpected warnings: %v", warnings)
		}
	})
	t.Run("remote missing rqmname", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			QueueName:     "REMOTE.Q",
			Type:          messagingv1alpha1.QueueTypeRemote,
			Attributes:    map[string]string{"xmitq": "XMIT.Q"},
		}
		_, errs := ValidateQueueSpec(context.Background(), cl, "ns", "remote-queue", spec)
		if len(errs) == 0 {
			t.Fatal("expected missing rqmname error")
		}
	})
	t.Run("unknown attribute warning", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.Q",
			Attributes:    map[string]string{"notreal": "x"},
		}
		warnings, errs := ValidateQueueSpec(context.Background(), cl, "ns", "app-queue", spec)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if len(warnings) != 1 {
			t.Fatalf("expected one warning, got %v", warnings)
		}
	})
	t.Run("invalid metadata name", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.Q",
		}
		_, errs := ValidateQueueSpec(context.Background(), cl, "ns", "APP_Invalid", spec)
		if len(errs) == 0 {
			t.Fatal("expected invalid metadata.name error")
		}
	})
	t.Run("reserved queue name amq", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
			QueueName:     "AMQ.ORDERS",
		}
		_, errs := ValidateQueueSpec(context.Background(), cl, "ns", "app-queue", spec)
		if len(errs) == 0 {
			t.Fatal("expected reserved queueName error")
		}
	})
}

func TestValidateChannelSpec(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	conn := sampleConnection("ns", "qm1")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()

	spec := &messagingv1alpha1.ChannelSpec{
		ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
		ChannelName:   "ORDERS.APP",
		Type:          messagingv1alpha1.ChannelType("sender"),
	}
	if _, errs := ValidateChannelSpec(context.Background(), cl, "ns", "orders-app", spec); len(errs) == 0 {
		t.Fatal("expected unsupported channel type error")
	}
}

func TestValidateTopicSpec(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	conn := sampleConnection("ns", "qm1")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()

	spec := &messagingv1alpha1.TopicSpec{
		ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
		TopicName:     "RETAIL.ORDERS",
		Attributes:    map[string]string{"topstr": "A.B"},
	}
	warnings, errs := ValidateTopicSpec(context.Background(), cl, "ns", "retail-orders", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestValidateQueueSpecRemote(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	conn := sampleConnection("ns", "qm1")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()

	spec := &messagingv1alpha1.QueueSpec{
		ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
		QueueName:     "REMOTE.Q",
		Type:          messagingv1alpha1.QueueTypeRemote,
		Attributes: map[string]string{
			"transmissionqueue": "XMIT.Q",
			"remotemanager":     "QM2",
			"remotequeue":       "R.Q",
		},
	}
	warnings, errs := ValidateQueueSpec(context.Background(), cl, "ns", "remote-queue", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestInvalidHelpers(t *testing.T) {
	t.Parallel()
	errs := field.ErrorList{field.Required(field.NewPath("spec"), "required")}
	if err := QueueInvalid("q1", errs); err == nil {
		t.Fatal("expected QueueInvalid error")
	}
	if err := TopicInvalid("t1", errs); err == nil {
		t.Fatal("expected TopicInvalid error")
	}
	if err := ChannelInvalid("c1", errs); err == nil {
		t.Fatal("expected ChannelInvalid error")
	}
	if err := QueueManagerConnectionInvalid("conn", errs); err == nil {
		t.Fatal("expected QueueManagerConnectionInvalid error")
	}
	if got := ObjectNameFromMeta(&metav1.ObjectMeta{GenerateName: "gen-"}); got != "gen-" {
		t.Fatalf("ObjectNameFromMeta = %q", got)
	}
}

func sampleConnection(ns, name string) *messagingv1alpha1.QueueManagerConnection {
	return &messagingv1alpha1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1alpha1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1alpha1.SecretReference{Name: "creds"},
		},
	}
}

func sampleManagedChannel(ns, name, connName, channelName string) *messagingv1alpha1.Channel {
	return &messagingv1alpha1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1alpha1.ChannelSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: connName},
			ChannelName:   channelName,
		},
	}
}

func fieldRoot(name string) *field.Path {
	return field.NewPath("spec").Child(name)
}
