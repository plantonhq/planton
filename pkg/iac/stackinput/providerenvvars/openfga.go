package providerenvvars

import (
	"github.com/pkg/errors"
	openfgaprovider "github.com/plantonhq/planton/apis/dev/planton/provider/openfga"
)

// loadOpenFgaEnvVars loads OpenFGA provider config and returns environment variables.
func loadOpenFgaEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(openfgaprovider.OpenFgaProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load OpenFGA provider config")
	}

	envVars := map[string]string{
		"FGA_API_URL": config.ApiUrl,
	}

	// Optional fields - only set if they have values
	if config.ApiToken != "" {
		envVars["FGA_API_TOKEN"] = config.ApiToken
	}
	if config.ClientId != "" {
		envVars["FGA_CLIENT_ID"] = config.ClientId
	}
	if config.ClientSecret != "" {
		envVars["FGA_CLIENT_SECRET"] = config.ClientSecret
	}
	if config.ApiTokenIssuer != "" {
		envVars["FGA_API_TOKEN_ISSUER"] = config.ApiTokenIssuer
	}
	if config.ApiScopes != "" {
		envVars["FGA_API_SCOPES"] = config.ApiScopes
	}
	if config.ApiAudience != "" {
		envVars["FGA_API_AUDIENCE"] = config.ApiAudience
	}

	return envVars, nil
}
