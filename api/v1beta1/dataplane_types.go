/*
Copyright 2022 Kong Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kong/gateway-operator/api/v1alpha1"
)

func init() {
	SchemeBuilder.Register(&DataPlane{}, &DataPlaneList{})
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=kodp,categories=kong;all
// +kubebuilder:validation:XValidation:message="Extension not allowed for DataPlane",rule="has(self.spec.extensions) ? self.spec.extensions.all(e, e.group == 'gateway-operator.konghq.com' && e.kind == 'DataPlaneKonnectExtension') : true"
// +kubebuilder:printcolumn:name="Ready",description="The Resource is ready",type=string,JSONPath=`.status.conditions[?(@.type=='Ready')].status`

// DataPlane is the Schema for the dataplanes API
type DataPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataPlaneSpec   `json:"spec,omitempty"`
	Status DataPlaneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DataPlaneList contains a list of DataPlane
type DataPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataPlane `json:"items"`
}

// DataPlaneSpec defines the desired state of DataPlane
type DataPlaneSpec struct {
	DataPlaneOptions `json:",inline"`
}

// DataPlaneOptions defines the information specifically needed to
// deploy the DataPlane.
type DataPlaneOptions struct {
	// +optional
	Deployment DataPlaneDeploymentOptions `json:"deployment"`

	// +optional
	Network DataPlaneNetworkOptions `json:"network"`

	// +optional
	Resources DataPlaneResources `json:"resources"`

	// Extensions provide additional or replacement features for the DataPlane
	// resources to influence or enhance functionality.
	// NOTE: since we have one extension only (DataPlaneKonnectExtension), we limit the amount of extensions to 1.
	//
	// +optional
	// +kubebuilder:validation:MinItems=0
	// +kubebuilder:validation:MaxItems=1
	Extensions []v1alpha1.ExtensionRef `json:"extensions,omitempty"`
	// PluginsToInstall is a list of KongPluginInstallation resources that
	// will be installed and available in the DataPlane.
	// +optional
	PluginsToInstall []NamespacedName `json:"pluginsToInstall,omitempty"`
}

// DataPlaneResources defines the resources that will be created and managed
// for the DataPlane.
type DataPlaneResources struct {
	// PodDisruptionBudget is the configuration for the PodDisruptionBudget
	// that will be created for the DataPlane.
	PodDisruptionBudget *PodDisruptionBudget `json:"podDisruptionBudget,omitempty"`
}

// PodDisruptionBudget defines the configuration for the PodDisruptionBudget.
type PodDisruptionBudget struct {
	// Spec defines the specification of the PodDisruptionBudget.
	// Selector is managed by the controller and cannot be set by the user.
	Spec PodDisruptionBudgetSpec `json:"spec,omitempty"`
}

// PodDisruptionBudgetSpec defines the specification of a PodDisruptionBudget.
//
// +kubebuilder:validation:XValidation:message="You can specify only one of maxUnavailable and minAvailable in a single PodDisruptionBudgetSpec.",rule="(has(self.minAvailable) && !has(self.maxUnavailable)) || (!has(self.minAvailable) && has(self.maxUnavailable))"
type PodDisruptionBudgetSpec struct {
	// An eviction is allowed if at least "minAvailable" pods selected by
	// "selector" will still be available after the eviction, i.e. even in the
	// absence of the evicted pod.  So for example you can prevent all voluntary
	// evictions by specifying "100%".
	// +optional
	MinAvailable *intstr.IntOrString `json:"minAvailable,omitempty" protobuf:"bytes,1,opt,name=minAvailable"`

	// An eviction is allowed if at most "maxUnavailable" pods selected by
	// "selector" are unavailable after the eviction, i.e. even in absence of
	// the evicted pod. For example, one can prevent all voluntary evictions
	// by specifying 0. This is a mutually exclusive setting with "minAvailable".
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"bytes,3,opt,name=maxUnavailable"`

	// UnhealthyPodEvictionPolicy defines the criteria for when unhealthy pods
	// should be considered for eviction. Current implementation considers healthy pods,
	// as pods that have status.conditions item with type="Ready",status="True".
	//
	// Valid policies are IfHealthyBudget and AlwaysAllow.
	// If no policy is specified, the default behavior will be used,
	// which corresponds to the IfHealthyBudget policy.
	//
	// IfHealthyBudget policy means that running pods (status.phase="Running"),
	// but not yet healthy can be evicted only if the guarded application is not
	// disrupted (status.currentHealthy is at least equal to status.desiredHealthy).
	// Healthy pods will be subject to the PDB for eviction.
	//
	// AlwaysAllow policy means that all running pods (status.phase="Running"),
	// but not yet healthy are considered disrupted and can be evicted regardless
	// of whether the criteria in a PDB is met. This means perspective running
	// pods of a disrupted application might not get a chance to become healthy.
	// Healthy pods will be subject to the PDB for eviction.
	//
	// Additional policies may be added in the future.
	// Clients making eviction decisions should disallow eviction of unhealthy pods
	// if they encounter an unrecognized policy in this field.
	//
	// This field is beta-level. The eviction API uses this field when
	// the feature gate PDBUnhealthyPodEvictionPolicy is enabled (enabled by default).
	// +optional
	UnhealthyPodEvictionPolicy *policyv1.UnhealthyPodEvictionPolicyType `json:"unhealthyPodEvictionPolicy,omitempty" protobuf:"bytes,4,opt,name=unhealthyPodEvictionPolicy"`
}

// DataPlaneDeploymentOptions specifies options for the Deployments (as in the Kubernetes
// resource "Deployment") which are created and managed for the DataPlane resource.
type DataPlaneDeploymentOptions struct {
	// Rollout describes a custom rollout strategy.
	//
	// +optional
	Rollout *Rollout `json:"rollout,omitempty"`

	DeploymentOptions `json:",inline"`
}

// DataPlaneNetworkOptions defines network related options for a DataPlane.
type DataPlaneNetworkOptions struct {
	// Services indicates the configuration of Kubernetes Services needed for
	// the topology of various forms of traffic (including ingress, e.t.c.) to
	// and from the DataPlane.
	Services *DataPlaneServices `json:"services,omitempty"`

	// KonnectCA is the certificate authority that the operator uses to provision client certificates the DataPlane
	// will use to authenticate itself to the Konnect API. Requires Enterprise.
	//
	// +optional
	KonnectCertificateOptions *KonnectCertificateOptions `json:"konnectCertificate,omitempty"`
}

// DataPlaneServices contains Services related DataPlane configuration, shared with the GatewayConfiguration.
type DataPlaneServices struct {
	// Ingress is the Kubernetes Service that will be used to expose ingress
	// traffic for the DataPlane. Here you can determine whether the DataPlane
	// will be exposed outside the cluster (e.g. using a LoadBalancer type
	// Services) or only internally (e.g. ClusterIP), and inject any additional
	// annotations you need on the service (for instance, if you need to
	// influence a cloud provider LoadBalancer configuration).
	//
	// +optional
	Ingress *DataPlaneServiceOptions `json:"ingress,omitempty"`
}

// DataPlaneServiceOptions contains Services related DataPlane configuration.
type DataPlaneServiceOptions struct {
	// Ports defines the list of ports that are exposed by the service.
	// The ports field allows defining the name, port and targetPort of
	// the underlying service ports, while the protocol is defaulted to TCP,
	// as it is the only protocol currently supported.
	Ports []DataPlaneServicePort `json:"ports,omitempty"`

	// ServiceOptions is the struct containing service options shared with
	// the GatewayConfiguration.
	ServiceOptions `json:",inline"`
}

// DataPlaneServicePort contains information on service's port.
type DataPlaneServicePort struct {
	// The name of this port within the service. This must be a DNS_LABEL.
	// All ports within a ServiceSpec must have unique names. When considering
	// the endpoints for a Service, this must match the 'name' field in the
	// EndpointPort.
	// Optional if only one ServicePort is defined on this service.
	// +optional
	Name string `json:"name,omitempty"`

	// The port that will be exposed by this service.
	Port int32 `json:"port"`

	// Number or name of the port to access on the pods targeted by the service.
	// Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
	// If this is a string, it will be looked up as a named port in the
	// target Pod's container ports. If this is not specified, the value
	// of the 'port' field is used (an identity map).
	// This field is ignored for services with clusterIP=None, and should be
	// omitted or set equal to the 'port' field.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service
	// +optional
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`
}

// ServiceOptions is used to includes options to customize the ingress service,
// such as the annotations.
type ServiceOptions struct {
	// Type determines how the Service is exposed.
	// Defaults to `LoadBalancer`.
	//
	// Valid options are `LoadBalancer` and `ClusterIP`.
	//
	// `ClusterIP` allocates a cluster-internal IP address for load-balancing
	// to endpoints.
	//
	// `LoadBalancer` builds on NodePort and creates an external load-balancer
	// (if supported in the current cloud) which routes to the same endpoints
	// as the clusterIP.
	//
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	//
	// +optional
	// +kubebuilder:default=LoadBalancer
	// +kubebuilder:validation:Enum=LoadBalancer;ClusterIP
	Type corev1.ServiceType `json:"type,omitempty" protobuf:"bytes,4,opt,name=type,casttype=ServiceType"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	//
	// More info: http://kubernetes.io/docs/user-guide/annotations
	//
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`

	// ExternalTrafficPolicy describes how nodes distribute service traffic they
	// receive on one of the Service's "externally-facing" addresses (NodePorts,
	// ExternalIPs, and LoadBalancer IPs). If set to "Local", the proxy will configure
	// the service in a way that assumes that external load balancers will take care
	// of balancing the service traffic between nodes, and so each node will deliver
	// traffic only to the node-local endpoints of the service, without masquerading
	// the client source IP. (Traffic mistakenly sent to a node with no endpoints will
	// be dropped.) The default value, "Cluster", uses the standard behavior of
	// routing to all endpoints evenly (possibly modified by topology and other
	// features). Note that traffic sent to an External IP or LoadBalancer IP from
	// within the cluster will always get "Cluster" semantics, but clients sending to
	// a NodePort from within the cluster may need to take traffic policy into account
	// when picking a node.
	//
	// More info: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip
	//
	// +optional
	// +kubebuilder:default=Cluster
	// +kubebuilder:validation:Enum=Cluster;Local
	ExternalTrafficPolicy corev1.ServiceExternalTrafficPolicy `json:"externalTrafficPolicy,omitempty"`
}

// DataPlaneStatus defines the observed state of DataPlane
type DataPlaneStatus struct {
	// Conditions describe the status of the DataPlane.
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Service indicates the Service that exposes the DataPlane's configured routes
	Service string `json:"service,omitempty"`

	// Addresses lists the addresses that have actually been bound to the DataPlane.
	//
	// +optional
	Addresses []Address `json:"addresses,omitempty"`

	// Selector contains a unique DataPlane identifier used as a deterministic
	// label selector that is used throughout its dependent resources.
	// This is used e.g. as a label selector for DataPlane's Services, Deployments and PodDisruptionBudgets.
	//
	// +kubebuilder:validation:MaxLength=512
	// +kubebuilder:validation:MinLength=8
	Selector string `json:"selector,omitempty"`

	// ReadyReplicas indicates how many replicas have reported to be ready.
	//
	// +kubebuilder:default=0
	ReadyReplicas int32 `json:"readyReplicas"`

	// Replicas indicates how many replicas have been set for the DataPlane.
	//
	// +kubebuilder:default=0
	Replicas int32 `json:"replicas"`

	// RolloutStatus contains information about the rollout.
	// It is set only if a rollout strategy was configured in the spec.
	//
	// +optional
	RolloutStatus *DataPlaneRolloutStatus `json:"rollout,omitempty"`
}

// DataPlaneRolloutStatus describes the DataPlane rollout status.
type DataPlaneRolloutStatus struct {
	// Services contain the information about the services which are available
	// through which user can access the preview deployment.
	Services *DataPlaneRolloutStatusServices `json:"services,omitempty"`

	// Deployment contains the information about the preview deployment.
	Deployment *DataPlaneRolloutStatusDeployment `json:"deployment,omitempty"`

	// Conditions contains the status conditions about the rollout.
	//
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// GetConditions retrieves the DataPlane Status Conditions
func (d *DataPlaneRolloutStatus) GetConditions() []metav1.Condition {
	if d == nil {
		return nil
	}
	return d.Conditions
}

// SetConditions sets the DataPlane Status Conditions
func (d *DataPlaneRolloutStatus) SetConditions(conditions []metav1.Condition) {
	if d == nil {
		return
	}
	d.Conditions = conditions
}

// DataPlaneRolloutStatusServices describes the status of the services during
// DataPlane rollout.
type DataPlaneRolloutStatusServices struct {
	// Ingress contains the name and the address of the preview service for ingress.
	// Using this service users can send requests that will hit the preview deployment.
	Ingress *RolloutStatusService `json:"ingress,omitempty"`

	// AdminAPI contains the name and the address of the preview service for Admin API.
	// Using this service users can send requests to configure the DataPlane's preview deployment.
	AdminAPI *RolloutStatusService `json:"adminAPI,omitempty"`
}

// DataPlaneRolloutStatusDeployment is a rollout status field which contains
// fields specific for Deployments during the rollout.
type DataPlaneRolloutStatusDeployment struct {
	// Selector is a stable label selector value assigned to a DataPlane rollout
	// status which is used throughout the rollout as a deterministic labels selector
	// for Services and Deployments.
	//
	// +kubebuilder:validation:MaxLength=512
	// +kubebuilder:validation:MinLength=8
	Selector string `json:"selector,omitempty"`
}

// RolloutStatusService is a struct which contains status information about
// services that are exposed as part of the rollout.
type RolloutStatusService struct {
	// Name indicates the name of the service.
	Name string `json:"name"`

	// Addresses contains the addresses of a Service.
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Addresses []Address `json:"addresses,omitempty"`
}

// Address describes an address which can be either an IP address or a hostname.
type Address struct {
	// Type of the address.
	//
	// +optional
	// +kubebuilder:default=IPAddress
	Type *AddressType `json:"type,omitempty"`

	// Value of the address. The validity of the values will depend
	// on the type and support by the controller.
	//
	// Examples: `1.2.3.4`, `128::1`, `my-ip-address`.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Value string `json:"value"`

	// Source type of the address.
	SourceType AddressSourceType `json:"sourceType"`
}

// AddressType defines how a network address is represented as a text string.
//
// Can be one of:
//
// * `IPAddress`
// * `Hostname`
//
// +kubebuilder:validation:Pattern=`^IPAddress|Hostname$`
type AddressType string

const (
	// IPAddressType is a textual representation of a numeric IP address. IPv4
	// addresses must be in dotted-decimal form. IPv6 addresses
	// must be in a standard IPv6 text representation
	// (see [RFC 5952](https://tools.ietf.org/html/rfc5952)).
	//
	// This type is intended for specific addresses. Address ranges are not
	// supported (e.g. you can not use a CIDR range like 127.0.0.0/24 as an
	// IPAddress).
	IPAddressType AddressType = "IPAddress"

	// HostnameAddressType represents a DNS based ingress point. This is similar to the
	// corresponding hostname field in Kubernetes load balancer status. For
	// example, this concept may be used for cloud load balancers where a DNS
	// name is used to expose a load balancer.
	HostnameAddressType AddressType = "Hostname"
)

// AddressSourceType defines the type of source this address represents.
//
// Can be one of:
//
// * `PublicLoadBalancer`
// * `PrivateLoadBalancer`
// * `PublicIP`
// * `PrivateIP`
//
// +kubebuilder:validation:Pattern=`^PublicLoadBalancer|PrivateLoadBalancer|PublicIP|PrivateIP$`
type AddressSourceType string

const (
	// PublicLoadBalancerAddressSourceType represents an address belonging to
	// a public Load Balancer.
	PublicLoadBalancerAddressSourceType AddressSourceType = "PublicLoadBalancer"

	// PrivateLoadBalancerAddressSourceType represents an address belonging to
	// a private Load Balancer.
	PrivateLoadBalancerAddressSourceType AddressSourceType = "PrivateLoadBalancer"

	// PublicIPAddressSourceType represents an address belonging to a public IP.
	PublicIPAddressSourceType AddressSourceType = "PublicIP"

	// PrivateIPAddressSourceType represents an address belonging to a private IP.
	PrivateIPAddressSourceType AddressSourceType = "PrivateIP"
)

// GetConditions retrieves the DataPlane Status Conditions
func (d *DataPlane) GetConditions() []metav1.Condition {
	return d.Status.Conditions
}

// SetConditions sets the DataPlane Status Conditions
func (d *DataPlane) SetConditions(conditions []metav1.Condition) {
	d.Status.Conditions = conditions
}
