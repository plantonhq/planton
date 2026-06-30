package providerenvvars

import (
	"github.com/pkg/errors"
	ociprovider "github.com/plantonhq/planton/apis/dev/planton/provider/oci"
)

// loadOciEnvVars loads OCI provider config and returns environment variables.
// Switches on the AuthenticationType enum to emit method-specific OCI_* env vars,
// then emits common fields (region) for all methods.
func loadOciEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(ociprovider.OciProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load OCI provider config")
	}

	envVars := map[string]string{}

	// Map AuthenticationType enum to the canonical Terraform provider auth string.
	// The OCI provider validates these values case-insensitively, but we use the canonical casing.
	switch config.AuthenticationType {
	case ociprovider.AuthenticationType_api_key:
		envVars["OCI_AUTH"] = "ApiKey"
		if config.ApiKey != nil {
			setNonEmpty(envVars, "OCI_TENANCY_OCID", config.ApiKey.TenancyOcid)
			setNonEmpty(envVars, "OCI_USER_OCID", config.ApiKey.UserOcid)
			setNonEmpty(envVars, "OCI_FINGERPRINT", config.ApiKey.Fingerprint)
			setNonEmpty(envVars, "OCI_PRIVATE_KEY", config.ApiKey.PrivateKey)
			setNonEmpty(envVars, "OCI_PRIVATE_KEY_PATH", config.ApiKey.PrivateKeyPath)
			setNonEmpty(envVars, "OCI_PRIVATE_KEY_PASSWORD", config.ApiKey.PrivateKeyPassword)
		}

	case ociprovider.AuthenticationType_instance_principal:
		envVars["OCI_AUTH"] = "InstancePrincipal"

	case ociprovider.AuthenticationType_security_token:
		envVars["OCI_AUTH"] = "SecurityToken"
		if config.SecurityToken != nil {
			setNonEmpty(envVars, "OCI_CONFIG_FILE_PROFILE", config.SecurityToken.ConfigFileProfile)
			setNonEmpty(envVars, "OCI_PRIVATE_KEY_PASSWORD", config.SecurityToken.PrivateKeyPassword)
		}

	case ociprovider.AuthenticationType_resource_principal:
		envVars["OCI_AUTH"] = "ResourcePrincipal"

	case ociprovider.AuthenticationType_oke_workload_identity:
		envVars["OCI_AUTH"] = "OKEWorkloadIdentity"
	}

	// Common field (emitted for all authentication methods)
	setNonEmpty(envVars, "OCI_REGION", config.Region)

	return envVars, nil
}
