/*
Copyright 2022.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&GatewayConfiguration{}, &GatewayConfigurationList{})
}

//+kubebuilder:object:root=true

// GatewayConfigurationList contains a list of GatewayConfiguration
type GatewayConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GatewayConfiguration `json:"items"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GatewayConfiguration is the Schema for the gatewayconfigurations API
type GatewayConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewayConfigurationSpec   `json:"spec,omitempty"`
	Status GatewayConfigurationStatus `json:"status,omitempty"`
}

// GatewayConfigurationSpec defines the desired state of GatewayConfiguration
type GatewayConfigurationSpec struct {
	// ContainerImage indicates the image name to use for the underlying Gateway
	// Deployment and pairs with the spec.Version option to allow specifying the
	// version of the image to use.
	//
	// +optional
	// +kubebuilder:default=DefaultGatewayContainerImage
	ContainerImage *string `json:"containerImage,omitempty"`

	// Version indicates the desired version of the ControlPlane which will also
	// correspond to the tag used for the ContainerImage.
	//
	// Not available when AutomaticUpgrades is in use.
	//
	// If omitted, the latest stable version will be used.
	//
	// +optional
	Version *string `json:"version,omitempty"`

	// Env provides environment variables that will be distributed to any Gateway
	// which is attached to this Configuration.
	//
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Env provides environment variables that will be distributed to any Gateway
	// which is attached to this Configuration with the values for for those
	// environment variables coming from a specified source.
	//
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`
}

// GatewayConfigurationStatus defines the observed state of GatewayConfiguration
type GatewayConfigurationStatus struct{}
