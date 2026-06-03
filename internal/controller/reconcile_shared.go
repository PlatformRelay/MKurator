package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	messagingv1alpha1 "github.com/konih/kurator/api/v1alpha1"
	"github.com/konih/kurator/internal/mqadmin"
)

func resolveConnection(
	ctx context.Context,
	c client.Client,
	namespace, name string,
) (*messagingv1alpha1.QueueManagerConnection, error) {
	conn := &messagingv1alpha1.QueueManagerConnection{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, conn); err != nil {
		return nil, fmt.Errorf("get connection %q: %w", name, err)
	}
	return conn, nil
}

func waitForConnectionReady(
	ctx context.Context,
	status client.StatusWriter,
	recorder events.EventRecorder,
	obj client.Object,
	conn *messagingv1alpha1.QueueManagerConnection,
	generation int64,
) (ctrl.Result, bool, error) {
	if connectionReady(conn) {
		return ctrl.Result{}, false, nil
	}
	msg := connectionWaitMessage(conn)
	if err := patchSyncedProgressing(ctx, status, recorder, obj, generation, msg); err != nil {
		return ctrl.Result{}, true, err
	}
	return ctrl.Result{RequeueAfter: 15 * time.Second}, true, nil
}

func syncedConditions(obj client.Object) []metav1.Condition {
	switch o := obj.(type) {
	case *messagingv1alpha1.Queue:
		return o.Status.Conditions
	case *messagingv1alpha1.Topic:
		return o.Status.Conditions
	case *messagingv1alpha1.Channel:
		return o.Status.Conditions
	case *messagingv1alpha1.ChannelAuthRule:
		return o.Status.Conditions
	case *messagingv1alpha1.AuthorityRecord:
		return o.Status.Conditions
	default:
		return nil
	}
}

func emitSyncedTransitionEvent(
	recorder events.EventRecorder,
	obj client.Object,
	newStatus metav1.ConditionStatus,
	newReason, message string,
) {
	if conditionChanged(syncedConditions(obj), messagingv1alpha1.ConditionSynced, newStatus, newReason) {
		recordNormalEvent(recorder, obj, newReason, message)
	}
}

//nolint:dupl // progressing vs deleting share the same per-kind status patch shape
func patchSyncedProgressing(
	ctx context.Context,
	status client.StatusWriter,
	recorder events.EventRecorder,
	obj client.Object,
	generation int64,
	message string,
) error {
	emitSyncedTransitionEvent(recorder, obj, metav1.ConditionFalse, messagingv1alpha1.ReasonProgressing, message)

	switch o := obj.(type) {
	case *messagingv1alpha1.Queue:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonProgressing, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.Topic:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonProgressing, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.Channel:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonProgressing, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.ChannelAuthRule:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonProgressing, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.AuthorityRecord:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonProgressing, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	default:
		return fmt.Errorf("patchSyncedProgressing: unsupported type %T", obj)
	}
}

func setSyncedError(
	ctx context.Context,
	status client.StatusWriter,
	recorder events.EventRecorder,
	obj client.Object,
	generation int64,
	err error,
	opts syncStatusOpts,
) (ctrl.Result, error) {
	recordReconcileWarning(recorder, obj, err)

	reason, message := classifyReconcileError(err)
	requeue := ctrl.Result{}
	if errors.Is(err, mqadmin.ErrTransient) {
		requeue = ctrl.Result{RequeueAfter: 30 * time.Second}
	}

	switch o := obj.(type) {
	case *messagingv1alpha1.Queue:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, reason, message, generation)
		applyMQObjectStatusFields(o, opts, message, nil)
		if statusErr := status.Update(ctx, o); statusErr != nil {
			return requeue, statusErr
		}
	case *messagingv1alpha1.Topic:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, reason, message, generation)
		applyMQObjectStatusFields(o, opts, message, nil)
		if statusErr := status.Update(ctx, o); statusErr != nil {
			return requeue, statusErr
		}
	case *messagingv1alpha1.Channel:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, reason, message, generation)
		applyMQObjectStatusFields(o, opts, message, nil)
		if statusErr := status.Update(ctx, o); statusErr != nil {
			return requeue, statusErr
		}
	case *messagingv1alpha1.ChannelAuthRule:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, reason, message, generation)
		applyMQObjectStatusFields(o, opts, message, nil)
		if statusErr := status.Update(ctx, o); statusErr != nil {
			return requeue, statusErr
		}
	case *messagingv1alpha1.AuthorityRecord:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, reason, message, generation)
		applyMQObjectStatusFields(o, opts, message, nil)
		if statusErr := status.Update(ctx, o); statusErr != nil {
			return requeue, statusErr
		}
	default:
		return ctrl.Result{}, fmt.Errorf("setSyncedError: unsupported type %T", obj)
	}

	if errors.Is(err, mqadmin.ErrTransient) {
		return requeue, err
	}
	return ctrl.Result{}, nil
}

func patchSyncedAvailable(
	ctx context.Context,
	status client.StatusWriter,
	recorder events.EventRecorder,
	obj client.Object,
	generation int64,
	message string,
	opts syncStatusOpts,
) error {
	emitSyncedTransitionEvent(recorder, obj, metav1.ConditionTrue, messagingv1alpha1.ReasonAvailable, message)
	now := metav1.Now()

	switch o := obj.(type) {
	case *messagingv1alpha1.Queue:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionTrue, messagingv1alpha1.ReasonAvailable, message, generation)
		o.Status.ObservedGeneration = generation
		applyMQObjectStatusFields(o, opts, message, &now)
		return status.Update(ctx, o)
	case *messagingv1alpha1.Topic:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionTrue, messagingv1alpha1.ReasonAvailable, message, generation)
		o.Status.ObservedGeneration = generation
		applyMQObjectStatusFields(o, opts, message, &now)
		return status.Update(ctx, o)
	case *messagingv1alpha1.Channel:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionTrue, messagingv1alpha1.ReasonAvailable, message, generation)
		o.Status.ObservedGeneration = generation
		applyMQObjectStatusFields(o, opts, message, &now)
		return status.Update(ctx, o)
	case *messagingv1alpha1.ChannelAuthRule:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionTrue, messagingv1alpha1.ReasonAvailable, message, generation)
		o.Status.ObservedGeneration = generation
		applyMQObjectStatusFields(o, opts, message, &now)
		return status.Update(ctx, o)
	case *messagingv1alpha1.AuthorityRecord:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionTrue, messagingv1alpha1.ReasonAvailable, message, generation)
		o.Status.ObservedGeneration = generation
		applyMQObjectStatusFields(o, opts, message, &now)
		return status.Update(ctx, o)
	default:
		return fmt.Errorf("patchSyncedAvailable: unsupported type %T", obj)
	}
}

//nolint:dupl // progressing vs deleting share the same per-kind status patch shape
func patchSyncedDeleting(
	ctx context.Context,
	status client.StatusWriter,
	recorder events.EventRecorder,
	obj client.Object,
	generation int64,
	message string,
) error {
	emitSyncedTransitionEvent(recorder, obj, metav1.ConditionFalse, messagingv1alpha1.ReasonDeleting, message)

	switch o := obj.(type) {
	case *messagingv1alpha1.Queue:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonDeleting, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.Topic:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonDeleting, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.Channel:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonDeleting, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.ChannelAuthRule:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonDeleting, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	case *messagingv1alpha1.AuthorityRecord:
		setCondition(&o.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonDeleting, message, generation)
		applyMQObjectStatusFields(o, syncStatusOpts{}, message, nil)
		return status.Update(ctx, o)
	default:
		return fmt.Errorf("patchSyncedDeleting: unsupported type %T", obj)
	}
}

func connectionRefName(obj client.Object) (string, error) {
	switch o := obj.(type) {
	case *messagingv1alpha1.Queue:
		return o.Spec.ConnectionRef.Name, nil
	case *messagingv1alpha1.Topic:
		return o.Spec.ConnectionRef.Name, nil
	case *messagingv1alpha1.Channel:
		return o.Spec.ConnectionRef.Name, nil
	case *messagingv1alpha1.ChannelAuthRule:
		return o.Spec.ConnectionRef.Name, nil
	case *messagingv1alpha1.AuthorityRecord:
		return o.Spec.ConnectionRef.Name, nil
	default:
		return "", fmt.Errorf("connectionRefName: unsupported type %T", obj)
	}
}

func requestsForConnection(
	ctx context.Context,
	c client.Client,
	conn *messagingv1alpha1.QueueManagerConnection,
) []reconcile.Request {
	var reqs []reconcile.Request
	ns := conn.Namespace
	connName := conn.Name

	queueList := &messagingv1alpha1.QueueList{}
	if err := c.List(ctx, queueList, client.InNamespace(ns)); err == nil {
		for i := range queueList.Items {
			if queueList.Items[i].Spec.ConnectionRef.Name == connName {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{Namespace: ns, Name: queueList.Items[i].Name},
				})
			}
		}
	}

	topicList := &messagingv1alpha1.TopicList{}
	if err := c.List(ctx, topicList, client.InNamespace(ns)); err == nil {
		for i := range topicList.Items {
			if topicList.Items[i].Spec.ConnectionRef.Name == connName {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{Namespace: ns, Name: topicList.Items[i].Name},
				})
			}
		}
	}

	channelList := &messagingv1alpha1.ChannelList{}
	if err := c.List(ctx, channelList, client.InNamespace(ns)); err == nil {
		for i := range channelList.Items {
			if channelList.Items[i].Spec.ConnectionRef.Name == connName {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{Namespace: ns, Name: channelList.Items[i].Name},
				})
			}
		}
	}

	authRuleList := &messagingv1alpha1.ChannelAuthRuleList{}
	if err := c.List(ctx, authRuleList, client.InNamespace(ns)); err == nil {
		for i := range authRuleList.Items {
			if authRuleList.Items[i].Spec.ConnectionRef.Name == connName {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{Namespace: ns, Name: authRuleList.Items[i].Name},
				})
			}
		}
	}

	authRecList := &messagingv1alpha1.AuthorityRecordList{}
	if err := c.List(ctx, authRecList, client.InNamespace(ns)); err == nil {
		for i := range authRecList.Items {
			if authRecList.Items[i].Spec.ConnectionRef.Name == connName {
				reqs = append(reqs, reconcile.Request{
					NamespacedName: types.NamespacedName{Namespace: ns, Name: authRecList.Items[i].Name},
				})
			}
		}
	}

	return reqs
}

func connectionEnqueueMapper(c client.Client) handler.MapFunc {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		conn, ok := obj.(*messagingv1alpha1.QueueManagerConnection)
		if !ok {
			return nil
		}
		return requestsForConnection(ctx, c, conn)
	}
}

func watchConnectionStatus(c client.Client) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(connectionEnqueueMapper(c))
}

func connectionReady(conn *messagingv1alpha1.QueueManagerConnection) bool {
	for _, c := range conn.Status.Conditions {
		if c.Type == messagingv1alpha1.ConditionReady && c.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func connectionReadyChanged(oldConn, newConn *messagingv1alpha1.QueueManagerConnection) bool {
	return connectionReady(oldConn) != connectionReady(newConn)
}

func ignoreNotFound(err error) bool {
	return k8serrors.IsNotFound(err)
}

func connectionWatchPredicates() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			conn, ok := e.Object.(*messagingv1alpha1.QueueManagerConnection)
			return ok && connectionReady(conn)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldConn, okOld := e.ObjectOld.(*messagingv1alpha1.QueueManagerConnection)
			newConn, okNew := e.ObjectNew.(*messagingv1alpha1.QueueManagerConnection)
			if !okOld || !okNew {
				return false
			}
			return connectionReadyChanged(oldConn, newConn) || oldConn.Generation != newConn.Generation
		},
	}
}

func setupMQObjectController(mgr ctrl.Manager, reconciler reconcile.Reconciler, forObj client.Object) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(forObj).
		WithOptions(controllerOptions()).
		Watches(
			&messagingv1alpha1.QueueManagerConnection{},
			watchConnectionStatus(mgr.GetClient()),
			builder.WithPredicates(connectionWatchPredicates()),
		).
		Complete(reconciler)
}
