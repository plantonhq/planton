package digitalocean

// CLI help constants for DigitalOcean provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the DigitalOcean provider.
// These are read by the Pulumi/Terraform DigitalOcean providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"DIGITALOCEAN_TOKEN",
	"DIGITALOCEAN_ACCESS_TOKEN",
	"SPACES_ACCESS_KEY_ID",
	"SPACES_SECRET_ACCESS_KEY",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export DIGITALOCEAN_TOKEN="<your-api-token>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `api_token: "<your-api-token>"
default_region: 1  # See DigitalOceanRegion enum for values`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "digitalocean-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "DigitalOcean"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://docs.digitalocean.com/reference/api/create-personal-access-token/"
