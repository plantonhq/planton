// Package pulumiawsnativeprovider is the convergent place where AWS pulumi-aws-native
// modules build their aws.Provider from the stack input's AwsProviderConfig. It mirrors
// pulumiawsprovider (the pulumi-aws "classic" builder) so a coding agent can learn both
// AWS credential-resolution paths from one shape.
//
// The pulumi-aws-native provider has NO web-identity support: its ProviderArgs exposes
// only static AccessKey/SecretKey/Token, Region, RoleArn, and a single AssumeRole -- there
// is no AssumeRoleWithWebIdentity field (upstream tracking issue
// pulumi/pulumi-aws-native#1042, open since 2023). So unlike the classic builder -- which
// hands the inline web-identity token to the provider and lets the provider plugin exchange
// it -- this builder performs the STS AssumeRoleWithWebIdentity exchange itself (via
// aws-sdk-go-v2) and injects the resulting short-lived credentials as static keys.
//
// This builder-side exchange is the only way to make pulumi-aws-native keyless today; it is
// issuer-agnostic (the web_identity_token is an opaque OIDC JWT minted by the caller, e.g.
// the Planton runner) and adds no Planton coupling. When #1042 ships, collapse this onto the
// same inline-token model the classic builder uses and delete the builder-side exchange.
//
// Dispatch on which fields of AwsProviderConfig are populated:
//   - web_identity set       -> exchange the JWT for temporary creds (single hop for oidc;
//     web-identity + chained AssumeRole for cross_account_trust) and inject them statically.
//   - static access keys set -> long-lived/temporary access-key credentials.
//   - neither                -> region only (the provider falls back to the SDK's ambient
//     credential chain).
package pulumiawsnativeprovider

import (
	"context"
	"fmt"
	"reflect"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	awsnative "github.com/pulumi/pulumi-aws-native/sdk/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// credentialResolver exchanges a web-identity config for temporary AWS credentials. It is an
// injectable seam so the credential dispatch (the security-critical part) is unit-testable
// without a live STS endpoint; production uses resolveWebIdentityCredentials.
type credentialResolver func(goCtx context.Context, region string,
	webIdentity *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error)

// Get builds an aws-native Provider from the given AwsProviderConfig. region is supplied by
// the caller (the resource's region). nameSuffixes disambiguate the provider resource name
// when a module needs more than one provider.
func Get(ctx *pulumi.Context, awsProviderConfig *awsprovider.AwsProviderConfig,
	region string, nameSuffixes ...string) (*awsnative.Provider, error) {
	// ctx.Context() is the stack job's Go context; the STS exchange (when needed) runs on it.
	providerArgs, err := buildProviderArgs(ctx.Context(), awsProviderConfig, region, resolveWebIdentityCredentials)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build aws-native provider args")
	}

	awsNativeProvider, err := awsnative.NewProvider(ctx, ProviderResourceName(nameSuffixes), providerArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws-native provider")
	}

	return awsNativeProvider, nil
}

// buildProviderArgs maps an AwsProviderConfig to aws-native ProviderArgs. For the web-identity
// arm it calls resolve to perform the STS exchange and injects the temporary credentials; this
// is the structural difference from the classic builder, forced by pulumi-aws-native#1042.
func buildProviderArgs(goCtx context.Context, awsProviderConfig *awsprovider.AwsProviderConfig,
	region string, resolve credentialResolver) (*awsnative.ProviderArgs, error) {
	providerArgs := &awsnative.ProviderArgs{}
	if region != "" {
		providerArgs.Region = pulumi.String(region)
	}

	// No config -> region-only provider (ambient credential chain).
	if awsProviderConfig == nil {
		return providerArgs, nil
	}

	switch {
	case awsProviderConfig.GetWebIdentity() != nil:
		webIdentity := awsProviderConfig.GetWebIdentity()
		if webIdentity.GetWebIdentityToken() == "" || webIdentity.GetRoleArn() == "" {
			return nil, errors.New("web_identity requires both web_identity_token and role_arn")
		}
		for i, hop := range webIdentity.GetChainedAssumeRoles() {
			if hop.GetRoleArn() == "" {
				return nil, errors.Errorf("chained_assume_roles[%d] requires role_arn", i)
			}
		}

		// pulumi-aws-native cannot exchange the JWT itself, so we resolve credentials here and
		// pass them as static keys. They are short-lived (the assumed-role session) and
		// connection-scoped -- far lower blast radius than long-lived keys.
		creds, err := resolve(goCtx, region, webIdentity)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve web-identity credentials via STS")
		}
		providerArgs.AccessKey = pulumi.String(creds.AccessKeyID)
		providerArgs.SecretKey = pulumi.String(creds.SecretAccessKey)
		if creds.SessionToken != "" {
			// The provider auto-secrets AccessKey/SecretKey but NOT Token, so secret it here
			// to keep the session token out of plaintext Pulumi state.
			providerArgs.Token = pulumi.ToSecret(pulumi.String(creds.SessionToken)).(pulumi.StringPtrInput)
		}

	case awsProviderConfig.GetAccessKeyId() != "":
		providerArgs.AccessKey = pulumi.String(awsProviderConfig.GetAccessKeyId())
		providerArgs.SecretKey = pulumi.String(awsProviderConfig.GetSecretAccessKey())
		if awsProviderConfig.GetSessionToken() != "" {
			providerArgs.Token = pulumi.String(awsProviderConfig.GetSessionToken())
		}

	default:
		// Region-only: no explicit credentials in the config.
	}

	return providerArgs, nil
}

// resolveWebIdentityCredentials performs the STS exchange: AssumeRoleWithWebIdentity into the
// first-hop role (oidc single hop), then any chained AssumeRole hops (cross_account_trust),
// and returns the final temporary credentials. AssumeRoleWithWebIdentity needs no ambient
// credentials (the JWT is the credential); the base config supplies only the region/HTTP client.
func resolveWebIdentityCredentials(goCtx context.Context, region string,
	webIdentity *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {
	baseCfg, err := awsconfig.LoadDefaultConfig(goCtx, awsconfig.WithRegion(region))
	if err != nil {
		return awssdk.Credentials{}, errors.Wrap(err, "loading base AWS config")
	}

	var provider awssdk.CredentialsProvider = stscreds.NewWebIdentityRoleProvider(
		sts.NewFromConfig(baseCfg),
		webIdentity.GetRoleArn(),
		identityToken(webIdentity.GetWebIdentityToken()),
		func(o *stscreds.WebIdentityRoleOptions) {
			if webIdentity.GetSessionName() != "" {
				o.RoleSessionName = webIdentity.GetSessionName()
			}
			if d := parseDuration(webIdentity.GetDuration()); d > 0 {
				o.Duration = d
			}
		},
	)

	// Each chained hop assumes the next role using the previous hop's credentials.
	for _, hop := range webIdentity.GetChainedAssumeRoles() {
		hopCfg := baseCfg.Copy()
		hopCfg.Credentials = awssdk.NewCredentialsCache(provider)
		h := hop
		provider = stscreds.NewAssumeRoleProvider(
			sts.NewFromConfig(hopCfg),
			h.GetRoleArn(),
			func(o *stscreds.AssumeRoleOptions) {
				if h.GetExternalId() != "" {
					o.ExternalID = awssdk.String(h.GetExternalId())
				}
				if h.GetSessionName() != "" {
					o.RoleSessionName = h.GetSessionName()
				}
				if d := parseDuration(h.GetDuration()); d > 0 {
					o.Duration = d
				}
			},
		)
	}

	return awssdk.NewCredentialsCache(provider).Retrieve(goCtx)
}

// parseDuration returns the parsed duration, or 0 when empty/invalid (the provider default applies).
func parseDuration(d string) time.Duration {
	if d == "" {
		return 0
	}
	parsed, err := time.ParseDuration(d)
	if err != nil {
		return 0
	}
	return parsed
}

// identityToken adapts an inline minted JWT to the AWS SDK's stscreds.IdentityTokenRetriever
// (the SDK calls GetIdentityToken each time it exchanges the token at STS).
type identityToken string

// GetIdentityToken returns the minted JWT bytes.
func (t identityToken) GetIdentityToken() ([]byte, error) {
	return []byte(t), nil
}

// ProviderResourceName returns the Pulumi resource name for the aws-native provider.
//
// The base is intentionally "native-provider": modules that use the aws-native provider create
// it with exactly this name. Pulumi tracks providers by resource name, so keeping it stable lets
// modules adopt this shared builder without triggering a provider replacement. Do not rename
// without a state-migration plan.
func ProviderResourceName(suffixes []string) string {
	name := "native-provider"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}

// PulumiOutputName builds a stable, prefixed output name for aws-native resources, mirroring the
// helper exposed by the other per-cloud provider builders.
func PulumiOutputName(r interface{}, name string, suffixes ...string) string {
	outputName := fmt.Sprintf("aws_%s", pulumioutput.Name(reflect.TypeOf(r), name))
	for _, s := range suffixes {
		outputName = fmt.Sprintf("%s_%s", outputName, s)
	}
	return outputName
}
