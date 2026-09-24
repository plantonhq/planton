package providerenvvars

import (
	"github.com/pkg/errors"
	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
)

// loadDigitalOceanEnvVars loads DigitalOcean provider config and returns environment variables.
//
// Unlike Cloudflare (whose tofu modules ship an empty provider block read from CLOUDFLARE_* env
// vars), the DigitalOcean tofu modules declare the credentials as Terraform variables
// (`token = var.digitalocean_token`, plus `var.spaces_access_id`/`var.spaces_secret_key` on the
// Spaces-backed kinds), so the bridge emits TF_VAR_* forms. Every one of those variables defaults
// to null, and a null token/key makes the provider fall back to its own environment defaults
// (DIGITALOCEAN_TOKEN, SPACES_ACCESS_KEY_ID, SPACES_SECRET_ACCESS_KEY) -- so a stack input without
// a provider config (a developer shell, the E2E harness) authenticates from the ambient
// environment exactly as the Pulumi module does. Terraform/OpenTofu silently ignores a TF_VAR_*
// env var the module does not declare, so the Spaces pair is safe to emit for every kind.
//
// default_region is intentionally not emitted: region is a resource property carried in each
// module's spec, not a provider-level input on the tofu path.
func loadDigitalOceanEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(digitaloceanprovider.DigitalOceanProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load DigitalOcean provider config")
	}

	envVars := map[string]string{}

	// Authentication (required by every DigitalOcean module).
	if config.ApiToken != "" {
		envVars["TF_VAR_digitalocean_token"] = config.ApiToken
	}

	// Spaces (S3-compatible object storage) key pair -- only the Spaces-backed
	// kinds declare these variables; emitted only when present so an absent pair
	// never materializes as empty-string variables.
	if config.SpacesAccessId != "" {
		envVars["TF_VAR_spaces_access_id"] = config.SpacesAccessId
	}
	if config.SpacesSecretKey != "" {
		envVars["TF_VAR_spaces_secret_key"] = config.SpacesSecretKey
	}

	return envVars, nil
}
