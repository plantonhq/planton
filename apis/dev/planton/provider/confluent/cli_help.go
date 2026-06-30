package confluent

// CLI help constants for Confluent Cloud provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Confluent provider.
// These are read by the Pulumi/Terraform Confluent providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"CONFLUENT_CLOUD_API_KEY",
	"CONFLUENT_CLOUD_API_SECRET",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export CONFLUENT_CLOUD_API_KEY="<your-api-key>"
export CONFLUENT_CLOUD_API_SECRET="<your-api-secret>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `api_key: "<your-api-key>"
api_secret: "<your-api-secret>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "confluent-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Confluent Cloud"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://docs.confluent.io/cloud/current/access-management/authenticate/api-keys/api-keys.html"
