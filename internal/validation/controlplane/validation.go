package controlplane

import (
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

// Validator validates ControlPlane objects.
type Validator struct{}

// NewValidator creates a ControlPlane validator.
func NewValidator(c client.Client) *Validator {
	return &Validator{}
}

// Validate validates a ControlPlane object and return the first validation error found.
func (v *Validator) Validate(controlplane *operatorv1beta1.ControlPlane) error {
	if err := v.ValidateDeploymentOptions(&controlplane.Spec.Deployment); err != nil {
		return err
	}

	// prepared for more validations
	return nil
}

// ValidateDeploymentOptions validates the DeploymentOptions field of ControlPlane object.
func (v *Validator) ValidateDeploymentOptions(opts *operatorv1beta1.ControlPlaneDeploymentOptions) error {
	if opts == nil || opts.PodTemplateSpec == nil {
		// Can't use empty DeploymentOptions because we still require users
		// to provide an image
		// Related: https://github.com/Kong/gateway-operator/issues/754.
		return errors.New("ControlPlane requires an image")
	}

	// Ref: https://github.com/Kong/gateway-operator/issues/736
	if opts.Replicas != nil && *opts.Replicas != 1 {
		return errors.New("ControlPlane only supports replicas of 1")
	}

	container := k8sutils.GetPodContainerByName(&opts.PodTemplateSpec.Spec, consts.ControlPlaneControllerContainerName)
	if container == nil {
		// We need the controller container for e.g. specifying an image which
		// is still required.
		// Ref: https://github.com/Kong/gateway-operator/issues/754.
		return errors.New("no controller container found in ControlPlane spec")
	}

	// Ref: https://github.com/Kong/gateway-operator/issues/754.
	if container.Image == "" {
		return errors.New("ControlPlane requires an image")
	}

	return nil
}
