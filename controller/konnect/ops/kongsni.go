package ops

import (
	"context"

	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
)

// SNIsSDK is the interface to operate Kong SNIs.
type SNIsSDK interface {
	CreateSniWithCertificate(context.Context, sdkkonnectops.CreateSniWithCertificateRequest, ...sdkkonnectops.Option) (*sdkkonnectops.CreateSniWithCertificateResponse, error)
	UpsertSniWithCertificate(ctx context.Context, request sdkkonnectops.UpsertSniWithCertificateRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.UpsertSniWithCertificateResponse, error)
	DeleteSniWithCertificate(ctx context.Context, request sdkkonnectops.DeleteSniWithCertificateRequest, opts ...sdkkonnectops.Option) (*sdkkonnectops.DeleteSniWithCertificateResponse, error)
}
