package providerenvvars

import (
	"github.com/pkg/errors"
	auth0provider "github.com/plantonhq/planton/apis/dev/planton/provider/auth0"
)

// loadAuth0EnvVars loads Auth0 provider config and returns environment variables.
func loadAuth0EnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(auth0provider.Auth0ProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Auth0 provider config")
	}

	envVars := map[string]string{
		"AUTH0_DOMAIN":        config.Domain,
		"AUTH0_CLIENT_ID":     config.ClientId,
		"AUTH0_CLIENT_SECRET": config.ClientSecret,
	}

	return envVars, nil
}
