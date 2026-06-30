package providerenvvars

import (
	"github.com/pkg/errors"
	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
)

// loadScalewayEnvVars loads Scaleway provider config and returns environment variables.
// It maps the structured ScalewayProviderConfig fields to the flat SCW_* environment variables
// expected by the Terraform Scaleway provider.
func loadScalewayEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(scalewayprovider.ScalewayProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Scaleway provider config")
	}

	envVars := map[string]string{}

	// Authentication (required)
	if config.AccessKey != "" {
		envVars["SCW_ACCESS_KEY"] = config.AccessKey
	}
	if config.SecretKey != "" {
		envVars["SCW_SECRET_KEY"] = config.SecretKey
	}

	// Project/organization scope
	if config.ProjectId != "" {
		envVars["SCW_DEFAULT_PROJECT_ID"] = config.ProjectId
	}
	if config.OrganizationId != "" {
		envVars["SCW_DEFAULT_ORGANIZATION_ID"] = config.OrganizationId
	}

	// Geographic defaults
	if config.Region != "" {
		envVars["SCW_DEFAULT_REGION"] = config.Region
	}
	if config.Zone != "" {
		envVars["SCW_DEFAULT_ZONE"] = config.Zone
	}

	return envVars, nil
}
