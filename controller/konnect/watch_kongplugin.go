package konnect

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/gateway-operator/controller/konnect/constraints"
	"github.com/kong/gateway-operator/controller/pkg/log"
	"github.com/kong/gateway-operator/pkg/annotations"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
	configurationv1alpha1 "github.com/kong/kubernetes-configuration/api/configuration/v1alpha1"
	configurationv1beta1 "github.com/kong/kubernetes-configuration/api/configuration/v1beta1"
)

// mapPluginsFromAnnotation enqueue requests for KongPlugins based on
// provided object's annotations.
func mapPluginsFromAnnotation[
	T interface {
		configurationv1alpha1.KongService |
			configurationv1alpha1.KongRoute |
			configurationv1.KongConsumer |
			configurationv1beta1.KongConsumerGroup
		GetTypeName() string
	},
](devMode bool) func(ctx context.Context, obj client.Object) []ctrl.Request {
	return func(ctx context.Context, obj client.Object) []ctrl.Request {
		_, ok := any(obj).(*T)
		if !ok {
			entityTypeName := constraints.EntityTypeName[T]()
			logger := log.GetLogger(ctx, entityTypeName, devMode)
			log.Error(logger,
				fmt.Errorf("cannot cast object to %s", entityTypeName),
				fmt.Sprintf("%s mapping handler", entityTypeName), obj,
			)
			return []ctrl.Request{}
		}

		var (
			namespace = obj.GetNamespace()
			plugins   = annotations.ExtractPlugins(obj)
			requests  = make([]ctrl.Request, 0, len(plugins))
		)

		for _, p := range plugins {
			requests = append(requests, ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: namespace,
					Name:      p,
				},
			})
		}
		return requests
	}
}

// mapKongPluginBindings enqueue requests for KongPlugins referenced by KongPluginBindings in their .spec.pluginRef field.
func (r *KongPluginReconciler) mapKongPluginBindings(ctx context.Context, obj client.Object) []ctrl.Request {
	logger := log.GetLogger(ctx, "KongPlugin", r.developmentMode)
	kongPluginBinding, ok := obj.(*configurationv1alpha1.KongPluginBinding)
	if !ok {
		log.Error(logger, errors.New("cannot cast object to KongPluginBinding"), "KongPluginBinding mapping handler", obj)
		return []ctrl.Request{}
	}

	return []ctrl.Request{
		{
			NamespacedName: types.NamespacedName{
				Namespace: kongPluginBinding.Namespace,
				Name:      kongPluginBinding.Spec.PluginReference.Name,
			},
		},
	}
}
