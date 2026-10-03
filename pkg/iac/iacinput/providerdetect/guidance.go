package providerdetect

import (
	"fmt"
	"strings"

	"github.com/plantonhq/planton/catalog/auth0"
	"github.com/plantonhq/planton/catalog/aws"
	"github.com/plantonhq/planton/catalog/azure"
	"github.com/plantonhq/planton/catalog/cloudflare"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/catalog/gcp"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/catalog/openfga"
	"github.com/plantonhq/planton/catalog/stripe"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// ProviderConfigExample returns an example YAML configuration for the given provider.
func ProviderConfigExample(provider catalogkind.CatalogProvider) string {
	switch provider {
	case catalogkind.CatalogProvider_auth0:
		return auth0.ConfigFileExample
	case catalogkind.CatalogProvider_aws:
		return aws.ConfigFileExample
	case catalogkind.CatalogProvider_azure:
		return azure.ConfigFileExample
	case catalogkind.CatalogProvider_cloudflare:
		return cloudflare.ConfigFileExample
	case catalogkind.CatalogProvider_digital_ocean:
		return digitalocean.ConfigFileExample
	case catalogkind.CatalogProvider_gcp:
		return gcp.ConfigFileExample
	case catalogkind.CatalogProvider_kubernetes:
		return kubernetes.ConfigFileExample
	case catalogkind.CatalogProvider_openfga:
		return openfga.ConfigFileExample
	case catalogkind.CatalogProvider_stripe:
		return stripe.ConfigFileExample
	default:
		return "# Provider config format not available"
	}
}

// ProviderConfigFilename returns the suggested filename for the provider config.
func ProviderConfigFilename(provider catalogkind.CatalogProvider) string {
	switch provider {
	case catalogkind.CatalogProvider_auth0:
		return auth0.ConfigFileName
	case catalogkind.CatalogProvider_aws:
		return aws.ConfigFileName
	case catalogkind.CatalogProvider_azure:
		return azure.ConfigFileName
	case catalogkind.CatalogProvider_cloudflare:
		return cloudflare.ConfigFileName
	case catalogkind.CatalogProvider_digital_ocean:
		return digitalocean.ConfigFileName
	case catalogkind.CatalogProvider_gcp:
		return gcp.ConfigFileName
	case catalogkind.CatalogProvider_kubernetes:
		return kubernetes.ConfigFileName
	case catalogkind.CatalogProvider_openfga:
		return openfga.ConfigFileName
	case catalogkind.CatalogProvider_stripe:
		return stripe.ConfigFileName
	default:
		return "provider-config.yaml"
	}
}

// ProviderEnvironmentVariablesHelp returns the environment variable export commands for the provider.
func ProviderEnvironmentVariablesHelp(provider catalogkind.CatalogProvider) string {
	switch provider {
	case catalogkind.CatalogProvider_auth0:
		return auth0.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_aws:
		return aws.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_azure:
		return azure.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_cloudflare:
		return cloudflare.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_digital_ocean:
		return digitalocean.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_gcp:
		return gcp.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_kubernetes:
		return kubernetes.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_openfga:
		return openfga.EnvironmentVariablesHelp
	case catalogkind.CatalogProvider_stripe:
		return stripe.EnvironmentVariablesHelp
	default:
		return "# Environment variables not available for this provider"
	}
}

// ProviderDocsURL returns the documentation URL for the provider.
func ProviderDocsURL(provider catalogkind.CatalogProvider) string {
	switch provider {
	case catalogkind.CatalogProvider_auth0:
		return auth0.ProviderDocsURL
	case catalogkind.CatalogProvider_aws:
		return aws.ProviderDocsURL
	case catalogkind.CatalogProvider_azure:
		return azure.ProviderDocsURL
	case catalogkind.CatalogProvider_cloudflare:
		return cloudflare.ProviderDocsURL
	case catalogkind.CatalogProvider_digital_ocean:
		return digitalocean.ProviderDocsURL
	case catalogkind.CatalogProvider_gcp:
		return gcp.ProviderDocsURL
	case catalogkind.CatalogProvider_kubernetes:
		return kubernetes.ProviderDocsURL
	case catalogkind.CatalogProvider_openfga:
		return openfga.ProviderDocsURL
	case catalogkind.CatalogProvider_stripe:
		return stripe.ProviderDocsURL
	default:
		return ""
	}
}

// MissingProviderConfigGuidance returns a helpful message when provider config is missing.
// It shows both options: environment variables (default) and explicit config file.
func MissingProviderConfigGuidance(result *DetectionResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("The %s resource requires %s credentials.\n\n",
		result.KindName, ProviderDisplayName(result.Provider)))

	// Option 1: Environment variables (recommended for local development)
	sb.WriteString("Option 1: Set environment variables\n\n")
	envHelp := ProviderEnvironmentVariablesHelp(result.Provider)
	for _, line := range strings.Split(envHelp, "\n") {
		sb.WriteString("  " + line + "\n")
	}

	// Option 2: Provider config file
	sb.WriteString("\nOption 2: Create a provider config file\n\n")
	sb.WriteString(fmt.Sprintf("  Create '%s' with:\n\n",
		ProviderConfigFilename(result.Provider)))

	example := ProviderConfigExample(result.Provider)
	for _, line := range strings.Split(example, "\n") {
		sb.WriteString("    " + line + "\n")
	}

	sb.WriteString("\n  Then run:\n\n")
	sb.WriteString(fmt.Sprintf("    planton plan -f manifest.yaml -p %s\n",
		ProviderConfigFilename(result.Provider)))

	// Add documentation link if available
	docsURL := ProviderDocsURL(result.Provider)
	if docsURL != "" {
		sb.WriteString(fmt.Sprintf("\nFor more information: %s\n", docsURL))
	}

	return sb.String()
}

// KindDetectionErrorGuidance returns a helpful message when kind detection fails.
func KindDetectionErrorGuidance() string {
	return `The manifest must contain valid 'apiVersion' and 'kind' fields:

  apiVersion: gcp.planton.dev/v1alpha1
  kind: GkeCluster
  metadata:
    name: my-cluster
  spec:
    # ... resource configuration

Check your manifest file for:
  - Missing or misspelled 'apiVersion'
  - Missing or misspelled 'kind'
  - Invalid YAML syntax

For supported resource kinds, see: https://planton.ai/docs`
}

// InvalidProviderConfigGuidance returns a helpful message when provider config is invalid.
func InvalidProviderConfigGuidance(result *DetectionResult, parseErr error) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("The provider config file could not be parsed as %s credentials.\n\n",
		ProviderDisplayName(result.Provider)))

	sb.WriteString("Parse error: " + parseErr.Error() + "\n\n")

	sb.WriteString(fmt.Sprintf("Expected format for %s provider config:\n\n",
		ProviderDisplayName(result.Provider)))

	example := ProviderConfigExample(result.Provider)
	for _, line := range strings.Split(example, "\n") {
		sb.WriteString("  " + line + "\n")
	}

	docsURL := ProviderDocsURL(result.Provider)
	if docsURL != "" {
		sb.WriteString(fmt.Sprintf("\nFor more information: %s\n", docsURL))
	}

	return sb.String()
}
