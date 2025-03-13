package konnect

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect/constraints"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func getCPAuthRefForRef(
	ctx context.Context,
	cl client.Client,
	cpRef commonv1alpha1.ControlPlaneRef,
	namespace string,
) (types.NamespacedName, error) {
	cp, err := getCPForRef(ctx, cl, cpRef, namespace)
	if err != nil {
		return types.NamespacedName{}, err
	}

	return types.NamespacedName{
		Name: cp.GetKonnectAPIAuthConfigurationRef().Name,
		// TODO(pmalek): enable if cross namespace refs are allowed
		Namespace: cp.GetNamespace(),
	}, nil
}

func getAPIAuthRefNN[T constraints.SupportedKonnectEntityType, TEnt constraints.EntityType[T]](
	ctx context.Context,
	cl client.Client,
	ent TEnt,
) (types.NamespacedName, error) {
	// If the entity has a KonnectAPIAuthConfigurationRef, return it.
	if ref, ok := any(ent).(constraints.EntityWithKonnectAPIAuthConfigurationRef); ok {
		return types.NamespacedName{
			Name: ref.GetKonnectAPIAuthConfigurationRef().Name,
			// TODO: enable if cross namespace refs are allowed
			Namespace: ent.GetNamespace(),
		}, nil
	}

	// If the entity has a ControlPlaneRef, get the KonnectAPIAuthConfiguration
	// ref from the referenced ControlPlane.
	cpRef, ok := getControlPlaneRef(ent).Get()
	if ok {
		cp, err := getCPForRef(ctx, cl, cpRef, ent.GetNamespace())
		if err != nil {
			return types.NamespacedName{}, fmt.Errorf("failed to get ControlPlane for %s: %w", client.ObjectKeyFromObject(ent), err)
		}

		cpNamespace := ent.GetNamespace()
		if ent.GetNamespace() == "" && cp.GetNamespace() != "" {
			cpNamespace = cp.GetNamespace()
		}
		return getCPAuthRefForRef(ctx, cl, cpRef, cpNamespace)
	}

	// If the entity has a KongServiceRef, get the KonnectAPIAuthConfiguration
	// ref from the referenced KongService.
	svcRef, ok := getServiceRef(ent).Get()
	if ok {
		if svcRef.Type != configurationv1alpha1.ServiceRefNamespacedRef {
			return types.NamespacedName{}, fmt.Errorf("unsupported KongService ref type %q", svcRef.Type)
		}
		// TODO(pmalek): handle cross namespace refs
		nn := types.NamespacedName{
			Name:      svcRef.NamespacedRef.Name,
			Namespace: ent.GetNamespace(),
		}

		var svc configurationv1alpha1.KongService
		if err := cl.Get(ctx, nn, &svc); err != nil {
			return types.NamespacedName{}, fmt.Errorf("failed to get KongService %s", nn)
		}

		cpRef, ok := getControlPlaneRef(&svc).Get()
		if !ok {
			return types.NamespacedName{}, fmt.Errorf("KongService %s does not have a ControlPlaneRef", nn)
		}
		return getCPAuthRefForRef(ctx, cl, cpRef, ent.GetNamespace())
	}

	// If the entity has a KongConsumerRef, get the KonnectAPIAuthConfiguration
	// ref from the referenced KongConsumer.
	consumerRef, ok := getConsumerRef(ent).Get()
	if ok {
		// TODO(pmalek): handle cross namespace refs
		nn := types.NamespacedName{
			Name:      consumerRef.Name,
			Namespace: ent.GetNamespace(),
		}

		var consumer configurationv1.KongConsumer
		if err := cl.Get(ctx, nn, &consumer); err != nil {
			return types.NamespacedName{}, fmt.Errorf("failed to get KongConsumer %s", nn)
		}

		cpRef, ok := getControlPlaneRef(&consumer).Get()
		if !ok {
			return types.NamespacedName{}, fmt.Errorf("KongConsumer %s does not have a ControlPlaneRef", nn)
		}
		return getCPAuthRefForRef(ctx, cl, cpRef, ent.GetNamespace())
	}

	// If the entity has a KongUpstreamRef, get the KonnectAPIAuthConfiguration
	// ref from the referenced KongUpstream.
	upstreamRef, ok := getKongUpstreamRef(ent).Get()
	if ok {
		nn := types.NamespacedName{
			Name:      upstreamRef.Name,
			Namespace: ent.GetNamespace(),
		}

		var upstream configurationv1alpha1.KongUpstream
		if err := cl.Get(ctx, nn, &upstream); err != nil {
			return types.NamespacedName{}, fmt.Errorf("failed to get KongUpstream %s", nn)
		}

		cpRef, ok := getControlPlaneRef(&upstream).Get()
		if !ok {
			return types.NamespacedName{}, fmt.Errorf("KongUpstream %s does not have a ControlPlaneRef", nn)
		}
		return getCPAuthRefForRef(ctx, cl, cpRef, ent.GetNamespace())
	}

	// If the entity has a KongCertificateRef, get the KonnectAPIAuthConfiguration
	// ref from the referenced KongUpstream.
	certificateRef, ok := getKongCertificateRef(ent).Get()
	if ok {
		nn := types.NamespacedName{
			Name:      certificateRef.Name,
			Namespace: ent.GetNamespace(),
		}

		var cert configurationv1alpha1.KongCertificate
		if err := cl.Get(ctx, nn, &cert); err != nil {
			return types.NamespacedName{}, fmt.Errorf("failed to get KongCertificate %s", nn)
		}

		cpRef, ok := getControlPlaneRef(&cert).Get()
		if !ok {
			return types.NamespacedName{}, fmt.Errorf("KongCertificate %s does not have a ControlPlaneRef", nn)
		}
		return getCPAuthRefForRef(ctx, cl, cpRef, ent.GetNamespace())
	}

	return types.NamespacedName{}, fmt.Errorf(
		"cannot get KonnectAPIAuthConfiguration for entity type %T %s",
		client.ObjectKeyFromObject(ent), ent,
	)
}
