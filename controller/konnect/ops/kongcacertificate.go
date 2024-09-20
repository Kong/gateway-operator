package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// CACertificatesSDK is the interface for the CACertificatesSDK.
type CACertificatesSDK interface {
	CreateCaCertificate(ctx context.Context, controlPlaneID string, caCertificate sdkkonnectcomp.CACertificateInput, opts ...sdkkonnectops.Option) (*sdkkonnectops.CreateCaCertificateResponse, error)
	UpsertCaCertificate(ctx context.Context, request sdkkonnectops.UpsertCaCertificateRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertCaCertificateResponse, error)
	DeleteCaCertificate(ctx context.Context, controlPlaneID string, caCertificateID string, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeleteCaCertificateResponse, error)
}
