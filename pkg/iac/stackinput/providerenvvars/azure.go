package providerenvvars

import (
	"github.com/pkg/errors"
	azureprovider "github.com/plantonhq/planton/apis/dev/planton/provider/azure"
)

// loadAzureEnvVars loads Azure provider config and returns environment variables.
func loadAzureEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(azureprovider.AzureProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Azure provider config")
	}

	envVars := map[string]string{
		"ARM_CLIENT_ID":       config.ClientId,
		"ARM_CLIENT_SECRET":   config.ClientSecret,
		"ARM_TENANT_ID":       config.TenantId,
		"ARM_SUBSCRIPTION_ID": config.SubscriptionId,
	}

	return envVars, nil
}
