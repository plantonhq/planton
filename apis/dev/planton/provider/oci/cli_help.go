package oci

// CLI help constants for OCI (Oracle Cloud Infrastructure) provider.
// These are used by the CLI to provide helpful guidance when credentials are missing or invalid.
// Source of truth: provider.proto in this package.
//
// OCI supports five authentication methods:
// - API Key: Long-lived API signing key (tenancy/user OCID + fingerprint + private key)
// - Instance Principal: Ambient auth from OCI compute instance metadata
// - Security Token: Session-based auth using an OCI CLI config file profile
// - Resource Principal: Ambient auth for OCI Functions and other OCI resources
// - OKE Workload Identity: Ambient auth for pods running on Oracle Kubernetes Engine
//
// All environment variables use the OCI_* prefix, which the Terraform OCI provider
// accepts as a fallback alongside TF_VAR_*.
//
// The quality of these constants directly drives the terminal experience when users
// encounter credential errors, via MissingProviderConfigGuidance and InvalidProviderConfigGuidance.

// EnvironmentVariables lists all environment variables supported by the OCI provider.
// These are read by the Terraform OCI provider when no explicit config file is provided.
var EnvironmentVariables = []string{
	// Auth method discriminator
	"OCI_AUTH",
	// Common
	"OCI_REGION",
	// API Key
	"OCI_TENANCY_OCID",
	"OCI_USER_OCID",
	"OCI_FINGERPRINT",
	"OCI_PRIVATE_KEY",
	"OCI_PRIVATE_KEY_PATH",
	"OCI_PRIVATE_KEY_PASSWORD",
	// Security Token
	"OCI_CONFIG_FILE_PROFILE",
}

// EnvironmentVariablesHelp provides export commands for the supported environment variables.
// Organized by authentication method so users can quickly identify which set of variables they need.
const EnvironmentVariablesHelp = `# Method 1: API Key (most common)
export OCI_AUTH="ApiKey"
export OCI_TENANCY_OCID="ocid1.tenancy.oc1..<unique-id>"
export OCI_USER_OCID="ocid1.user.oc1..<unique-id>"
export OCI_FINGERPRINT="aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"
export OCI_PRIVATE_KEY_PATH="~/.oci/oci_api_key.pem"
export OCI_REGION="us-ashburn-1"
# export OCI_PRIVATE_KEY_PASSWORD="<passphrase>"  # if key is encrypted

# Method 2: Instance Principal (on OCI compute instances)
export OCI_AUTH="InstancePrincipal"
export OCI_REGION="us-ashburn-1"

# Method 3: Security Token (OCI CLI session)
export OCI_AUTH="SecurityToken"
export OCI_CONFIG_FILE_PROFILE="<profile-name>"
export OCI_REGION="us-ashburn-1"
# export OCI_PRIVATE_KEY_PASSWORD="<passphrase>"  # if key is encrypted

# Method 4: Resource Principal (OCI Functions / resources)
export OCI_AUTH="ResourcePrincipal"
export OCI_REGION="us-ashburn-1"

# Method 5: OKE Workload Identity (pods on OKE)
export OCI_AUTH="OKEWorkloadIdentity"
export OCI_REGION="us-ashburn-1"`

// ConfigFileExample provides an example YAML configuration file.
// Shows API Key as the primary method (most common for stored credentials),
// with other authentication methods as commented alternatives.
const ConfigFileExample = `# API Key (most common)
authentication_type: "api_key"
region: "us-ashburn-1"
api_key:
  tenancy_ocid: "ocid1.tenancy.oc1..<unique-id>"
  user_ocid: "ocid1.user.oc1..<unique-id>"
  fingerprint: "aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99"
  private_key_path: "~/.oci/oci_api_key.pem"
  # private_key_password: "<passphrase>"  # if key is encrypted

# Alternative: Instance Principal (on OCI compute instances)
# authentication_type: "instance_principal"
# region: "us-ashburn-1"

# Alternative: Security Token (OCI CLI session)
# authentication_type: "security_token"
# region: "us-ashburn-1"
# security_token:
#   config_file_profile: "<profile-name>"`

// ConfigFileName is the suggested filename for the provider config.
const ConfigFileName = "oci-provider-config.yaml"

// ProviderDisplayName is the human-readable name for this provider.
const ProviderDisplayName = "OCI"

// ProviderDocsURL points to the Terraform provider documentation.
const ProviderDocsURL = "https://registry.terraform.io/providers/oracle/oci/latest/docs"
