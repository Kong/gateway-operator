package controlplane

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	admregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakectrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kong/gateway-operator/controller/pkg/op"
	"github.com/kong/gateway-operator/modules/manager/scheme"
	"github.com/kong/gateway-operator/pkg/consts"
	k8sresources "github.com/kong/gateway-operator/pkg/utils/kubernetes/resources"

	operatorv1beta1 "github.com/kong/kubernetes-configuration/api/gateway-operator/v1beta1"
)

func Test_ensureValidatingWebhookConfiguration(t *testing.T) {
	webhookSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "webhook-svc",
		},
	}

	testCases := []struct {
		name    string
		cp      *operatorv1beta1.ControlPlane
		webhook *admregv1.ValidatingWebhookConfiguration

		testBody func(*testing.T, *Reconciler, *operatorv1beta1.ControlPlane)
	}{
		{
			name: "creating validating webhook configuration",
			cp: &operatorv1beta1.ControlPlane{
				TypeMeta: metav1.TypeMeta{
					Kind: "ControlPlane",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cp",
				},
				Spec: operatorv1beta1.ControlPlaneSpec{
					ControlPlaneOptions: operatorv1beta1.ControlPlaneOptions{
						Deployment: operatorv1beta1.ControlPlaneDeploymentOptions{
							Replicas: lo.ToPtr(int32(1)),
							PodTemplateSpec: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										func() corev1.Container {
											c := k8sresources.GenerateControlPlaneContainer(
												k8sresources.GenerateContainerForControlPlaneParams{
													Image:                          consts.DefaultControlPlaneImage,
													AdmissionWebhookCertSecretName: lo.ToPtr("cert-secret"),
												})
											// Envs are set elsewhere so fill in the CONTROLLER_ADMISSION_WEBHOOK_LISTEN
											// here so that the webhook is enabled.
											c.Env = append(c.Env, corev1.EnvVar{
												Name:  "CONTROLLER_ADMISSION_WEBHOOK_LISTEN",
												Value: "0.0.0.0:8080",
											})
											return c
										}(),
									},
								},
							},
						},
					},
				},
			},
			testBody: func(t *testing.T, r *Reconciler, cp *operatorv1beta1.ControlPlane) {
				var (
					ctx      = t.Context()
					webhooks admregv1.ValidatingWebhookConfigurationList
				)
				require.NoError(t, r.List(ctx, &webhooks))
				require.Empty(t, webhooks.Items)

				certSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cert-secret",
					},
					Data: map[string][]byte{
						"ca.crt": []byte("ca"), // dummy
					},
				}

				res, err := r.ensureValidatingWebhookConfiguration(ctx, cp, certSecret, webhookSvc)
				require.NoError(t, err)
				require.Equal(t, op.Created, res)

				require.NoError(t, r.List(ctx, &webhooks))
				require.Len(t, webhooks.Items, 1)

				res, err = r.ensureValidatingWebhookConfiguration(ctx, cp, certSecret, webhookSvc)
				require.NoError(t, err)
				require.Equal(t, op.Noop, res)
			},
		},
		{
			name: "updating validating webhook configuration enforces ObjectMeta",
			cp: &operatorv1beta1.ControlPlane{
				TypeMeta: metav1.TypeMeta{
					Kind: "ControlPlane",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "cp",
				},
				Spec: operatorv1beta1.ControlPlaneSpec{
					ControlPlaneOptions: operatorv1beta1.ControlPlaneOptions{
						Deployment: operatorv1beta1.ControlPlaneDeploymentOptions{
							Replicas: lo.ToPtr(int32(1)),
							PodTemplateSpec: &corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										func() corev1.Container {
											c := k8sresources.GenerateControlPlaneContainer(
												k8sresources.GenerateContainerForControlPlaneParams{
													Image:                          consts.DefaultControlPlaneImage,
													AdmissionWebhookCertSecretName: lo.ToPtr("cert-secret"),
												})
											// Envs are set elsewhere so fill in the CONTROLLER_ADMISSION_WEBHOOK_LISTEN
											// here so that the webhook is enabled.
											c.Env = append(c.Env, corev1.EnvVar{
												Name:  "CONTROLLER_ADMISSION_WEBHOOK_LISTEN",
												Value: "0.0.0.0:8080",
											})
											return c
										}(),
									},
								},
							},
						},
					},
				},
			},
			testBody: func(t *testing.T, r *Reconciler, cp *operatorv1beta1.ControlPlane) {
				var (
					ctx      = t.Context()
					webhooks admregv1.ValidatingWebhookConfigurationList
				)
				require.NoError(t, r.List(ctx, &webhooks))
				require.Empty(t, webhooks.Items)

				certSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cert-secret",
					},
					Data: map[string][]byte{
						"ca.crt": []byte("ca"), // dummy
					},
				}

				res, err := r.ensureValidatingWebhookConfiguration(ctx, cp, certSecret, webhookSvc)
				require.NoError(t, err)
				require.Equal(t, op.Created, res)

				require.NoError(t, r.List(ctx, &webhooks))
				require.Len(t, webhooks.Items, 1, "webhook configuration should be created")

				res, err = r.ensureValidatingWebhookConfiguration(ctx, cp, certSecret, webhookSvc)
				require.NoError(t, err)
				require.Equal(t, op.Noop, res)

				t.Log("updating webhook configuration outside of the controller")
				{
					w := webhooks.Items[0]
					w.Labels["foo"] = "bar"
					require.NoError(t, r.Update(ctx, &w))
				}

				t.Log("running ensureValidatingWebhookConfiguration to enforce ObjectMeta")
				res, err = r.ensureValidatingWebhookConfiguration(ctx, cp, certSecret, webhookSvc)
				require.NoError(t, err)
				require.Equal(t, op.Updated, res)

				require.NoError(t, r.List(ctx, &webhooks))
				require.Len(t, webhooks.Items, 1)
				require.NotContains(t, webhooks.Items[0].Labels, "foo",
					"labels should be updated by the controller so that changes applied by 3rd parties are overwritten",
				)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := fakectrlruntimeclient.
				NewClientBuilder().
				WithScheme(scheme.Get()).
				WithObjects(tc.cp).
				Build()

			r := &Reconciler{
				Client: fakeClient,
			}

			tc.testBody(t, r, tc.cp)
		})
	}
}

func TestEnsureReferenceGrantsForNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		cp        *operatorv1beta1.ControlPlane
		refGrants []client.Object
		wantErr   bool
	}{
		{
			name:      "reference grant exists and matches",
			namespace: "test-ns",
			cp: &operatorv1beta1.ControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cp",
					Namespace: "cp-ns",
				},
			},
			refGrants: []client.Object{
				&gatewayv1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "valid-grant",
						Namespace: "test-ns",
					},
					Spec: gatewayv1beta1.ReferenceGrantSpec{
						From: []gatewayv1beta1.ReferenceGrantFrom{
							{
								Group:     "gateway-operator.konghq.com",
								Kind:      "ControlPlane",
								Namespace: "cp-ns",
							},
						},
						To: []gatewayv1beta1.ReferenceGrantTo{
							{
								Group: "",
								Kind:  "Namespace",
								Name:  lo.ToPtr(gatewayv1beta1.ObjectName("test-ns")),
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:      "reference grant doesn't exist",
			namespace: "test-ns",
			cp: &operatorv1beta1.ControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cp",
					Namespace: "cp-ns",
				},
			},
			refGrants: []client.Object{},
			wantErr:   true,
		},
		{
			name:      "reference grant exists but from namespace doesn't match",
			namespace: "test-ns",
			cp: &operatorv1beta1.ControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cp",
					Namespace: "cp-ns",
				},
			},
			refGrants: []client.Object{
				&gatewayv1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-from-namespace",
						Namespace: "test-ns",
					},
					Spec: gatewayv1beta1.ReferenceGrantSpec{
						From: []gatewayv1beta1.ReferenceGrantFrom{
							{
								Group:     "gateway-operator.konghq.com",
								Kind:      "ControlPlane",
								Namespace: "wrong-namespace",
							},
						},
						To: []gatewayv1beta1.ReferenceGrantTo{
							{
								Group: "",
								Kind:  "Namespace",
								Name:  lo.ToPtr(gatewayv1beta1.ObjectName("test-ns")),
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "reference grant exists but to name doesn't match",
			namespace: "test-ns",
			cp: &operatorv1beta1.ControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cp",
					Namespace: "cp-ns",
				},
			},
			refGrants: []client.Object{
				&gatewayv1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-to-name",
						Namespace: "test-ns",
					},
					Spec: gatewayv1beta1.ReferenceGrantSpec{
						From: []gatewayv1beta1.ReferenceGrantFrom{
							{
								Group:     "gateway-operator.konghq.com",
								Kind:      "ControlPlane",
								Namespace: "cp-ns",
							},
						},
						To: []gatewayv1beta1.ReferenceGrantTo{
							{
								Group: "",
								Kind:  "Namespace",
								Name:  lo.ToPtr(gatewayv1beta1.ObjectName("wrong-name")),
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:      "multiple reference grants with only one valid",
			namespace: "test-ns",
			cp: &operatorv1beta1.ControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cp",
					Namespace: "cp-ns",
				},
			},
			refGrants: []client.Object{
				&gatewayv1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-grant",
						Namespace: "test-ns",
					},
					Spec: gatewayv1beta1.ReferenceGrantSpec{
						From: []gatewayv1beta1.ReferenceGrantFrom{
							{
								Group:     "wrong.group",
								Kind:      "WrongKind",
								Namespace: "wrong-ns",
							},
						},
						To: []gatewayv1beta1.ReferenceGrantTo{
							{
								Group: "",
								Kind:  "Namespace",
								Name:  lo.ToPtr(gatewayv1beta1.ObjectName("wrong-name")),
							},
						},
					},
				},
				&gatewayv1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "valid-grant",
						Namespace: "test-ns",
					},
					Spec: gatewayv1beta1.ReferenceGrantSpec{
						From: []gatewayv1beta1.ReferenceGrantFrom{
							{
								Group:     "gateway-operator.konghq.com",
								Kind:      "ControlPlane",
								Namespace: "cp-ns",
							},
						},
						To: []gatewayv1beta1.ReferenceGrantTo{
							{
								Group: "",
								Kind:  "Namespace",
								Name:  lo.ToPtr(gatewayv1beta1.ObjectName("test-ns")),
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fakectrlruntimeclient.
				NewClientBuilder().
				WithScheme(scheme.Get()).
				WithObjects(tt.refGrants...).
				Build()

			err := ensureReferenceGrantsForNamespace(t.Context(), fakeClient, tt.cp, tt.namespace)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
