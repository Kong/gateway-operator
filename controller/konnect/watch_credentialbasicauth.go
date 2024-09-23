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

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// TODO(pmalek): this can be extracted and used in reconciler.go
// as every Konnect entity will have a reference to the KonnectAPIAuthConfiguration.
// This would require:
// - mapping function from non List types to List types
// - a function on each Konnect entity type to get the API Auth configuration
//   reference from the object
// - lists have their items stored in Items field, not returned via a method

// kongCredentialBasicAuthReconciliationWatchOptions returns the watch options for
// the KongCredentialBasicAuth.
func kongCredentialBasicAuthReconciliationWatchOptions(
	cl client.Client,
) []func(*ctrl.Builder) *ctrl.Builder {
	return []func(*ctrl.Builder) *ctrl.Builder{
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.For(&configurationv1alpha1.KongCredentialBasicAuth{},
				builder.WithPredicates(
					predicate.NewPredicateFuncs(kongCredentialBasicAuthRefersToKonnectGatewayControlPlane(cl)),
				),
			)
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&configurationv1.KongConsumer{},
				handler.EnqueueRequestsFromMapFunc(
					kongCredentialBasicAuthForKongConsumer(cl),
				),
			)
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&konnectv1alpha1.KonnectAPIAuthConfiguration{},
				handler.EnqueueRequestsFromMapFunc(
					kongCredentialBasicAuthForKonnectAPIAuthConfiguration(cl),
				),
			)
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&konnectv1alpha1.KonnectGatewayControlPlane{},
				handler.EnqueueRequestsFromMapFunc(
					kongCredentialBasicAuthForKonnectGatewayControlPlane(cl),
				),
			)
		},
	}
}

// kongCredentialBasicAuthRefersToKonnectGatewayControlPlane returns true if the KongCredentialBasicAuth
// refers to a KonnectGatewayControlPlane.
func kongCredentialBasicAuthRefersToKonnectGatewayControlPlane(cl client.Client) func(obj client.Object) bool {
	return func(obj client.Object) bool {
		KongCredentialBasicAuth, ok := obj.(*configurationv1alpha1.KongCredentialBasicAuth)
		if !ok {
			ctrllog.FromContext(context.Background()).Error(
				operatorerrors.ErrUnexpectedObject,
				"failed to run predicate function",
				"expected", "KongCredentialBasicAuth", "found", reflect.TypeOf(obj),
			)
			return false
		}

		consumerRef := KongCredentialBasicAuth.Spec.ConsumerRef
		nn := types.NamespacedName{
			Namespace: KongCredentialBasicAuth.Namespace,
			Name:      consumerRef.Name,
		}
		consumer := configurationv1.KongConsumer{}
		if err := cl.Get(context.Background(), nn, &consumer); client.IgnoreNotFound(err) != nil {
			return true
		}

		cpRef := consumer.Spec.ControlPlaneRef
		return cpRef != nil && cpRef.Type == configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef
	}
}

func kongCredentialBasicAuthForKonnectAPIAuthConfiguration(
	cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		auth, ok := obj.(*konnectv1alpha1.KonnectAPIAuthConfiguration)
		if !ok {
			return nil
		}

		var l configurationv1.KongConsumerList
		if err := cl.List(ctx, &l,
			// TODO: change this when cross namespace refs are allowed.
			client.InNamespace(auth.GetNamespace()),
		); err != nil {
			return nil
		}

		var ret []reconcile.Request
		for _, consumer := range l.Items {
			cpRef := consumer.Spec.ControlPlaneRef
			if cpRef == nil ||
				cpRef.Type != configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef ||
				cpRef.KonnectNamespacedRef == nil ||
				cpRef.KonnectNamespacedRef.Name != auth.GetName() {
				continue
			}

			cpNN := types.NamespacedName{
				Name:      cpRef.KonnectNamespacedRef.Name,
				Namespace: consumer.Namespace,
			}
			var cp konnectv1alpha1.KonnectGatewayControlPlane
			if err := cl.Get(ctx, cpNN, &cp); err != nil {
				ctrllog.FromContext(ctx).Error(
					err,
					"failed to get KonnectGatewayControlPlane",
					"KonnectGatewayControlPlane", cpNN,
				)
				continue
			}

			// TODO: change this when cross namespace refs are allowed.
			if cp.GetKonnectAPIAuthConfigurationRef().Name != auth.Name {
				continue
			}

			var credList configurationv1alpha1.KongCredentialBasicAuthList
			if err := cl.List(ctx, &credList,
				client.MatchingFields{
					IndexFieldKongCredentialBasicAuthReferencesKongConsumer: consumer.Name,
				},
				client.InNamespace(auth.GetNamespace()),
			); err != nil {
				return nil
			}

			for _, cred := range credList.Items {
				ret = append(ret, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: cred.Namespace,
						Name:      cred.Name,
					},
				},
				)
			}
		}
		return ret
	}
}

func kongCredentialBasicAuthForKonnectGatewayControlPlane(
	cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		cp, ok := obj.(*konnectv1alpha1.KonnectGatewayControlPlane)
		if !ok {
			return nil
		}
		var l configurationv1.KongConsumerList
		if err := cl.List(ctx, &l,
			// TODO: change this when cross namespace refs are allowed.
			client.InNamespace(cp.GetNamespace()),
		); err != nil {
			return nil
		}

		var ret []reconcile.Request
		for _, consumer := range l.Items {
			cpRef := consumer.Spec.ControlPlaneRef
			if cpRef.Type != configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef ||
				cpRef.KonnectNamespacedRef == nil ||
				cpRef.KonnectNamespacedRef.Name != cp.GetName() {
				continue
			}

			var credList configurationv1alpha1.KongCredentialBasicAuthList
			if err := cl.List(ctx, &credList,
				client.MatchingFields{
					IndexFieldKongCredentialBasicAuthReferencesKongConsumer: consumer.Name,
				},
				client.InNamespace(cp.GetNamespace()),
			); err != nil {
				return nil
			}

			for _, cred := range credList.Items {
				ret = append(ret, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: cred.Namespace,
						Name:      cred.Name,
					},
				},
				)
			}
		}
		return ret
	}
}

func kongCredentialBasicAuthForKongConsumer(
	cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		consumer, ok := obj.(*configurationv1.KongConsumer)
		if !ok {
			return nil
		}
		var l configurationv1alpha1.KongCredentialBasicAuthList
		if err := cl.List(ctx, &l,
			client.MatchingFields{
				IndexFieldKongCredentialBasicAuthReferencesKongConsumer: consumer.Name,
			},
			// TODO: change this when cross namespace refs are allowed.
			client.InNamespace(consumer.GetNamespace()),
		); err != nil {
			return nil
		}

		var ret []reconcile.Request
		for _, cred := range l.Items {
			ret = append(ret, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: cred.Namespace,
					Name:      cred.Name,
				},
			},
			)
		}
		return ret
	}
}
