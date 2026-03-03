package gcp

// CLI help constants for GCP provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the GCP provider.
// These are read by the Pulumi/Terraform GCP providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"GOOGLE_APPLICATION_CREDENTIALS",
	"GOOGLE_CLOUD_PROJECT",
	"GOOGLE_PROJECT",
	"GCLOUD_PROJECT",
	"CLOUDSDK_CORE_PROJECT",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
export GOOGLE_CLOUD_PROJECT="<your-project-id>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `service_account_key: "<service-account-json>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "gcp-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "GCP"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://cloud.google.com/docs/authentication/application-default-credentials"
