package providerenvvars

import (
	"testing"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// digitalOceanConfigYaml renders a DigitalOceanProviderConfig the way the runner injects it
// (protojson, camelCase); loadProviderConfigProto reads it back through YAMLToJSON -> protojson,
// so this exercises the same round-trip the live IaC input takes.
func digitalOceanConfigYaml(t *testing.T, cfg *digitaloceanprovider.DigitalOceanProviderConfig) []byte {
	t.Helper()
	b, err := protojson.Marshal(cfg)
	require.NoError(t, err)
	return b
}

func TestLoadDigitalOceanEnvVars_Token_EmittedAsTfVar(t *testing.T) {
	// The DO tofu modules declare the token as a required Terraform variable
	// (token = var.digitalocean_token), so the TF_VAR_* form is the contract.
	env, err := loadDigitalOceanEnvVars(digitalOceanConfigYaml(t, &digitaloceanprovider.DigitalOceanProviderConfig{
		ApiToken: "dop_v1_test",
	}))
	require.NoError(t, err)

	assert.Equal(t, "dop_v1_test", env["TF_VAR_digitalocean_token"])
	_, spacesAccessPresent := env["TF_VAR_spaces_access_id"]
	assert.False(t, spacesAccessPresent)
	_, spacesSecretPresent := env["TF_VAR_spaces_secret_key"]
	assert.False(t, spacesSecretPresent)
}

func TestLoadDigitalOceanEnvVars_SpacesPair_EmittedWhenPresent(t *testing.T) {
	env, err := loadDigitalOceanEnvVars(digitalOceanConfigYaml(t, &digitaloceanprovider.DigitalOceanProviderConfig{
		ApiToken:        "dop_v1_test",
		SpacesAccessId:  "SPACESKEY",
		SpacesSecretKey: "spacessecret",
	}))
	require.NoError(t, err)

	assert.Equal(t, "SPACESKEY", env["TF_VAR_spaces_access_id"])
	assert.Equal(t, "spacessecret", env["TF_VAR_spaces_secret_key"])
}

func TestLoadDigitalOceanEnvVars_EmptyConfig_NoEmptyEnvVars(t *testing.T) {
	// Empty-string TF_VARs would satisfy a required variable with a broken value;
	// absent fields must produce absent variables so tofu fails loudly instead.
	env, err := loadDigitalOceanEnvVars(digitalOceanConfigYaml(t, &digitaloceanprovider.DigitalOceanProviderConfig{}))
	require.NoError(t, err)

	assert.Empty(t, env)
}
