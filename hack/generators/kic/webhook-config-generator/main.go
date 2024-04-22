package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/samber/lo"
	admregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/kong/gateway-operator/hack/generators/kic"
	kicversions "github.com/kong/gateway-operator/internal/versions"
)

const (
	validatingWebhookConfigurationPath                          = "config/webhook/manifests.yaml"
	validatingWebhookConfigurationGeneratorForVersionOutputPath = "pkg/utils/kubernetes/resources/validatingwebhookconfig/zz_generated_kic_%s.go"
	validatingWebhookConfigurationGeneratorMasterOutputPath     = "pkg/utils/kubernetes/resources/zz_generated_kic_validatingwebhookconfig.go"
)

func main() {
	generateHelpersForAllConfiguredVersions()
	generateMasterHelper()
}

// generateHelpersForAllConfiguredVersions iterates over kicversions.ManifestsVersionsForKICVersions map and generates
// GenerateValidatingWebhookConfigurationForKIC_{versionConstraint} function for each configured version.
func generateHelpersForAllConfiguredVersions() {
	for versionConstraint, version := range kicversions.ManifestsVersionsForKICVersions {
		log.Printf("Generating ValidatingWebhook Configuration for KIC versions %s (using manifests: %s)\n", versionConstraint, version)

		// Download KIC-generated ValidatingWebhookConfiguration.
		manifestContent, err := kic.GetFileFromKICRepositoryForVersion(validatingWebhookConfigurationPath, version)
		if err != nil {
			log.Fatalf("Failed to download %s from KIC repository: %s", validatingWebhookConfigurationPath, err)
		}

		// Get rid of the YAML objects separator as we know there's only one ValidatingWebhookConfiguration in the file.
		manifestContent = bytes.ReplaceAll(manifestContent, []byte("---"), []byte(""))

		// Unmarshal the manifest.
		cfg := admregv1.ValidatingWebhookConfiguration{}
		if err := yaml.Unmarshal(manifestContent, &cfg); err != nil {
			log.Fatalf("Failed to unmarshal ValidatingWebhookConfiguration: %s", err)
		}

		// Render template.
		tpl, err := template.New("tpl").Parse(generateValidatingWebhookConfigurationForKICVersionTemplate)
		if err != nil {
			log.Fatalf("Failed to parse 'generateValidatingWebhookConfigurationForKICTemplate' template: %s", err)
		}
		sanitizedConstraint := kic.SanitizeVersionConstraint(versionConstraint)

		// Filter out webhooks that KGO implements on its own.
		filteredWebhooks := lo.Reject(cfg.Webhooks, func(webhook admregv1.ValidatingWebhook, _ int) bool {
			isGateway := lo.ContainsBy(webhook.Rules, func(rule admregv1.RuleWithOperations) bool {
				return lo.ContainsBy(rule.Resources, func(resource string) bool {
					return resource == "gateways"
				}) && lo.ContainsBy(rule.APIGroups, func(apiGroup string) bool {
					return apiGroup == gatewayv1.GroupVersion.Group
				})
			})
			return isGateway
		})
		buf := &bytes.Buffer{}
		if err := tpl.Execute(buf, singleVersionTemplateData{
			SanitizedVersionConstraint: sanitizedConstraint,
			VersionConstraint:          versionConstraint,
			Webhooks:                   filteredWebhooks,
		}); err != nil {
			log.Fatalf("Failed to execute template: %s", err)
		}

		// Write the output to a file.
		outPath := fmt.Sprintf(validatingWebhookConfigurationGeneratorForVersionOutputPath, sanitizedConstraint)
		if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
			log.Fatalf("Failed to write output to %s: %s", outPath, err)
		}
		log.Printf("Successfully generated %s\n", outPath)
	}
}

// generateMasterHelper generates a GenerateValidatingWebhookConfigurationForControlPlane function that is to be used
// to get ValidatingWebhookConfiguration for a dynamically passed KIC version.
func generateMasterHelper() {
	// Prepare a map with constraints mapped to their sanitized forms.
	constraints := lo.SliceToMap(lo.Keys(kicversions.ManifestsVersionsForKICVersions), func(c string) (string, string) {
		return c, kic.SanitizeVersionConstraint(c)
	})

	// Render template.
	tpl, err := template.New("tpl").Parse(generateValidatingWebhookConfigurationForKICVersionMasterTemplate)
	if err != nil {
		log.Fatalf("Failed to parse 'generateValidatingWebhookConfigurationForKICTemplate' template: %s", err)
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, masterTemplateData{
		Constraints: constraints,
	}); err != nil {
		log.Fatalf("Failed to execute template: %s", err)
	}

	// Write the output to a file.
	outPath := validatingWebhookConfigurationGeneratorMasterOutputPath
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		log.Fatalf("Failed to write output to %s: %s", outPath, err)
	}
	log.Printf("Successfully generated %s\n", outPath)
}
