package atlas

// CLI help constants for MongoDB Atlas provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the MongoDB Atlas provider.
// These are read by the Pulumi/Terraform MongoDB Atlas providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"MONGODB_ATLAS_PUBLIC_KEY",
	"MONGODB_ATLAS_PRIVATE_KEY",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export MONGODB_ATLAS_PUBLIC_KEY="<your-public-key>"
export MONGODB_ATLAS_PRIVATE_KEY="<your-private-key>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `public_key: "<your-public-key>"
private_key: "<your-private-key>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "atlas-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "MongoDB Atlas"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://www.pulumi.com/registry/packages/mongodbatlas/installation-configuration/#configuring-credentials"
