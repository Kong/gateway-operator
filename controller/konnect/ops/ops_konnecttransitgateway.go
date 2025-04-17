package ops

import (
	context "context"
	"fmt"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	"github.com/samber/lo"

	sdkops "github.com/kong/gateway-operator/controller/konnect/ops/sdk"

	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

// createKonnectTransitGateway creates a transit gateway on the Konnect side.
func createKonnectTransitGateway(
	ctx context.Context,
	sdk sdkops.TransitGatewaysSDK,
	tg *konnectv1alpha1.KonnectCloudGatewayTransitGateway,
) error {
	networkID := tg.GetNetworkID()
	if networkID == "" {
		return CantPerformOperationWithoutNetworkIDError{
			Entity: tg,
			Op:     CreateOp,
		}
	}

	resp, err := sdk.CreateTransitGateway(ctx, networkID, transitGatewaySpecToTransitGatewayInput(tg.Spec.KonnectTransitGatewayAPISpec))
	if err != nil {
		return err
	}

	if errWrap := wrapErrIfKonnectOpFailed(err, CreateOp, tg); errWrap != nil {
		return errWrap
	}

	if resp == nil || resp.TransitGatewayResponse == nil {
		return fmt.Errorf("failed creating %s: %w", tg.GetTypeName(), ErrNilResponse)
	}

	tg.SetKonnectID(extractKonnectIDFromTransitGatewayResponse(resp.TransitGatewayResponse))
	tg.Status.State = extractStateFromTransitGatewayResponse(resp.TransitGatewayResponse)
	return nil
}

// updateKonnectTransitGateway is called when an "Update" operation is called in reconciling a Konnect transit gateway.
// Since Konnect does not provide API to update an existing transit gateway, here we can only update the status of the
// KonnectCloudGatewayTransitGateway resource based on the state of the transit gateway on the Konnect side.
func updateKonnectTransitGateway(
	ctx context.Context,
	sdk sdkops.TransitGatewaysSDK,
	tg *konnectv1alpha1.KonnectCloudGatewayTransitGateway,
) error {
	networkID := tg.GetNetworkID()
	if networkID == "" {
		return CantPerformOperationWithoutNetworkIDError{
			Entity: tg,
			Op:     UpdateOp,
		}
	}

	resp, err := sdk.GetTransitGateway(ctx, networkID, tg.GetKonnectID())
	if errWrap := wrapErrIfKonnectOpFailed(err, UpdateOp, tg); errWrap != nil {
		return errWrap
	}

	if resp == nil || resp.TransitGatewayResponse == nil {
		return fmt.Errorf("failed updating %s: %w", tg.GetTypeName(), ErrNilResponse)
	}

	tg.Status.State = extractStateFromTransitGatewayResponse(resp.TransitGatewayResponse)
	return nil
}

// deleteKonnectTransitGateway deletes a Konnect transit gateway.
func deleteKonnectTransitGateway(
	ctx context.Context,
	sdk sdkops.TransitGatewaysSDK,
	tg *konnectv1alpha1.KonnectCloudGatewayTransitGateway,
) error {
	networkID := tg.GetNetworkID()
	if networkID == "" {
		return CantPerformOperationWithoutNetworkIDError{
			Entity: tg,
			Op:     DeleteOp,
		}
	}

	resp, err := sdk.DeleteTransitGateway(ctx, networkID, tg.GetKonnectID())

	if errWrap := wrapErrIfKonnectOpFailed(err, DeleteOp, tg); errWrap != nil {
		return errWrap
	}

	if resp == nil {
		return fmt.Errorf("failed deleting %s: %w", tg.GetTypeName(), ErrNilResponse)
	}

	return nil
}

var trasitGatewayTypeToSDKTransitGatewayType = map[konnectv1alpha1.TransitGatewayType]sdkkonnectcomp.CreateTransitGatewayRequestType{
	konnectv1alpha1.TransitGatewayTypeAWSTransitGateway:   sdkkonnectcomp.CreateTransitGatewayRequestTypeAWSTransitGateway,
	konnectv1alpha1.TransitGatewayTypeAzureTransitGateway: sdkkonnectcomp.CreateTransitGatewayRequestTypeAzureTransitGateway,
}

func transitGatewaySpecToTransitGatewayInput(
	spec konnectv1alpha1.KonnectTransitGatewayAPISpec,
) sdkkonnectcomp.CreateTransitGatewayRequest {
	typ := trasitGatewayTypeToSDKTransitGatewayType[spec.Type]

	req := sdkkonnectcomp.CreateTransitGatewayRequest{
		Type: typ,
	}

	switch spec.Type {
	case konnectv1alpha1.TransitGatewayTypeAWSTransitGateway:
		req.AWSTransitGateway = &sdkkonnectcomp.AWSTransitGateway{
			Name: spec.AWSTransitGateway.Name,
			DNSConfig: lo.Map(spec.AWSTransitGateway.DNSConfig, func(dnsConf konnectv1alpha1.TransitGatewayDNSConfig, _ int) sdkkonnectcomp.TransitGatewayDNSConfig {
				return sdkkonnectcomp.TransitGatewayDNSConfig{
					RemoteDNSServerIPAddresses: dnsConf.RemoteDNSServerIPAddresses,
					DomainProxyList:            dnsConf.DomainProxyList,
				}
			}),
			CidrBlocks: spec.AWSTransitGateway.CIDRBlocks,
			TransitGatewayAttachmentConfig: sdkkonnectcomp.AwsTransitGatewayAttachmentConfig{
				Kind:             sdkkonnectcomp.AWSTransitGatewayAttachmentTypeAwsTransitGatewayAttachment,
				TransitGatewayID: spec.AWSTransitGateway.AttachmentConfig.TransitGatewayID,
				RAMShareArn:      spec.AWSTransitGateway.AttachmentConfig.RAMShareArn,
			},
		}
	case konnectv1alpha1.TransitGatewayTypeAzureTransitGateway:
		req.AzureTransitGateway = &sdkkonnectcomp.AzureTransitGateway{
			Name: spec.AzureTransitGateway.Name,
			DNSConfig: lo.Map(spec.AWSTransitGateway.DNSConfig, func(dnsConf konnectv1alpha1.TransitGatewayDNSConfig, _ int) sdkkonnectcomp.TransitGatewayDNSConfig {
				return sdkkonnectcomp.TransitGatewayDNSConfig{
					RemoteDNSServerIPAddresses: dnsConf.RemoteDNSServerIPAddresses,
					DomainProxyList:            dnsConf.DomainProxyList,
				}
			}),
			TransitGatewayAttachmentConfig: sdkkonnectcomp.AzureVNETPeeringAttachmentConfig{
				Kind:              sdkkonnectcomp.AzureVNETPeeringAttachmentTypeAzureVnetPeeringAttachment,
				TenantID:          spec.AzureTransitGateway.AttachmentConfig.TenantID,
				SubscriptionID:    spec.AzureTransitGateway.AttachmentConfig.SubscriptionID,
				ResourceGroupName: spec.AzureTransitGateway.AttachmentConfig.ResourceGroupName,
				VnetName:          spec.AzureTransitGateway.AttachmentConfig.VnetName,
			},
		}
	}

	return req
}

func extractKonnectIDFromTransitGatewayResponse(resp *sdkkonnectcomp.TransitGatewayResponse) string {
	switch resp.Type {
	case sdkkonnectcomp.TransitGatewayResponseTypeAwsTransitGatewayResponse:
		return resp.AwsTransitGatewayResponse.ID
	case sdkkonnectcomp.TransitGatewayResponseTypeAzureTransitGatewayResponse:
		return resp.AzureTransitGatewayResponse.ID
	case sdkkonnectcomp.TransitGatewayResponseTypeAwsVpcPeeringGatewayResponse:
		// AWS VPC peering gateway is not supported yet.
		return ""
	}
	return ""
}

func extractStateFromTransitGatewayResponse(resp *sdkkonnectcomp.TransitGatewayResponse) sdkkonnectcomp.TransitGatewayState {
	switch resp.Type {
	case sdkkonnectcomp.TransitGatewayResponseTypeAwsTransitGatewayResponse:
		return resp.AwsTransitGatewayResponse.State
	case sdkkonnectcomp.TransitGatewayResponseTypeAzureTransitGatewayResponse:
		return resp.AzureTransitGatewayResponse.State
	case sdkkonnectcomp.TransitGatewayResponseTypeAwsVpcPeeringGatewayResponse:
		// AWS VPC peering gateway is not supported yet.
		return sdkkonnectcomp.TransitGatewayState("")
	}
	return sdkkonnectcomp.TransitGatewayState("")
}
