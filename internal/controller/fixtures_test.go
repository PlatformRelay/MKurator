package controller

import (
	"context"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/konradheimel/kurator/api/v1alpha1"
)

const (
	testQueueManager = "QM1"
	testEndpoint     = "https://mq.example:9443"
	testSecretName   = "mq-credentials"
	testQueueName    = "APP.ORDERS"
	testMaxDepth     = "5000"
	testAttrMaxDepth = "maxdepth"
)

func cleanupNamespace(ctx context.Context, ns string) {
	deleteAllOf(ctx, &messagingv1alpha1.QueueList{}, ns)
	deleteAllOf(ctx, &messagingv1alpha1.TopicList{}, ns)
	deleteAllOf(ctx, &messagingv1alpha1.ChannelList{}, ns)
	deleteAllOf(ctx, &messagingv1alpha1.QueueManagerConnectionList{}, ns)
	deleteAllOf(ctx, &corev1.SecretList{}, ns)
}

func deleteAllOf(ctx context.Context, list client.ObjectList, ns string) {
	Expect(k8sClient.List(ctx, list, client.InNamespace(ns))).To(Succeed())
	switch items := list.(type) {
	case *messagingv1alpha1.QueueList:
		for i := range items.Items {
			obj := &items.Items[i]
			obj.Finalizers = nil
			Expect(k8sClient.Update(ctx, obj)).To(Succeed())
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, obj))).To(Succeed())
		}
	case *messagingv1alpha1.TopicList:
		for i := range items.Items {
			obj := &items.Items[i]
			obj.Finalizers = nil
			Expect(k8sClient.Update(ctx, obj)).To(Succeed())
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, obj))).To(Succeed())
		}
	case *messagingv1alpha1.ChannelList:
		for i := range items.Items {
			obj := &items.Items[i]
			obj.Finalizers = nil
			Expect(k8sClient.Update(ctx, obj)).To(Succeed())
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, obj))).To(Succeed())
		}
	case *messagingv1alpha1.QueueManagerConnectionList:
		for i := range items.Items {
			obj := &items.Items[i]
			obj.Finalizers = nil
			Expect(k8sClient.Update(ctx, obj)).To(Succeed())
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, obj))).To(Succeed())
		}
	case *corev1.SecretList:
		for i := range items.Items {
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &items.Items[i]))).To(Succeed())
		}
	}
}
