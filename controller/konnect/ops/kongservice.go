package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// ServicesSDK is the interface for the Konnect Service SDK.
type ServicesSDK interface {
	CreateService(ctx context.Context, controlPlaneID string, service sdkkonnectcomp.ServiceInput, opts ...sdkkonnectops.Option) (*sdkkonnectops.CreateServiceResponse, error)
	UpsertService(ctx context.Context, req sdkkonnectops.UpsertServiceRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertServiceResponse, error)
	DeleteService(ctx context.Context, controlPlaneID, serviceID string, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeleteServiceResponse, error)
}
