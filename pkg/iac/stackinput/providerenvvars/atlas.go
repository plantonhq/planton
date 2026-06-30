package providerenvvars

import (
	"github.com/pkg/errors"
	atlasprovider "github.com/plantonhq/planton/apis/dev/planton/provider/atlas"
)

// loadAtlasEnvVars loads MongoDB Atlas provider config and returns environment variables.
func loadAtlasEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(atlasprovider.AtlasProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Atlas provider config")
	}

	envVars := map[string]string{
		"MONGODB_ATLAS_PUBLIC_KEY":  config.PublicKey,
		"MONGODB_ATLAS_PRIVATE_KEY": config.PrivateKey,
	}

	return envVars, nil
}
