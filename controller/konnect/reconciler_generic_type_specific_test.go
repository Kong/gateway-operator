package konnect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kong/gateway-operator/modules/manager/scheme"

	configurationv1 "github.com/kong/kubernetes-configuration/api/configuration/v1"
)

func TestHandleKongConsumerSpecific(t *testing.T) {
	t.Run("KongConsumer", func(t *testing.T) {
		testCases := []struct {
			name            string
			consumer        *configurationv1.KongConsumer
			existingSecrets []client.Object
			wantStop        bool
			wantIsProblem   bool
			wantCondition   metav1.Condition
		}{
			{
				name: "no credentials",
				consumer: &configurationv1.KongConsumer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-consumer",
						Namespace: "test",
					},
					Credentials: []string{},
				},
				wantStop:      false,
				wantIsProblem: false,
			},
			{
				name: "all credentials exist",
				consumer: &configurationv1.KongConsumer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-consumer",
						Namespace: "test",
					},
					Credentials: []string{"secret1", "secret2"},
				},
				existingSecrets: []client.Object{
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "secret1",
							Namespace: "test",
						},
					},
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "secret2",
							Namespace: "test",
						},
					},
				},
				wantStop:      false,
				wantIsProblem: false,
			},
			{
				name: "some credentials missing",
				consumer: &configurationv1.KongConsumer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-consumer",
						Namespace: "test",
					},
					Credentials: []string{"secret1", "secret2", "missing-secret"},
				},
				existingSecrets: []client.Object{
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "secret1",
							Namespace: "test",
						},
					},
					&corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "secret2",
							Namespace: "test",
						},
					},
				},
				wantStop:      true,
				wantIsProblem: true,
				wantCondition: metav1.Condition{
					Type:    "CredentialSecretRefValid",
					Status:  metav1.ConditionFalse,
					Reason:  "CredentialSecretRefInvalid",
					Message: "secrets \"missing-secret\" not found",
				},
			},
			{
				name: "all credentials missing",
				consumer: &configurationv1.KongConsumer{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-consumer",
						Namespace: "test",
					},
					Credentials: []string{"missing-secret1", "missing-secret2"},
				},
				wantStop:      true,
				wantIsProblem: true,
				wantCondition: metav1.Condition{
					Type:    "CredentialSecretRefValid",
					Status:  metav1.ConditionFalse,
					Reason:  "CredentialSecretRefInvalid",
					Message: "secrets \"missing-secret1\" not found\nsecrets \"missing-secret2\" not found",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cl := fake.NewClientBuilder().
					WithObjects(tc.existingSecrets...).
					WithScheme(scheme.Get()).
					Build()

				stop, isProblem := handleKongConsumerSpecific(t.Context(), cl, tc.consumer)

				assert.Equal(t, tc.wantStop, stop)
				assert.Equal(t, tc.wantIsProblem, isProblem)

				// Check condition if a problem was expected
				if tc.wantIsProblem {
					hasCondition := false
					for _, cond := range tc.consumer.Status.Conditions {
						if cond.Type == tc.wantCondition.Type {
							hasCondition = true
							assert.Equal(t, tc.wantCondition.Status, cond.Status)
							assert.Equal(t, tc.wantCondition.Reason, cond.Reason)
							assert.Equal(t, tc.wantCondition.Message, cond.Message)
						}
					}
					assert.True(t, hasCondition, "Expected condition not found")
				}
			})
		}
	})
}
