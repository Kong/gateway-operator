// This file is generated by /hack/generators/kic/webhook-config-generator. DO NOT EDIT.

package validatingwebhookconfig

import (
	"github.com/samber/lo"
	admregv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateValidatingWebhookConfigurationForKIC_ge3_1 generates a ValidatingWebhookConfiguration for KIC >=3.1.
func GenerateValidatingWebhookConfigurationForKIC_ge3_1(name string, clientConfig admregv1.WebhookClientConfig) *admregv1.ValidatingWebhookConfiguration {
	return &admregv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []admregv1.ValidatingWebhook{
			{
				Name: "httproutes.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"gateway.networking.k8s.io",
							},
							APIVersions: []string{
								"v1",
								"v1beta1",
							},
							Resources: []string{
								"httproutes",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
			{
				Name: "ingresses.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"networking.k8s.io",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"ingresses",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
			{
				Name: "kongclusterplugins.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"configuration.konghq.com",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"kongclusterplugins",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
			{
				Name: "kongconsumergroups.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"configuration.konghq.com",
							},
							APIVersions: []string{
								"v1beta1",
							},
							Resources: []string{
								"kongconsumergroups",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
			{
				Name: "kongconsumers.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"configuration.konghq.com",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"kongconsumers",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
						},
					},
				},
			},
			{
				Name: "kongingresses.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"configuration.konghq.com",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"kongingresses",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
			{
				Name: "kongplugins.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"configuration.konghq.com",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"kongplugins",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
			{
				Name: "kongvaults.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"configuration.konghq.com",
							},
							APIVersions: []string{
								"v1alpha1",
							},
							Resources: []string{
								"kongvaults",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
			{
				Name: "secrets.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"secrets",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
						},
					},
				},
			},
			{
				Name: "services.validation.ingress-controller.konghq.com",
				ClientConfig: clientConfig,
				FailurePolicy: lo.ToPtr(admregv1.FailurePolicyType("Fail")),
				MatchPolicy: lo.ToPtr(admregv1.MatchPolicyType("Equivalent")),
				SideEffects:   lo.ToPtr(admregv1.SideEffectClass("None")),
				AdmissionReviewVersions: []string{
					"v1",
				},
				Rules: []admregv1.RuleWithOperations{
					{
						Rule: admregv1.Rule{
							APIGroups: []string{
								"",
							},
							APIVersions: []string{
								"v1",
							},
							Resources: []string{
								"services",
							},
						},
						Operations: []admregv1.OperationType{
							"CREATE",
							"UPDATE",
							"DELETE",
						},
					},
				},
			},
		},
	}
}