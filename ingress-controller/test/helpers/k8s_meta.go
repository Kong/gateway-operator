package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kong/kong-operator/ingress-controller/internal/util"
	"github.com/kong/kong-operator/ingress-controller/pkg/manager/scheme"
)

// WithTypeMeta adds type meta to the given object based on its Go type.
func WithTypeMeta[T runtime.Object](t *testing.T, obj T) T {
	err := util.PopulateTypeMeta(obj, scheme.Get())
	require.NoError(t, err)
	return obj
}
