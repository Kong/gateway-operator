package konnect

import (
	"context"
	"errors"
	"fmt"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectgoops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/Kong/sdk-konnect-go/models/sdkerrors"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func createConsumer(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	c *configurationv1.KongConsumer,
) error {
	resp, err := sdk.Consumers.CreateConsumer(ctx, c.Status.Konnect.ControlPlaneID, sdkkonnectgocomp.ConsumerInput{
		CustomID: &c.CustomID,
		// TODO(pmalek): handle tags via resource annotation as per: https://docs.konghq.com/kubernetes-ingress-controller/latest/reference/annotations/#konghqcomtags
		// Tags:     ...
		Username: &c.Username,
	})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errHandled := handleResp[configurationv1.KongConsumer](err, resp, CreateOp); errHandled != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				errHandled.Error(),
				c.GetGeneration(),
			),
			c,
		)
		return errHandled
	}

	c.Status.Konnect.SetKonnectID(*resp.Consumer.ID)
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			c.GetGeneration(),
		),
		c,
	)

	return nil
}

func updateConsumer(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	c *configurationv1.KongConsumer,
) error {
	// TODO(pmalek) handle other types of CP ref
	nnCP := types.NamespacedName{
		Namespace: c.Spec.ControlPlaneRef.KonnectNamespacedRef.Namespace,
		Name:      c.Spec.ControlPlaneRef.KonnectNamespacedRef.Name,
	}
	if nnCP.Namespace == "" {
		nnCP.Namespace = c.Namespace
	}
	var cp konnectv1alpha1.KonnectControlPlane
	if err := cl.Get(ctx, nnCP, &cp); err != nil {
		return fmt.Errorf("failed to get KonnectControlPlane %s: for KongConsumer %s: %w",
			nnCP, client.ObjectKeyFromObject(c), err,
		)
	}

	resp, err := sdk.Consumers.UpsertConsumer(ctx, sdkkonnectgoops.UpsertConsumerRequest{
		ControlPlaneID: cp.Status.ID,
		ConsumerID:     c.Status.Konnect.ID,
		Consumer: sdkkonnectgocomp.ConsumerInput{
			CustomID: &c.CustomID,
			// TODO(pmalek): handle tags via resource annotation as per: https://docs.konghq.com/kubernetes-ingress-controller/latest/reference/annotations/#konghqcomtags
			// Tags:     ...
			Username: &c.Username,
		},
	})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errHandled := handleResp[configurationv1.KongConsumer](err, resp, UpdateOp); errHandled != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				errHandled.Error(),
				c.GetGeneration(),
			),
			c,
		)
		return errHandled
	}

	c.Status.Konnect.SetKonnectID(*resp.Consumer.ID)
	c.Status.Konnect.ControlPlaneID = cp.Status.ID
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			c.GetGeneration(),
		),
		c,
	)

	return nil
}

func deleteConsumer(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	c *configurationv1.KongConsumer,
) error {
	id := c.Status.Konnect.GetKonnectID()
	if id == "" {
		return fmt.Errorf("can't remove %T without a Konnect ID", c)
	}

	resp, err := sdk.Consumers.DeleteConsumer(ctx, c.Status.Konnect.ControlPlaneID, id)
	if errHandled := handleResp[configurationv1.KongConsumer](err, resp, DeleteOp); errHandled != nil {
		var sdkError *sdkerrors.SDKError
		if errors.As(errHandled, &sdkError) && sdkError.StatusCode == 404 {
			logger.Info("entity not found in Konnect, skipping delete",
				"op", DeleteOp, "type", c.GetTypeName(), "id", id,
			)
			return nil
		}
		return FailedKonnectOpError[configurationv1.KongConsumer]{
			Op:  DeleteOp,
			Err: errHandled,
		}
	}

	return nil
}
