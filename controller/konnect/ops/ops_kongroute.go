package ops

import (
	"context"
	"fmt"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/samber/lo"

	sdkops "github.com/kong/gateway-operator/controller/konnect/ops/sdk"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func createRoute(
	ctx context.Context,
	sdk sdkops.RoutesSDK,
	route *configurationv1alpha1.KongRoute,
) error {
	if route.GetControlPlaneID() == "" {
		return CantPerformOperationWithoutControlPlaneIDError{Entity: route, Op: CreateOp}
	}

	resp, err := sdk.CreateRoute(ctx, route.Status.Konnect.ControlPlaneID, kongRouteToSDKRouteInput(route))

	if errWrap := wrapErrIfKonnectOpFailed(err, CreateOp, route); errWrap != nil {
		return errWrap
	}

	if resp == nil || resp.Route == nil || resp.Route.RouteJSON.ID == nil {
		return fmt.Errorf("failed creating %s: %w", route.GetTypeName(), ErrNilResponse)
	}

	route.SetKonnectID(*resp.Route.RouteJSON.ID)

	return nil
}

// updateRoute updates the Konnect Route entity.
// It is assumed that provided KongRoute has Konnect ID set in status.
// It returns an error if the KongRoute does not have a ControlPlaneRef or
// if the operation fails.
func updateRoute(
	ctx context.Context,
	sdk sdkops.RoutesSDK,
	route *configurationv1alpha1.KongRoute,
) error {
	cpID := route.GetControlPlaneID()
	if cpID == "" {
		return CantPerformOperationWithoutControlPlaneIDError{Entity: route, Op: UpdateOp}
	}

	id := route.GetKonnectStatus().GetKonnectID()
	_, err := sdk.UpsertRoute(ctx, sdkkonnectops.UpsertRouteRequest{
		ControlPlaneID: cpID,
		RouteID:        id,
		Route:          kongRouteToSDKRouteInput(route),
	})

	if errWrap := wrapErrIfKonnectOpFailed(err, UpdateOp, route); errWrap != nil {
		return errWrap
	}

	return nil
}

// deleteRoute deletes a KongRoute in Konnect.
// It is assumed that provided KongRoute has Konnect ID set in status.
// It returns an error if the operation fails.
func deleteRoute(
	ctx context.Context,
	sdk sdkops.RoutesSDK,
	route *configurationv1alpha1.KongRoute,
) error {
	id := route.GetKonnectStatus().GetKonnectID()
	_, err := sdk.DeleteRoute(ctx, route.Status.Konnect.ControlPlaneID, id)
	if errWrap := wrapErrIfKonnectOpFailed(err, DeleteOp, route); errWrap != nil {
		return handleDeleteError(ctx, err, route)
	}

	return nil
}

func kongRouteToSDKRouteInput(
	route *configurationv1alpha1.KongRoute,
) sdkkonnectcomp.RouteInput {
	r := sdkkonnectcomp.RouteInput{
		RouteJSONInput: &sdkkonnectcomp.RouteJSONInput{
			Destinations:            route.Spec.KongRouteAPISpec.Destinations,
			Headers:                 route.Spec.Headers,
			Hosts:                   route.Spec.KongRouteAPISpec.Hosts,
			HTTPSRedirectStatusCode: route.Spec.KongRouteAPISpec.HTTPSRedirectStatusCode,
			Methods:                 route.Spec.KongRouteAPISpec.Methods,
			Name:                    route.Spec.KongRouteAPISpec.Name,
			PathHandling:            route.Spec.KongRouteAPISpec.PathHandling,
			Paths:                   route.Spec.KongRouteAPISpec.Paths,
			PreserveHost:            route.Spec.KongRouteAPISpec.PreserveHost,
			Protocols: lo.Map(route.Spec.KongRouteAPISpec.Protocols, func(protocol sdkkonnectcomp.RouteWithoutParentsProtocols, _ int) sdkkonnectcomp.RouteJSONProtocols {
				return sdkkonnectcomp.RouteJSONProtocols(protocol)
			}),
			RegexPriority:     route.Spec.KongRouteAPISpec.RegexPriority,
			RequestBuffering:  route.Spec.KongRouteAPISpec.RequestBuffering,
			ResponseBuffering: route.Spec.KongRouteAPISpec.ResponseBuffering,
			Snis:              route.Spec.KongRouteAPISpec.Snis,
			Sources:           route.Spec.KongRouteAPISpec.Sources,
			StripPath:         route.Spec.KongRouteAPISpec.StripPath,
			Tags:              GenerateTagsForObject(route, route.Spec.KongRouteAPISpec.Tags...),
		},
	}
	if route.Status.Konnect != nil && route.Status.Konnect.ServiceID != "" {
		r.RouteJSONInput.Service = &sdkkonnectcomp.RouteJSONService{
			ID: sdkkonnectgo.String(route.Status.Konnect.ServiceID),
		}
	}
	return r
}

// getKongRouteForUID returns the Konnect ID of the KongRoute
// that matches the UID of the provided KongRoute.
func getKongRouteForUID(
	ctx context.Context,
	sdk sdkops.RoutesSDK,
	r *configurationv1alpha1.KongRoute,
) (string, error) {
	reqList := sdkkonnectops.ListRouteRequest{
		// NOTE: only filter on object's UID.
		// Other fields like name might have changed in the meantime but that's OK.
		// Those will be enforced via subsequent updates.
		Tags:           lo.ToPtr(UIDLabelForObject(r)),
		ControlPlaneID: r.GetControlPlaneID(),
	}

	resp, err := sdk.ListRoute(ctx, reqList)
	if err != nil {
		return "", fmt.Errorf("failed listing %s: %w", r.GetTypeName(), err)
	}

	if resp == nil || resp.Object == nil {
		return "", fmt.Errorf("failed listing %s: %w", r.GetTypeName(), ErrNilResponse)
	}

	return getMatchingEntryFromListResponseData(
		sliceToEntityWithIDPtrSlice(
			lo.Map(resp.Object.Data, func(route sdkkonnectcomp.Route, _ int) sdkkonnectcomp.RouteJSON {
				return *route.RouteJSON
			}),
		), r)
}
