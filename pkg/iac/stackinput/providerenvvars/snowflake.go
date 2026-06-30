package providerenvvars

import (
	"github.com/pkg/errors"
	snowflakeprovider "github.com/plantonhq/planton/apis/dev/planton/provider/snowflake"
)

// loadSnowflakeEnvVars loads Snowflake provider config and returns environment variables.
func loadSnowflakeEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(snowflakeprovider.SnowflakeProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Snowflake provider config")
	}

	envVars := map[string]string{
		"SNOWFLAKE_ACCOUNT":  config.Account,
		"SNOWFLAKE_REGION":   config.Region,
		"SNOWFLAKE_USERNAME": config.Username,
		"SNOWFLAKE_PASSWORD": config.Password,
	}

	return envVars, nil
}
