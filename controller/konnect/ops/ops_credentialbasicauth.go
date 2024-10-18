package ops

import (
	"context"

	sdkkonnectcomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/samber/lo"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

func createKongCredentialBasicAuth(
	ctx context.Context,
	sdk KongCredentialBasicAuthSDK,
	cred *configurationv1alpha1.KongCredentialBasicAuth,
) error {
	cpID := cred.GetControlPlaneID()
	if cpID == "" {
		return CantPerformOperationWithoutControlPlaneIDError{Entity: cred, Op: CreateOp}
	}

	resp, err := sdk.CreateBasicAuthWithConsumer(ctx,
		sdkkonnectops.CreateBasicAuthWithConsumerRequest{
			ControlPlaneID:              cpID,
			ConsumerIDForNestedEntities: cred.Status.Konnect.GetConsumerID(),
			BasicAuthWithoutParents:     kongCredentialBasicAuthToBasicAuthWithoutParents(cred),
		},
	)

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errWrap := wrapErrIfKonnectOpFailed(err, CreateOp, cred); errWrap != nil {
		SetKonnectEntityProgrammedConditionFalse(cred, "FailedToCreate", errWrap.Error())
		return errWrap
	}

	cred.Status.Konnect.SetKonnectID(*resp.BasicAuth.ID)
	SetKonnectEntityProgrammedCondition(cred)

	return nil
}

// updateKongCredentialBasicAuth updates the Konnect BasicAuth entity.
// It is assumed that the provided BasicAuth has Konnect ID set in status.
// It returns an error if the BasicAuth does not have a ControlPlaneRef or
// if the operation fails.
func updateKongCredentialBasicAuth(
	ctx context.Context,
	sdk KongCredentialBasicAuthSDK,
	cred *configurationv1alpha1.KongCredentialBasicAuth,
) error {
	cpID := cred.GetControlPlaneID()
	if cpID == "" {
		return CantPerformOperationWithoutControlPlaneIDError{Entity: cred, Op: UpdateOp}
	}

	_, err := sdk.UpsertBasicAuthWithConsumer(ctx, sdkkonnectops.UpsertBasicAuthWithConsumerRequest{
		ControlPlaneID:              cpID,
		ConsumerIDForNestedEntities: cred.Status.Konnect.GetConsumerID(),
		BasicAuthID:                 cred.GetKonnectStatus().GetKonnectID(),
		BasicAuthWithoutParents:     kongCredentialBasicAuthToBasicAuthWithoutParents(cred),
	})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errWrap := wrapErrIfKonnectOpFailed(err, UpdateOp, cred); errWrap != nil {
		SetKonnectEntityProgrammedConditionFalse(cred, "FailedToUpdate", errWrap.Error())
		return errWrap
	}

	SetKonnectEntityProgrammedCondition(cred)

	return nil
}

// deleteKongCredentialBasicAuth deletes a BasicAuth credential in Konnect.
// It is assumed that the provided BasicAuth has Konnect ID set in status.
// It returns an error if the operation fails.
func deleteKongCredentialBasicAuth(
	ctx context.Context,
	sdk KongCredentialBasicAuthSDK,
	cred *configurationv1alpha1.KongCredentialBasicAuth,
) error {
	cpID := cred.GetControlPlaneID()
	id := cred.GetKonnectStatus().GetKonnectID()
	_, err := sdk.DeleteBasicAuthWithConsumer(ctx,
		sdkkonnectops.DeleteBasicAuthWithConsumerRequest{
			ControlPlaneID:              cpID,
			ConsumerIDForNestedEntities: cred.Status.Konnect.GetConsumerID(),
			BasicAuthID:                 id,
		})
	if errWrap := wrapErrIfKonnectOpFailed(err, DeleteOp, cred); errWrap != nil {
		return handleDeleteError(ctx, err, cred)
	}

	return nil
}

func kongCredentialBasicAuthToBasicAuthWithoutParents(
	cred *configurationv1alpha1.KongCredentialBasicAuth,
) sdkkonnectcomp.BasicAuthWithoutParents {
	return sdkkonnectcomp.BasicAuthWithoutParents{
		Password: lo.ToPtr(cred.Spec.Password),
		Username: lo.ToPtr(cred.Spec.Username),
		Tags:     GenerateTagsForObject(cred, cred.Spec.Tags...),
	}
}
