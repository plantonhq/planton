package snowflake

// CLI help constants for Snowflake provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the Snowflake provider.
// These are read by the Pulumi/Terraform Snowflake providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"SNOWFLAKE_ACCOUNT",
	"SNOWFLAKE_REGION",
	"SNOWFLAKE_USER",
	"SNOWFLAKE_PASSWORD",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export SNOWFLAKE_ACCOUNT="<your-account-identifier>"
export SNOWFLAKE_REGION="<your-region>"
export SNOWFLAKE_USER="<your-username>"
export SNOWFLAKE_PASSWORD="<your-password>"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `account: "<your-account-identifier>"
region: "<your-region>"
username: "<your-username>"
password: "<your-password>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "snowflake-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Snowflake"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://www.pulumi.com/registry/packages/snowflake/installation-configuration/#configuring-credentials"
