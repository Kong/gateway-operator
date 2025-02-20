package konnect

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// TODO(pmalek): this can be extracted and used in reconciler.go
// as every Konnect entity will have a reference to the KonnectAPIAuthConfiguration.
// This would require:
// - mapping function from non List types to List types
// - a function on each Konnect entity type to get the API Auth configuration
//   reference from the object
// - lists have their items stored in Items field, not returned via a method

// KonnectCloudGatewayDataPlaneGroupConfigurationReconciliationWatchOptions returns
// the watch options for the KonnectCloudGatewayDataPlaneGroupConfiguration.
func KonnectCloudGatewayDataPlaneGroupConfigurationReconciliationWatchOptions(
	cl client.Client,
) []func(*ctrl.Builder) *ctrl.Builder {
	return []func(*ctrl.Builder) *ctrl.Builder{
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.For(&konnectv1alpha1.KonnectCloudGatewayDataPlaneGroupConfiguration{})
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&konnectv1alpha1.KonnectAPIAuthConfiguration{},
				handler.EnqueueRequestsFromMapFunc(
					enqueueKonnectCloudGatewayDataPlaneGroupConfigurationForKonnectAPIAuthConfiguration(cl),
				),
			)
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&konnectv1alpha1.KonnectGatewayControlPlane{},
				handler.EnqueueRequestsFromMapFunc(
					enqueueObjectForKonnectGatewayControlPlane[konnectv1alpha1.KonnectCloudGatewayDataPlaneGroupConfigurationList](
						cl, IndexFieldKonnectCloudGatewayDataPlaneGroupConfigurationOnKonnectGatewayControlPlane,
					),
				),
			)
		},
	}
}

func enqueueKonnectCloudGatewayDataPlaneGroupConfigurationForKonnectAPIAuthConfiguration(
	cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		auth, ok := obj.(*konnectv1alpha1.KonnectAPIAuthConfiguration)
		if !ok {
			return nil
		}
		var l konnectv1alpha1.KonnectCloudGatewayDataPlaneGroupConfigurationList
		if err := cl.List(ctx, &l,
			// TODO: change this when cross namespace refs are allowed.
			client.InNamespace(auth.GetNamespace()),
			client.MatchingFields{
				IndexFieldKonnectCloudGatewayDataPlaneGroupConfigurationOnAPIAuthConfiguration: auth.Name,
			},
		); err != nil {
			return nil
		}
		return objectListToReconcileRequests(l.Items)
	}
}
