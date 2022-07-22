package kubernetes

import (
	"fmt"
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// -----------------------------------------------------------------------------
// Kubernetes Utils - Object Metadata
// -----------------------------------------------------------------------------

// GetAPIVersionForObject provides the string of the full group and version for
// the provided object, e.g. "apps/v1"
func GetAPIVersionForObject(obj client.Object) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Group, obj.GetObjectKind().GroupVersionKind().Version)
}

// EnsureObjectMetaIsUpdated ensures that the existing object metadata has
// all the needed fields set. The source of truth is the second argument of
// the function, a generated object metadata.
func EnsureObjectMetaIsUpdated(
	existingObjMeta metav1.ObjectMeta,
	generatedObjMeta metav1.ObjectMeta,
) (toUpdate bool, updatedMeta metav1.ObjectMeta) {
	newObjectMeta := existingObjMeta.DeepCopy()
	newObjectMeta.SetOwnerReferences(generatedObjMeta.GetOwnerReferences())

	for k, v := range generatedObjMeta.GetLabels() {
		newObjectMeta.Labels[k] = v
	}

	if !reflect.DeepEqual(existingObjMeta.OwnerReferences, newObjectMeta.OwnerReferences) ||
		!reflect.DeepEqual(existingObjMeta.Labels, newObjectMeta.Labels) {
		return true, *newObjectMeta
	}

	return false, *newObjectMeta
}
