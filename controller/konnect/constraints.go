package konnect

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kong/gateway-operator/api/v1alpha1"
)

type SupportedKonnectEntityType interface {
	operatorv1alpha1.KonnectControlPlane
	// TODO: add other types

	GetTypeName() string
}

type EntityType[
	T SupportedKonnectEntityType,
] interface {
	*T

	// Kubernetes Object methods

	GetObjectMeta() metav1.Object
	client.Object

	// Added methods

	GetStatusID() string
	SetKonnectLabels(labels map[string]string)
	GetReconciliationWatchOptions(client.Client) []func(*ctrl.Builder) *ctrl.Builder
	GetKonnectAPIAuthConfigurationRef() operatorv1alpha1.KonnectAPIAuthConfigurationRef
}
