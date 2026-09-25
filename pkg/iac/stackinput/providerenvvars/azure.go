package providerenvvars

import (
	"github.com/pkg/errors"
	azureprovider "github.com/plantonhq/planton/catalog/azure"
)

// loadAzureEnvVars loads Azure provider config and returns environment variables. A runner-mode
// connection carries only the tenant and subscription; the credential is the runner's own (a
// service principal, managed identity or workload identity in its environment), which an empty
// ARM_CLIENT_ID would erase.
func loadAzureEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(azureprovider.AzureProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Azure provider config")
	}

	envVars := map[string]string{}
	putIfSet(envVars, "ARM_CLIENT_ID", config.ClientId)
	putIfSet(envVars, "ARM_CLIENT_SECRET", config.ClientSecret)
	putIfSet(envVars, "ARM_TENANT_ID", config.TenantId)
	putIfSet(envVars, "ARM_SUBSCRIPTION_ID", config.SubscriptionId)

	return envVars, nil
}
