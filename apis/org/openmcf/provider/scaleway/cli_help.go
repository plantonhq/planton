package scaleway

// CLI help constants for Scaleway provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.
//
// Scaleway uses a simple access key / secret key authentication model.
// The key pair is created in the Scaleway console under IAM > API Keys.
// An optional project_id scopes resources to a specific project.
//
// The quality of these constants directly drives the terminal experience when users
// encounter credential errors, via MissingProviderConfigGuidance and InvalidProviderConfigGuidance.

// EnvironmentVariables lists all environment variables supported by the Scaleway provider.
// These are read by the Terraform Scaleway provider when no explicit config file is provided.
var EnvironmentVariables = []string{
	"SCW_ACCESS_KEY",
	"SCW_SECRET_KEY",
	"SCW_DEFAULT_PROJECT_ID",
	"SCW_DEFAULT_ORGANIZATION_ID",
	"SCW_DEFAULT_REGION",
	"SCW_DEFAULT_ZONE",
}

// EnvironmentVariablesHelp provides export commands for the supported environment variables.
const EnvironmentVariablesHelp = `export SCW_ACCESS_KEY="<your-access-key>"
export SCW_SECRET_KEY="<your-secret-key>"
export SCW_DEFAULT_PROJECT_ID="<your-project-id>"

# Optional
export SCW_DEFAULT_ORGANIZATION_ID="<your-organization-id>"
export SCW_DEFAULT_REGION="fr-par"
export SCW_DEFAULT_ZONE="fr-par-1"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `access_key: "<your-access-key>"
secret_key: "<your-secret-key>"
project_id: "<your-project-id>"

# Optional
# organization_id: "<your-organization-id>"
# region: "fr-par"
# zone: "fr-par-1"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "scaleway-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Scaleway"

// ProviderDocsURL points to the Terraform provider documentation.
const ProviderDocsURL = "https://registry.terraform.io/providers/scaleway/scaleway/latest/docs"
