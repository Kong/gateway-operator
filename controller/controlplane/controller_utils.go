package controlplane

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	konnectextensions "github.com/kong/gateway-operator/internal/extensions/konnect"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"
)

func getKonnectExtension(ctx context.Context, cl client.Client, controlplane operatorv1beta1.ControlPlane) (*operatorv1alpha1.KonnectExtension, error) {
	var konnectExtensionRef *operatorv1alpha1.ExtensionRef
	for _, extensionRef := range controlplane.Spec.Extensions {
		if extensionRef.Group == operatorv1alpha1.SchemeGroupVersion.Group && extensionRef.Kind == operatorv1alpha1.KonnectExtensionKind {
			konnectExtensionRef = &extensionRef
			break
		}
	}
	if konnectExtensionRef == nil {
		return nil, nil
	}
	if konnectExtensionRef.Namespace != nil && *konnectExtensionRef.Namespace != controlplane.Namespace {
		return nil, fmt.Errorf("cross-namespace reference to the extension %s/%s is not permitted", *konnectExtensionRef.Namespace, konnectExtensionRef.Name)
	}

	konnectExtension := &operatorv1alpha1.KonnectExtension{}
	if err := cl.Get(ctx, client.ObjectKey{Namespace: controlplane.Namespace, Name: konnectExtensionRef.Name}, konnectExtension); err != nil {
		return nil, fmt.Errorf("failed fetching KonnectExtension %s/%s: %w", controlplane.Namespace, konnectExtension.Name, err)
	}

	return konnectExtension, nil
}

// -----------------------------------------------------------------------------
// ControlPlane - Private Functions - extensions
// -----------------------------------------------------------------------------

// applyExtensions patches the controlplane spec by taking into account customizations from the referenced extensions.
// In case any extension is referenced, it adds a resolvedRefs condition to the controlplane, indicating the status of the
// extension reference. it returns 3 values:
//   - patched: a boolean indicating if the controlplane was patched. If the controlplane was patched, a reconciliation loop will be automatically re-triggered.
//   - requeue: a boolean indicating if the controlplane should be requeued. If the error was unexpected (e.g., because of API server error), the controlplane should be requeued.
//     In case the error is related to a misconfiguration, the controlplane does not need to be requeued, and feedback is provided into the controlplane status.
//   - err: an error in case of failure.
func applyExtensions(ctx context.Context, cl client.Client, logger logr.Logger, controlplane *operatorv1beta1.ControlPlane, konnectEnabled bool) (patched bool, requeue bool, err error) {
	if len(controlplane.Spec.Extensions) == 0 {
		return false, false, nil
	}

	// the konnect extension is the only one implemented at the moment. In case konnect is not enabled, we return early.
	if !konnectEnabled {
		return false, false, nil
	}

	condition := k8sutils.NewConditionWithGeneration(consts.ResolvedRefsType, metav1.ConditionTrue, consts.ResolvedRefsReason, "", controlplane.GetGeneration())
	err = konnectextensions.ApplyControlPlaneKonnectExtension(ctx, cl, controlplane)
	if err != nil {
		switch {
		case errors.Is(err, konnectextensions.ErrCrossNamespaceReference):
			condition.Status = metav1.ConditionFalse
			condition.Reason = string(consts.RefNotPermittedReason)
			condition.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
		case errors.Is(err, konnectextensions.ErrKonnectExtensionNotFound):
			condition.Status = metav1.ConditionFalse
			condition.Reason = string(consts.InvalidExtensionRefReason)
			condition.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
		case errors.Is(err, konnectextensions.ErrClusterCertificateNotFound):
			condition.Status = metav1.ConditionFalse
			condition.Reason = string(consts.InvalidSecretRefReason)
			condition.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
		default:
			return patched, true, err
		}
	}
	newControlPlane := controlplane.DeepCopy()
	k8sutils.SetCondition(condition, newControlPlane)
	patched, patchErr := patchControlPlaneStatus(ctx, cl, logger, newControlPlane)
	if patchErr != nil {
		return false, true, fmt.Errorf("failed patching status for DataPlane %s/%s: %w", controlplane.Namespace, controlplane.Name, patchErr)
	}
	return patched, false, err
}
