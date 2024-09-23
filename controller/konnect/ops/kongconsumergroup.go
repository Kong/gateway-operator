package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// ConsumerGroupSDK is the interface for the Konnect ConsumerGroups SDK.
type ConsumerGroupSDK interface {
	CreateConsumerGroup(ctx context.Context, controlPlaneID string, consumerInput sdkkonnectcomp.ConsumerGroupInput, opts ...sdkkonnectops.Option) (*sdkkonnectops.CreateConsumerGroupResponse, error)
	UpsertConsumerGroup(ctx context.Context, upsertConsumerRequest sdkkonnectops.UpsertConsumerGroupRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertConsumerGroupResponse, error)
	DeleteConsumerGroup(ctx context.Context, controlPlaneID string, consumerID string, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeleteConsumerGroupResponse, error)
	AddConsumerToGroup(ctx context.Context, request sdkkonnectops.AddConsumerToGroupRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.AddConsumerToGroupResponse, error)
	RemoveConsumerFromGroup(ctx context.Context, request sdkkonnectops.RemoveConsumerFromGroupRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.RemoveConsumerFromGroupResponse, error)
	ListConsumerGroupsForConsumer(ctx context.Context, request sdkkonnectops.ListConsumerGroupsForConsumerRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.ListConsumerGroupsForConsumerResponse, error)
}
