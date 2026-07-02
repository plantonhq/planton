package providerenvvars

import (
	"testing"

	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// gcpConfigYaml renders a GcpProviderConfig the way the runner injects it (protojson, camelCase);
// loadProviderConfigProto reads it back through YAMLToJSON -> protojson, so this exercises the
// same round-trip the live stack input takes.
func gcpConfigYaml(t *testing.T, cfg *gcpprovider.GcpProviderConfig) []byte {
	t.Helper()
	b, err := protojson.Marshal(cfg)
	require.NoError(t, err)
	return b
}

func TestLoadGcpEnvVars_ServiceAccountKey_Emitted(t *testing.T) {
	env, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{
		ServiceAccountKey: `{"type":"service_account"}`,
	}))
	require.NoError(t, err)

	assert.Equal(t, `{"type":"service_account"}`, env["GOOGLE_CREDENTIALS"])
}

func TestLoadGcpEnvVars_EmptyKey_NoEmptyEnvVar(t *testing.T) {
	// An empty GOOGLE_CREDENTIALS would poison the ambient credential chain; the variable
	// must be absent, not empty.
	env, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{}))
	require.NoError(t, err)

	_, present := env["GOOGLE_CREDENTIALS"]
	assert.False(t, present)
	assert.Empty(t, env)
}

func TestLoadGcpEnvVars_WebIdentity_FailsLoudly(t *testing.T) {
	// Keyless auth has no terraform env-var form; silently degrading to ambient credentials
	// would run the deploy as the wrong identity, so the loader must reject it outright.
	_, err := loadGcpEnvVars(gcpConfigYaml(t, &gcpprovider.GcpProviderConfig{
		WebIdentity: &gcpprovider.GcpWebIdentityProviderConfig{
			WebIdentityToken:    "eyJhbGciOiJSUzI1NiJ9.payload.sig",
			Audience:            "//iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/test-pool/providers/test-provider",
			ServiceAccountEmail: "provisioner@test-project.iam.gserviceaccount.com",
		},
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported on the tofu/terraform path")
}
