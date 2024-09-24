package konnect

import (
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
)

// KongPluginBindingBuilder helps to build KongPluginBinding objects.
type KongPluginBindingBuilder struct {
	binding *configurationv1alpha1.KongPluginBinding
}

// NewKongPluginBindingBuilder creates a new KongPluginBindingBuilder.
func NewKongPluginBindingBuilder() *KongPluginBindingBuilder {
	return &KongPluginBindingBuilder{
		binding: &configurationv1alpha1.KongPluginBinding{},
	}
}

// WithName sets the name of the KongPluginBinding.
func (b *KongPluginBindingBuilder) WithName(name string) *KongPluginBindingBuilder {
	b.binding.Name = name
	return b
}

// WithGenerateName sets the generate name of the KongPluginBinding.
func (b *KongPluginBindingBuilder) WithGenerateName(name string) *KongPluginBindingBuilder {
	b.binding.GenerateName = name
	return b
}

// WithNamespace sets the namespace of the KongPluginBinding.
func (b *KongPluginBindingBuilder) WithNamespace(namespace string) *KongPluginBindingBuilder {
	b.binding.Namespace = namespace
	return b
}

// WithPluginRef sets the plugin reference of the KongPluginBinding.
func (b *KongPluginBindingBuilder) WithPluginRef(pluginName string) *KongPluginBindingBuilder {
	b.binding.Spec.PluginReference.Name = pluginName
	return b
}

// WithControlPlaneRef sets the control plane reference of the KongPluginBinding.
func (b *KongPluginBindingBuilder) WithControlPlaneRef(ref *configurationv1alpha1.ControlPlaneRef) *KongPluginBindingBuilder {
	// TODO: Cross check this with other types of ControlPlaneRefs
	// used by Route, Consumer and/or ConsumerGroups that also bind this plugin
	// in this KongPluginBinding spec.
	b.binding.Spec.ControlPlaneRef = ref
	return b
}

// WithServiceTarget sets the service target of the KongPluginBinding.
func (b *KongPluginBindingBuilder) WithServiceTarget(serviceName string) *KongPluginBindingBuilder {
	b.binding.Spec.Targets.ServiceReference = &configurationv1alpha1.TargetRefWithGroupKind{
		Group: configurationv1alpha1.GroupVersion.Group,
		Kind:  "KongService",
		Name:  serviceName,
	}
	return b
}

// WithRouteTarget sets the route target of the KongPluginBinding.
func (b *KongPluginBindingBuilder) WithRouteTarget(routeName string) *KongPluginBindingBuilder {
	b.binding.Spec.Targets.RouteReference = &configurationv1alpha1.TargetRefWithGroupKind{
		Group: configurationv1alpha1.GroupVersion.Group,
		Kind:  "KongRoute",
		Name:  routeName,
	}
	return b
}

// Build returns the KongPluginBinding.
func (b *KongPluginBindingBuilder) Build() *configurationv1alpha1.KongPluginBinding {
	return b.binding
}
