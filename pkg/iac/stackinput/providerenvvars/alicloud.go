package providerenvvars

import (
	"strconv"

	"github.com/pkg/errors"
	alicloudprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud"
)

// loadAliCloudEnvVars loads Alibaba Cloud provider config and returns environment variables.
// It switches on the AuthenticationType enum to emit method-specific ALIBABA_CLOUD_*
// environment variables expected by the Terraform Alibaba Cloud provider.
func loadAliCloudEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(alicloudprovider.AliCloudProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Alibaba Cloud provider config")
	}

	envVars := map[string]string{}

	// Method-specific environment variables
	switch config.AuthenticationType {
	case alicloudprovider.AuthenticationType_static_credentials:
		if config.StaticCredentials != nil {
			setNonEmpty(envVars, "ALIBABA_CLOUD_ACCESS_KEY_ID", config.StaticCredentials.AccessKey)
			setNonEmpty(envVars, "ALIBABA_CLOUD_ACCESS_KEY_SECRET", config.StaticCredentials.SecretKey)
		}

	case alicloudprovider.AuthenticationType_sts_token:
		if config.StsToken != nil {
			setNonEmpty(envVars, "ALIBABA_CLOUD_ACCESS_KEY_ID", config.StsToken.AccessKey)
			setNonEmpty(envVars, "ALIBABA_CLOUD_ACCESS_KEY_SECRET", config.StsToken.SecretKey)
			setNonEmpty(envVars, "ALIBABA_CLOUD_SECURITY_TOKEN", config.StsToken.SecurityToken)
		}

	case alicloudprovider.AuthenticationType_ecs_role:
		if config.EcsRole != nil {
			setNonEmpty(envVars, "ALIBABA_CLOUD_ECS_METADATA", config.EcsRole.EcsRoleName)
		}

	case alicloudprovider.AuthenticationType_assume_role:
		if config.AssumeRole != nil {
			setNonEmpty(envVars, "ALIBABA_CLOUD_ACCESS_KEY_ID", config.AssumeRole.AccessKey)
			setNonEmpty(envVars, "ALIBABA_CLOUD_ACCESS_KEY_SECRET", config.AssumeRole.SecretKey)
			setNonEmpty(envVars, "ALIBABA_CLOUD_ROLE_ARN", config.AssumeRole.RoleArn)
			setNonEmpty(envVars, "ALIBABA_CLOUD_ROLE_SESSION_NAME", config.AssumeRole.SessionName)
			if config.AssumeRole.SessionExpiration != 0 {
				envVars["ALICLOUD_ASSUME_ROLE_SESSION_EXPIRATION"] = strconv.Itoa(int(config.AssumeRole.SessionExpiration))
			}
			// policy and external_id have no environment variable mapping in the Terraform provider
		}

	case alicloudprovider.AuthenticationType_assume_role_with_oidc:
		if config.AssumeRoleWithOidc != nil {
			setNonEmpty(envVars, "ALIBABA_CLOUD_OIDC_PROVIDER_ARN", config.AssumeRoleWithOidc.OidcProviderArn)
			setNonEmpty(envVars, "ALIBABA_CLOUD_ROLE_ARN", config.AssumeRoleWithOidc.RoleArn)
			setNonEmpty(envVars, "ALIBABA_CLOUD_OIDC_TOKEN", config.AssumeRoleWithOidc.OidcToken)
			setNonEmpty(envVars, "ALIBABA_CLOUD_OIDC_TOKEN_FILE", config.AssumeRoleWithOidc.OidcTokenFile)
			setNonEmpty(envVars, "ALIBABA_CLOUD_ROLE_SESSION_NAME", config.AssumeRoleWithOidc.SessionName)
			// policy and session_expiration have no environment variable mapping for OIDC
		}

	case alicloudprovider.AuthenticationType_shared_credentials:
		if config.SharedCredentials != nil {
			setNonEmpty(envVars, "ALIBABA_CLOUD_CREDENTIALS_FILE", config.SharedCredentials.CredentialsFile)
			setNonEmpty(envVars, "ALIBABA_CLOUD_PROFILE", config.SharedCredentials.Profile)
		}

	case alicloudprovider.AuthenticationType_sidecar_credentials:
		if config.SidecarCredentials != nil {
			setNonEmpty(envVars, "ALIBABA_CLOUD_CREDENTIALS_URI", config.SidecarCredentials.CredentialsUri)
		}
	}

	// Common fields (emitted for all authentication methods)
	setNonEmpty(envVars, "ALIBABA_CLOUD_REGION", config.Region)
	setNonEmpty(envVars, "ALIBABA_CLOUD_ACCOUNT_ID", config.AccountId)
	setNonEmpty(envVars, "ALIBABA_CLOUD_ACCOUNT_TYPE", config.AccountType)

	return envVars, nil
}

// setNonEmpty adds the key-value pair to the map only if the value is non-empty.
func setNonEmpty(envVars map[string]string, key, value string) {
	if value != "" {
		envVars[key] = value
	}
}
