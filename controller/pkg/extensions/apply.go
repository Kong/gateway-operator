package extensions

import (
	"context"
	"errors"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	extensionserrors "github.com/kong/gateway-operator/controller/pkg/extensions/errors"
	"github.com/kong/gateway-operator/controller/pkg/extensions/konnect"
	"github.com/kong/gateway-operator/controller/pkg/patch"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	operatorv1beta1 "github.com/kong/kubernetes-configuration/api/gateway-operator/v1beta1"
)

// ExtendableT is the interface implemented by the objects which implementation
// can be extended through extensions.
type ExtendableT interface {
	client.Object
	withExtensions
	k8sutils.ConditionsAware

	*operatorv1beta1.DataPlane |
		*operatorv1beta1.ControlPlane
}

type withExtensions interface {
	GetExtensions() []commonv1alpha1.ExtensionRef
}

// applyExtensions patches the dataplane or controlplane spec by taking into account customizations from the referenced extensions.
// In case any extension is referenced, it adds a resolvedRefs condition to the dataplane, indicating the status of the
// extension reference. it returns 3 values:
//   - stop: a boolean indicating if the caller must return. It's true when the dataplane status has been patched.
//   - requeue: a boolean indicating if the dataplane should be requeued. If the error was unexpected (e.g., because of API server error), the dataplane should be requeued.
//     In case the error is related to a misconfiguration, the dataplane does not need to be requeued, and feedback is provided into the dataplane status.
//   - err: an error in case of failure.
func ApplyExtensions[t ExtendableT](ctx context.Context, cl client.Client, logger logr.Logger, o t, konnectEnabled bool) (stop bool, res ctrl.Result, err error) {
	// extensionsCondition can be nil. In that case, no extensions are referenced by the object.
	extensionsCondition := validateExtensions(o)
	if extensionsCondition == nil {
		return false, ctrl.Result{}, nil
	}

	if res, err := patch.StatusWithCondition(
		ctx,
		cl,
		o,
		consts.ConditionType(extensionsCondition.Type),
		extensionsCondition.Status,
		consts.ConditionReason(extensionsCondition.Reason),
		extensionsCondition.Message,
	); err != nil || !res.IsZero() {
		return true, res, err
	}
	if extensionsCondition.Status == metav1.ConditionFalse {
		return false, ctrl.Result{}, extensionserrors.ErrInvalidExtensions
	}

	// the konnect extension is the only one implemented at the moment. In case konnect is not enabled, we return early.
	if !konnectEnabled {
		return false, ctrl.Result{}, nil
	}

	// in case the extensionsCondition is true, let's apply the extensions.
	konnectExtensionApplied := k8sutils.NewConditionWithGeneration(consts.KonnectExtensionAppliedType, metav1.ConditionTrue, consts.KonnectExtensionAppliedReason, "The Konnect extension has been successsfully applied", o.GetGeneration())
	if extensionsCondition.Status == metav1.ConditionTrue {
		var (
			extensionRefFound bool
			err               error
		)

		switch obj := any(o).(type) {
		case *operatorv1beta1.DataPlane:
			extensionRefFound, err = konnect.ApplyDataPlaneKonnectExtension(ctx, cl, obj)
		case *operatorv1beta1.ControlPlane:
			extensionRefFound, err = konnect.ApplyControlPlaneKonnectExtension(ctx, cl, obj)
		default:
			return false, ctrl.Result{}, errors.New("unsupported object type")
		}
		if err != nil {
			switch {
			case errors.Is(err, extensionserrors.ErrCrossNamespaceReference):
				konnectExtensionApplied.Status = metav1.ConditionFalse
				konnectExtensionApplied.Reason = string(consts.RefNotPermittedReason)
				konnectExtensionApplied.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
			case errors.Is(err, extensionserrors.ErrKonnectExtensionNotFound):
				konnectExtensionApplied.Status = metav1.ConditionFalse
				konnectExtensionApplied.Reason = string(consts.InvalidExtensionRefReason)
				konnectExtensionApplied.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
			case errors.Is(err, extensionserrors.ErrClusterCertificateNotFound):
				konnectExtensionApplied.Status = metav1.ConditionFalse
				konnectExtensionApplied.Reason = string(consts.InvalidSecretRefReason)
				konnectExtensionApplied.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
			case errors.Is(err, extensionserrors.ErrKonnectExtensionNotReady):
				konnectExtensionApplied.Status = metav1.ConditionFalse
				konnectExtensionApplied.Reason = string(consts.KonnectExtensionNotReadyReason)
				konnectExtensionApplied.Message = strings.ReplaceAll(err.Error(), "\n", " - ")
			default:
				return true, ctrl.Result{}, err
			}
		}
		if !extensionRefFound {
			return false, ctrl.Result{}, nil
		}
	}

	if res, err := patch.StatusWithCondition(
		ctx,
		cl,
		o,
		consts.ConditionType(konnectExtensionApplied.Type),
		konnectExtensionApplied.Status,
		consts.ConditionReason(konnectExtensionApplied.Reason),
		konnectExtensionApplied.Message,
	); err != nil || !res.IsZero() {
		return true, res, err
	}

	return false, ctrl.Result{}, err
}
