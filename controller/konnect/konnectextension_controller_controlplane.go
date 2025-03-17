package konnect

import (
	"context"

	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func enqueueKonnectExtensionsForKonnectGatewayControlPlane(cl client.Client) func(context.Context, client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		cp, ok := obj.(*konnectv1alpha1.KonnectGatewayControlPlane)
		if !ok {
			return nil
		}

		konnectExtensionList := konnectv1alpha1.KonnectExtensionList{}
		if err := cl.List(
			ctx,
			&konnectExtensionList,
			client.InNamespace(cp.Namespace),
			client.MatchingFields{
				IndexFieldKonnectExtensionOnKonnectGatewayControlPlane: cp.Name,
			},
		); err != nil {
			return nil
		}

		return lo.Map(konnectExtensionList.Items, func(ext konnectv1alpha1.KonnectExtension, _ int) reconcile.Request {
			return reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&ext),
			}
		})
	}
}
