package fullhybrid

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kong/kong-operator/controller/fullhybrid/converter"
	"github.com/kong/kong-operator/modules/manager/scheme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	configurationv1alpha1 "github.com/kong/kubernetes-configuration/v2/api/configuration/v1alpha1"
)

func TestGetOwnedResources(t *testing.T) {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-kongservice",
			Namespace: "test-namespace",
			UID:       "12345",
		},
	}
	testCases := []struct {
		name            string
		owner           client.Object
		existingObjects []client.Object
		expectedMap     map[string][]client.Object
	}{
		{
			name:  "no owned resources",
			owner: service,
			existingObjects: []client.Object{
				&configurationv1alpha1.KongService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "owned-resource-1",
						Namespace: "test-namespace",
					},
				},
			},
			expectedMap: map[string][]client.Object{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, obj := range tc.existingObjects {
				controllerutil.SetOwnerReference(tc.owner, obj, scheme.Get(), controllerutil.WithBlockOwnerDeletion(true))
			}
			cl := fake.NewClientBuilder().
				WithScheme(scheme.Get()).
				WithObjects(tc.owner).
				WithObjects(tc.existingObjects...).
				Build()

			conv := converter.NewDummyConverter(cl)
			objects, err := conv.ListExistingObjects(context.Background(), tc.owner)
			assert.NoError(t, err)
			resourceMap, err := mapOwnedResources(tc.owner, objects)
			require.NoError(t, err)
			require.Equal(t, tc.expectedMap, resourceMap)
		})
	}
}
