package utils

import (
	"fmt"
	"hash/fnv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/hash"
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

func Hash(obj interface{}) string {
	hasher := fnv.New32a()
	hash.DeepHashObject(hasher, obj)
	hashValue := fmt.Sprintf("%x", hasher.Sum32())
	return hashValue
}
