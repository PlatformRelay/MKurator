package webhookv1beta1

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

func webhookTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := messagingv1beta1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add corev1 scheme: %v", err)
	}
	return scheme
}

func sampleWebhookConnV1Beta1(ns string) *messagingv1beta1.QueueManagerConnection {
	return &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "qm1", Namespace: ns},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "creds"},
		},
	}
}

func sampleWebhookChannelV1Beta1(ns, name, channelName string) *messagingv1beta1.Channel {
	return &messagingv1beta1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: messagingv1beta1.ChannelSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   channelName,
		},
	}
}

func TestQueueWebhookValidateCreateDeprecatedAttributes(t *testing.T) {
	t.Parallel()
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &queueCustomValidator{Client: cl}

	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "q1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
			Attributes:    map[string]string{"maxdepth": "500"},
		},
	}
	warnings, err := v.ValidateCreate(context.Background(), queue)
	if err != nil {
		t.Fatalf("ValidateCreate: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v", warnings)
	}
	if !strings.Contains(warnings[0], "maxdepth") || !strings.Contains(warnings[0], "spec.maxDepth") {
		t.Fatalf("unexpected warning: %q", warnings[0])
	}
}

func TestQueueWebhookValidateUpdateDelete(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &queueCustomValidator{Client: cl}

	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "q1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			QueueName:     "APP.ORDERS",
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), queue, queue); err != nil {
		t.Fatalf("ValidateUpdate: %v", err)
	}
	if _, err := v.ValidateDelete(context.Background(), queue); err != nil {
		t.Fatalf("ValidateDelete: %v", err)
	}
}

func TestQueueWebhookValidateCreateMissingConnection(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	v := &queueCustomValidator{Client: cl}

	queue := &messagingv1beta1.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: "q1", Namespace: "ns"},
		Spec: messagingv1beta1.QueueSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "missing"},
			QueueName:     "APP.ORDERS",
		},
	}
	if _, err := v.ValidateCreate(context.Background(), queue); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestTopicWebhookValidateCreate(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &topicCustomValidator{Client: cl}

	topic := &messagingv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "t1", Namespace: "ns"},
		Spec: messagingv1beta1.TopicSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			TopicName:     "RETAIL.ORDERS",
		},
	}
	if _, err := v.ValidateCreate(context.Background(), topic); err != nil {
		t.Fatalf("ValidateCreate: %v", err)
	}
}

func TestTopicWebhookValidateUpdateDelete(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &topicCustomValidator{Client: cl}

	topic := &messagingv1beta1.Topic{
		ObjectMeta: metav1.ObjectMeta{Name: "t1", Namespace: "ns"},
		Spec: messagingv1beta1.TopicSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			TopicName:     "RETAIL.ORDERS",
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), topic, topic); err != nil {
		t.Fatalf("ValidateUpdate: %v", err)
	}
	if _, err := v.ValidateDelete(context.Background(), topic); err != nil {
		t.Fatalf("ValidateDelete: %v", err)
	}
}

func TestChannelWebhookValidateCreate(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &channelCustomValidator{Client: cl}

	channel := &messagingv1beta1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: "c1", Namespace: "ns"},
		Spec: messagingv1beta1.ChannelSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
		},
	}
	if _, err := v.ValidateCreate(context.Background(), channel); err != nil {
		t.Fatalf("ValidateCreate: %v", err)
	}
}

func TestChannelWebhookValidateUpdateDelete(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &channelCustomValidator{Client: cl}

	channel := &messagingv1beta1.Channel{
		ObjectMeta: metav1.ObjectMeta{Name: "c1", Namespace: "ns"},
		Spec: messagingv1beta1.ChannelSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), channel, channel); err != nil {
		t.Fatalf("ValidateUpdate: %v", err)
	}
	if _, err := v.ValidateDelete(context.Background(), channel); err != nil {
		t.Fatalf("ValidateDelete: %v", err)
	}
}

func TestQueueManagerConnectionWebhookValidateCreate(t *testing.T) {
	scheme := webhookTestScheme(t)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"},
		Data:       map[string][]byte{"username": []byte("mquser"), "password": []byte("x")},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
	v := &queueManagerConnectionCustomValidator{Client: cl}

	conn := sampleWebhookConnV1Beta1("ns")
	if _, err := v.ValidateCreate(context.Background(), conn); err != nil {
		t.Fatalf("ValidateCreate: %v", err)
	}
}

func TestQueueManagerConnectionWebhookValidateUpdateDelete(t *testing.T) {
	scheme := webhookTestScheme(t)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"},
		Data:       map[string][]byte{"username": []byte("mquser"), "password": []byte("x")},
	}
	conn := sampleWebhookConnV1Beta1("ns")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret, conn).Build()
	v := &queueManagerConnectionCustomValidator{Client: cl}

	if _, err := v.ValidateUpdate(context.Background(), conn, conn); err != nil {
		t.Fatalf("ValidateUpdate: %v", err)
	}
	if _, err := v.ValidateDelete(context.Background(), conn); err != nil {
		t.Fatalf("ValidateDelete: %v", err)
	}
}

func TestQueueManagerConnectionWebhookValidateInvalidSpec(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	v := &queueManagerConnectionCustomValidator{Client: cl}

	conn := &messagingv1beta1.QueueManagerConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "bad", Namespace: "ns"},
		Spec: messagingv1beta1.QueueManagerConnectionSpec{
			QueueManager:         "QM1",
			Endpoint:             "https://mq.example:9443",
			CredentialsSecretRef: messagingv1beta1.SecretReference{Name: "missing"},
		},
	}
	if _, err := v.validate(context.Background(), conn); err == nil {
		t.Fatal("expected validation error for missing credentials secret")
	}
}

func TestQueueManagerConnectionWebhookValidateInsecureTLS(t *testing.T) {
	scheme := webhookTestScheme(t)
	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "ns"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(secret).Build()
	v := &queueManagerConnectionCustomValidator{Client: cl}

	conn := sampleWebhookConnV1Beta1("ns")
	conn.Spec.TLS = &messagingv1beta1.TLSConfig{InsecureSkipVerify: true}
	if _, err := v.ValidateCreate(context.Background(), conn); err == nil {
		t.Fatal("expected deny without allow-insecure-tls annotation")
	}

	conn.Annotations = map[string]string{
		messagingv1beta1.AllowInsecureTLSAnnotation: "true",
	}
	if _, err := v.ValidateCreate(context.Background(), conn); err != nil {
		t.Fatalf("ValidateCreate with opt-in annotation: %v", err)
	}
}

func TestChannelAuthRuleWebhookValidateCreate(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(sampleWebhookConnV1Beta1("ns"), sampleWebhookChannelV1Beta1("ns", "orders-app", "ORDERS.APP")).
		Build()
	v := &channelAuthRuleCustomValidator{Client: cl}

	rule := &messagingv1beta1.ChannelAuthRule{
		ObjectMeta: metav1.ObjectMeta{Name: "car1", Namespace: "ns"},
		Spec: messagingv1beta1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1beta1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		},
	}
	if _, err := v.ValidateCreate(context.Background(), rule); err != nil {
		t.Fatalf("ValidateCreate: %v", err)
	}
}

func TestChannelAuthRuleWebhookValidateUpdateDelete(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(sampleWebhookConnV1Beta1("ns"), sampleWebhookChannelV1Beta1("ns", "orders-app", "ORDERS.APP")).
		Build()
	v := &channelAuthRuleCustomValidator{Client: cl}

	rule := &messagingv1beta1.ChannelAuthRule{
		ObjectMeta: metav1.ObjectMeta{Name: "car1", Namespace: "ns"},
		Spec: messagingv1beta1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1beta1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), rule, rule); err != nil {
		t.Fatalf("ValidateUpdate: %v", err)
	}
	if _, err := v.ValidateDelete(context.Background(), rule); err != nil {
		t.Fatalf("ValidateDelete: %v", err)
	}
}

func TestChannelAuthRuleWebhookValidateCreateMissingChannel(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &channelAuthRuleCustomValidator{Client: cl}

	rule := &messagingv1beta1.ChannelAuthRule{
		ObjectMeta: metav1.ObjectMeta{Name: "car1", Namespace: "ns"},
		Spec: messagingv1beta1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1beta1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		},
	}
	if _, err := v.ValidateCreate(context.Background(), rule); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestChannelAuthRuleWebhookValidateUpdateDuringDeleteSkipsSpec(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &channelAuthRuleCustomValidator{Client: cl}
	now := metav1.Now()
	rule := &messagingv1beta1.ChannelAuthRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "car1",
			Namespace:         "ns",
			DeletionTimestamp: &now,
			Finalizers:        []string{messagingv1beta1.ChannelAuthRuleFinalizer},
		},
		Spec: messagingv1beta1.ChannelAuthRuleSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			ChannelName:   "ORDERS.APP",
			RuleType:      messagingv1beta1.ChannelAuthRuleTypeAddressMap,
			Address:       "*",
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), rule, rule); err != nil {
		t.Fatalf("ValidateUpdate during delete: %v", err)
	}
}

func TestAuthorityRecordWebhookValidateCreate(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &authorityRecordCustomValidator{Client: cl}

	auth := &messagingv1beta1.AuthorityRecord{
		ObjectMeta: metav1.ObjectMeta{Name: "auth1", Namespace: "ns"},
		Spec: messagingv1beta1.AuthorityRecordSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1beta1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   []string{"GET", "PUT"},
		},
	}
	if _, err := v.ValidateCreate(context.Background(), auth); err != nil {
		t.Fatalf("ValidateCreate: %v", err)
	}
}

func TestAuthorityRecordWebhookValidateUpdateDelete(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(sampleWebhookConnV1Beta1("ns")).Build()
	v := &authorityRecordCustomValidator{Client: cl}

	auth := &messagingv1beta1.AuthorityRecord{
		ObjectMeta: metav1.ObjectMeta{Name: "auth1", Namespace: "ns"},
		Spec: messagingv1beta1.AuthorityRecordSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1beta1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   []string{"GET", "PUT"},
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), auth, auth); err != nil {
		t.Fatalf("ValidateUpdate: %v", err)
	}
	if _, err := v.ValidateDelete(context.Background(), auth); err != nil {
		t.Fatalf("ValidateDelete: %v", err)
	}
}

func TestAuthorityRecordWebhookValidateUpdateDuringDeleteSkipsSpec(t *testing.T) {
	scheme := webhookTestScheme(t)
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	v := &authorityRecordCustomValidator{Client: cl}
	now := metav1.Now()
	auth := &messagingv1beta1.AuthorityRecord{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "auth1",
			Namespace:         "ns",
			DeletionTimestamp: &now,
			Finalizers:        []string{messagingv1beta1.AuthorityRecordFinalizer},
		},
		Spec: messagingv1beta1.AuthorityRecordSpec{
			ConnectionRef: messagingv1beta1.LocalObjectReference{Name: "qm1"},
			Profile:       "APP.ORDERS",
			ObjectType:    messagingv1beta1.AuthorityObjectTypeQueue,
			Principal:     "app",
			Authorities:   []string{"GET", "PUT"},
		},
	}
	if _, err := v.ValidateUpdate(context.Background(), auth, auth); err != nil {
		t.Fatalf("ValidateUpdate during delete: %v", err)
	}
}
