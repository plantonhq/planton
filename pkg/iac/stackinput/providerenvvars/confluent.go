package providerenvvars

import (
	"github.com/pkg/errors"
	confluentprovider "github.com/plantonhq/planton/apis/dev/planton/provider/confluent"
)

// loadConfluentEnvVars loads Confluent provider config and returns environment variables.
func loadConfluentEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(confluentprovider.ConfluentProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Confluent provider config")
	}

	envVars := map[string]string{
		"CONFLUENT_API_KEY":    config.ApiKey,
		"CONFLUENT_API_SECRET": config.ApiSecret,
	}

	return envVars, nil
}
