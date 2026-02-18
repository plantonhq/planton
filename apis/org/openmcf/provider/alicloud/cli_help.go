package alicloud

// CLI help constants for Alibaba Cloud provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.
//
// Alibaba Cloud supports seven authentication methods:
// - Static Credentials: Long-lived IAM access key pair (simplest, most common)
// - STS Token: Temporary credentials with a security token
// - ECS Role: Automatic credentials from ECS instance metadata
// - Assume Role: RAM role assumption using base access keys
// - Assume Role with OIDC: RAM role assumption via OIDC federation (no access key needed)
// - Shared Credentials: Credentials loaded from a shared file / named profile
// - Sidecar Credentials: Credentials fetched from a sidecar HTTP endpoint
//
// All environment variables use the modern ALIBABA_CLOUD_* prefix.
//
// The quality of these constants directly drives the terminal experience when users
// encounter credential errors, via MissingProviderConfigGuidance and InvalidProviderConfigGuidance.

// EnvironmentVariables lists all environment variables supported by the Alibaba Cloud provider.
// These are read by the Terraform Alibaba Cloud provider when no explicit config file is provided.
var EnvironmentVariables = []string{
	// Common
	"ALIBABA_CLOUD_REGION",
	"ALIBABA_CLOUD_ACCOUNT_ID",
	"ALIBABA_CLOUD_ACCOUNT_TYPE",
	// Static credentials
	"ALIBABA_CLOUD_ACCESS_KEY_ID",
	"ALIBABA_CLOUD_ACCESS_KEY_SECRET",
	// STS token
	"ALIBABA_CLOUD_SECURITY_TOKEN",
	// ECS role
	"ALIBABA_CLOUD_ECS_METADATA",
	// Assume role
	"ALIBABA_CLOUD_ROLE_ARN",
	"ALIBABA_CLOUD_ROLE_SESSION_NAME",
	"ALICLOUD_ASSUME_ROLE_SESSION_EXPIRATION",
	// Assume role with OIDC
	"ALIBABA_CLOUD_OIDC_PROVIDER_ARN",
	"ALIBABA_CLOUD_OIDC_TOKEN",
	"ALIBABA_CLOUD_OIDC_TOKEN_FILE",
	// Shared credentials
	"ALIBABA_CLOUD_CREDENTIALS_FILE",
	"ALIBABA_CLOUD_PROFILE",
	// Sidecar credentials
	"ALIBABA_CLOUD_CREDENTIALS_URI",
}

// EnvironmentVariablesHelp provides export commands for the supported environment variables.
// Organized by authentication method so users can quickly identify which set of variables they need.
const EnvironmentVariablesHelp = `# Method 1: Static Credentials (simplest)
export ALIBABA_CLOUD_ACCESS_KEY_ID="<your-access-key-id>"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="<your-access-key-secret>"
export ALIBABA_CLOUD_REGION="cn-hangzhou"

# Method 2: STS Token (temporary credentials)
export ALIBABA_CLOUD_ACCESS_KEY_ID="<your-temporary-access-key-id>"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="<your-temporary-access-key-secret>"
export ALIBABA_CLOUD_SECURITY_TOKEN="<your-security-token>"
export ALIBABA_CLOUD_REGION="cn-hangzhou"

# Method 3: ECS Instance Role
export ALIBABA_CLOUD_ECS_METADATA="<your-ecs-role-name>"
export ALIBABA_CLOUD_REGION="cn-hangzhou"

# Method 4: Assume RAM Role
export ALIBABA_CLOUD_ACCESS_KEY_ID="<your-access-key-id>"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="<your-access-key-secret>"
export ALIBABA_CLOUD_ROLE_ARN="acs:ram::<account-id>:role/<role-name>"
export ALIBABA_CLOUD_ROLE_SESSION_NAME="terraform"
export ALIBABA_CLOUD_REGION="cn-hangzhou"

# Method 5: Assume Role with OIDC
export ALIBABA_CLOUD_OIDC_PROVIDER_ARN="acs:ram::<account-id>:oidc-provider/<provider-name>"
export ALIBABA_CLOUD_ROLE_ARN="acs:ram::<account-id>:role/<role-name>"
export ALIBABA_CLOUD_OIDC_TOKEN_FILE="/path/to/oidc-token"
export ALIBABA_CLOUD_ROLE_SESSION_NAME="terraform"
export ALIBABA_CLOUD_REGION="cn-hangzhou"

# Method 6: Shared Credentials File
export ALIBABA_CLOUD_CREDENTIALS_FILE="~/.aliyun/config.json"
export ALIBABA_CLOUD_PROFILE="default"
export ALIBABA_CLOUD_REGION="cn-hangzhou"

# Method 7: Sidecar Credentials
export ALIBABA_CLOUD_CREDENTIALS_URI="http://localhost:8080/credentials"
export ALIBABA_CLOUD_REGION="cn-hangzhou"

# Optional (all methods)
# export ALIBABA_CLOUD_ACCOUNT_ID="<your-account-id>"
# export ALIBABA_CLOUD_ACCOUNT_TYPE="Domestic"`

// ConfigFileExample provides an example YAML configuration file.
// Shows static credentials as the primary method (most common),
// with other authentication methods as commented alternatives.
const ConfigFileExample = `# Static Credentials (most common)
authentication_type: "static_credentials"
region: "cn-hangzhou"
static_credentials:
  access_key: "<your-access-key-id>"
  secret_key: "<your-access-key-secret>"

# Optional
# account_id: "<your-account-id>"
# account_type: "Domestic"

# Alternative: Assume RAM Role
# authentication_type: "assume_role"
# region: "cn-hangzhou"
# assume_role:
#   access_key: "<your-access-key-id>"
#   secret_key: "<your-access-key-secret>"
#   role_arn: "acs:ram::<account-id>:role/<role-name>"
#   session_name: "terraform"

# Alternative: ECS Instance Role
# authentication_type: "ecs_role"
# region: "cn-hangzhou"
# ecs_role:
#   ecs_role_name: "<your-ecs-role-name>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "alicloud-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "Alibaba Cloud"

// ProviderDocsURL points to the Terraform provider documentation.
const ProviderDocsURL = "https://registry.terraform.io/providers/aliyun/alicloud/latest/docs"
