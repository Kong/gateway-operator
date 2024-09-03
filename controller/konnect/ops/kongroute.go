package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// RoutesSDK is the interface for the Konnect Routes SDK.
type RoutesSDK interface {
	CreateRoute(ctx context.Context, controlPlaneID string, route sdkkonnectcomp.RouteInput, opts ...sdkkonnectops.Option) (*sdkkonnectops.CreateRouteResponse, error)
	UpsertRoute(ctx context.Context, req sdkkonnectops.UpsertRouteRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertRouteResponse, error)
	DeleteRoute(ctx context.Context, controlPlaneID, routeID string, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeleteRouteResponse, error)
}
