package konnect

import (
	"context"
	"errors"
	"fmt"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectgoops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func createRoute(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	route *configurationv1alpha1.KongRoute,
) error {
	resp, err := sdk.Routes.CreateRoute(ctx, route.Status.Konnect.ControlPlaneID, sdkkonnectgocomp.RouteInput{
		Destinations:            route.Spec.KongRouteAPISpec.Destinations,
		Headers:                 route.Spec.KongRouteAPISpec.Headers,
		Hosts:                   route.Spec.KongRouteAPISpec.Hosts,
		HTTPSRedirectStatusCode: route.Spec.KongRouteAPISpec.HTTPSRedirectStatusCode,
		Methods:                 route.Spec.KongRouteAPISpec.Methods,
		Name:                    route.Spec.KongRouteAPISpec.Name,
		PathHandling:            route.Spec.KongRouteAPISpec.PathHandling,
		Paths:                   route.Spec.KongRouteAPISpec.Paths,
		PreserveHost:            route.Spec.KongRouteAPISpec.PreserveHost,
		Protocols:               route.Spec.KongRouteAPISpec.Protocols,
		RegexPriority:           route.Spec.KongRouteAPISpec.RegexPriority,
		RequestBuffering:        route.Spec.KongRouteAPISpec.RequestBuffering,
		ResponseBuffering:       route.Spec.KongRouteAPISpec.ResponseBuffering,
		Snis:                    route.Spec.KongRouteAPISpec.Snis,
		Sources:                 route.Spec.KongRouteAPISpec.Sources,
		StripPath:               route.Spec.KongRouteAPISpec.StripPath,
		Tags:                    route.Spec.KongRouteAPISpec.Tags,
		Service: &sdkkonnectgocomp.RouteService{
			ID: sdkkonnectgo.String(route.Status.Konnect.ServiceID),
		},
	})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errHandled := handleResp[configurationv1alpha1.KongRoute](err, resp, CreateOp); errHandled != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				errHandled.Error(),
				route.GetGeneration(),
			),
			route,
		)
		return errHandled
	}

	route.GetKonnectStatus().SetKonnectID(*resp.Route.ID)
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			route.GetGeneration(),
		),
		route,
	)

	return nil
}

func updateRoute(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	route *configurationv1alpha1.KongRoute,
) error {
	// TODO(pmalek) handle other types of CP ref
	nnCP := types.NamespacedName{
		Namespace: route.Spec.ControlPlaneRef.KonnectNamespacedRef.Namespace,
		Name:      route.Spec.ControlPlaneRef.KonnectNamespacedRef.Name,
	}
	if nnCP.Namespace == "" {
		nnCP.Namespace = route.Namespace
	}
	var cp konnectv1alpha1.KonnectControlPlane
	if err := cl.Get(ctx, nnCP, &cp); err != nil {
		return fmt.Errorf("failed to get KonnectControlPlane %s: for KongRoute %s: %w",
			nnCP, client.ObjectKeyFromObject(route), err,
		)
	}

	resp, err := sdk.Routes.UpsertRoute(ctx, sdkkonnectgoops.UpsertRouteRequest{
		ControlPlaneID: cp.Status.ID,
		RouteID:        route.Status.Konnect.ID,
		Route: sdkkonnectgocomp.RouteInput{
			Destinations:            route.Spec.KongRouteAPISpec.Destinations,
			Headers:                 route.Spec.KongRouteAPISpec.Headers,
			Hosts:                   route.Spec.KongRouteAPISpec.Hosts,
			HTTPSRedirectStatusCode: route.Spec.KongRouteAPISpec.HTTPSRedirectStatusCode,
			Methods:                 route.Spec.KongRouteAPISpec.Methods,
			Name:                    route.Spec.KongRouteAPISpec.Name,
			PathHandling:            route.Spec.KongRouteAPISpec.PathHandling,
			Paths:                   route.Spec.KongRouteAPISpec.Paths,
			PreserveHost:            route.Spec.KongRouteAPISpec.PreserveHost,
			Protocols:               route.Spec.KongRouteAPISpec.Protocols,
			RegexPriority:           route.Spec.KongRouteAPISpec.RegexPriority,
			RequestBuffering:        route.Spec.KongRouteAPISpec.RequestBuffering,
			ResponseBuffering:       route.Spec.KongRouteAPISpec.ResponseBuffering,
			Snis:                    route.Spec.KongRouteAPISpec.Snis,
			Sources:                 route.Spec.KongRouteAPISpec.Sources,
			StripPath:               route.Spec.KongRouteAPISpec.StripPath,
			Tags:                    route.Spec.KongRouteAPISpec.Tags,
			Service: &sdkkonnectgocomp.RouteService{
				ID: sdkkonnectgo.String(route.Status.Konnect.ServiceID),
			},
		},
	})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errHandled := handleResp[configurationv1alpha1.KongRoute](err, resp, UpdateOp); errHandled != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				errHandled.Error(),
				route.GetGeneration(),
			),
			route,
		)
		return errHandled
	}

	route.GetKonnectStatus().SetKonnectID(*resp.Route.ID)
	route.Status.Konnect.ControlPlaneID = cp.Status.ID
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			route.GetGeneration(),
		),
		route,
	)

	return nil
}

func deleteRoute(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	route *configurationv1alpha1.KongRoute,
) error {
	id := route.GetKonnectStatus().GetKonnectID()
	if id == "" {
		return fmt.Errorf("can't remove %T without a Konnect ID", route)
	}

	resp, err := sdk.Routes.DeleteRoute(ctx, route.Status.Konnect.ControlPlaneID, id)
	if errHandled := handleResp[configurationv1alpha1.KongRoute](err, resp, DeleteOp); errHandled != nil {
		var sdkError *sdkerrors.SDKError
		if errors.As(errHandled, &sdkError) && sdkError.StatusCode == 404 {
			logger.Info("entity not found in Konnect, skipping delete",
				"op", DeleteOp, "type", route.GetTypeName(), "id", id,
			)
			return nil
		}
		return FailedKonnectOpError[configurationv1alpha1.KongRoute]{
			Op:  DeleteOp,
			Err: errHandled,
		}
	}

	return nil
}
