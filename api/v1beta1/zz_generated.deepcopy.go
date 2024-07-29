//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"github.com/kong/gateway-operator/api/v1alpha1"
	"k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/gateway-api/apis/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Address) DeepCopyInto(out *Address) {
	*out = *in
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(AddressType)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Address.
func (in *Address) DeepCopy() *Address {
	if in == nil {
		return nil
	}
	out := new(Address)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BlueGreenStrategy) DeepCopyInto(out *BlueGreenStrategy) {
	*out = *in
	out.Promotion = in.Promotion
	out.Resources = in.Resources
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BlueGreenStrategy.
func (in *BlueGreenStrategy) DeepCopy() *BlueGreenStrategy {
	if in == nil {
		return nil
	}
	out := new(BlueGreenStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlane) DeepCopyInto(out *ControlPlane) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlane.
func (in *ControlPlane) DeepCopy() *ControlPlane {
	if in == nil {
		return nil
	}
	out := new(ControlPlane)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ControlPlane) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneDeploymentOptions) DeepCopyInto(out *ControlPlaneDeploymentOptions) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.PodTemplateSpec != nil {
		in, out := &in.PodTemplateSpec, &out.PodTemplateSpec
		*out = new(corev1.PodTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneDeploymentOptions.
func (in *ControlPlaneDeploymentOptions) DeepCopy() *ControlPlaneDeploymentOptions {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneDeploymentOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneList) DeepCopyInto(out *ControlPlaneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ControlPlane, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneList.
func (in *ControlPlaneList) DeepCopy() *ControlPlaneList {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ControlPlaneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneOptions) DeepCopyInto(out *ControlPlaneOptions) {
	*out = *in
	in.Deployment.DeepCopyInto(&out.Deployment)
	if in.DataPlane != nil {
		in, out := &in.DataPlane, &out.DataPlane
		*out = new(string)
		**out = **in
	}
	if in.Extensions != nil {
		in, out := &in.Extensions, &out.Extensions
		*out = make([]v1alpha1.ExtensionRef, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneOptions.
func (in *ControlPlaneOptions) DeepCopy() *ControlPlaneOptions {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneSpec) DeepCopyInto(out *ControlPlaneSpec) {
	*out = *in
	in.ControlPlaneOptions.DeepCopyInto(&out.ControlPlaneOptions)
	if in.GatewayClass != nil {
		in, out := &in.GatewayClass, &out.GatewayClass
		*out = new(v1.ObjectName)
		**out = **in
	}
	if in.IngressClass != nil {
		in, out := &in.IngressClass, &out.IngressClass
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneSpec.
func (in *ControlPlaneSpec) DeepCopy() *ControlPlaneSpec {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControlPlaneStatus) DeepCopyInto(out *ControlPlaneStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControlPlaneStatus.
func (in *ControlPlaneStatus) DeepCopy() *ControlPlaneStatus {
	if in == nil {
		return nil
	}
	out := new(ControlPlaneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlane) DeepCopyInto(out *DataPlane) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlane.
func (in *DataPlane) DeepCopy() *DataPlane {
	if in == nil {
		return nil
	}
	out := new(DataPlane)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataPlane) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneDeploymentOptions) DeepCopyInto(out *DataPlaneDeploymentOptions) {
	*out = *in
	if in.Rollout != nil {
		in, out := &in.Rollout, &out.Rollout
		*out = new(Rollout)
		(*in).DeepCopyInto(*out)
	}
	in.DeploymentOptions.DeepCopyInto(&out.DeploymentOptions)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneDeploymentOptions.
func (in *DataPlaneDeploymentOptions) DeepCopy() *DataPlaneDeploymentOptions {
	if in == nil {
		return nil
	}
	out := new(DataPlaneDeploymentOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneList) DeepCopyInto(out *DataPlaneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DataPlane, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneList.
func (in *DataPlaneList) DeepCopy() *DataPlaneList {
	if in == nil {
		return nil
	}
	out := new(DataPlaneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataPlaneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneNetworkOptions) DeepCopyInto(out *DataPlaneNetworkOptions) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = new(DataPlaneServices)
		(*in).DeepCopyInto(*out)
	}
	if in.KonnectCertificateOptions != nil {
		in, out := &in.KonnectCertificateOptions, &out.KonnectCertificateOptions
		*out = new(KonnectCertificateOptions)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneNetworkOptions.
func (in *DataPlaneNetworkOptions) DeepCopy() *DataPlaneNetworkOptions {
	if in == nil {
		return nil
	}
	out := new(DataPlaneNetworkOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneOptions) DeepCopyInto(out *DataPlaneOptions) {
	*out = *in
	in.Deployment.DeepCopyInto(&out.Deployment)
	in.Network.DeepCopyInto(&out.Network)
	in.Resources.DeepCopyInto(&out.Resources)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneOptions.
func (in *DataPlaneOptions) DeepCopy() *DataPlaneOptions {
	if in == nil {
		return nil
	}
	out := new(DataPlaneOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneResources) DeepCopyInto(out *DataPlaneResources) {
	*out = *in
	if in.PodDisruptionBudget != nil {
		in, out := &in.PodDisruptionBudget, &out.PodDisruptionBudget
		*out = new(PodDisruptionBudget)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneResources.
func (in *DataPlaneResources) DeepCopy() *DataPlaneResources {
	if in == nil {
		return nil
	}
	out := new(DataPlaneResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneRolloutStatus) DeepCopyInto(out *DataPlaneRolloutStatus) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = new(DataPlaneRolloutStatusServices)
		(*in).DeepCopyInto(*out)
	}
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(DataPlaneRolloutStatusDeployment)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneRolloutStatus.
func (in *DataPlaneRolloutStatus) DeepCopy() *DataPlaneRolloutStatus {
	if in == nil {
		return nil
	}
	out := new(DataPlaneRolloutStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneRolloutStatusDeployment) DeepCopyInto(out *DataPlaneRolloutStatusDeployment) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneRolloutStatusDeployment.
func (in *DataPlaneRolloutStatusDeployment) DeepCopy() *DataPlaneRolloutStatusDeployment {
	if in == nil {
		return nil
	}
	out := new(DataPlaneRolloutStatusDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneRolloutStatusServices) DeepCopyInto(out *DataPlaneRolloutStatusServices) {
	*out = *in
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(RolloutStatusService)
		(*in).DeepCopyInto(*out)
	}
	if in.AdminAPI != nil {
		in, out := &in.AdminAPI, &out.AdminAPI
		*out = new(RolloutStatusService)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneRolloutStatusServices.
func (in *DataPlaneRolloutStatusServices) DeepCopy() *DataPlaneRolloutStatusServices {
	if in == nil {
		return nil
	}
	out := new(DataPlaneRolloutStatusServices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneServiceOptions) DeepCopyInto(out *DataPlaneServiceOptions) {
	*out = *in
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]DataPlaneServicePort, len(*in))
		copy(*out, *in)
	}
	in.ServiceOptions.DeepCopyInto(&out.ServiceOptions)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneServiceOptions.
func (in *DataPlaneServiceOptions) DeepCopy() *DataPlaneServiceOptions {
	if in == nil {
		return nil
	}
	out := new(DataPlaneServiceOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneServicePort) DeepCopyInto(out *DataPlaneServicePort) {
	*out = *in
	out.TargetPort = in.TargetPort
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneServicePort.
func (in *DataPlaneServicePort) DeepCopy() *DataPlaneServicePort {
	if in == nil {
		return nil
	}
	out := new(DataPlaneServicePort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneServices) DeepCopyInto(out *DataPlaneServices) {
	*out = *in
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(DataPlaneServiceOptions)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneServices.
func (in *DataPlaneServices) DeepCopy() *DataPlaneServices {
	if in == nil {
		return nil
	}
	out := new(DataPlaneServices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneSpec) DeepCopyInto(out *DataPlaneSpec) {
	*out = *in
	in.DataPlaneOptions.DeepCopyInto(&out.DataPlaneOptions)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneSpec.
func (in *DataPlaneSpec) DeepCopy() *DataPlaneSpec {
	if in == nil {
		return nil
	}
	out := new(DataPlaneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneStatus) DeepCopyInto(out *DataPlaneStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]Address, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RolloutStatus != nil {
		in, out := &in.RolloutStatus, &out.RolloutStatus
		*out = new(DataPlaneRolloutStatus)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneStatus.
func (in *DataPlaneStatus) DeepCopy() *DataPlaneStatus {
	if in == nil {
		return nil
	}
	out := new(DataPlaneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeploymentOptions) DeepCopyInto(out *DeploymentOptions) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Scaling != nil {
		in, out := &in.Scaling, &out.Scaling
		*out = new(Scaling)
		(*in).DeepCopyInto(*out)
	}
	if in.PodTemplateSpec != nil {
		in, out := &in.PodTemplateSpec, &out.PodTemplateSpec
		*out = new(corev1.PodTemplateSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeploymentOptions.
func (in *DeploymentOptions) DeepCopy() *DeploymentOptions {
	if in == nil {
		return nil
	}
	out := new(DeploymentOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfigDataPlaneNetworkOptions) DeepCopyInto(out *GatewayConfigDataPlaneNetworkOptions) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = new(GatewayConfigDataPlaneServices)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfigDataPlaneNetworkOptions.
func (in *GatewayConfigDataPlaneNetworkOptions) DeepCopy() *GatewayConfigDataPlaneNetworkOptions {
	if in == nil {
		return nil
	}
	out := new(GatewayConfigDataPlaneNetworkOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfigDataPlaneOptions) DeepCopyInto(out *GatewayConfigDataPlaneOptions) {
	*out = *in
	in.Deployment.DeepCopyInto(&out.Deployment)
	in.Network.DeepCopyInto(&out.Network)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfigDataPlaneOptions.
func (in *GatewayConfigDataPlaneOptions) DeepCopy() *GatewayConfigDataPlaneOptions {
	if in == nil {
		return nil
	}
	out := new(GatewayConfigDataPlaneOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfigDataPlaneServices) DeepCopyInto(out *GatewayConfigDataPlaneServices) {
	*out = *in
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new(GatewayConfigServiceOptions)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfigDataPlaneServices.
func (in *GatewayConfigDataPlaneServices) DeepCopy() *GatewayConfigDataPlaneServices {
	if in == nil {
		return nil
	}
	out := new(GatewayConfigDataPlaneServices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfigServiceOptions) DeepCopyInto(out *GatewayConfigServiceOptions) {
	*out = *in
	in.ServiceOptions.DeepCopyInto(&out.ServiceOptions)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfigServiceOptions.
func (in *GatewayConfigServiceOptions) DeepCopy() *GatewayConfigServiceOptions {
	if in == nil {
		return nil
	}
	out := new(GatewayConfigServiceOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfiguration) DeepCopyInto(out *GatewayConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfiguration.
func (in *GatewayConfiguration) DeepCopy() *GatewayConfiguration {
	if in == nil {
		return nil
	}
	out := new(GatewayConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GatewayConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfigurationList) DeepCopyInto(out *GatewayConfigurationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GatewayConfiguration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfigurationList.
func (in *GatewayConfigurationList) DeepCopy() *GatewayConfigurationList {
	if in == nil {
		return nil
	}
	out := new(GatewayConfigurationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GatewayConfigurationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfigurationSpec) DeepCopyInto(out *GatewayConfigurationSpec) {
	*out = *in
	if in.DataPlaneOptions != nil {
		in, out := &in.DataPlaneOptions, &out.DataPlaneOptions
		*out = new(GatewayConfigDataPlaneOptions)
		(*in).DeepCopyInto(*out)
	}
	if in.ControlPlaneOptions != nil {
		in, out := &in.ControlPlaneOptions, &out.ControlPlaneOptions
		*out = new(ControlPlaneOptions)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfigurationSpec.
func (in *GatewayConfigurationSpec) DeepCopy() *GatewayConfigurationSpec {
	if in == nil {
		return nil
	}
	out := new(GatewayConfigurationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayConfigurationStatus) DeepCopyInto(out *GatewayConfigurationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayConfigurationStatus.
func (in *GatewayConfigurationStatus) DeepCopy() *GatewayConfigurationStatus {
	if in == nil {
		return nil
	}
	out := new(GatewayConfigurationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalScaling) DeepCopyInto(out *HorizontalScaling) {
	*out = *in
	if in.MinReplicas != nil {
		in, out := &in.MinReplicas, &out.MinReplicas
		*out = new(int32)
		**out = **in
	}
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]v2.MetricSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Behavior != nil {
		in, out := &in.Behavior, &out.Behavior
		*out = new(v2.HorizontalPodAutoscalerBehavior)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalScaling.
func (in *HorizontalScaling) DeepCopy() *HorizontalScaling {
	if in == nil {
		return nil
	}
	out := new(HorizontalScaling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KonnectCertificateOptions) DeepCopyInto(out *KonnectCertificateOptions) {
	*out = *in
	out.Issuer = in.Issuer
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KonnectCertificateOptions.
func (in *KonnectCertificateOptions) DeepCopy() *KonnectCertificateOptions {
	if in == nil {
		return nil
	}
	out := new(KonnectCertificateOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NamespacedName) DeepCopyInto(out *NamespacedName) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NamespacedName.
func (in *NamespacedName) DeepCopy() *NamespacedName {
	if in == nil {
		return nil
	}
	out := new(NamespacedName)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodDisruptionBudget) DeepCopyInto(out *PodDisruptionBudget) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodDisruptionBudget.
func (in *PodDisruptionBudget) DeepCopy() *PodDisruptionBudget {
	if in == nil {
		return nil
	}
	out := new(PodDisruptionBudget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodDisruptionBudgetSpec) DeepCopyInto(out *PodDisruptionBudgetSpec) {
	*out = *in
	if in.MinAvailable != nil {
		in, out := &in.MinAvailable, &out.MinAvailable
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.MaxUnavailable != nil {
		in, out := &in.MaxUnavailable, &out.MaxUnavailable
		*out = new(intstr.IntOrString)
		**out = **in
	}
	if in.UnhealthyPodEvictionPolicy != nil {
		in, out := &in.UnhealthyPodEvictionPolicy, &out.UnhealthyPodEvictionPolicy
		*out = new(policyv1.UnhealthyPodEvictionPolicyType)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodDisruptionBudgetSpec.
func (in *PodDisruptionBudgetSpec) DeepCopy() *PodDisruptionBudgetSpec {
	if in == nil {
		return nil
	}
	out := new(PodDisruptionBudgetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Promotion) DeepCopyInto(out *Promotion) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Promotion.
func (in *Promotion) DeepCopy() *Promotion {
	if in == nil {
		return nil
	}
	out := new(Promotion)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rollout) DeepCopyInto(out *Rollout) {
	*out = *in
	in.Strategy.DeepCopyInto(&out.Strategy)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rollout.
func (in *Rollout) DeepCopy() *Rollout {
	if in == nil {
		return nil
	}
	out := new(Rollout)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RolloutResourcePlan) DeepCopyInto(out *RolloutResourcePlan) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RolloutResourcePlan.
func (in *RolloutResourcePlan) DeepCopy() *RolloutResourcePlan {
	if in == nil {
		return nil
	}
	out := new(RolloutResourcePlan)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RolloutResources) DeepCopyInto(out *RolloutResources) {
	*out = *in
	out.Plan = in.Plan
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RolloutResources.
func (in *RolloutResources) DeepCopy() *RolloutResources {
	if in == nil {
		return nil
	}
	out := new(RolloutResources)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RolloutStatusService) DeepCopyInto(out *RolloutStatusService) {
	*out = *in
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]Address, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RolloutStatusService.
func (in *RolloutStatusService) DeepCopy() *RolloutStatusService {
	if in == nil {
		return nil
	}
	out := new(RolloutStatusService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RolloutStrategy) DeepCopyInto(out *RolloutStrategy) {
	*out = *in
	if in.BlueGreen != nil {
		in, out := &in.BlueGreen, &out.BlueGreen
		*out = new(BlueGreenStrategy)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RolloutStrategy.
func (in *RolloutStrategy) DeepCopy() *RolloutStrategy {
	if in == nil {
		return nil
	}
	out := new(RolloutStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Scaling) DeepCopyInto(out *Scaling) {
	*out = *in
	if in.HorizontalScaling != nil {
		in, out := &in.HorizontalScaling, &out.HorizontalScaling
		*out = new(HorizontalScaling)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Scaling.
func (in *Scaling) DeepCopy() *Scaling {
	if in == nil {
		return nil
	}
	out := new(Scaling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceOptions) DeepCopyInto(out *ServiceOptions) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceOptions.
func (in *ServiceOptions) DeepCopy() *ServiceOptions {
	if in == nil {
		return nil
	}
	out := new(ServiceOptions)
	in.DeepCopyInto(out)
	return out
}
