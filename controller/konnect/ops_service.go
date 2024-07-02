package konnect

import (
	"context"

	sdkkonnectgo "github.com/Kong/sdk-konnect-go"
	sdkkonnectgocomp "github.com/Kong/sdk-konnect-go/models/components"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configurationv1alpha1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1alpha1"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

func createService(
	ctx context.Context,
	sdk *sdkkonnectgo.SDK,
	logger logr.Logger,
	svc *configurationv1alpha1.Service,
) error {
	resp, err := sdk.Services.CreateService(ctx, sdkkonnectgocomp.CreateService{
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
				cp.GetGeneration(),
			),
			&cp.Status,
		)
		return errHandled
	}

	cp.Status.SetKonnectID(*resp.ControlPlane.ID)
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
