package konnect

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gwtypes "github.com/kong/gateway-operator/internal/types"

	konnectv1alpha2 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha2"
)

// ApplyControlPlaneKonnectExtension gets the DataPlane as argument, and in case it references a KonnectExtension, it
// fetches the referenced extension and applies the necessary changes to the DataPlane spec.
func ApplyControlPlaneKonnectExtension(ctx context.Context, cl client.Client, controlPlane *gwtypes.ControlPlane) (bool, error) {
	var konnectExtension *konnectv1alpha2.KonnectExtension
	for _, extensionRef := range controlPlane.Spec.Extensions {
		extension, err := getExtension(ctx, cl, controlPlane.Namespace, extensionRef)
		if err != nil {
			return false, err
		}
		if extension != nil {
			konnectExtension = extension
			break
		}
	}
	if konnectExtension == nil {
		return false, nil
	}

	// TODO: implement KonnectExtension for ControlPlane v2alpha1: https://github.com/Kong/gateway-operator/issues/1730

	return true, nil
}
