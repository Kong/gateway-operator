package admission

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	operatorv1beta1 "github.com/kong/gateway-operator/api/v1beta1"
	"github.com/kong/gateway-operator/pkg/consts"
)

func TestHandleDataPlaneValidation(t *testing.T) {
	b := fakeclient.NewClientBuilder()
	b.WithObjects(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test-cm"},
			Data: map[string]string{
				"off": "off",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test-secret"},
			// fake client does not encode fields in StringData to Data,
			// so here we should usebase64 encoded value in Data.
			Data: map[string][]byte{
				"postgres": []byte(base64.StdEncoding.EncodeToString([]byte("postgres"))),
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test-cm-2"},
			// fake client does not encode fields in StringData to Data,
			// so here we should usebase64 encoded value in Data.
			Data: map[string]string{
				"KONG_DATABASE": "xxx",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "test-secret-2"},
			// fake client does not encode fields in StringData to Data,
			// so here we should usebase64 encoded value in Data.
			Data: map[string][]byte{
				"DATABASE": []byte(base64.StdEncoding.EncodeToString([]byte("xxx"))),
			},
		},
	)
	c := b.Build()

	handler := NewRequestHandler(c, logr.Discard())
	server := httptest.NewServer(handler)

	testCases := []struct {
		name      string
		dataplane *operatorv1beta1.DataPlane
		hasError  bool
		errMsg    string
	}{
		{
			name: "validate_ok:dbmode=off",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-off",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												Env: []corev1.EnvVar{
													{
														Name:  consts.EnvVarKongDatabase,
														Value: "off",
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name: "validate_ok:dbmode=empty",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-off",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												Env: []corev1.EnvVar{
													{
														Name:  consts.EnvVarKongDatabase,
														Value: "",
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name: "validate_error:database=postgres",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-postgres",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												Env: []corev1.EnvVar{
													{
														Name:  consts.EnvVarKongDatabase,
														Value: "postgres",
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: true,
			errMsg:   "database backend postgres of DataPlane not supported currently",
		},
		{
			name: "validate_error:database=xxx",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-xxx",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												Env: []corev1.EnvVar{
													{
														Name:  consts.EnvVarKongDatabase,
														Value: "xxx",
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: true,
			errMsg:   "database backend xxx of DataPlane not supported currently",
		},
		{
			name: "validate_ok:db=off_in_configmap",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-off-in-cm",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												Env: []corev1.EnvVar{
													{
														Name: consts.EnvVarKongDatabase,
														ValueFrom: &corev1.EnvVarSource{
															ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
																LocalObjectReference: corev1.LocalObjectReference{Name: "test-cm"},
																Key:                  "off",
															},
														},
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: false,
		},
		{
			name: "validate_error:db=postgres_in_secret",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-postgres-in-secret",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												Env: []corev1.EnvVar{
													{
														Name: consts.EnvVarKongDatabase,
														ValueFrom: &corev1.EnvVarSource{
															SecretKeyRef: &corev1.SecretKeySelector{
																LocalObjectReference: corev1.LocalObjectReference{Name: "test-secret"},
																Key:                  "postgres",
															},
														},
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: true,
			errMsg:   "database backend postgres of DataPlane not supported currently",
		},
		{
			name: "validate_error:db=xxx_in_cm_envFrom",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-xxx-in-cm",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												EnvFrom: []corev1.EnvFromSource{
													{
														Prefix: "",
														ConfigMapRef: &corev1.ConfigMapEnvSource{
															LocalObjectReference: corev1.LocalObjectReference{Name: "test-cm-2"},
														},
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: true,
			errMsg:   "database backend xxx of DataPlane not supported currently",
		},
		{
			name: "validate_ok:db=off_in_secret_envfrom",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-off-in-secret",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
												EnvFrom: []corev1.EnvFromSource{
													{
														Prefix: "KONG_",
														SecretRef: &corev1.SecretEnvSource{
															LocalObjectReference: corev1.LocalObjectReference{Name: "test-secret-2"},
														},
													},
												},
												Image: consts.DefaultDataPlaneImage,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: true,
			errMsg:   "database backend xxx of DataPlane not supported currently",
		},
		{
			name: "validate_error:missing_container_image",
			dataplane: &operatorv1beta1.DataPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-off-in-secret",
					Namespace: "default",
				},
				Spec: operatorv1beta1.DataPlaneSpec{
					DataPlaneOptions: operatorv1beta1.DataPlaneOptions{
						Deployment: operatorv1beta1.DataPlaneDeploymentOptions{
							DeploymentOptions: operatorv1beta1.DeploymentOptions{
								PodTemplateSpec: &corev1.PodTemplateSpec{
									Spec: corev1.PodSpec{
										Containers: []corev1.Container{
											{
												Name: consts.DataPlaneProxyContainerName,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			hasError: true,
			errMsg:   "DataPlane requires an image",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			review := &admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID: "",
					Kind: metav1.GroupVersionKind{
						Group:   operatorv1beta1.SchemeGroupVersion.Group,
						Version: operatorv1beta1.SchemeGroupVersion.Version,
						Kind:    "dataplanes",
					},
					Resource:  dataPlaneGVResource,
					Name:      tc.dataplane.Name,
					Namespace: tc.dataplane.Namespace,
					Operation: admissionv1.Create,
					Object: runtime.RawExtension{
						Object: tc.dataplane,
					},
				},
			}

			buf, err := json.Marshal(review)
			require.NoErrorf(t, err, "there should be error in marshaling into JSON")
			req, err := http.NewRequest("POST", server.URL, bytes.NewReader(buf))
			require.NoError(t, err, "there should be no error in making HTTP request")
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err, "there should be no error in getting response")
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "there should be no error in reading body")
			resp.Body.Close()
			respReview := &admissionv1.AdmissionReview{}
			err = json.Unmarshal(body, respReview)
			require.NoError(t, err, "there should be no error in unmarshalling body")
			validationResp := respReview.Response

			if !tc.hasError {
				// code in http package is in type int, but int32 in Result.Code
				// so EqualValues used instead of Equal
				require.EqualValues(t, http.StatusOK, validationResp.Result.Code, "response code should be 200 OK")
			} else {
				require.EqualValues(t, http.StatusBadRequest, validationResp.Result.Code, "response code should be 400 Bad Request")
				require.Equal(t, validationResp.Result.Message, tc.errMsg, "result message should contain expected content")
			}
		})
	}
}
