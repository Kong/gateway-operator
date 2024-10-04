package ops

import (
	"context"
	"errors"
	"fmt"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	sdkkonnecterrs "github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func createKongCredentialHMAC(
	ctx context.Context,
	sdk KongCredentialHMACSDK,
	cred *configurationv1alpha1.KongCredentialHMAC,
) error {
	cpID := cred.GetControlPlaneID()
	if cpID == "" {
		return fmt.Errorf("can't create %T %s without a Konnect ControlPlane ID", cred, client.ObjectKeyFromObject(cred))
	}

	resp, err := sdk.CreateHmacAuthWithConsumer(ctx,
		sdkkonnectops.CreateHmacAuthWithConsumerRequest{
			ControlPlaneID:              cpID,
			ConsumerIDForNestedEntities: cred.Status.Konnect.GetConsumerID(),
			HMACAuthWithoutParents:      kongCredentialHMACToHMACWithoutParents(cred),
		},
	)

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errWrap := wrapErrIfKonnectOpFailed(err, CreateOp, cred); errWrap != nil {
		SetKonnectEntityProgrammedConditionFalse(cred, "FailedToCreate", errWrap.Error())
		return errWrap
	}

	cred.Status.Konnect.SetKonnectID(*resp.HMACAuth.ID)
	SetKonnectEntityProgrammedCondition(cred)

	return nil
}

// updateKongCredentialHMAC updates the Konnect HMAC entity.
// It is assumed that the provided HMAC has Konnect ID set in status.
// It returns an error if the HMAC does not have a ControlPlaneRef or
// if the operation fails.
func updateKongCredentialHMAC(
	ctx context.Context,
	sdk KongCredentialHMACSDK,
	cred *configurationv1alpha1.KongCredentialHMAC,
) error {
	cpID := cred.GetControlPlaneID()
	if cpID == "" {
		return fmt.Errorf("can't update %T %s without a Konnect ControlPlane ID", cred, client.ObjectKeyFromObject(cred))
	}

	_, err := sdk.UpsertHmacAuthWithConsumer(ctx,
		sdkkonnectops.UpsertHmacAuthWithConsumerRequest{
			ControlPlaneID:              cpID,
			ConsumerIDForNestedEntities: cred.Status.Konnect.GetConsumerID(),
			HMACAuthID:                  cred.GetKonnectStatus().GetKonnectID(),
			HMACAuthWithoutParents:      kongCredentialHMACToHMACWithoutParents(cred),
		})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errWrap := wrapErrIfKonnectOpFailed(err, UpdateOp, cred); errWrap != nil {
		// HMAC update operation returns an SDKError instead of a NotFoundError.
		var sdkError *sdkkonnecterrs.SDKError
		if errors.As(errWrap, &sdkError) {
			switch sdkError.StatusCode {
			case 404:
				if err := createKongCredentialHMAC(ctx, sdk, cred); err != nil {
					return FailedKonnectOpError[configurationv1alpha1.KongCredentialHMAC]{
						Op:  UpdateOp,
						Err: err,
					}
				}
				return nil
			default:
				return FailedKonnectOpError[configurationv1alpha1.KongCredentialHMAC]{
					Op:  UpdateOp,
					Err: sdkError,
				}

			}
		}

		SetKonnectEntityProgrammedConditionFalse(cred, "FailedToUpdate", errWrap.Error())
		return errWrap
	}

	SetKonnectEntityProgrammedCondition(cred)

	return nil
}

// deleteKongCredentialHMAC deletes an HMAC credential in Konnect.
// It is assumed that the provided HMAC has Konnect ID set in status.
// It returns an error if the operation fails.
func deleteKongCredentialHMAC(
	ctx context.Context,
	sdk KongCredentialHMACSDK,
	cred *configurationv1alpha1.KongCredentialHMAC,
) error {
	cpID := cred.GetControlPlaneID()
	id := cred.GetKonnectStatus().GetKonnectID()
	_, err := sdk.DeleteHmacAuthWithConsumer(ctx,
		sdkkonnectops.DeleteHmacAuthWithConsumerRequest{
			ControlPlaneID:              cpID,
			ConsumerIDForNestedEntities: cred.Status.Konnect.GetConsumerID(),
			HMACAuthID:                  id,
		})
	if errWrap := wrapErrIfKonnectOpFailed(err, DeleteOp, cred); errWrap != nil {
		// Service delete operation returns an SDKError instead of a NotFoundError.
		var sdkError *sdkkonnecterrs.SDKError
		if errors.As(errWrap, &sdkError) {
			if sdkError.StatusCode == 404 {
				ctrllog.FromContext(ctx).
					Info("entity not found in Konnect, skipping delete",
						"op", DeleteOp, "type", cred.GetTypeName(), "id", id,
					)
				return nil
			}
			return FailedKonnectOpError[configurationv1alpha1.KongCredentialHMAC]{
				Op:  DeleteOp,
				Err: sdkError,
			}
		}
		return FailedKonnectOpError[configurationv1alpha1.KongService]{
			Op:  DeleteOp,
			Err: errWrap,
		}
	}

	return nil
}

func kongCredentialHMACToHMACWithoutParents(
	cred *configurationv1alpha1.KongCredentialHMAC,
) sdkkonnectcomp.HMACAuthWithoutParents {
	ret := sdkkonnectcomp.HMACAuthWithoutParents{
		Username: cred.Spec.Username,
		Secret:   cred.Spec.Secret,
		Tags:     GenerateTagsForObject(cred, cred.Spec.Tags...),
	}
	return ret
}
