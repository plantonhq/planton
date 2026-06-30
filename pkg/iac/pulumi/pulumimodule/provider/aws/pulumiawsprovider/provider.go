// Package pulumiawsprovider is the single, convergent place where every AWS pulumi-aws
// "classic" module builds its aws.Provider from the stack input's AwsProviderConfig. It mirrors
// the sibling per-cloud builders (e.g. pulumiawsnativeprovider, pulumigoogleprovider) so a coding
// agent can learn the AWS credential-resolution path by reading one file.
//
// It dispatches on which fields of AwsProviderConfig are populated, supporting every auth mode
// with a single seam:
//   - web_identity set            -> keyless OIDC federation. The minted JWT is exchanged for
//     temporary AWS credentials via STS AssumeRoleWithWebIdentity (single hop for `oidc`;
//     web-identity + chained AssumeRole for `cross_account_trust`) and the result is injected as
//     static credentials. See "Why a builder-side exchange" below.
//   - static access keys set      -> long-lived/temporary access-key credentials.
//   - neither                     -> region only (the provider falls back to the SDK's ambient
//     credential chain, e.g. a self-hosted runner's instance role).
//
// Why a builder-side exchange for web identity (and not the provider-native inline token):
// AWS STS and the pulumi-aws provider both natively support a provider-level web-identity token
// (inline or file), so the obvious path is to hand the inline token to the provider and let its
// plugin exchange it. We deliberately do NOT do that, because pulumi-aws classic runs a
// pre-configure credential-validation step (validateCredentials) that is currently broken for
// AssumeRoleWithWebIdentity: it fails provider initialization BEFORE any STS call is made,
// surfacing as "Invalid credentials configured." with zero AssumeRoleWithWebIdentity events in
// CloudTrail. This is an upstream bug -- github.com/pulumi/pulumi-aws/issues/6228 (history in
// #2252) -- that reproduces even with a token FILE, so it is not an inline-token contract problem
// (the token we pass is a plain string, never a secret-wrapped value that could read as empty).
// Exchanging the token ourselves and passing the resulting temporary AccessKey/SecretKey/Token
// takes the provider's normal, working static-credential path and sidesteps the broken validation
// entirely. It also converges this builder with the aws-native and OpenTofu paths, which already
// resolve credentials through the same engine-neutral awswebidentity package.
//
// Freshness: the exchanged credentials do not auto-refresh, so each pulumi operation must run on a
// freshly minted token. The runner re-mints the web_identity_token before every operation and the
// module program re-runs per operation, so the exchange here always sees a fresh token whose
// assumed-role session (up to 1h; chained cross_account_trust capped at 1h by AWS) covers that one
// operation. No token is ever written to disk.
//
// SWITCH BACK TO PROVIDER-NATIVE WEB IDENTITY once pulumi-aws#6228 is fixed: verify the fix with a
// real keyless apply that reaches STS (CloudTrail shows AssumeRoleWithWebIdentity), then replace
// the awswebidentity.ResolveCredentials call in the web_identity arm of buildProviderArgs with the
// provider-native form -- set providerArgs.AssumeRoleWithWebIdentity to an
// aws.ProviderAssumeRoleWithWebIdentityArgs{RoleArn, WebIdentityToken, SessionName, Duration} and
// providerArgs.AssumeRoles to the chained hops (the cross_account_trust second hop) -- and drop the
// awswebidentity dependency. The provider then refreshes credentials itself, which lets the
// runner's per-operation re-mint relax to per-job. (The aws-native builder carries the analogous
// note for its own upstream gap, pulumi-aws-native#1042.)
//
// It deliberately never passes empty-string AccessKey/SecretKey to aws.NewProvider -- doing so is
// what produced "Invalid credentials configured." for keyless connections before this builder
// existed.
package pulumiawsprovider

import (
	"context"
	"fmt"
	"reflect"

	awsprovider "github.com/plantonhq/planton/apis/dev/planton/provider/aws"
	"github.com/plantonhq/planton/pkg/iac/provider/aws/awswebidentity"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds an aws.Provider from the given AwsProviderConfig. region is supplied by the
// caller (the resource's region) rather than read from the config so the provider's region
// matches the resource being provisioned. nameSuffixes disambiguate the provider resource
// name when a module needs more than one provider (e.g. multi-region).
func Get(ctx *pulumi.Context, awsProviderConfig *awsprovider.AwsProviderConfig,
	region string, nameSuffixes ...string) (*aws.Provider, error) {
	// ctx.Context() is the stack job's Go context; the STS exchange (when needed) runs on it.
	providerArgs, err := buildProviderArgs(ctx.Context(), awsProviderConfig, region, awswebidentity.ResolveCredentials)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build aws provider args")
	}

	awsProvider, err := aws.NewProvider(ctx, ProviderResourceName(nameSuffixes), providerArgs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws provider")
	}

	return awsProvider, nil
}

// buildProviderArgs is the pure, side-effect-free core of the builder: it maps an
// AwsProviderConfig to aws.ProviderArgs. It is split out from Get so the credential
// dispatch (the security-critical part) is unit-testable without a Pulumi context; resolve is
// an injectable seam so tests exercise the dispatch without a live STS endpoint (production
// passes awswebidentity.ResolveCredentials).
func buildProviderArgs(goCtx context.Context, awsProviderConfig *awsprovider.AwsProviderConfig,
	region string, resolve awswebidentity.CredentialResolver) (*aws.ProviderArgs, error) {
	providerArgs := &aws.ProviderArgs{}
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
		if err := awswebidentity.Validate(webIdentity); err != nil {
			return nil, err
		}

		// We exchange the JWT ourselves (rather than letting the provider plugin do it) to bypass
		// pulumi-aws#6228; see the package doc. resolve performs the single-hop oidc exchange plus
		// any chained cross_account_trust hops and returns the final temporary credentials.
		creds, err := resolve(goCtx, region, webIdentity)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve web-identity credentials via STS")
		}
		providerArgs.AccessKey = pulumi.String(creds.AccessKeyID)
		providerArgs.SecretKey = pulumi.String(creds.SecretAccessKey)
		if creds.SessionToken != "" {
			// The provider auto-secrets AccessKey/SecretKey but NOT Token, so secret it here to
			// keep the session token out of plaintext Pulumi state.
			providerArgs.Token = pulumi.ToSecret(pulumi.String(creds.SessionToken)).(pulumi.StringPtrInput)
		}

	case awsProviderConfig.GetAccessKeyId() != "":
		// Static credentials (long-lived AKIA or temporary ASIA + session token).
		providerArgs.AccessKey = pulumi.String(awsProviderConfig.GetAccessKeyId())
		providerArgs.SecretKey = pulumi.String(awsProviderConfig.GetSecretAccessKey())
		if awsProviderConfig.GetSessionToken() != "" {
			providerArgs.Token = pulumi.String(awsProviderConfig.GetSessionToken())
		}

	default:
		// Region-only: no explicit credentials in the config. The provider resolves
		// credentials from the SDK's ambient chain (e.g. a self-hosted runner's role).
	}

	return providerArgs, nil
}

// ProviderResourceName returns the Pulumi resource name for the AWS provider.
//
// The base is intentionally "classic-provider": every AWS module historically created
// its provider with exactly this name (aws.NewProvider(ctx, "classic-provider", ...)).
// Pulumi tracks providers by resource name, so keeping it stable lets existing modules
// adopt this shared builder without triggering a provider replacement -- and the
// resource churn that would follow -- in already-provisioned stacks. Do not rename
// without a state-migration plan.
func ProviderResourceName(suffixes []string) string {
	name := "classic-provider"
	for _, s := range suffixes {
		name = fmt.Sprintf("%s-%s", name, s)
	}
	return name
}

// PulumiOutputName builds a stable, prefixed output name for AWS resources, mirroring
// the helper exposed by the other per-cloud provider builders.
func PulumiOutputName(r interface{}, name string, suffixes ...string) string {
	outputName := fmt.Sprintf("aws_%s", pulumioutput.Name(reflect.TypeOf(r), name))
	for _, s := range suffixes {
		outputName = fmt.Sprintf("%s_%s", outputName, s)
	}
	return outputName
}
