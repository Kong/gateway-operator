package konnect

import (
	"context"
	"errors"
	"fmt"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/components"
	"github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

func createControlPlane(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cp *operatorv1alpha1.KonnectControlPlane,
) error {
	setKonnectLabels(cp, &cp.Spec)

	resp, err := sdk.ControlPlanes.CreateControlPlane(ctx, cp.Spec.CreateControlPlaneRequest)
	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if err != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				err.Error(),
				cp.GetGeneration(),
			),
			&cp.Status,
		)
		return err
	}
	if err := handleStatusCode[operatorv1alpha1.KonnectControlPlane](resp, CreateOp); err != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				err.Error(),
				cp.GetGeneration(),
			),
			&cp.Status,
		)
		return err
	}

	cp.Status.KonnectID = *resp.ControlPlane.ID
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			cp.GetGeneration(),
		),
		&cp.Status,
	)

	return nil
}

func deleteControlPlane(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cp *operatorv1alpha1.KonnectControlPlane,
) error {
	if cp.GetStatusID() == "" {
		return fmt.Errorf("can't remove %T without a Konnect ID", cp)
	}
	id := cp.GetStatusID()
	resp, err := sdk.ControlPlanes.DeleteControlPlane(ctx, id)
	if err != nil {
		var errNotFound *sdkerrors.NotFoundError
		if errors.As(err, &errNotFound) {
			logger.Info("entity not found in Konnect, skipping delete",
				"op", DeleteOp, "type", cp.GetTypeName(), "id", id,
			)
			return nil
		}
		return FailedKonnectOpError[operatorv1alpha1.KonnectControlPlane]{
			Op:  DeleteOp,
			Err: err,
		}
	}

	if err := handleStatusCode[operatorv1alpha1.KonnectControlPlane](resp, DeleteOp); err != nil {
		return err
	}

	return nil
}

func updateControlPlane(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cp *operatorv1alpha1.KonnectControlPlane,
) error {
	if cp.GetStatusID() == "" {
		return fmt.Errorf("can't update %T without a Konnect ID", cp)
	}
	id := cp.GetStatusID()

	setKonnectLabels(cp, &cp.Spec)
	req := components.UpdateControlPlaneRequest{
		Name:        sdkkonnectgo.String(cp.Spec.Name),
		Description: cp.Spec.Description,
		AuthType:    (*components.UpdateControlPlaneRequestAuthType)(cp.Spec.AuthType),
		ProxyUrls:   cp.Spec.ProxyUrls,
		Labels:      cp.Spec.Labels,
	}

	resp, err := sdk.ControlPlanes.UpdateControlPlane(ctx, id, req)
	if err != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToUpdate",
				err.Error(),
				cp.GetGeneration(),
			),
			&cp.Status,
		)
		return FailedKonnectOpError[operatorv1alpha1.KonnectControlPlane]{
			Op:  UpdateOp,
			Err: err,
		}
	}

	if err := handleStatusCode[operatorv1alpha1.KonnectControlPlane](resp, UpdateOp); err != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToUpdate",
				err.Error(),
				cp.GetGeneration(),
			),
			&cp.Status,
		)
		return FailedKonnectOpError[operatorv1alpha1.KonnectControlPlane]{
			Op:  UpdateOp,
			Err: err,
		}
	}

	cp.Status.KonnectID = *resp.ControlPlane.ID
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			cp.GetGeneration(),
		),
		&cp.Status,
	)

	return nil
}
