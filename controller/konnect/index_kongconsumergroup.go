package konnect

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	configurationv1beta1 "github.com/kong/kubernetes-configuration/api/configuration/v1beta1"
	"github.com/kong/kubernetes-configuration/pkg/metadata"
)

const (
	// IndexFieldKongConsumerGroupOnPlugin is the index field for KongConsumerGroup -> KongPlugin.
	IndexFieldKongConsumerGroupOnPlugin = "consumerGroupPluginRef"
	// IndexFieldKongConsumerGroupOnKonnectGatewayControlPlane is the index field for KongConsumerGroup -> KonnectGatewayControlPlane.
	IndexFieldKongConsumerGroupOnKonnectGatewayControlPlane = "consumerGroupKonnectGatewayControlPlaneRef"
)

// IndexOptionsForKongConsumerGroup returns required Index options for KongConsumerGroup reconciler.
func IndexOptionsForKongConsumerGroup(cl client.Client) []ReconciliationIndexOption {
	return []ReconciliationIndexOption{
		{
			IndexObject:  &configurationv1beta1.KongConsumerGroup{},
			IndexField:   IndexFieldKongConsumerGroupOnPlugin,
			ExtractValue: kongConsumerGroupReferencesKongPluginsViaAnnotation,
		},
		{
			IndexObject:  &configurationv1beta1.KongConsumerGroup{},
			IndexField:   IndexFieldKongConsumerGroupOnKonnectGatewayControlPlane,
			ExtractValue: indexKonnectGatewayControlPlaneRef[configurationv1beta1.KongConsumerGroup](cl),
		},
	}
}

func kongConsumerGroupReferencesKongPluginsViaAnnotation(object client.Object) []string {
	consumerGroup, ok := object.(*configurationv1beta1.KongConsumerGroup)
	if !ok {
		return nil
	}
	return metadata.ExtractPluginsWithNamespaces(consumerGroup)
}
