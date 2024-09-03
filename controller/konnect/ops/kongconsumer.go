package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// ConsumersSDK is the interface for the Konnect Consumers SDK.
type ConsumersSDK interface {
	CreateConsumer(ctx context.Context, controlPlaneID string, consumerInput sdkkonnectcomp.ConsumerInput, opts ...sdkkonnectops.Option) (*sdkkonnectops.CreateConsumerResponse, error)
	UpsertConsumer(ctx context.Context, upsertConsumerRequest sdkkonnectops.UpsertConsumerRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertConsumerResponse, error)
	DeleteConsumer(ctx context.Context, controlPlaneID string, consumerID string, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeleteConsumerResponse, error)
}
