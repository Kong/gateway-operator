package konnect

import (
	"context"
	"errors"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	extensionserrors "github.com/kong/gateway-operator/controller/pkg/extensions/errors"
	k8sutils "github.com/kong/gateway-operator/pkg/utils/kubernetes"

	commonv1alpha1 "github.com/kong/kubernetes-configuration/api/common/v1alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/api/konnect/v1alpha1"
)

func getExtension(ctx context.Context, cl client.Client, objNamespace string, extRef commonv1alpha1.ExtensionRef) (*konnectv1alpha1.KonnectExtension, error) {
	if extRef.Group != konnectv1alpha1.SchemeGroupVersion.Group || extRef.Kind != konnectv1alpha1.KonnectExtensionKind {
		return nil, nil
	}

	if extRef.Namespace != nil && *extRef.Namespace != objNamespace {
		return nil, errors.Join(extensionserrors.ErrCrossNamespaceReference, fmt.Errorf("the cross-namespace reference to the extension %s/%s is not permitted", *extRef.Namespace, extRef.Name))
	}

	konnectExt := konnectv1alpha1.KonnectExtension{}
	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: objNamespace,
		Name:      extRef.Name,
	}, &konnectExt); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errors.Join(extensionserrors.ErrKonnectExtensionNotFound, fmt.Errorf("the extension %s/%s is not found", objNamespace, extRef.Name))
		}
		return nil, err
	}
	if !k8sutils.HasConditionTrue(konnectv1alpha1.KonnectExtensionReadyConditionType, &konnectExt) {
		return nil, errors.Join(extensionserrors.ErrKonnectExtensionNotReady, fmt.Errorf("the extension %s/%s is not ready", objNamespace, extRef.Name))
	}

	return &konnectExt, nil
}
