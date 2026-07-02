package providerenvvars

import (
	"github.com/pkg/errors"
	gcpprovider "github.com/plantonhq/planton/apis/dev/planton/provider/gcp"
)

// loadGcpEnvVars builds the GCP provider environment variables from the resolved provider
// config:
//
//   - service_account_key set -> emit it as GOOGLE_CREDENTIALS (the terraform google provider
//     reads the key JSON from that variable).
//   - web_identity set -> fail loudly. The terraform google provider has no env-var form of
//     its external_credentials block, so keyless auth on the tofu/terraform path needs its own
//     deliberate design (per-module HCL vs a pre-exchanged access token) -- until that lands,
//     silently falling through to ambient credentials would run the deploy as whatever
//     identity the environment happens to hold, which is exactly the failure keyless auth
//     exists to prevent. The pulumi path supports web identity via the shared provider
//     builder.
//   - neither -> no credential env vars; the provider resolves credentials from the ambient
//     Application Default Credentials chain.
//
// Credential env vars are never emitted with empty values: an empty GOOGLE_CREDENTIALS would
// poison the provider's ambient credential chain.
func loadGcpEnvVars(providerConfigYaml []byte) (map[string]string, error) {
	config := new(gcpprovider.GcpProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load GCP provider config")
	}

	if config.GetWebIdentity() != nil {
		return nil, errors.New("GCP web-identity (keyless OIDC) auth is not supported on the " +
			"tofu/terraform path yet; use the pulumi provisioner or a service account key")
	}

	envVars := map[string]string{}
	if config.GetServiceAccountKey() != "" {
		envVars["GOOGLE_CREDENTIALS"] = config.GetServiceAccountKey()
	}

	return envVars, nil
}
