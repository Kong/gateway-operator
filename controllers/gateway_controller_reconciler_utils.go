package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	operatorv1alpha1 "github.com/kong/gateway-operator/apis/v1alpha1"
	"github.com/kong/gateway-operator/internal/consts"
	operatorerrors "github.com/kong/gateway-operator/internal/errors"
	gatewayutils "github.com/kong/gateway-operator/internal/utils/gateway"
	k8sutils "github.com/kong/gateway-operator/internal/utils/kubernetes"
	k8sresources "github.com/kong/gateway-operator/internal/utils/kubernetes/resources"
	"github.com/kong/gateway-operator/pkg/vars"
)

// -----------------------------------------------------------------------------
// GatewayReconciler - Reconciler Helpers
// -----------------------------------------------------------------------------

func (r *GatewayReconciler) ensureDataPlaneForGateway(ctx context.Context,
	gateway *gatewayv1alpha2.Gateway,
	gatewayConfig *operatorv1alpha1.GatewayConfiguration,
) (*operatorv1alpha1.DataPlane, error) {
	dataplanes, err := gatewayutils.ListDataPlanesForGateway(
		ctx,
		r.Client,
		gateway,
	)
	if err != nil {
		return nil, err
	}

	count := len(dataplanes)
	if count > 1 {
		// if there is more than one Dataplane owned by the same Gateway,
		// delete all of them and recreate only one as follows
		if err := r.Client.DeleteAllOf(ctx, &operatorv1alpha1.DataPlane{},
			client.InNamespace(gateway.Namespace),
			client.MatchingLabels{consts.GatewayOperatorControlledLabel: consts.GatewayManagedLabelValue},
		); err != nil {
			return nil, err
		}
	}

	generatedDataplane := k8sresources.GenerateNewDataPlaneForGateway(*gateway, gatewayConfig)
	k8sutils.SetOwnerForObject(generatedDataplane, gateway)
	gatewayutils.LabelObjectAsGatewayManaged(generatedDataplane)

	if count == 1 {
		var updated bool
		existingDataplane := &dataplanes[0]
		updated, existingDataplane.ObjectMeta = k8sutils.EnsureObjectMetaIsUpdated(existingDataplane.ObjectMeta, generatedDataplane.ObjectMeta)

		if gatewayConfig.Spec.DataPlaneDeploymentOptions != nil {
			if !dataplaneSpecDeepEqual(&existingDataplane.Spec.DataPlaneDeploymentOptions, gatewayConfig.Spec.DataPlaneDeploymentOptions) {
				existingDataplane.Spec.DataPlaneDeploymentOptions = *gatewayConfig.Spec.DataPlaneDeploymentOptions
				updated = true
			}
		}
		if updated {
			return existingDataplane, r.Client.Update(ctx, existingDataplane)
		}
		return existingDataplane, nil
	}

	return generatedDataplane, r.Client.Create(ctx, generatedDataplane)
}

func (r *GatewayReconciler) ensureControlPlaneForGateway(
	ctx context.Context,
	gatewayClass *gatewayv1alpha2.GatewayClass,
	gateway *gatewayv1alpha2.Gateway,
	gatewayConfig *operatorv1alpha1.GatewayConfiguration,
	dataplaneName string,
) (*operatorv1alpha1.ControlPlane, error) {
	controlplanes, err := gatewayutils.ListControlPlanesForGateway(ctx, r.Client, gateway)
	if err != nil {
		return nil, err
	}

	count := len(controlplanes)
	if count > 1 {
		// if there is more than one ControlPlane owned by the same Gateway,
		// delete all of them and recreate only one as follows
		if err := r.Client.DeleteAllOf(ctx, &operatorv1alpha1.DataPlane{},
			client.InNamespace(gateway.Namespace),
			client.MatchingLabels{consts.GatewayOperatorControlledLabel: consts.GatewayManagedLabelValue},
		); err != nil {
			return nil, err
		}
	}

	generatedControlplane := k8sresources.GenerateNewControlPlaneForGateway(gatewayClass, gateway, gatewayConfig, dataplaneName)
	k8sutils.SetOwnerForObject(generatedControlplane, gateway)
	gatewayutils.LabelObjectAsGatewayManaged(generatedControlplane)

	if count == 1 {
		var updated bool
		existingControlplane := &controlplanes[0]
		updated, existingControlplane.ObjectMeta = k8sutils.EnsureObjectMetaIsUpdated(existingControlplane.ObjectMeta, generatedControlplane.ObjectMeta)

		if gatewayConfig.Spec.ControlPlaneDeploymentOptions != nil {
			if !controlplaneSpecDeepEqual(&existingControlplane.Spec.ControlPlaneDeploymentOptions, gatewayConfig.Spec.ControlPlaneDeploymentOptions) {
				existingControlplane.Spec.ControlPlaneDeploymentOptions = *gatewayConfig.Spec.ControlPlaneDeploymentOptions
				updated = true
			}
		}
		if updated {
			return existingControlplane, r.Client.Update(ctx, existingControlplane)
		}
		return existingControlplane, nil
	}

	return generatedControlplane, r.Client.Create(ctx, generatedControlplane)
}

func (r *GatewayReconciler) ensureGatewayMarkedReady(ctx context.Context, gateway *gatewayv1alpha2.Gateway, dataplane *operatorv1alpha1.DataPlane) error {
	if !gatewayutils.IsGatewayReady(gateway) {
		services, err := k8sutils.ListServicesForOwner(
			ctx,
			r.Client,
			consts.GatewayOperatorControlledLabel,
			consts.DataPlaneManagedLabelValue,
			dataplane.Namespace,
			dataplane.UID,
		)
		if err != nil {
			return err
		}

		count := len(services)
		if count > 1 {
			return fmt.Errorf("found %d services for DataPlane currently unsupported: expected 1 or less", count)
		}

		if count == 0 {
			return fmt.Errorf("no services found for dataplane %s/%s", dataplane.Namespace, dataplane.Name)
		}
		svc := services[0]
		if svc.Spec.ClusterIP == "" {
			return fmt.Errorf("service %s doesn't have a ClusterIP yet, not ready", svc.Name)
		}

		gatewayIPs := make([]string, 0)
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			gatewayIPs = append(gatewayIPs, svc.Status.LoadBalancer.Ingress[0].IP) // TODO: handle hostnames https://github.com/Kong/gateway-operator/issues/24
		}

		newAddresses := make([]gatewayv1alpha2.GatewayAddress, 0, len(gatewayIPs))
		ipaddrT := gatewayv1alpha2.IPAddressType
		for _, ip := range append(gatewayIPs, svc.Spec.ClusterIP) {
			newAddresses = append(newAddresses, gatewayv1alpha2.GatewayAddress{
				Type:  &ipaddrT,
				Value: ip,
			})
		}

		gateway.Status.Addresses = newAddresses

		gateway = gatewayutils.PruneGatewayStatusConds(gateway)
		newConditions := make([]metav1.Condition, 0, len(gateway.Status.Conditions))
		newConditions = append(newConditions, metav1.Condition{
			Type:               string(gatewayv1alpha2.GatewayConditionReady),
			Status:             metav1.ConditionTrue,
			ObservedGeneration: gateway.Generation,
			LastTransitionTime: metav1.Now(),
			Reason:             string(gatewayv1alpha2.GatewayReasonReady),
		})
		gateway.Status.Conditions = newConditions
		return r.Client.Status().Update(ctx, gateway)
	}

	return nil
}

func (r *GatewayReconciler) verifyGatewayClassSupport(ctx context.Context, gateway *gatewayv1alpha2.Gateway) (*gatewayv1alpha2.GatewayClass, error) {
	if gateway.Spec.GatewayClassName == "" {
		return nil, operatorerrors.ErrUnsupportedGateway
	}

	gwc := new(gatewayv1alpha2.GatewayClass)
	if err := r.Client.Get(ctx, client.ObjectKey{Name: string(gateway.Spec.GatewayClassName)}, gwc); err != nil {
		return nil, err
	}

	if string(gwc.Spec.ControllerName) != vars.ControllerName {
		return nil, operatorerrors.ErrUnsupportedGateway
	}

	return gwc, nil
}

func (r *GatewayReconciler) getOrCreateGatewayConfiguration(ctx context.Context, gatewayClass *gatewayv1alpha2.GatewayClass) (*operatorv1alpha1.GatewayConfiguration, error) {
	gatewayConfig, err := r.getGatewayConfigForGatewayClass(ctx, gatewayClass)
	if err != nil {
		if errors.Is(err, operatorerrors.ErrObjectMissingParametersRef) {
			return new(operatorv1alpha1.GatewayConfiguration), nil
		}
		return nil, err
	}

	return gatewayConfig, nil
}

func (r *GatewayReconciler) getGatewayConfigForGatewayClass(ctx context.Context, gatewayClass *gatewayv1alpha2.GatewayClass) (*operatorv1alpha1.GatewayConfiguration, error) {
	if gatewayClass.Spec.ParametersRef == nil {
		return nil, fmt.Errorf("%w, gatewayClass = %s", operatorerrors.ErrObjectMissingParametersRef, gatewayClass.Name)
	}

	if string(gatewayClass.Spec.ParametersRef.Group) != operatorv1alpha1.SchemeGroupVersion.Group ||
		string(gatewayClass.Spec.ParametersRef.Kind) != "GatewayConfiguration" {
		return nil, &k8serrors.StatusError{
			ErrStatus: metav1.Status{
				Status: metav1.StatusFailure,
				Code:   http.StatusBadRequest,
				Reason: metav1.StatusReasonInvalid,
				Details: &metav1.StatusDetails{
					Kind: string(gatewayClass.Spec.ParametersRef.Kind),
					Causes: []metav1.StatusCause{{
						Type: metav1.CauseTypeFieldValueNotSupported,
						Message: fmt.Sprintf("controller only supports %s %s resources for GatewayClass parametersRef",
							operatorv1alpha1.SchemeGroupVersion.Group, "GatewayConfiguration"),
					}},
				},
			}}
	}

	if gatewayClass.Spec.ParametersRef.Namespace == nil ||
		*gatewayClass.Spec.ParametersRef.Namespace == "" ||
		gatewayClass.Spec.ParametersRef.Name == "" {
		return nil, fmt.Errorf("GatewayClass %s has invalid ParametersRef: both namespace and name must be provided", gatewayClass.Name)
	}

	gatewayConfig := new(operatorv1alpha1.GatewayConfiguration)
	return gatewayConfig, r.Client.Get(ctx, client.ObjectKey{
		Namespace: string(*gatewayClass.Spec.ParametersRef.Namespace),
		Name:      gatewayClass.Spec.ParametersRef.Name,
	}, gatewayConfig)
}
