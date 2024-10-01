package ops

import (
	"context"
	"errors"
	"fmt"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	sdkkonnecterrs "github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

// createKeySet creates a KongKeySet in Konnect.
// It sets the KonnectID and the Programmed condition in the KongKeySet status.
func createKeySet(
	ctx context.Context,
	sdk KeySetsSDK,
	keySet *configurationv1alpha1.KongKeySet,
) error {
	cpID := keySet.GetControlPlaneID()
	if cpID == "" {
		return fmt.Errorf("can't create %T %s without a Konnect ControlPlane ID", keySet, client.ObjectKeyFromObject(keySet))
	}

	resp, err := sdk.CreateKeySet(ctx,
		cpID,
		kongKeySetToKeySetInput(keySet),
	)

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errWrap := wrapErrIfKonnectOpFailed(err, CreateOp, keySet); errWrap != nil {
		SetKonnectEntityProgrammedConditionFalse(keySet, "FailedToCreate", errWrap.Error())
		return errWrap
	}

	keySet.Status.Konnect.SetKonnectID(*resp.KeySet.ID)
	SetKonnectEntityProgrammedCondition(keySet)

	return nil
}

// updateKeySet updates a KongKeySet in Konnect.
// The KongKeySet must have a KonnectID set in its status.
// It returns an error if the KongKeySet does not have a KonnectID.
func updateKeySet(
	ctx context.Context,
	sdk KeySetsSDK,
	keySet *configurationv1alpha1.KongKeySet,
) error {
	cpID := keySet.GetControlPlaneID()
	if cpID == "" {
		return fmt.Errorf("can't update %T without a ControlPlaneID", keySet)
	}

	_, err := sdk.UpsertKeySet(ctx,
		sdkkonnectops.UpsertKeySetRequest{
			ControlPlaneID: cpID,
			KeySetID:       keySet.GetKonnectStatus().GetKonnectID(),
			KeySet:         kongKeySetToKeySetInput(keySet),
		},
	)

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errWrap := wrapErrIfKonnectOpFailed(err, UpdateOp, keySet); errWrap != nil {
		var sdkError *sdkkonnecterrs.SDKError
		if errors.As(errWrap, &sdkError) {
			if sdkError.StatusCode == 404 {
				if err := createKeySet(ctx, sdk, keySet); err != nil {
					return FailedKonnectOpError[configurationv1alpha1.KongKeySet]{
						Op:  UpdateOp,
						Err: err,
					}
				}
				return nil // createKeySet sets the status so we can return here.
			}
			return FailedKonnectOpError[configurationv1alpha1.KongKeySet]{
				Op:  UpdateOp,
				Err: sdkError,
			}
		}
		SetKonnectEntityProgrammedConditionFalse(keySet, "FailedToUpdate", errWrap.Error())
		return errWrap
	}

	SetKonnectEntityProgrammedCondition(keySet)

	return nil
}

// deleteKeySet deletes a KongKeySet in Konnect.
// The KongKeySet must have a KonnectID set in its status.
// It returns an error if the operation fails.
func deleteKeySet(
	ctx context.Context,
	sdk KeySetsSDK,
	keySet *configurationv1alpha1.KongKeySet,
) error {
	id := keySet.Status.Konnect.GetKonnectID()
	_, err := sdk.DeleteKeySet(ctx, keySet.GetControlPlaneID(), id)
	if errWrap := wrapErrIfKonnectOpFailed(err, DeleteOp, keySet); errWrap != nil {
		var sdkError *sdkkonnecterrs.SDKError
		if errors.As(errWrap, &sdkError) {
			if sdkError.StatusCode == 404 {
				ctrllog.FromContext(ctx).
					Info("entity not found in Konnect, skipping delete",
						"op", DeleteOp, "type", keySet.GetTypeName(), "id", id,
					)
				return nil
			}
			return FailedKonnectOpError[configurationv1alpha1.KongKeySet]{
				Op:  DeleteOp,
				Err: sdkError,
			}
		}
		return FailedKonnectOpError[configurationv1alpha1.KongKeySet]{
			Op:  DeleteOp,
			Err: errWrap,
		}
	}

	return nil
}

func kongKeySetToKeySetInput(keySet *configurationv1alpha1.KongKeySet) sdkkonnectcomp.KeySetInput {
	return sdkkonnectcomp.KeySetInput{
		Name: lo.ToPtr(keySet.Spec.Name),
		Tags: GenerateTagsForObject(keySet, keySet.Spec.Tags...),
	}
}
