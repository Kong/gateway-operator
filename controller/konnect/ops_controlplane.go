package konnect

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	"github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

func createControlPlane(
	ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, cp *operatorv1alpha1.KonnectControlPlane,
) error {
	start := time.Now()

	// TODO(pmalek): move setKonnnectLabels out of this switch type so that
	// it's shared across types
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

	// TODO(pmalek): move out of so that it's shared across types
	logOpComplete(logger, start, CreateOp, cp)
	return nil
}

func deleteControlPlane(
	ctx context.Context, sdk *sdkkonnectgo.SDK, logger logr.Logger, cp *operatorv1alpha1.KonnectControlPlane,
) error {
	if cp.GetStatusID() == "" {
		return fmt.Errorf("can't remove %T without a Konnect ID", cp)
	}
	start := time.Now()
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
	// TODO(pmalek): move out of so that it's shared across types
	logOpComplete(logger, start, DeleteOp, cp)

	if err := handleStatusCode[operatorv1alpha1.KonnectControlPlane](resp, DeleteOp); err != nil {
		return err
	}

	return nil
}
