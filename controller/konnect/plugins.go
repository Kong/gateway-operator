package konnect

import (
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kong/gateway-operator/pkg/consts"
)

var kongPluginsAnnotationChangedPredicate = predicate.Funcs{
	CreateFunc: func(e event.TypedCreateEvent[client.Object]) bool {
		_, ok := e.Object.GetAnnotations()[consts.PluginsAnnotationKey]
		return ok
	},
	UpdateFunc: func(e event.TypedUpdateEvent[client.Object]) bool {
		if e.ObjectOld == nil || e.ObjectNew == nil {
			return false
		}
		return e.ObjectNew.GetAnnotations()[consts.PluginsAnnotationKey] !=
			e.ObjectOld.GetAnnotations()[consts.PluginsAnnotationKey]
	},
	DeleteFunc: func(e event.TypedDeleteEvent[client.Object]) bool {
		_, ok := e.Object.GetAnnotations()[consts.PluginsAnnotationKey]
		return ok
	},
}

func ownerRefIsAnyKongPlugin(obj client.Object) bool {
	return lo.ContainsBy(
		obj.GetOwnerReferences(),
		func(ownerRef metav1.OwnerReference) bool {
			return ownerRef.Kind == "KongPlugin" ||
				// NOTE: We currently do not support KongClusterPlugin, but we keep this here for future use.
				ownerRef.Kind == "KongClusterPlugin"
		},
	)
}
