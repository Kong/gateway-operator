package kubernetes

import (
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
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
	existingMeta metav1.ObjectMeta,
	generatedMeta metav1.ObjectMeta,
	options ...func(existingMeta, generatedMeta metav1.ObjectMeta) (bool, metav1.ObjectMeta),
) (toUpdate bool, updatedMeta metav1.ObjectMeta) {
	var metaToUpdate bool

	// compare and enforce labels
	if !maps.Equal(existingMeta.Labels, generatedMeta.Labels) {
		existingMeta.SetLabels(generatedMeta.GetLabels())
		metaToUpdate = true
	}

	// compare and enforce ownerReferences
	if !slices.EqualFunc(existingMeta.OwnerReferences, generatedMeta.OwnerReferences, func(newObjRef metav1.OwnerReference, genObjRef metav1.OwnerReference) bool {
		sameController := (newObjRef.Controller != nil && genObjRef.Controller != nil && *newObjRef.Controller == *genObjRef.Controller) ||
			(newObjRef.Controller == nil && genObjRef.Controller == nil)
		sameBlockOwnerDeletion := (newObjRef.BlockOwnerDeletion != nil && genObjRef.BlockOwnerDeletion != nil && *newObjRef.BlockOwnerDeletion == *genObjRef.BlockOwnerDeletion) ||
			(newObjRef.BlockOwnerDeletion == nil && genObjRef.BlockOwnerDeletion == nil)
		return newObjRef.APIVersion == genObjRef.APIVersion &&
			newObjRef.Kind == genObjRef.Kind &&
			newObjRef.Name == genObjRef.Name &&
			newObjRef.UID == genObjRef.UID &&
			sameController &&
			sameBlockOwnerDeletion
	}) {
		existingMeta.SetOwnerReferences(generatedMeta.GetOwnerReferences())
		metaToUpdate = true
	}

	// apply all the passed options
	for _, opt := range options {
		var changed bool
		changed, existingMeta = opt(existingMeta, generatedMeta)
		if changed {
			metaToUpdate = true
		}
	}

	return metaToUpdate, existingMeta
}

// TrimGenerateName cut the string to 63 chars, in case it is longer,
// to be compliant with the GenerateName length maximum size of 63 chars.
func TrimGenerateName(name string) string {
	if len(name) > 63 {
		name = name[:62]
	}
	if !strings.HasSuffix(name, "-") {
		name += "-"
	}
	return name
}
