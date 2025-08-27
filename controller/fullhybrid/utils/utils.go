package utils

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ToUnstructured(obj client.Object) (*unstructured.Unstructured, error) {
	u := &unstructured.Unstructured{}
	gvk := obj.GetObjectKind().GroupVersionKind()
	u.SetGroupVersionKind(gvk)
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u.Object = objMap
	return u, nil
}
