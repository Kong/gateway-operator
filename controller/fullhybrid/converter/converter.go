package converter

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/kong-operator/controller/fullhybrid/utils"
	gwtypes "github.com/kong/kong-operator/internal/types"
)

// APIConverter is an interface that groups the methods needed to convert a
// Kubernetes API object into Kong configuration objects.
type APIConverter[t RootObject] interface {
	GetRootObject() t
	SetRootObject(obj t)
	LoadStore(ctx context.Context) error
	Translate() error
	GetStore(ctx context.Context) []client.Object
	Reduct() []utils.ReductFunc
	ListExistingObjects(ctx context.Context, owner client.Object) ([]unstructured.Unstructured, error)
}

// RootObject is an interface that represents all resource types that can be loaded
// as root by the APIConverter.
type RootObject interface {
	*corev1.Service |
		*gwtypes.HTTPRoute

	client.Object
}
