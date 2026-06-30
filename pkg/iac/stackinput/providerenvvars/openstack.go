package providerenvvars

import (
	"strconv"

	"github.com/pkg/errors"
	openstackprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openstack"
)

// loadOpenStackEnvVars loads OpenStack provider config and returns environment variables.
// It flattens the structured oneof credentials into the flat OS_* environment variables
// expected by the Terraform OpenStack provider.
func loadOpenStackEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(openstackprovider.OpenStackProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load OpenStack provider config")
	}

	envVars := map[string]string{}

	// Connection (common to all auth methods)
	if config.AuthUrl != "" {
		envVars["OS_AUTH_URL"] = config.AuthUrl
	}
	if config.Region != "" {
		envVars["OS_REGION_NAME"] = config.Region
	}

	// Authentication method (oneof credentials)
	switch creds := config.Credentials.(type) {
	case *openstackprovider.OpenStackProviderConfig_Password:
		if creds.Password != nil {
			if creds.Password.UserName != "" {
				envVars["OS_USERNAME"] = creds.Password.UserName
			}
			if creds.Password.Password != "" {
				envVars["OS_PASSWORD"] = creds.Password.Password
			}
		}
	case *openstackprovider.OpenStackProviderConfig_ApplicationCredential:
		if creds.ApplicationCredential != nil {
			if creds.ApplicationCredential.Id != "" {
				envVars["OS_APPLICATION_CREDENTIAL_ID"] = creds.ApplicationCredential.Id
			}
			if creds.ApplicationCredential.Name != "" {
				envVars["OS_APPLICATION_CREDENTIAL_NAME"] = creds.ApplicationCredential.Name
			}
			if creds.ApplicationCredential.Secret != "" {
				envVars["OS_APPLICATION_CREDENTIAL_SECRET"] = creds.ApplicationCredential.Secret
			}
		}
	case *openstackprovider.OpenStackProviderConfig_Token:
		if creds.Token != nil {
			if creds.Token.Token != "" {
				envVars["OS_TOKEN"] = creds.Token.Token
			}
		}
	}

	// Project/tenant context
	if config.TenantName != "" {
		envVars["OS_TENANT_NAME"] = config.TenantName
	}
	if config.TenantId != "" {
		envVars["OS_TENANT_ID"] = config.TenantId
	}

	// Domain context (Identity v3)
	if config.UserDomainName != "" {
		envVars["OS_USER_DOMAIN_NAME"] = config.UserDomainName
	}
	if config.UserDomainId != "" {
		envVars["OS_USER_DOMAIN_ID"] = config.UserDomainId
	}
	if config.ProjectDomainName != "" {
		envVars["OS_PROJECT_DOMAIN_NAME"] = config.ProjectDomainName
	}
	if config.ProjectDomainId != "" {
		envVars["OS_PROJECT_DOMAIN_ID"] = config.ProjectDomainId
	}

	// TLS configuration
	if config.Insecure {
		envVars["OS_INSECURE"] = strconv.FormatBool(config.Insecure)
	}
	if config.CacertFile != "" {
		envVars["OS_CACERT"] = config.CacertFile
	}

	// Advanced
	if config.EndpointType != "" {
		envVars["OS_ENDPOINT_TYPE"] = config.EndpointType
	}

	return envVars, nil
}
