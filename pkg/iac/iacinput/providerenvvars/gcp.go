package providerenvvars

import (
	"context"

	"github.com/pkg/errors"
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"github.com/plantonhq/planton/pkg/iac/provider/gcp/gcpwebidentity"
)

// loadGcpEnvVars builds the GCP provider environment variables from the resolved provider
// config:
//
//   - web_identity, EngineReadsEnvironment -> exchange the minted token for the impersonated
//     service account's access token (gcpwebidentity) and emit it as GOOGLE_OAUTH_ACCESS_TOKEN,
//     the same variable the access_token arm emits. The google provider's only other keyless form
//     is a credentials file, which would put the token on disk and re-exchange a minutes-long
//     token in every engine process; the package doc of gcpwebidentity has the full reasoning.
//   - web_identity, EngineBuildsProviders -> nothing; pulumi-gcp exchanges the token inside the
//     plugin from the provider config itself.
//   - web_identity, no engine named -> refused (see Engine).
//   - access_token set -> emit it as GOOGLE_OAUTH_ACCESS_TOKEN (the terraform google
//     provider's env form of its access_token argument). It deliberately wins over a stale
//     service_account_key, and only ONE credential variable is ever emitted: the token was
//     minted for this run, and emitting both would leave the effective identity to provider
//     precedence rules instead of this dispatch.
//   - service_account_key set -> emit it as GOOGLE_CREDENTIALS (the terraform google provider
//     reads the key JSON from that variable).
//   - none -> no credential env vars; the provider resolves credentials from the ambient
//     Application Default Credentials chain (runner mode: the runner's own identity).
//
// Credential env vars are never emitted with empty values: an empty GOOGLE_CREDENTIALS or
// GOOGLE_OAUTH_ACCESS_TOKEN would poison the provider's ambient credential chain.
func loadGcpEnvVars(providerConfigYaml []byte, opts Options, resolve gcpwebidentity.TokenResolver) (map[string]string, error) {
	config := new(gcpprovider.GcpProviderConfig)
	if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
		return nil, errors.Wrap(err, "failed to load GCP provider config")
	}

	envVars := map[string]string{}
	if webIdentity := config.GetWebIdentity(); webIdentity != nil {
		switch opts.Engine {
		case EngineBuildsProviders:
			return envVars, nil
		case EngineReadsEnvironment:
		default:
			return nil, errEngineUnset("Google Cloud")
		}
		if err := gcpwebidentity.Validate(webIdentity); err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), keylessExchangeTimeout)
		defer cancel()
		token, err := resolve(ctx, webIdentity)
		if err != nil {
			return nil, errors.Wrap(err, "resolving the Google Cloud keyless credential")
		}
		envVars["GOOGLE_OAUTH_ACCESS_TOKEN"] = token
		return envVars, nil
	}
	if config.GetAccessToken() != "" {
		envVars["GOOGLE_OAUTH_ACCESS_TOKEN"] = config.GetAccessToken()
		return envVars, nil
	}
	if config.GetServiceAccountKey() != "" {
		envVars["GOOGLE_CREDENTIALS"] = config.GetServiceAccountKey()
	}

	return envVars, nil
}
