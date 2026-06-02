package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	messagingv1alpha1 "github.com/konradheimel/kurator/api/v1alpha1"
	"github.com/konradheimel/kurator/internal/metrics"
	"github.com/konradheimel/kurator/internal/mqadmin"
)

// ChannelReconciler reconciles Channel objects into MQSC on IBM MQ.
type ChannelReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	MQFactory mqadmin.Factory
}

// +kubebuilder:rbac:groups=messaging.kurator.dev,resources=channels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=messaging.kurator.dev,resources=channels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=messaging.kurator.dev,resources=channels/finalizers,verbs=update
// +kubebuilder:rbac:groups=messaging.kurator.dev,resources=queuemanagerconnections,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile ensures the MQ channel matches spec.
func (r *ChannelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	result, err := r.reconcile(ctx, req)
	metrics.RecordReconcile(metrics.ControllerChannel, err)
	return result, err
}

func (r *ChannelReconciler) reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	channel := &messagingv1alpha1.Channel{}
	if err := r.Get(ctx, req.NamespacedName, channel); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("get Channel: %w", err)
	}

	conn, err := r.getConnection(ctx, channel)
	if err != nil {
		return r.setSyncedError(ctx, channel, err)
	}

	if !connectionReady(conn) {
		setCondition(&channel.Status.Conditions, messagingv1alpha1.ConditionSynced,
			metav1.ConditionFalse, messagingv1alpha1.ReasonProgressing,
			fmt.Sprintf("waiting for connection %q to become Ready", conn.Name), channel.Generation)
		if statusErr := r.Status().Update(ctx, channel); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	admin, err := r.MQFactory.ForConnection(ctx, conn)
	if err != nil {
		return r.setSyncedError(ctx, channel, err)
	}

	if !channel.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, channel, admin)
	}

	if !controllerutil.ContainsFinalizer(channel, messagingv1alpha1.ChannelFinalizer) {
		controllerutil.AddFinalizer(channel, messagingv1alpha1.ChannelFinalizer)
		if err := r.Update(ctx, channel); err != nil {
			return ctrl.Result{}, fmt.Errorf("add finalizer: %w", err)
		}
		return ctrl.Result{}, nil
	}

	if channel.Spec.Type != "" && channel.Spec.Type != messagingv1alpha1.ChannelTypeSvrconn {
		return r.setSyncedError(ctx, channel, &mqadmin.TerminalError{
			Reason:  "UnsupportedChannelType",
			Message: fmt.Sprintf("channel type %q is not supported in v1alpha1", channel.Spec.Type),
		})
	}

	spec := toMQChannelSpec(channel)
	if err := r.ensureChannel(ctx, admin, spec); err != nil {
		return r.setSyncedError(ctx, channel, err)
	}

	setCondition(&channel.Status.Conditions, messagingv1alpha1.ConditionSynced,
		metav1.ConditionTrue, messagingv1alpha1.ReasonAvailable, "Channel matches spec", channel.Generation)
	channel.Status.ObservedGeneration = channel.Generation
	if err := r.Status().Update(ctx, channel); err != nil {
		return ctrl.Result{}, fmt.Errorf("update status: %w", err)
	}
	logger.Info("Channel synced", "channel", channel.Spec.ChannelName, "type", spec.Type)
	return ctrl.Result{}, nil
}

func (r *ChannelReconciler) getConnection(
	ctx context.Context,
	channel *messagingv1alpha1.Channel,
) (*messagingv1alpha1.QueueManagerConnection, error) {
	conn := &messagingv1alpha1.QueueManagerConnection{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: channel.Namespace,
		Name:      channel.Spec.ConnectionRef.Name,
	}, conn); err != nil {
		return nil, fmt.Errorf("get connection %q: %w", channel.Spec.ConnectionRef.Name, err)
	}
	return conn, nil
}

func (r *ChannelReconciler) ensureChannel(ctx context.Context, admin mqadmin.Admin, spec mqadmin.ChannelSpec) error {
	observed, err := admin.GetChannel(ctx, spec)
	if err != nil && !errors.Is(err, mqadmin.ErrNotFound) {
		return err
	}
	if observed == nil || channelNeedsUpdate(spec, observed) {
		if err := admin.DefineChannel(ctx, spec); err != nil {
			return err
		}
	}
	return nil
}

func channelNeedsUpdate(desired mqadmin.ChannelSpec, observed *mqadmin.ChannelState) bool {
	for k, v := range desired.Attributes {
		key := strings.ToLower(k)
		if observed.Attributes[key] != v {
			return true
		}
	}
	return false
}

func (r *ChannelReconciler) handleDeletion(
	ctx context.Context,
	channel *messagingv1alpha1.Channel,
	admin mqadmin.Admin,
) (ctrl.Result, error) {
	setCondition(&channel.Status.Conditions, messagingv1alpha1.ConditionSynced,
		metav1.ConditionFalse, messagingv1alpha1.ReasonDeleting, "Deleting channel from IBM MQ", channel.Generation)
	if err := r.Status().Update(ctx, channel); err != nil {
		return ctrl.Result{}, err
	}

	spec := toMQChannelSpec(channel)
	if err := admin.DeleteChannel(ctx, spec); err != nil {
		return r.setSyncedError(ctx, channel, err)
	}

	controllerutil.RemoveFinalizer(channel, messagingv1alpha1.ChannelFinalizer)
	if err := r.Update(ctx, channel); err != nil {
		return ctrl.Result{}, fmt.Errorf("remove finalizer: %w", err)
	}
	return ctrl.Result{}, nil
}

func (r *ChannelReconciler) setSyncedError(
	ctx context.Context,
	channel *messagingv1alpha1.Channel,
	err error,
) (ctrl.Result, error) {
	reason := messagingv1alpha1.ReasonError
	requeue := ctrl.Result{}
	if errors.Is(err, mqadmin.ErrTransient) {
		requeue = ctrl.Result{RequeueAfter: 30 * time.Second}
	}
	setCondition(&channel.Status.Conditions, messagingv1alpha1.ConditionSynced,
		metav1.ConditionFalse, reason, err.Error(), channel.Generation)
	if statusErr := r.Status().Update(ctx, channel); statusErr != nil {
		return requeue, statusErr
	}
	if errors.Is(err, mqadmin.ErrTransient) {
		return requeue, err
	}
	return ctrl.Result{}, nil
}

func toMQChannelSpec(channel *messagingv1alpha1.Channel) mqadmin.ChannelSpec {
	attrs := map[string]string{}
	for k, v := range channel.Spec.Attributes {
		attrs[strings.ToLower(k)] = v
	}
	chType := mqadmin.ChannelTypeSvrconn
	if channel.Spec.Type != "" {
		chType = mqadmin.ChannelType(channel.Spec.Type)
	}
	return mqadmin.ChannelSpec{
		Name:       channel.Spec.ChannelName,
		Type:       chType,
		Attributes: attrs,
	}
}

// SetupWithManager wires the reconciler.
func (r *ChannelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&messagingv1alpha1.Channel{}).
		Complete(r)
}
