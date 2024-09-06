package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// PluginSDK is the interface for Konnect plugin SDK.
type PluginSDK interface {
	CreatePlugin(ctx context.Context, controlPlaneID string, plugin sdkkonnectcomp.PluginInput, opts ...sdkkonnectops.Option) (*sdkkonnectops.CreatePluginResponse, error)
	UpsertPlugin(ctx context.Context, request sdkkonnectops.UpsertPluginRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertPluginResponse, error)
	DeletePlugin(ctx context.Context, controlPlaneID string, pluginID string, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeletePluginResponse, error)
}
