package providerenvvars

import (
	"testing"

	azureprovider "github.com/plantonhq/planton/catalog/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func azureKeylessConfig(t *testing.T) []byte {
	t.Helper()
	b, err := protojson.Marshal(&azureprovider.AzureProviderConfig{
		ClientId: "client", TenantId: "tenant", SubscriptionId: "subscription",
		WebIdentity: &azureprovider.AzureWebIdentityProviderConfig{WebIdentityToken: "eyJhbGciOiJSUzI1NiJ9.payload.sig"},
	})
	require.NoError(t, err)
	return b
}

func TestLoadAzureEnvVars_WebIdentity_ReadsEnvironment_HandsAzurermTheToken(t *testing.T) {
	env, err := loadAzureEnvVars(azureKeylessConfig(t), Options{Engine: EngineReadsEnvironment})
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"ARM_CLIENT_ID": "client", "ARM_TENANT_ID": "tenant", "ARM_SUBSCRIPTION_ID": "subscription",
		"ARM_USE_OIDC": "true", "ARM_OIDC_TOKEN": "eyJhbGciOiJSUzI1NiJ9.payload.sig",
	}, env)
}

func TestLoadAzureEnvVars_WebIdentity_BuildsProviders_OnlyTheCoordinates(t *testing.T) {
	env, err := loadAzureEnvVars(azureKeylessConfig(t), Options{Engine: EngineBuildsProviders})
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"ARM_TENANT_ID": "tenant", "ARM_SUBSCRIPTION_ID": "subscription"}, env)
}

func TestLoadAzureEnvVars_WebIdentity_NoEngine_Refused(t *testing.T) {
	_, err := loadAzureEnvVars(azureKeylessConfig(t), Options{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Options.Engine")
}

func TestLoadAzureEnvVars_WebIdentity_WithoutClientId_Refused(t *testing.T) {
	b, err := protojson.Marshal(&azureprovider.AzureProviderConfig{
		TenantId:    "tenant",
		WebIdentity: &azureprovider.AzureWebIdentityProviderConfig{WebIdentityToken: "token"},
	})
	require.NoError(t, err)

	_, err = loadAzureEnvVars(b, Options{Engine: EngineReadsEnvironment})
	assert.Error(t, err)
}
