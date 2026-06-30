package hetznercloud

// CLI help constants for Hetzner Cloud provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.
//
// Hetzner Cloud uses a single API token for authentication. Tokens are 64-character strings
// created per-project in the Hetzner Cloud Console under Security > API Tokens.
// All API requests are scoped to the project that owns the token.
//
// The quality of these constants directly drives the terminal experience when users
// encounter credential errors, via MissingProviderConfigGuidance and InvalidProviderConfigGuidance.

// EnvironmentVariables lists all environment variables supported by the Hetzner Cloud provider.
// These are read by the Terraform hcloud provider when no explicit config file is provided.
var EnvironmentVariables = []string{
	"HCLOUD_TOKEN",
	"HCLOUD_ENDPOINT",
	"HETZNER_ENDPOINT",
}

// EnvironmentVariablesHelp provides export commands for the supported environment variables.
const EnvironmentVariablesHelp = `export HCLOUD_TOKEN="<your-api-token>"

# Optional
export HCLOUD_ENDPOINT="https://api.hetzner.cloud/v1"
export HETZNER_ENDPOINT="https://api.hetzner.com/v1"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `token: "<your-api-token>"

# Optional
# endpoint: "https://api.hetzner.cloud/v1"
# endpoint_hetzner: "https://api.hetzner.com/v1"
# poll_interval: "500ms"
# poll_function: "exponential"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "hetznercloud-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Hetzner Cloud"

// ProviderDocsURL points to the Terraform provider documentation.
const ProviderDocsURL = "https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs"
