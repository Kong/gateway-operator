package fullhybrid

import (
	"context"
	"fmt"

	"github.com/kong/kong-operator/controller/fullhybrid/converter"
	"github.com/kong/kong-operator/pkg/consts"
	k8sutils "github.com/kong/kong-operator/pkg/utils/kubernetes"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnforceState[t converter.RootObject](ctx context.Context, cl client.Client, conv converter.APIConverter[t]) (requeue bool, err error) {
	store := conv.GetStore(ctx)
	rootObject := conv.GetRootObject()
	// Get the resources owned by the root object
	resources, err := conv.ListExistingObjects(ctx, rootObject)
	if err != nil {
		return true, err
	}
	ownedResourceMap, err := mapOwnedResources(rootObject, resources)
	for _, expectedObject := range store {
		expectedHash := expectedObject.GetLabels()[consts.GatewayOperatorHashSpecLabel]
		existingObject, found := ownedResourceMap[expectedHash]
		if !found {
			if err := cl.Create(ctx, expectedObject); err != nil {
				return true, err
			}
		}

		// TODO: ensure the spec is up to date
		fmt.Println(existingObject)
	}

	return false, nil
}

func mapOwnedResources(owner client.Object, resources []unstructured.Unstructured) (map[string]unstructured.Unstructured, error) {
	OwnerRef := k8sutils.GenerateOwnerReferenceForObject(owner)
	hasOwnerRef := func(r unstructured.Unstructured) bool {
		refs, found, err := unstructured.NestedSlice(r.Object, "metadata", "ownerReferences")
		if !found || err != nil {
			return false
		}
		for _, ref := range refs {
			refMap, ok := ref.(map[string]interface{})
			if !ok {
				continue
			}
			if refMap["uid"] == string(OwnerRef.UID) &&
				refMap["kind"] == OwnerRef.Kind &&
				refMap["name"] == OwnerRef.Name &&
				refMap["apiVersion"] == OwnerRef.APIVersion {
				return true
			}
		}
		return false
	}

	return lo.FilterSliceToMap(resources, func(r unstructured.Unstructured) (string, unstructured.Unstructured, bool) {
		toKeep := hasOwnerRef(r)
		labels := r.GetLabels()
		if len(labels) == 0 || labels[consts.GatewayOperatorHashSpecLabel] == "" {
			return "", r, false
		}
		hashLabel := r.GetLabels()[consts.GatewayOperatorHashSpecLabel]
		return hashLabel, r, toKeep
	}), nil
}
