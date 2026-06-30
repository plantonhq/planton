package aws

// CLI help constants for AWS provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.

// EnvironmentVariables lists the environment variables supported by the AWS provider.
// These are read by the Pulumi/Terraform AWS providers when no explicit config file is provided.
var EnvironmentVariables = []string{
	"AWS_ACCESS_KEY_ID",
	"AWS_SECRET_ACCESS_KEY",
	"AWS_DEFAULT_REGION",
	"AWS_REGION",
	"AWS_SESSION_TOKEN",
	"AWS_PROFILE",
}

// EnvironmentVariablesHelp provides export commands for the required environment variables.
const EnvironmentVariablesHelp = `export AWS_ACCESS_KEY_ID="<your-access-key-id>"
export AWS_SECRET_ACCESS_KEY="<your-secret-access-key>"
export AWS_DEFAULT_REGION="us-west-2"`

// ConfigFileExample provides an example YAML configuration file.
const ConfigFileExample = `account_id: "<your-aws-account-id>"
access_key_id: "<your-access-key-id>"
secret_access_key: "<your-secret-access-key>"
region: "us-west-2"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "aws-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "AWS"

// ProviderDocsURL points to the provider documentation.
const ProviderDocsURL = "https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html"
