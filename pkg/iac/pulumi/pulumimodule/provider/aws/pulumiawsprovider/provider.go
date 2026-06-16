// Package pulumiawsprovider is the single, convergent place where every AWS Pulumi
// module builds its aws.Provider from the stack input's AwsProviderConfig. It mirrors
// the per-cloud builders for other providers (e.g. pulumigoogleprovider) so that a
// coding agent can learn the AWS credential-resolution path by reading one file.
//
// It dispatches on which fields of AwsProviderConfig are populated, supporting every
// auth mode with a single seam:
//   - web_identity set            -> keyless OIDC federation via STS AssumeRoleWithWebIdentity
//     (single hop for `oidc`; web-identity + chained AssumeRole for `cross_account_trust`).
//   - static access keys set      -> long-lived/temporary access-key credentials.
//   - neither                     -> region only (the provider falls back to the SDK's
//     ambient credential chain, e.g. a self-hosted runner's instance role).
//
// It deliberately never passes empty-string AccessKey/SecretKey to aws.NewProvider --
// doing so is what produced "Invalid credentials configured." for keyless connections
// before this builder existed.
package pulumiawsprovider

import (
	"fmt"
	"reflect"

	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

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
	providerArgs, err := buildProviderArgs(awsProviderConfig, region)
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
// dispatch (the security-critical part) is unit-testable without a Pulumi context.
func buildProviderArgs(awsProviderConfig *awsprovider.AwsProviderConfig,
	region string) (*aws.ProviderArgs, error) {
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
		// Keyless OIDC federation. The caller (e.g. the runner) supplies the minted JWT and the
		// provider exchanges it for credentials via STS AssumeRoleWithWebIdentity. The token is
		// supplied either inline in memory (web_identity_token) or as a path the provider re-reads
		// (web_identity_token_file) -- the latter lets a long-running job outlive a single
		// assumed-role session, since the classic provider re-reads the file to refresh.
		webIdentity := awsProviderConfig.GetWebIdentity()
		token := webIdentity.GetWebIdentityToken()
		tokenFile := webIdentity.GetWebIdentityTokenFile()
		if webIdentity.GetRoleArn() == "" {
			return nil, errors.New("web_identity requires role_arn")
		}
		// Exactly one token source. Both-or-neither is a malformed config; reject it here as
		// defense in depth alongside the proto message-level CEL (token_xor_file).
		if (token == "") == (tokenFile == "") {
			return nil, errors.New("web_identity requires exactly one of web_identity_token or web_identity_token_file")
		}

		assumeRoleWithWebIdentity := aws.ProviderAssumeRoleWithWebIdentityArgs{
			RoleArn: pulumi.String(webIdentity.GetRoleArn()),
		}
		if tokenFile != "" {
			assumeRoleWithWebIdentity.WebIdentityTokenFile = pulumi.String(tokenFile)
		} else {
			assumeRoleWithWebIdentity.WebIdentityToken = pulumi.String(token)
		}
		if webIdentity.GetSessionName() != "" {
			assumeRoleWithWebIdentity.SessionName = pulumi.String(webIdentity.GetSessionName())
		}
		if webIdentity.GetDuration() != "" {
			assumeRoleWithWebIdentity.Duration = pulumi.String(webIdentity.GetDuration())
		}
		providerArgs.AssumeRoleWithWebIdentity = assumeRoleWithWebIdentity

		// Optional chained hops applied after the web-identity assumption (the
		// cross_account_trust second hop into the customer role). Empty for oidc.
		if len(webIdentity.GetChainedAssumeRoles()) > 0 {
			chainedAssumeRoles := make(aws.ProviderAssumeRoleArray, 0, len(webIdentity.GetChainedAssumeRoles()))
			for i, hop := range webIdentity.GetChainedAssumeRoles() {
				if hop.GetRoleArn() == "" {
					return nil, errors.Errorf("chained_assume_roles[%d] requires role_arn", i)
				}
				assumeRole := aws.ProviderAssumeRoleArgs{
					RoleArn: pulumi.String(hop.GetRoleArn()),
				}
				if hop.GetExternalId() != "" {
					assumeRole.ExternalId = pulumi.String(hop.GetExternalId())
				}
				if hop.GetSessionName() != "" {
					assumeRole.SessionName = pulumi.String(hop.GetSessionName())
				}
				if hop.GetDuration() != "" {
					assumeRole.Duration = pulumi.String(hop.GetDuration())
				}
				chainedAssumeRoles = append(chainedAssumeRoles, assumeRole)
			}
			providerArgs.AssumeRoles = chainedAssumeRoles
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
