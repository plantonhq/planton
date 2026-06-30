package civo

// CLI help constants for Civo provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Civo provider.
// These are read by the Pulumi/Terraform Civo providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"CIVO_TOKEN",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export CIVO_TOKEN="<your-api-token>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `api_token: "<your-api-token>"
default_region: 1  # See CivoRegion enum for values`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "civo-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Civo"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://dashboard.civo.com/security"
