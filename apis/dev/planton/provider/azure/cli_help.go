package azure

// CLI help constants for Azure provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Azure provider.
// These are read by the Pulumi/Terraform Azure providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"ARM_CLIENT_ID",
	"ARM_CLIENT_SECRET",
	"ARM_TENANT_ID",
	"ARM_SUBSCRIPTION_ID",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export ARM_CLIENT_ID="<your-client-id>"
export ARM_CLIENT_SECRET="<your-client-secret>"
export ARM_TENANT_ID="<your-tenant-id>"
export ARM_SUBSCRIPTION_ID="<your-subscription-id>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `client_id: "<your-client-id>"
client_secret: "<your-client-secret>"
tenant_id: "<your-tenant-id>"
subscription_id: "<your-subscription-id>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "azure-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Azure"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://learn.microsoft.com/en-us/azure/developer/terraform/authenticate-to-azure"
