package providerenvvars

import (
	"sort"
	"testing"

	auth0provider "github.com/plantonhq/planton/catalog/auth0"
	azureprovider "github.com/plantonhq/planton/catalog/azure"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// Every provider's builder, handed a provider config with nothing set, emits
// no variable with an empty value. That config is the runner-mode shape: the
// runner's own ambient identity is the credential, so the connection carries
// none. An empty ARM_CLIENT_ID or AUTH0_CLIENT_SECRET is not "unset" -- the
// runner appends these variables after its own environment, the last
// duplicate wins, and the identity the runner's machine holds (a service
// principal, a managed identity's client id, a workload-identity binding) is
// erased before the engine starts. The providers are read from the enum, so a
// provider added later is held to the same rule without anyone listing it.
// AWS is pinned in aws_test.go, because its builder runs outside
// loadProviderEnvVars.
func TestEveryProviderBuilder_EmptyConfig_EmitsNoEmptyVariable(t *testing.T) {
	names := make([]string, 0, len(cloudresourcekind.CloudResourceProvider_value))
	for name := range cloudresourcekind.CloudResourceProvider_value {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		provider := cloudresourcekind.CloudResourceProvider(cloudresourcekind.CloudResourceProvider_value[name])
		t.Run(name, func(t *testing.T) {
			env, err := loadProviderEnvVars([]byte("{}"), provider, Options{FileCacheLoc: t.TempDir()})
			if err != nil {
				// A builder that refuses an empty config emits nothing at all, which keeps the rule.
				return
			}
			for key, value := range env {
				assert.NotEmptyf(t, value, "%s emits %s with an empty value, which overrides whatever the runner's own environment holds for it", name, key)
			}
		})
	}
}

func TestLoadAzureEnvVars_ServicePrincipal_Emitted(t *testing.T) {
	cfg, err := protojson.Marshal(&azureprovider.AzureProviderConfig{
		ClientId: "client", ClientSecret: "secret", TenantId: "tenant", SubscriptionId: "subscription",
	})
	require.NoError(t, err)

	env, err := loadAzureEnvVars(cfg)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"ARM_CLIENT_ID": "client", "ARM_CLIENT_SECRET": "secret", "ARM_TENANT_ID": "tenant", "ARM_SUBSCRIPTION_ID": "subscription",
	}, env)
}

func TestLoadAzureEnvVars_RunnerMode_OnlyTheCoordinates(t *testing.T) {
	// Runner mode carries the tenant and subscription the deploy targets, and no credential.
	cfg, err := protojson.Marshal(&azureprovider.AzureProviderConfig{TenantId: "tenant", SubscriptionId: "subscription"})
	require.NoError(t, err)

	env, err := loadAzureEnvVars(cfg)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"ARM_TENANT_ID": "tenant", "ARM_SUBSCRIPTION_ID": "subscription"}, env)
}

func TestLoadAuth0EnvVars_ClientCredentials_Emitted(t *testing.T) {
	cfg, err := protojson.Marshal(&auth0provider.Auth0ProviderConfig{Domain: "tenant.auth0.com", ClientId: "client", ClientSecret: "secret"})
	require.NoError(t, err)

	env, err := loadAuth0EnvVars(cfg)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"AUTH0_DOMAIN": "tenant.auth0.com", "AUTH0_CLIENT_ID": "client", "AUTH0_CLIENT_SECRET": "secret"}, env)
}
