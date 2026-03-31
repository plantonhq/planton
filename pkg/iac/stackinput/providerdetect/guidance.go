package providerdetect

import (
	"fmt"
	"strings"

	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/atlas"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/civo"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/confluent"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/digitalocean"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/openfga"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/snowflake"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
)

// ProviderConfigExample returns an example YAML configuration for the given provider.
func ProviderConfigExample(provider cloudresourcekind.CloudResourceProvider) string {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_atlas:
		return atlas.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_auth0:
		return auth0.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_aws:
		return aws.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_azure:
		return azure.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_civo:
		return civo.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return cloudflare.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_confluent:
		return confluent.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_digital_ocean:
		return digitalocean.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_gcp:
		return gcp.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return kubernetes.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_openfga:
		return openfga.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return snowflake.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_openstack:
		return openstack.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_scaleway:
		return scaleway.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_alicloud:
		return alicloud.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_oci:
		return oci.ConfigFileExample
	case cloudresourcekind.CloudResourceProvider_hetzner_cloud:
		return hetznercloud.ConfigFileExample
	default:
		return "# Provider config format not available"
	}
}

// ProviderConfigFilename returns the suggested filename for the provider config.
func ProviderConfigFilename(provider cloudresourcekind.CloudResourceProvider) string {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_atlas:
		return atlas.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_auth0:
		return auth0.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_aws:
		return aws.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_azure:
		return azure.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_civo:
		return civo.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return cloudflare.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_confluent:
		return confluent.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_digital_ocean:
		return digitalocean.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_gcp:
		return gcp.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return kubernetes.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_openfga:
		return openfga.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return snowflake.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_openstack:
		return openstack.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_scaleway:
		return scaleway.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_alicloud:
		return alicloud.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_oci:
		return oci.ConfigFileName
	case cloudresourcekind.CloudResourceProvider_hetzner_cloud:
		return hetznercloud.ConfigFileName
	default:
		return "provider-config.yaml"
	}
}

// ProviderEnvironmentVariablesHelp returns the environment variable export commands for the provider.
func ProviderEnvironmentVariablesHelp(provider cloudresourcekind.CloudResourceProvider) string {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_atlas:
		return atlas.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_auth0:
		return auth0.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_aws:
		return aws.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_azure:
		return azure.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_civo:
		return civo.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return cloudflare.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_confluent:
		return confluent.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_digital_ocean:
		return digitalocean.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_gcp:
		return gcp.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return kubernetes.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_openfga:
		return openfga.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return snowflake.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_openstack:
		return openstack.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_scaleway:
		return scaleway.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_alicloud:
		return alicloud.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_oci:
		return oci.EnvironmentVariablesHelp
	case cloudresourcekind.CloudResourceProvider_hetzner_cloud:
		return hetznercloud.EnvironmentVariablesHelp
	default:
		return "# Environment variables not available for this provider"
	}
}

// ProviderDocsURL returns the documentation URL for the provider.
func ProviderDocsURL(provider cloudresourcekind.CloudResourceProvider) string {
	switch provider {
	case cloudresourcekind.CloudResourceProvider_atlas:
		return atlas.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_auth0:
		return auth0.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_aws:
		return aws.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_azure:
		return azure.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_civo:
		return civo.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_cloudflare:
		return cloudflare.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_confluent:
		return confluent.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_digital_ocean:
		return digitalocean.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_gcp:
		return gcp.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_kubernetes:
		return kubernetes.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_openfga:
		return openfga.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_snowflake:
		return snowflake.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_openstack:
		return openstack.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_scaleway:
		return scaleway.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_alicloud:
		return alicloud.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_oci:
		return oci.ProviderDocsURL
	case cloudresourcekind.CloudResourceProvider_hetzner_cloud:
		return hetznercloud.ProviderDocsURL
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
	sb.WriteString(fmt.Sprintf("    openmcf plan -f manifest.yaml -p %s\n",
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

  apiVersion: gcp.openmcf.org/v1
  kind: GkeCluster
  metadata:
    name: my-cluster
  spec:
    # ... resource configuration

Check your manifest file for:
  - Missing or misspelled 'apiVersion'
  - Missing or misspelled 'kind'
  - Invalid YAML syntax

For supported resource kinds, see: https://openmcf.org/docs/resources`
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
