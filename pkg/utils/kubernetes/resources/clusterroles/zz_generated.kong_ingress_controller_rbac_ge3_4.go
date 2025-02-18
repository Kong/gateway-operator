// This file is generated by /hack/generators/kic/role-generator. DO NOT EDIT.

package clusterroles

// -----------------------------------------------------------------------------
// Kong Ingress Controller - RBAC
// -----------------------------------------------------------------------------

//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=list;watch

//+kubebuilder:rbac:groups=core,resources=configmaps;nodes;secrets,verbs=list;watch
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=core,resources=pods;services,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get;patch;update
//+kubebuilder:rbac:groups=configuration.konghq.com,resources=ingressclassparameterses;kongclusterplugins;kongconsumergroups;kongconsumers;kongcustomentities;kongingresses;konglicenses;kongplugins;kongupstreampolicies;kongvaults;tcpingresses;udpingresses,verbs=get;list;watch
//+kubebuilder:rbac:groups=configuration.konghq.com,resources=kongclusterplugins/status;kongconsumergroups/status;kongconsumers/status;kongcustomentities/status;kongingresses/status;konglicenses/status;kongplugins/status;kongupstreampolicies/status;kongvaults/status;tcpingresses/status;udpingresses/status,verbs=get;patch;update
//+kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;list;watch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;list;watch
//+kubebuilder:rbac:groups=incubator.ingress-controller.konghq.com,resources=kongservicefacades,verbs=get;list;watch
//+kubebuilder:rbac:groups=incubator.ingress-controller.konghq.com,resources=kongservicefacades/status,verbs=get;patch;update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses;ingresses,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;patch;update

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=core,resources=configmaps;namespaces;secrets;services,verbs=get;list;watch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=backendtlspolicies;gatewayclasses;grpcroutes;httproutes;referencegrants;tcproutes;tlsroutes;udproutes,verbs=get;list;watch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=backendtlspolicies/status,verbs=patch;update
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gatewayclasses/status;gateways/status;httproutes/status;tcproutes/status;tlsroutes/status;udproutes/status,verbs=get;update
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=gateways,verbs=get;list;update;watch
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=grpcroutes/status,verbs=get;patch;update
//+kubebuilder:rbac:groups=gateway.networking.k8s.io,resources=referencegrants/status,verbs=get
