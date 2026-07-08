package validation

import (
	"context"
	"fmt"

	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	messagingv1alpha1 "github.com/conduit-ops/mkurator/api/v1alpha1"
	messagingv1beta1 "github.com/conduit-ops/mkurator/api/v1beta1"
)

// ValidateManagedChannelRef ensures a Channel CR exists in the same namespace with matching
// spec.channelName and spec.connectionRef.name. CHLAUTH rules target channels MKurator manages
// via Channel CRs; pre-existing MQ-only channels are out of scope for this check.
func ValidateManagedChannelRef(
	ctx context.Context,
	reader client.Reader,
	namespace, connectionRefName, channelName string,
	path *field.Path,
) field.ErrorList {
	var errs field.ErrorList
	if channelName == "" {
		return errs
	}

	var channels messagingv1alpha1.ChannelList
	if err := reader.List(ctx, &channels, client.InNamespace(namespace)); err != nil {
		if !k8sruntime.IsNotRegisteredError(err) {
			return field.ErrorList{
				field.InternalError(path, fmt.Errorf("list Channels: %w", err)),
			}
		}
	}

	var match *messagingv1alpha1.Channel
	for i := range channels.Items {
		ch := &channels.Items[i]
		if ch.Spec.ChannelName != channelName {
			continue
		}
		if connectionRefName != "" && ch.Spec.ConnectionRef.Name != connectionRefName {
			continue
		}
		match = ch
		break
	}
	if match == nil {
		var channelsV1Beta1 messagingv1beta1.ChannelList
		if err := reader.List(ctx, &channelsV1Beta1, client.InNamespace(namespace)); err != nil {
			if !k8sruntime.IsNotRegisteredError(err) {
				return field.ErrorList{
					field.InternalError(path, fmt.Errorf("list Channels (v1beta1): %w", err)),
				}
			}
		}
		for i := range channelsV1Beta1.Items {
			ch := &channelsV1Beta1.Items[i]
			if ch.Spec.ChannelName != channelName {
				continue
			}
			if connectionRefName != "" && ch.Spec.ConnectionRef.Name != connectionRefName {
				continue
			}
			if ch.DeletionTimestamp != nil {
				return field.ErrorList{
					field.Invalid(path, channelName, fmt.Sprintf(
						"Channel %q is deleting; wait for deletion to finish or point channelName at another Channel",
						ch.Name,
					)),
				}
			}
			return errs
		}
	}
	if match == nil {
		return field.ErrorList{
			field.NotFound(path, fmt.Sprintf(
				"Channel with channelName %q and connectionRef %q not found in namespace %q; create a Channel CR first",
				channelName, connectionRefName, namespace,
			)),
		}
	}
	if match.DeletionTimestamp != nil {
		return field.ErrorList{
			field.Invalid(path, channelName, fmt.Sprintf(
				"Channel %q is deleting; wait for deletion to finish or point channelName at another Channel",
				match.Name,
			)),
		}
	}
	return errs
}
