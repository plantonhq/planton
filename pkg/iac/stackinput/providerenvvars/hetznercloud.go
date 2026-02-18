package providerenvvars

import (
	"github.com/pkg/errors"
	hetznercloudprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud"
)

// loadHetznercloudEnvVars loads Hetzner Cloud provider config and returns environment variables.
// It maps the structured HetznercloudProviderConfig fields to the HCLOUD_*/HETZNER_* environment
// variables expected by the Terraform hcloud provider.
//
// poll_interval and poll_function are intentionally not emitted because the upstream Terraform
// provider does not read them from environment variables.
func loadHetznercloudEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(hetznercloudprovider.HetznercloudProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Hetzner Cloud provider config")
	}

	envVars := map[string]string{}

	// Authentication (required)
	if config.Token != "" {
		envVars["HCLOUD_TOKEN"] = config.Token
	}

	// Endpoint overrides
	if config.Endpoint != "" {
		envVars["HCLOUD_ENDPOINT"] = config.Endpoint
	}
	if config.EndpointHetzner != "" {
		envVars["HETZNER_ENDPOINT"] = config.EndpointHetzner
	}

	return envVars, nil
}
