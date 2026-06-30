package cloudflare

// CLI help constants for Cloudflare provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Cloudflare provider.
// These are read by the Pulumi/Terraform Cloudflare providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"CLOUDFLARE_API_TOKEN",
	"CLOUDFLARE_API_KEY",
	"CLOUDFLARE_EMAIL",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
// API Token is recommended over legacy API Key + Email.
const EnvironmentVariablesHelp = `export CLOUDFLARE_API_TOKEN="<your-cloudflare-api-token>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `auth_scheme: 1  # 1=API_TOKEN (recommended), 2=LEGACY_API_KEY
api_token: "<your-cloudflare-api-token>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "cloudflare-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Cloudflare"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://developers.cloudflare.com/fundamentals/api/get-started/create-token/"
