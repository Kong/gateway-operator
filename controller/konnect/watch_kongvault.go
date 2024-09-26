package konnect

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kong/gateway-operator/modules/manager/logging"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// KongVaultReconciliationWatchOptions returns the watch options for KongVault.
func KongVaultReconciliationWatchOptions(cl client.Client) []func(*ctrl.Builder) *ctrl.Builder {
	return []func(*ctrl.Builder) *ctrl.Builder{
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.For(&configurationv1alpha1.KongVault{},
				builder.WithPredicates(
					predicate.NewPredicateFuncs(objRefersToKonnectGatewayControlPlane[configurationv1alpha1.KongVault]),
				),
			)
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&konnectv1alpha1.KonnectAPIAuthConfiguration{},
				handler.EnqueueRequestsFromMapFunc(
					enqueueKongVaultForKonnectAPIAuthConfiguration(cl),
				),
			)
		},
		func(b *ctrl.Builder) *ctrl.Builder {
			return b.Watches(
				&konnectv1alpha1.KonnectGatewayControlPlane{},
				handler.EnqueueRequestsFromMapFunc(
					enqueueKongVaultForKonnectGatewayControlPlane(cl),
				),
			)
		},
	}
}

// enqueueKongVaultForKonnectAPIAuthConfiguration enqueues KongVaults
// when KonnectAPIAuthConfiguration which is associated with the Konnect Control plane referenced by the KongVault.
func enqueueKongVaultForKonnectAPIAuthConfiguration(
	cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		auth, ok := obj.(*konnectv1alpha1.KonnectAPIAuthConfiguration)
		if !ok {
			return nil
		}

		l := configurationv1alpha1.KongVaultList{}

		if err := cl.List(ctx, &l); err != nil {
			return nil
		}

		var ret []reconcile.Request
		for _, vault := range l.Items {
			cpRef, ok := getControlPlaneRef(&vault).Get()
			if !ok {
				continue
			}
			switch cpRef.Type {
			case configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef:
				// Need to get namespace from controlPlaneRef because KongVault is cluster scoped.
				nn := types.NamespacedName{
					Name:      cpRef.KonnectNamespacedRef.Name,
					Namespace: cpRef.KonnectNamespacedRef.Namespace,
				}
				// TODO: change this when cross namespace refs are allowed.
				if nn.Namespace != auth.Namespace {
					continue
				}
				var cp konnectv1alpha1.KonnectGatewayControlPlane
				if err := cl.Get(ctx, nn, &cp); err != nil {
					ctrllog.FromContext(ctx).Error(
						err,
						"failed to get KonnectGatewayControlPlane",
						"KonnectGatewayControlPlane", nn,
					)
					continue
				}

				if cp.GetKonnectAPIAuthConfigurationRef().Name != auth.Name {
					continue
				}

				// Append the KongVault to reconcile request list when the controlPlaneRef of the KongVault is pointing to the control plane
				// which references the affected API auth configuration.
				ret = append(ret, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name: vault.Name,
					},
				})

			case configurationv1alpha1.ControlPlaneRefKonnectID:
				ctrllog.FromContext(ctx).Error(
					fmt.Errorf("unimplemented ControlPlaneRef type %q", cpRef.Type),
					"unimplemented ControlPlaneRef for KongVault",
					"KongVault", vault, "refType", cpRef.Type,
				)
				continue

			default:
				ctrllog.FromContext(ctx).V(logging.DebugLevel.Value()).Info(
					"unsupported ControlPlaneRef for KongVault",
					"KongVault", vault, "refType", cpRef.Type,
				)
				continue
			}

		}
		return ret
	}
}

func enqueueKongVaultForKonnectGatewayControlPlane(
	cl client.Client,
) func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		cp, ok := obj.(*konnectv1alpha1.KonnectGatewayControlPlane)
		if !ok {
			return nil
		}

		l := configurationv1alpha1.KongVaultList{}

		if err := cl.List(ctx, &l); err != nil {
			return nil
		}
		var ret []reconcile.Request
		for _, vault := range l.Items {
			cpRef, ok := getControlPlaneRef(&vault).Get()
			if !ok {
				continue
			}
			switch cpRef.Type {
			case configurationv1alpha1.ControlPlaneRefKonnectNamespacedRef:
				// Need to check namespace in controlPlaneRef because KongVault is cluster scoped.
				if cp.Namespace != vault.Spec.ControlPlaneRef.KonnectNamespacedRef.Namespace ||
					cp.Name != vault.Spec.ControlPlaneRef.KonnectNamespacedRef.Name {
					continue
				}

				// Append the KongVault to reconcile request list when the controlPlaneRef of the KongVault is pointing to the control plane.
				ret = append(ret, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name: vault.Name,
					},
				})

			case configurationv1alpha1.ControlPlaneRefKonnectID:
				ctrllog.FromContext(ctx).Error(
					fmt.Errorf("unimplemented ControlPlaneRef type %q", cpRef.Type),
					"unimplemented ControlPlaneRef for KongVault",
					"KongVault", vault, "refType", cpRef.Type,
				)
				continue

			default:
				ctrllog.FromContext(ctx).V(logging.DebugLevel.Value()).Info(
					"unsupported ControlPlaneRef for KongVault",
					"KongVault", vault, "refType", cpRef.Type,
				)
				continue
			}

		}
		return ret
	}
}
