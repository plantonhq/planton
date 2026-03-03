package providerenvvars

import (
	"github.com/pkg/errors"
	gcpprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp"
)

// loadGcpEnvVars loads GCP provider config and returns environment variables.
func loadGcpEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(gcpprovider.GcpProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load GCP provider config")
	}

	envVars := map[string]string{
		"GOOGLE_CREDENTIALS": config.ServiceAccountKey,
	}

	return envVars, nil
}
