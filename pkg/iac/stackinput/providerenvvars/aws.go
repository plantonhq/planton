package providerenvvars

import (
	"context"
	"time"

	"github.com/pkg/errors"
	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	"github.com/plantonhq/openmcf/pkg/iac/provider/aws/awswebidentity"
)

// awsWebIdentityExchangeTimeout bounds the synchronous STS exchange done on the tofu path.
// The exchange is a single AssumeRoleWithWebIdentity (+ optional chained AssumeRole) run once,
// before any tofu command; this ceiling protects the stack job from a hung STS endpoint. We use
// a fresh context.Background() rather than threading a caller context so the public
// providerenvvars/tofumodule signatures stay stable (keeping this change wholly within
// openmcf); the minted JWT's own short TTL bounds credential validity independently.
const awsWebIdentityExchangeTimeout = 60 * time.Second

// loadAwsEnvVars builds the AWS provider environment variables from the resolved provider config.
// The AWS tofu/terraform modules ship an empty `provider "aws" {}` block, so BOTH region and
// credentials are injected here as env vars:
//
//   - AWS_REGION is always set from the resource's region (the connection region in
//     provider_config is only a fallback) -- region is a resource property, mirroring how the
//     pulumi builders take the resource's spec.Region rather than the connection's.
//   - web identity + ResolveAwsWebIdentity -> exchange the JWT via STS and emit the resulting
//     temporary AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_SESSION_TOKEN (the tofu path).
//   - web identity + !ResolveAwsWebIdentity -> region only; the pulumi in-program builder owns
//     the exchange, so emitting (shadowed) creds here would be wasteful.
//   - static keys -> emit them, including AWS_SESSION_TOKEN for temporary (ASIA) credentials.
//   - neither -> region only; the provider resolves credentials from the ambient chain.
//
// Credential env vars are never emitted with empty values: an empty AWS_ACCESS_KEY_ID would
// poison the SDK's ambient credential chain (the runner / region-only mode).
func loadAwsEnvVars(providerConfigYaml []byte, hasProviderConfig bool, resourceRegion string,
	opts Options, resolve awswebidentity.CredentialResolver) (map[string]string, error) {

	envVars := map[string]string{}
	region := resourceRegion

	var config *awsprovider.AwsProviderConfig
	if hasProviderConfig {
		config = new(awsprovider.AwsProviderConfig)
		if err := loadProviderConfigProto(providerConfigYaml, config); err != nil {
			return nil, errors.Wrap(err, "failed to load AWS provider config")
		}
		if region == "" {
			region = config.GetRegion()
		}
	}

	if region != "" {
		envVars["AWS_REGION"] = region
	}

	// No provider_config -> ambient credential chain; region only.
	if config == nil {
		return envVars, nil
	}

	switch {
	case config.GetWebIdentity() != nil:
		if !opts.ResolveAwsWebIdentity {
			// Pulumi path: the in-program builder performs the exchange; nothing to inject here.
			break
		}
		if err := awswebidentity.Validate(config.GetWebIdentity()); err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), awsWebIdentityExchangeTimeout)
		defer cancel()
		creds, err := resolve(ctx, region, config.GetWebIdentity())
		if err != nil {
			return nil, errors.Wrap(err, "resolving AWS web-identity credentials via STS")
		}
		envVars["AWS_ACCESS_KEY_ID"] = creds.AccessKeyID
		envVars["AWS_SECRET_ACCESS_KEY"] = creds.SecretAccessKey
		if creds.SessionToken != "" {
			envVars["AWS_SESSION_TOKEN"] = creds.SessionToken
		}

	case config.GetAccessKeyId() != "":
		envVars["AWS_ACCESS_KEY_ID"] = config.GetAccessKeyId()
		envVars["AWS_SECRET_ACCESS_KEY"] = config.GetSecretAccessKey()
		if config.GetSessionToken() != "" {
			envVars["AWS_SESSION_TOKEN"] = config.GetSessionToken()
		}

	default:
		// Region-only: ambient credential chain.
	}

	return envVars, nil
}
