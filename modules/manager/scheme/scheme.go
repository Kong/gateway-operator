package scheme

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	configurationv1 "github.com/kong/kubernetes-configuration/v2/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/v2/api/configuration/v1alpha1"
	configurationv1beta1 "github.com/kong/kubernetes-configuration/v2/api/configuration/v1beta1"
	operatorv1alpha1 "github.com/kong/kubernetes-configuration/v2/api/gateway-operator/v1alpha1"
	operatorv1beta1 "github.com/kong/kubernetes-configuration/v2/api/gateway-operator/v1beta1"
	operatorv2alpha1 "github.com/kong/kubernetes-configuration/v2/api/gateway-operator/v2alpha1"
	konnectv1alpha1 "github.com/kong/kubernetes-configuration/v2/api/konnect/v1alpha1"
	konnectv1alpha2 "github.com/kong/kubernetes-configuration/v2/api/konnect/v1alpha2"
)

// Get returns a scheme aware of all types the manager can interact with.
func Get() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(operatorv2alpha1.AddToScheme(scheme))
	utilruntime.Must(operatorv1beta1.AddToScheme(scheme))
	utilruntime.Must(operatorv1alpha1.AddToScheme(scheme))

	utilruntime.Must(gatewayv1.Install(scheme))
	utilruntime.Must(gatewayv1beta1.Install(scheme))

	utilruntime.Must(configurationv1.AddToScheme(scheme))
	utilruntime.Must(configurationv1alpha1.AddToScheme(scheme))
	utilruntime.Must(configurationv1beta1.AddToScheme(scheme))

	utilruntime.Must(konnectv1alpha1.AddToScheme(scheme))
	utilruntime.Must(konnectv1alpha2.AddToScheme(scheme))

	utilruntime.Must(certmanagerv1.AddToScheme(scheme))

	return scheme
}
