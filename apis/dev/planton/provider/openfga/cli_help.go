package openfga

// CLI help constants for OpenFGA provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the OpenFGA provider.
// These are read by the Terraform provider when no explicit config file is provided.
var EnvironmentVariables = []string{
	"FGA_API_URL",
	"FGA_API_TOKEN",
	"FGA_CLIENT_ID",
	"FGA_CLIENT_SECRET",
	"FGA_API_TOKEN_ISSUER",
	"FGA_API_SCOPES",
	"FGA_API_AUDIENCE",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export FGA_API_URL="<your-openfga-api-url>"
export FGA_API_TOKEN="<your-openfga-api-token>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `api_url: "<your-openfga-api-url>"
api_token: "<your-openfga-api-token>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "openfga-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "OpenFGA"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://registry.terraform.io/providers/openfga/openfga/latest/docs"
