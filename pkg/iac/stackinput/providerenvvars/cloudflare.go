package providerenvvars

import (
	"github.com/pkg/errors"
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
)

// loadCloudflareEnvVars loads Cloudflare provider config and returns environment variables.
// Supports both API Token and Legacy API Key authentication schemes.
// R2 credentials are not mapped here -- they use S3-compatible auth and are
// separate from the Cloudflare provider plugin's API authentication.
func loadCloudflareEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(cloudflareprovider.CloudflareProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load Cloudflare provider config")
	}

	envVars := map[string]string{}

	switch config.AuthScheme {
	case cloudflareprovider.CloudflareAuthScheme_api_token:
		if config.ApiToken != "" {
			envVars["CLOUDFLARE_API_TOKEN"] = config.ApiToken
		}
	case cloudflareprovider.CloudflareAuthScheme_legacy_api_key:
		if config.ApiKey != "" {
			envVars["CLOUDFLARE_API_KEY"] = config.ApiKey
		}
		if config.Email != "" {
			envVars["CLOUDFLARE_EMAIL"] = config.Email
		}
	}

	return envVars, nil
}
