package auth0

// CLI help constants for Auth0 provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Auth0 provider.
// These are read by the Pulumi/Terraform Auth0 providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"AUTH0_DOMAIN",
	"AUTH0_CLIENT_ID",
	"AUTH0_CLIENT_SECRET",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export AUTH0_DOMAIN="<your-tenant>.auth0.com"
export AUTH0_CLIENT_ID="<your-client-id>"
export AUTH0_CLIENT_SECRET="<your-client-secret>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `domain: "<your-tenant>.auth0.com"
client_id: "<your-client-id>"
client_secret: "<your-client-secret>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "auth0-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Auth0"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://www.pulumi.com/registry/packages/auth0/installation-configuration/#configuring-credentials"
