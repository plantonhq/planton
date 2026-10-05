package providerenvvars

import (
	"context"
	"testing"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"github.com/plantonhq/planton/pkg/iac/provider/gcp/gcpwebidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// gcpConfigYaml renders a GcpProviderConfig the way the runner injects it (protojson, camelCase);
// loadProviderConfigProto reads it back through YAMLToJSON -> protojson, so this exercises the
// same round-trip the live IaC input takes.
func gcpConfigYaml(t *testing.T, cfg *gcpprovider.GcpProviderConfig) []byte {
	t.Helper()
	b, err := protojson.Marshal(cfg)
	require.NoError(t, err)
	return b
}

func TestLoadGcpEnvVars_ServiceAccountKey_Emitted(t *testing.T) {
	env, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{
		ServiceAccountKey: `{"type":"service_account"}`,
	}), Options{}, failingGcpResolver(t))
	require.NoError(t, err)

	assert.Equal(t, `{"type":"service_account"}`, env["GOOGLE_CREDENTIALS"])
}

func TestLoadGcpEnvVars_EmptyKey_NoEmptyEnvVar(t *testing.T) {
	// An empty GOOGLE_CREDENTIALS would poison the ambient credential chain; the variable
	// must be absent, not empty.
	env, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{}), Options{}, failingGcpResolver(t))
	require.NoError(t, err)

	_, present := env["GOOGLE_CREDENTIALS"]
	assert.False(t, present)
	assert.Empty(t, env)
}

func TestLoadGcpEnvVars_AccessToken_Emitted(t *testing.T) {
	env, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{
		AccessToken: "ya29.test-token",
	}), Options{}, failingGcpResolver(t))
	require.NoError(t, err)

	assert.Equal(t, "ya29.test-token", env["GOOGLE_OAUTH_ACCESS_TOKEN"])
	_, credentialsPresent := env["GOOGLE_CREDENTIALS"]
	assert.False(t, credentialsPresent)
}

func TestLoadGcpEnvVars_AccessToken_WinsOverStaleKey(t *testing.T) {
	// A pre-minted token is the deliberate credential for this run; a lingering
	// service_account_key must neither win nor be co-emitted (two credential variables
	// would leave the effective identity to provider precedence rules).
	env, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{
		AccessToken:       "ya29.test-token",
		ServiceAccountKey: `{"type":"service_account"}`,
	}), Options{}, failingGcpResolver(t))
	require.NoError(t, err)

	assert.Equal(t, "ya29.test-token", env["GOOGLE_OAUTH_ACCESS_TOKEN"])
	_, credentialsPresent := env["GOOGLE_CREDENTIALS"]
	assert.False(t, credentialsPresent)
	assert.Len(t, env, 1)
}

func TestLoadGcpEnvVars_EmptyAccessToken_NoEmptyEnvVar(t *testing.T) {
	// An empty GOOGLE_OAUTH_ACCESS_TOKEN would poison the ambient credential chain the
	// same way an empty GOOGLE_CREDENTIALS would.
	env, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{
		AccessToken: "",
	}), Options{}, failingGcpResolver(t))
	require.NoError(t, err)

	_, present := env["GOOGLE_OAUTH_ACCESS_TOKEN"]
	assert.False(t, present)
	assert.Empty(t, env)
}

func gcpKeylessConfig() *gcpprovider.GcpProviderConfig {
	return &gcpprovider.GcpProviderConfig{
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			WebIdentityToken:    "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			Audience:            "//iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/test-pool/providers/test-provider",
			ServiceAccountEmail: "provisioner@test-project.iam.gserviceaccount.com",
		},
	}
}

// failingGcpResolver fails the test if the keyless exchange is reached.
func failingGcpResolver(t *testing.T) gcpwebidentity.TokenResolver {
	t.Helper()
	return func(context.Context, *gcpprovider.GcpWebIdentityProviderConfig) (string, error) {
		t.Fatal("the Google Cloud keyless exchange must not be called for this case")
		return "", nil
	}
}

func TestLoadGcpEnvVars_WebIdentity_ReadsEnvironment_EmitsTheExchangedToken(t *testing.T) {
	var exchanged *gcpprovider.GcpWebIdentityProviderConfig
	resolve := func(_ context.Context, wi *gcpprovider.GcpWebIdentityProviderConfig) (string, error) {
		exchanged = wi
		return "ya29.impersonated", nil
	}

	env, err := loadGcpEnvVars(gcpConfigYaml(t, gcpKeylessConfig()), Options{Engine: EngineReadsEnvironment}, resolve)
	require.NoError(t, err)

	// Exactly one credential: the impersonated account's token, in the access-token arm's variable.
	assert.Equal(t, map[string]string{"GOOGLE_OAUTH_ACCESS_TOKEN": "ya29.impersonated"}, env)
	assert.Equal(t, "provisioner@test-project.iam.gserviceaccount.com", exchanged.GetServiceAccountEmail())
}

func TestLoadGcpEnvVars_WebIdentity_BuildsProviders_EmitsNothing(t *testing.T) {
	// Pulumi: pulumi-gcp exchanges the token inside the plugin from the provider config.
	env, err := loadGcpEnvVars(gcpConfigYaml(t, gcpKeylessConfig()), Options{Engine: EngineBuildsProviders}, failingGcpResolver(t))
	require.NoError(t, err)
	assert.Empty(t, env)
}

func TestLoadGcpEnvVars_WebIdentity_NoEngine_Refused(t *testing.T) {
	_, err := loadGcpEnvVars(gcpConfigYaml(t, gcpKeylessConfig()), Options{}, failingGcpResolver(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Options.Engine")
}

func TestLoadGcpEnvVars_WebIdentity_ResolverError_Propagates(t *testing.T) {
	resolve := func(context.Context, *gcpprovider.GcpWebIdentityProviderConfig) (string, error) {
		return "", assert.AnError
	}
	_, err := loadGcpEnvVars(gcpConfigYaml(t, gcpKeylessConfig()), Options{Engine: EngineReadsEnvironment}, resolve)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestLoadGcpEnvVars_WebIdentity_Invalid_RefusedBeforeAnyExchange(t *testing.T) {
	cfg := gcpKeylessConfig()
	cfg.WebIdentity.ServiceAccountEmail = ""
	_, err := loadGcpEnvVars(gcpConfigYaml(t, cfg), Options{Engine: EngineReadsEnvironment}, failingGcpResolver(t))
	assert.Error(t, err)
}
