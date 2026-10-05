package providerenvvars

import (
	"github.com/pkg/errors"
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
)

// loadCloudflareEnvVars loads Cloudflare provider config and returns environment variables.
//
// Two credential planes, two delivery forms:
//
//   - The API credential (API Token, or the legacy Global API Key with its email) is what the
//     Cloudflare provider plugin authenticates with. The tofu modules ship an empty provider block,
//     so it rides the provider's own CLOUDFLARE_* environment names.
//   - The R2 key pair is S3-compatible and separate from the API credential; the API credential
//     cannot sign an S3 request. Only the kinds that read R2 objects through the S3 API declare it
//     (the Worker's r2_bundle source), as nullable Terraform variables in their iac/tf/credentials.tf,
//     so the bridge emits the TF_VAR_* forms. It never becomes AWS_ACCESS_KEY_ID /
//     AWS_SECRET_ACCESS_KEY: an S3 state backend authenticated from the environment reads those
//     same names, and the R2 pair would silently replace the backend's credentials.
//
// Terraform/OpenTofu silently ignores a TF_VAR_* env var the module does not declare, so the R2
// variables are safe to emit for every Cloudflare kind. Each is emitted only when present: an
// absent value must stay absent so a module's null default (the provider's own environment chain)
// applies, never an empty string that the provider would try to sign with.
func loadCloudflareEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(cloudflareprovider.CloudflareProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Cloudflare provider config")
	}

	envVars := map[string]string{}

	switch config.AuthScheme {
	case cloudflareprovider.CloudflareAuthScheme_api_token:
		putIfSet(envVars, "CLOUDFLARE_API_TOKEN", config.ApiToken)
	case cloudflareprovider.CloudflareAuthScheme_legacy_api_key:
		putIfSet(envVars, "CLOUDFLARE_API_KEY", config.ApiKey)
		putIfSet(envVars, "CLOUDFLARE_EMAIL", config.Email)
	}

	if r2 := config.GetR2(); r2 != nil {
		putIfSet(envVars, "TF_VAR_r2_access_key_id", r2.GetAccessKeyId())
		putIfSet(envVars, "TF_VAR_r2_secret_access_key", r2.GetSecretAccessKey())
		putIfSet(envVars, "TF_VAR_r2_endpoint", r2.GetEndpoint())
	}

	return envVars, nil
}
