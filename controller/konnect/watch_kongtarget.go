package konnect

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorerrors "github.com/kong/gateway-operator/internal/errors"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

// KongTargetReconciliationWatchOptions  returns the watch options for
// the KongTarget.
func KongTargetReconciliationWatchOptions(cl client.Client,
) []func(*ctrl.Builder) *ctrl.Builder {
	return []func(*ctrl.Builder) *ctrl.Builder{
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.For(
				&configurationv1alpha1.KongTarget{},
				builder.WithPredicates(
					predicate.NewPredicateFuncs(kongTargetRefersToKonnectGatewayControlPlane(cl)),
				),
			)
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&configurationv1alpha1.KongUpstream{},
				handler.EnqueueRequestsFromMapFunc(enqueueKongTargetForKongUpstream(cl)),
			)
		},
	}
}

// kongTargetRefersToKonnectGatewayControlPlane returns the predict
// that checks whether a KongTarget is referring a Konnect Control Plane via upstream.
func kongTargetRefersToKonnectGatewayControlPlane(cl client.Client) func(obj client.Object) bool {
	return func(obj client.Object) bool {
		kongTarget, ok := obj.(*configurationv1alpha1.KongTarget)
		if !ok {
			ctrllog.FromContext(context.Background()).Error(
				operatorerrors.ErrUnexpectedObject,
				"failed to run predicate function",
				"expected", "KongTarget", "found", reflect.TypeOf(obj),
			)
			return false
		}

		upstream := configurationv1alpha1.KongUpstream{}
		nn := types.NamespacedName{
			Namespace: kongTarget.Namespace,
			Name:      kongTarget.Spec.UpstreamRef.Name,
		}
		if err := cl.Get(context.Background(), nn, &upstream); client.IgnoreNotFound(err) != nil {
			return true
		}
		cpRef := upstream.Spec.ControlPlaneRef
		return cpRef != nil && cpRef.Type == configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef
	}
}

func enqueueKongTargetForKongUpstream(cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		kongUpstream, ok := obj.(*configurationv1alpha1.KongUpstream)
		if !ok {
			return nil
		}
		cpRef := kongUpstream.Spec.ControlPlaneRef
		if cpRef == nil || cpRef.Type != configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef {
			return nil
		}
		var targetList configurationv1alpha1.KongTargetList
		if err := cl.List(ctx, &targetList, &client.ListOptions{
			// TODO: change this when cross namespace refs are allowed.
			Namespace: kongUpstream.GetNamespace(),
		}); err != nil {
			return nil
		}

		var ret []reconcile.Request
		for _, target := range targetList.Items {
			if target.Spec.UpstreamRef.Name == kongUpstream.Name {
				ret = append(ret, reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&target),
				})
			}
		}
		return ret
	}
}