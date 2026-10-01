package providerenvvars

import (
	"github.com/pkg/errors"
	azureprovider "github.com/plantonhq/planton/catalog/azure"
)

// loadAzureEnvVars loads Azure provider config and returns environment variables. A runner-mode
// connection carries only the tenant and subscription; the credential is the runner's own (a
// service principal, managed identity or workload identity in its environment), which an empty
// ARM_CLIENT_ID would erase.
//
// A keyless (web_identity) connection on an engine that reads its environment hands azurerm the
// minted token with ARM_USE_OIDC, and azurerm exchanges it at Entra ID for the connection's
// application when the engine starts -- minutes after the mint, well inside the token's life. It
// never carries a client secret, and it must never be left without its token: the provider would
// then sign in as whatever identity the runner's machine holds. On Pulumi the builder does the same
// exchange in the plugin, so nothing credential-shaped is emitted; with no engine named, it is
// refused (see Engine).
func loadAzureEnvVars(providerConfigYaml []byte, opts Options) (map[string]string, error) {
	config := new(azureprovider.AzureProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Azure provider config")
	}

	envVars := map[string]string{}
	putIfSet(envVars, "ARM_TENANT_ID", config.TenantId)
	putIfSet(envVars, "ARM_SUBSCRIPTION_ID", config.SubscriptionId)

	if webIdentity := config.GetWebIdentity(); webIdentity != nil {
		switch opts.Engine {
		case EngineBuildsProviders:
			return envVars, nil
		case EngineReadsEnvironment:
		default:
			return nil, errEngineUnset("Azure")
		}
		if webIdentity.GetWebIdentityToken() == "" || config.GetClientId() == "" {
			return nil, errors.New("web_identity requires web_identity_token and the application's client_id")
		}
		envVars["ARM_CLIENT_ID"] = config.GetClientId()
		envVars["ARM_USE_OIDC"] = "true"
		envVars["ARM_OIDC_TOKEN"] = webIdentity.GetWebIdentityToken()
		return envVars, nil
	}

	putIfSet(envVars, "ARM_CLIENT_ID", config.ClientId)
	putIfSet(envVars, "ARM_CLIENT_SECRET", config.ClientSecret)
	return envVars, nil
}
