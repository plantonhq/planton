package openstack

// CLI help constants for OpenStack provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.
//
// OpenStack supports three authentication methods:
// - Password: Username/password with project/domain context (common for interactive use)
// - Application Credentials: Pre-scoped ID + secret (recommended for automation)
// - Token: Pre-authenticated Keystone token (for delegated/short-lived access)
//
// The quality of these constants directly drives the terminal experience when users
// encounter credential errors, via MissingProviderConfigGuidance and InvalidProviderConfigGuidance.

// EnvironmentVariables lists all environment variables supported by the OpenStack provider.
// These are read by the Terraform OpenStack provider when no explicit config file is provided.
var EnvironmentVariables = []string{
	// Connection
	"OS_AUTH_URL",
	"OS_REGION_NAME",
	// Password authentication
	"OS_USERNAME",
	"OS_PASSWORD",
	// Application credential authentication
	"OS_APPLICATION_CREDENTIAL_ID",
	"OS_APPLICATION_CREDENTIAL_NAME",
	"OS_APPLICATION_CREDENTIAL_SECRET",
	// Token authentication
	"OS_TOKEN",
	// Project/tenant context
	"OS_TENANT_NAME",
	"OS_TENANT_ID",
	// Domain context (Identity v3)
	"OS_USER_DOMAIN_NAME",
	"OS_USER_DOMAIN_ID",
	"OS_PROJECT_DOMAIN_NAME",
	"OS_PROJECT_DOMAIN_ID",
	// TLS
	"OS_INSECURE",
	"OS_CACERT",
	// Advanced
	"OS_ENDPOINT_TYPE",
}

// EnvironmentVariablesHelp provides export commands for the supported environment variables.
// Organized by authentication method so users can quickly identify which set of variables they need.
const EnvironmentVariablesHelp = `# Method 1: Password Authentication
export OS_AUTH_URL="https://cloud.example.com:5000/v3"
export OS_REGION_NAME="RegionOne"
export OS_USERNAME="<your-username>"
export OS_PASSWORD="<your-password>"
export OS_USER_DOMAIN_NAME="Default"
export OS_PROJECT_NAME="<your-project>"
export OS_PROJECT_DOMAIN_NAME="Default"

# Method 2: Application Credentials (recommended for automation)
export OS_AUTH_URL="https://cloud.example.com:5000/v3"
export OS_REGION_NAME="RegionOne"
export OS_APPLICATION_CREDENTIAL_ID="<your-credential-id>"
export OS_APPLICATION_CREDENTIAL_SECRET="<your-credential-secret>"

# Method 3: Token Authentication
export OS_AUTH_URL="https://cloud.example.com:5000/v3"
export OS_REGION_NAME="RegionOne"
export OS_TOKEN="<your-auth-token>"
export OS_PROJECT_NAME="<your-project>"`

// ConfigFileExample provides an example YAML configuration file.
// Shows the recommended method (application credentials) as primary,
// with password authentication as a commented alternative.
const ConfigFileExample = `# Application Credentials (recommended for automation)
auth_url: "https://cloud.example.com:5000/v3"
region: "RegionOne"
application_credential:
  id: "<your-credential-id>"
  secret: "<your-credential-secret>"

# Alternative: Password Authentication
# auth_url: "https://cloud.example.com:5000/v3"
# region: "RegionOne"
# password:
#   user_name: "<your-username>"
#   password: "<your-password>"
# user_domain_name: "Default"
# tenant_name: "<your-project>"
# project_domain_name: "Default"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "openstack-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "OpenStack"

// ProviderDocsURL points to the Terraform provider documentation.
const ProviderDocsURL = "https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs"
