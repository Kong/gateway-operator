package konnect

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kong/gateway-operator/controller/konnect/constraints"
	operatorerrors "github.com/kong/gateway-operator/internal/errors"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	configurationv1beta1 "github.com/kong/kubernetes-configuration/api/configuration/v1beta1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// ReconciliationWatchOptionsForEntity returns the watch options for the given
// Konnect entity type.
func ReconciliationWatchOptionsForEntity[
	T constraints.SupportedKonnectEntityType,
	TEnt constraints.EntityType[T],
](
	cl client.Client,
	ent TEnt,
) []func(*ctrl.Builder) *ctrl.Builder {
	switch any(ent).(type) {
	case *configurationv1beta1.KongConsumerGroup:
		return KongConsumerGroupReconciliationWatchOptions(cl)
	case *configurationv1.KongConsumer:
		return KongConsumerReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongRoute:
		return KongRouteReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongService:
		return KongServiceReconciliationWatchOptions(cl)
	case *konnectv1alpha1.KonnectGatewayControlPlane:
		return KonnectGatewayControlPlaneReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongPluginBinding:
		return KongPluginBindingReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongUpstream:
		return KongUpstreamReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongCredentialBasicAuth:
		return kongCredentialBasicAuthReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongCredentialAPIKey:
		return kongCredentialAPIKeyReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongCredentialACL:
		return kongCredentialACLReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongCredentialJWT:
		return kongCredentialJWTReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongCredentialHMAC:
		return kongCredentialHMACReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongCACertificate:
		return KongCACertificateReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongCertificate:
		return KongCertificateReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongTarget:
		return KongTargetReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongVault:
		return KongVaultReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongKey:
		return KongKeyReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongKeySet:
		return KongKeySetReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongSNI:
		return KongSNIReconciliationWatchOptions(cl)
	case *configurationv1alpha1.KongDataPlaneClientCertificate:
		return KongDataPlaneClientCertificateReconciliationWatchOptions(cl)
	default:
		panic(fmt.Sprintf("unsupported entity type %T", ent))
	}
}

// objRefersToKonnectGatewayControlPlane returns true if the object
// refers to a KonnectGatewayControlPlane.
func objRefersToKonnectGatewayControlPlane[
	T constraints.SupportedKonnectEntityType,
	TEnt constraints.EntityType[T],
](obj client.Object) bool {
	ent, ok := obj.(TEnt)
	if !ok {
		ctrllog.FromContext(context.Background()).Error(
			operatorerrors.ErrUnexpectedObject,
			"failed to run predicate function",
			"expected", constraints.EntityTypeName[T](), "found", reflect.TypeOf(obj),
		)
		return false
	}

	return objHasControlPlaneRefKonnectNamespacedRef(ent)
}

func objHasControlPlaneRefKonnectNamespacedRef[
	T constraints.SupportedKonnectEntityType,
	TEnt constraints.EntityType[T],
](ent TEnt) bool {
	_, ok := controlPlaneRefIsKonnectNamespacedRef(ent)
	return ok
}

// controlPlaneRefIsKonnectNamespacedRef returns:
// - the ControlPlane KonnectNamespacedRef of the object if it is a KonnectNamespacedRef.
// - a boolean indicating if the object has a KonnectNamespacedRef.
func controlPlaneRefIsKonnectNamespacedRef[
	T constraints.SupportedKonnectEntityType,
	TEnt constraints.EntityType[T],
](ent TEnt) (configurationv1alpha1.ControlPlaneRef, bool) {
	cpRef, ok := getControlPlaneRef(ent).Get()
	if !ok {
		return configurationv1alpha1.ControlPlaneRef{}, false
	}
	return cpRef, cpRef.KonnectNamespacedRef != nil &&
		cpRef.Type == configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef
}

// objectListToReconcileRequests converts a list of objects to a list of reconcile requests.
func objectListToReconcileRequests[
	T constraints.SupportedKonnectEntityType,
	TPtr constraints.EntityType[T],
](
	items []T,
	filters ...func(TPtr) bool,
) []ctrl.Request {
	ret := make([]ctrl.Request, 0, len(items))
	for _, item := range items {
		var e TPtr = &item
		for _, filter := range filters {
			if filter != nil && !filter(e) {
				continue
			}
		}
		ret = append(ret, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: e.GetNamespace(),
				Name:      e.GetName(),
			},
		})
	}

	return ret
}

// enqueueObjectForKonnectGatewayControlPlane returns a function that enqueues
// reconcile requests for objects matching the provided list type, so for example
// providing KongConsumerList, this function will enqueue reconcile requests for
// KongConsumers that refer to the KonnectGatewayControlPlane that was provided
// as the object.
func enqueueObjectForKonnectGatewayControlPlane[
	TList interface {
		client.ObjectList
		GetItems() []T
	},
	T constraints.SupportedKonnectEntityType,
	TT constraints.EntityType[T],
](
	cl client.Client,
	index string,
) func(context.Context, client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		cp, ok := obj.(*konnectv1alpha1.KonnectGatewayControlPlane)
		if !ok {
			return nil
		}
		var l TList
		if err := cl.List(ctx, l,
			// TODO: change this when cross namespace refs are allowed.
			client.InNamespace(cp.GetNamespace()),
			client.MatchingFields{
				index: cp.GetNamespace() + "/" + cp.GetName(),
			},
		); err != nil {
			return nil
		}

		return objectListToReconcileRequests[T, TT](l.GetItems())
	}
}
