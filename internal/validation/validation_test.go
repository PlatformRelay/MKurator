package validation

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1alpha1 "github.com/conduit-ops/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

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
	t.Run("missing connection ref", func(t *testing.T) {
		t.Parallel()
		spec := &messagingv1alpha1.QueueSpec{
			ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "missing"},
			QueueName:     "APP.Q",
		}
		_, errs := ValidateQueueSpec(context.Background(), cl, "ns", "app-queue", spec)
		if len(errs) == 0 {
			t.Fatal("expected connection ref error")
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
	}
	warnings, errs := ValidateChannelSpec(context.Background(), cl, "ns", "orders-app", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestValidateChannelSpecSdrRequiresConnectionAttrs(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1alpha1.AddToScheme(scheme)
	conn := sampleConnection("ns", "qm1")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()

	spec := &messagingv1alpha1.ChannelSpec{
		ConnectionRef: messagingv1alpha1.LocalObjectReference{Name: "qm1"},
		ChannelName:   "QM1.TO.QM2",
		Type:          messagingv1alpha1.ChannelTypeSdr,
	}
	_, errs := ValidateChannelSpec(context.Background(), cl, "ns", "qm1-to-qm2", spec)
	if len(errs) != 2 {
		t.Fatalf("expected connName and xmitQueue required, got %v", errs)
	}

	spec.ConnName = "qm2.example.com(1414)"
	spec.XmitQueue = "SYSTEM.DEFAULT.XMIT.QUEUE"
	_, errs = ValidateChannelSpec(context.Background(), cl, "ns", "qm1-to-qm2", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
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
	if err := QueueInvalidV1Beta1("q1", errs); err == nil {
		t.Fatal("expected QueueInvalidV1Beta1 error")
	}
	if err := TopicInvalidV1Beta1("t1", errs); err == nil {
		t.Fatal("expected TopicInvalidV1Beta1 error")
	}
	if err := ChannelInvalidV1Beta1("c1", errs); err == nil {
		t.Fatal("expected ChannelInvalidV1Beta1 error")
	}
	if err := ChannelAuthRuleInvalidV1Beta1("car1", errs); err == nil {
		t.Fatal("expected ChannelAuthRuleInvalidV1Beta1 error")
	}
	if err := AuthorityRecordInvalidV1Beta1("auth1", errs); err == nil {
		t.Fatal("expected AuthorityRecordInvalidV1Beta1 error")
	}
	if err := QueueManagerConnectionInvalidV1Beta1("conn", errs); err == nil {
		t.Fatal("expected QueueManagerConnectionInvalidV1Beta1 error")
	}
	if got := ObjectNameFromMeta(&metav1.ObjectMeta{GenerateName: "gen-"}); got != "gen-" {
		t.Fatalf("ObjectNameFromMeta = %q", got)
	}
}

func TestValidateQueueSpecV1Beta1(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	spec := &messagingv1beta1.QueueSpec{
		ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
		QueueName:     "APP.Q",
		Attributes:    map[string]string{"maxdepth": "10"},
	}
	warnings, errs := ValidateQueueSpecV1Beta1(context.Background(), cl, "ns", "queue", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warnings) == 0 {
		t.Fatal("expected deprecated warning")
	}
}

func TestValidateQueueManagerConnectionSpecV1Beta1(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"},
		Data:       map[string][]byte{"username": []byte("user"), "password": []byte("pass")},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
	spec := &messagingv1beta1.QueueManagerConnectionSpec{
		QueueManager:         "QM1",
		Endpoint:             "https://mq.example:9443",
		CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
	}
	warnings, errs := ValidateQueueManagerConnectionSpecV1Beta1(context.Background(), cl, "ns", nil, spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestValidateTopicSpecV1Beta1(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	spec := &messagingv1beta1.TopicSpec{
		ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
		TopicName:     "TOPIC.ONE",
		Attributes:    map[string]string{"topicstr": "orders"},
	}
	warnings, errs := ValidateTopicSpecV1Beta1(context.Background(), cl, "ns", "topic", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warnings) == 0 {
		t.Fatal("expected deprecated warning")
	}
}

func TestValidateChannelSpecV1Beta1(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	spec := &messagingv1beta1.ChannelSpec{
		ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
		ChannelName:   "QM1.TO.QM2",
		Type:          messagingv1beta1.ChannelTypeSdr,
	}
	_, errs := ValidateChannelSpecV1Beta1(context.Background(), cl, "ns", "channel", spec)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %v", errs)
	}
	spec.ConnName = "qm2.example.com(1414)"
	spec.XmitQueue = "SYSTEM.DEFAULT.XMIT.QUEUE"
	warnings, errs := ValidateChannelSpecV1Beta1(context.Background(), cl, "ns", "channel", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
}

func TestValidateChannelAuthRuleSpecV1Beta1(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	spec := &messagingv1beta1.ChannelAuthRuleSpec{
		ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
		ChannelName:   "*",
		RuleType:      messagingv1beta1.ChannelAuthRuleTypeBlockAddr,
	}
	errs := ValidateChannelAuthRuleSpecV1Beta1(context.Background(), cl, "ns", "car", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateAuthorityRecordSpecV1Beta1(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	spec := &messagingv1beta1.AuthorityRecordSpec{
		ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
		Profile:       "APP.ORDERS",
		ObjectType:    messagingv1beta1.AuthorityObjectTypeQueue,
		Principal:     "app",
		Authorities:   []string{"GET"},
	}
	errs := ValidateAuthorityRecordSpecV1Beta1(context.Background(), cl, "ns", "auth", spec)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateQueueManagerConnectionDeleteV1Beta1(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	_ = messagingv1beta1.AddToScheme(scheme)
	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "q1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn, queue).Build()

	errs := ValidateQueueManagerConnectionDeleteV1Beta1(context.Background(), cl, conn)
	if len(errs) == 0 {
		t.Fatal("expected dependent validation error")
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
