package fullhybrid

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/kong/kong-operator/controller/fullhybrid/converter"
	"github.com/kong/kong-operator/controller/fullhybrid/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func EnforceState[t converter.RootObject](ctx context.Context, cl client.Client, conv converter.APIConverter[t]) (requeue bool, err error) {
	store := conv.GetStore(ctx)

	for _, expectedObject := range store {
		var existingObject client.Object
		if existingObject, requeue, err = ensureSingleObject(ctx, cl, expectedObject, conv); requeue || err != nil {
			return requeue, err
		}

		existing, err := utils.ToUnstructured(existingObject)
		if err != nil {
			return false, err
		}
		expected, err := utils.ToUnstructured(expectedObject)
		if err != nil {
			return false, err
		}
		existingSpec, _, err := unstructured.NestedFieldCopy(existing.Object, "spec")
		if err != nil {
			return false, err
		}
		expectedSpec, _, err := unstructured.NestedFieldCopy(expected.Object, "spec")
		if err != nil {
			return false, err
		}

		if !cmp.Equal(expectedSpec, existingSpec) {
			if err := cl.Patch(ctx, existingObject, client.MergeFrom(expectedObject)); err != nil {
				return false, err
			}
		}
	}

	return false, nil
}

func ensureSingleObject[t converter.RootObject](ctx context.Context, cl client.Client, obj client.Object, conv converter.APIConverter[t]) (object client.Object, requeue bool, err error) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	listGVK := gvk
	listGVK.Kind += "List"
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(listGVK)

	ns := obj.GetNamespace()
	labels := obj.GetLabels()
	opts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels(labels),
	}
	if err := cl.List(ctx, list, opts...); err != nil {
		return nil, false, err
	}

	count := len(list.Items)
	if count > 1 {
		for _, fn := range conv.Reduct() {
			filtered := fn(list.Items)
			for _, objToDelete := range filtered {
				if err := cl.Delete(ctx, &objToDelete); err != nil {
					return nil, true, err
				}
			}
		}
		return nil, true, fmt.Errorf("reduced multiple objects of kind %s/%s", gvk.Kind, ns)
	}
	if count == 0 {
		if err := cl.Create(ctx, obj); err != nil {
			return nil, false, err
		}
		return obj, true, nil
	}

	found := &unstructured.Unstructured{Object: list.Items[0].Object}
	found.SetGroupVersionKind(gvk)
	return found, false, nil
}
