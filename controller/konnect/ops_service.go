package konnect

import (
	"context"
	"fmt"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	sdkkonnectgoops "github.com/Kong/sdk-konnect-go/models/operations"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configurationv1alpha1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1alpha1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

func createService(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	svc *configurationv1alpha1.Service,
) error {
	// TODO(pmalek) handle other types of CP ref
	nnCP := types.NamespacedName{
		Namespace: svc.Spec.ControlPlaneRef.KonnectNamespacedRef.Namespace,
		Name:      svc.Spec.ControlPlaneRef.KonnectNamespacedRef.Name,
	}
	if nnCP.Namespace == "" {
		nnCP.Namespace = svc.Namespace
	}
	var cp operatorv1alpha1.KonnectControlPlane
	if err := cl.Get(ctx, nnCP, &cp); err != nil {
		return fmt.Errorf("failed to get KonnectControlPlane %s: for Service %s: %w",
			nnCP, client.ObjectKeyFromObject(svc), err,
		)
	}

	resp, err := sdk.Services.CreateService(ctx, cp.Status.KonnectID, sdkkonnectgocomp.CreateService{
		URL:            svc.Spec.URL,
		CaCertificates: svc.Spec.CaCertificates,
		ConnectTimeout: svc.Spec.ConnectTimeout,
		Enabled:        svc.Spec.Enabled,
		Host:           svc.Spec.Host,
		Name:           svc.Spec.Name,
		Path:           svc.Spec.Path,
		Port:           svc.Spec.Port,
		Protocol:       svc.Spec.Protocol,
		ReadTimeout:    svc.Spec.ReadTimeout,
		Retries:        svc.Spec.Retries,
		Tags:           svc.Spec.Tags,
		TLSVerify:      svc.Spec.TLSVerify,
		TLSVerifyDepth: svc.Spec.TLSVerifyDepth,
		WriteTimeout:   svc.Spec.WriteTimeout,
	})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errHandled := handleResp[operatorv1alpha1.KonnectControlPlane](err, resp, CreateOp); errHandled != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				errHandled.Error(),
				svc.GetGeneration(),
			),
			&svc.Status,
		)
		return errHandled
	}

	svc.Status.SetKonnectID(resp.Service.ID)
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			svc.GetGeneration(),
		),
		&svc.Status,
	)

	return nil
}

func updateService(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	cl client.Client,
	svc *configurationv1alpha1.Service,
) error {
	// TODO(pmalek) handle other types of CP ref
	nnCP := types.NamespacedName{
		Namespace: svc.Spec.ControlPlaneRef.KonnectNamespacedRef.Namespace,
		Name:      svc.Spec.ControlPlaneRef.KonnectNamespacedRef.Name,
	}
	if nnCP.Namespace == "" {
		nnCP.Namespace = svc.Namespace
	}
	var cp operatorv1alpha1.KonnectControlPlane
	if err := cl.Get(ctx, nnCP, &cp); err != nil {
		return fmt.Errorf("failed to get KonnectControlPlane %s: for Service %s: %w",
			nnCP, client.ObjectKeyFromObject(svc), err,
		)
	}

	resp, err := sdk.Services.UpsertService(ctx, sdkkonnectgoops.UpsertServiceRequest{
		ControlPlaneID: cp.Status.KonnectID,
		ServiceID:      svc.Status.KonnectID,
		CreateService: sdkkonnectgocomp.CreateService{
			URL:            svc.Spec.URL,
			CaCertificates: svc.Spec.CaCertificates,
			ConnectTimeout: svc.Spec.ConnectTimeout,
			Enabled:        svc.Spec.Enabled,
			Host:           svc.Spec.Host,
			Name:           svc.Spec.Name,
			Path:           svc.Spec.Path,
			Port:           svc.Spec.Port,
			Protocol:       svc.Spec.Protocol,
			ReadTimeout:    svc.Spec.ReadTimeout,
			Retries:        svc.Spec.Retries,
			Tags:           svc.Spec.Tags,
			TLSVerify:      svc.Spec.TLSVerify,
			TLSVerifyDepth: svc.Spec.TLSVerifyDepth,
			WriteTimeout:   svc.Spec.WriteTimeout,
		},
	})

	// TODO: handle already exists
	// Can't adopt it as it will cause conflicts between the controller
	// that created that entity and already manages it, hm
	if errHandled := handleResp[operatorv1alpha1.KonnectControlPlane](err, resp, UpdateOp); errHandled != nil {
		k8sutils.SetCondition(
			k8sutils.NewConditionWithGeneration(
				KonnectEntityProgrammedConditionType,
				metav1.ConditionFalse,
				"FailedToCreate",
				errHandled.Error(),
				svc.GetGeneration(),
			),
			&svc.Status,
		)
		return errHandled
	}

	svc.Status.SetKonnectID(resp.Service.ID)
	k8sutils.SetCondition(
		k8sutils.NewConditionWithGeneration(
			KonnectEntityProgrammedConditionType,
			metav1.ConditionTrue,
			KonnectEntityProgrammedReason,
			"",
			svc.GetGeneration(),
		),
		&svc.Status,
	)

	return nil
}
