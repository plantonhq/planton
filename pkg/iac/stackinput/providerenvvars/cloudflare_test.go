package providerenvvars

import (
	"testing"

	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// cloudflareConfigYaml renders a CloudflareProviderConfig the way the runner injects it (protojson,
// camelCase); loadProviderConfigProto reads it back through YAMLToJSON -> protojson, so this
// exercises the same round-trip the live stack input takes.
func cloudflareConfigYaml(t *testing.T, cfg *cloudflareprovider.CloudflareProviderConfig) []byte {
	t.Helper()
	b, err := protojson.Marshal(cfg)
	require.NoError(t, err)
	return b
}

const (
	testCloudflareToken = "cf-test-token-0123456789abcdef"
	testR2AccessKeyID   = "r2-access-key-id-0123456789"
	testR2SecretKey     = "r2-secret-access-key-0123456789abcdef"
)

func TestLoadCloudflareEnvVars_TokenOnly_NoR2Variables(t *testing.T) {
	env, err := loadCloudflareEnvVars(cloudflareConfigYaml(t, &cloudflareprovider.CloudflareProviderConfig{
		AuthScheme: cloudflareprovider.CloudflareAuthScheme_api_token,
		ApiToken:   testCloudflareToken,
	}))
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"CLOUDFLARE_API_TOKEN": testCloudflareToken}, env)
}

func TestLoadCloudflareEnvVars_R2Pair_EmittedAsTfVars(t *testing.T) {
	// The R2 pair reaches the tofu modules as the nullable variables their credentials.tf declares,
	// never as AWS_* names an S3 state backend would also read.
	env, err := loadCloudflareEnvVars(cloudflareConfigYaml(t, &cloudflareprovider.CloudflareProviderConfig{
		AuthScheme: cloudflareprovider.CloudflareAuthScheme_api_token,
		ApiToken:   testCloudflareToken,
		R2: &cloudflareprovider.CloudflareCredentialsR2Spec{
			AccessKeyId:     testR2AccessKeyID,
			SecretAccessKey: testR2SecretKey,
			Endpoint:        "https://example.eu.r2.cloudflarestorage.com",
		},
	}))
	require.NoError(t, err)

	assert.Equal(t, testCloudflareToken, env["CLOUDFLARE_API_TOKEN"])
	assert.Equal(t, testR2AccessKeyID, env["TF_VAR_r2_access_key_id"])
	assert.Equal(t, testR2SecretKey, env["TF_VAR_r2_secret_access_key"])
	assert.Equal(t, "https://example.eu.r2.cloudflarestorage.com", env["TF_VAR_r2_endpoint"])
	for _, name := range []string{"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_ENDPOINT_URL_S3"} {
		_, present := env[name]
		assert.Falsef(t, present, "%s must never carry the R2 pair", name)
	}
}

func TestLoadCloudflareEnvVars_R2WithoutEndpoint_EndpointAbsent(t *testing.T) {
	// No endpoint means the module derives the account's default endpoint; an empty variable
	// would override that derivation with nothing.
	env, err := loadCloudflareEnvVars(cloudflareConfigYaml(t, &cloudflareprovider.CloudflareProviderConfig{
		AuthScheme: cloudflareprovider.CloudflareAuthScheme_api_token,
		ApiToken:   testCloudflareToken,
		R2: &cloudflareprovider.CloudflareCredentialsR2Spec{
			AccessKeyId:     testR2AccessKeyID,
			SecretAccessKey: testR2SecretKey,
		},
	}))
	require.NoError(t, err)

	_, endpointPresent := env["TF_VAR_r2_endpoint"]
	assert.False(t, endpointPresent)
	assert.Equal(t, testR2AccessKeyID, env["TF_VAR_r2_access_key_id"])
}

func TestLoadCloudflareEnvVars_LegacyKey_EmitsKeyAndEmail(t *testing.T) {
	env, err := loadCloudflareEnvVars(cloudflareConfigYaml(t, &cloudflareprovider.CloudflareProviderConfig{
		AuthScheme: cloudflareprovider.CloudflareAuthScheme_legacy_api_key,
		ApiKey:     "legacy-global-api-key-0123456789",
		Email:      "owner@example.com",
	}))
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"CLOUDFLARE_API_KEY": "legacy-global-api-key-0123456789",
		"CLOUDFLARE_EMAIL":   "owner@example.com",
	}, env)
}
