package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// VaultSDK is the interface for Konnect Vault SDK.
type VaultSDK interface {
	CreateVault(ctx context.Context, controlPlaneID string, vault sdkkonnectcomp.VaultInput, opts ...sdkkonnectops.Option) (*sdkkonnectops.CreateVaultResponse, error)
	UpsertVault(ctx context.Context, request sdkkonnectops.UpsertVaultRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertVaultResponse, error)
	DeleteVault(ctx context.Context, controlPlaneID string, vaultID string, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeleteVaultResponse, error)
}
