// Package pulumiawsnativeprovider is the convergent place where AWS pulumi-aws-native
// modules build their aws.Provider from the IaC input's AwsProviderConfig. It mirrors
// pulumiawsprovider (the pulumi-aws "classic" builder) so a coding agent can learn both
// AWS credential-resolution paths from one shape.
//
// The pulumi-aws-native provider has NO web-identity support: its ProviderArgs exposes
// only static AccessKey/SecretKey/Token, Region, RoleArn, and a single AssumeRole -- there
// is no AssumeRoleWithWebIdentity field (upstream tracking issue
// pulumi/pulumi-aws-native#1042, open since 2023). So unlike the classic builder -- which
// hands the inline web-identity token to the provider and lets the provider plugin exchange
// it -- this builder performs the STS exchange itself (via the engine-neutral
// awswebidentity package) and injects the resulting short-lived credentials as static keys.
//
// This builder-side exchange is the only way to make pulumi-aws-native keyless today; it is
// issuer-agnostic (the web_identity_token is an opaque OIDC JWT minted by the caller, e.g.
// the Planton runner) and adds no Planton coupling. When #1042 ships, collapse this onto the
// same inline-token model the classic builder uses and delete the builder-side exchange.
//
// Dispatch on which fields of AwsProviderConfig are populated:
//   - web_identity set       -> exchange the JWT for temporary creds via STS
//     AssumeRoleWithWebIdentity and inject them statically.
//   - static access keys set -> long-lived/temporary access-key credentials.
//   - neither                -> region only (the provider falls back to the SDK's ambient
//     credential chain).
package pulumiawsnativeprovider

import (
	"context"
	"fmt"
	"reflect"
	"time"

	awsprovider "github.com/plantonhq/planton/catalog/aws"
	"github.com/plantonhq/planton/pkg/iac/provider/aws/awswebidentity"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/pulumi/pulumioutput"

	"github.com/pkg/errors"
	awsnative "github.com/pulumi/pulumi-aws-native/sdk/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Get builds an aws-native Provider from the given AwsProviderConfig. region is supplied by
// the caller (the resource's region). nameSuffixes disambiguate the provider resource name
// when a module needs more than one provider.
func Get(ctx *pulumi.Context, awsProviderConfig *awsprovider.AwsProviderConfig,
	region string, nameSuffixes ...string) (*awsnative.Provider, error) {
	// ctx.Context() is the infra job's Go context; the STS exchange (when needed) runs on it.
	providerArgs, err := buildProviderArgs(ctx.Context(), awsProviderConfig, region, awswebidentity.ResolveCredentials)
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
	region string, resolve awswebidentity.CredentialResolver) (*awsnative.ProviderArgs, error) {
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
		if err := awswebidentity.Validate(webIdentity); err != nil {
			return nil, err
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

	// The provider-block surface composes with every base-credential mode (the
	// classic builder and the OpenTofu override file carry the same contract).
	if err := applyProviderBlockArgs(providerArgs, awsProviderConfig); err != nil {
		return nil, err
	}

	return providerArgs, nil
}

// applyProviderBlockArgs maps the config's provider-block surface onto aws-native
// ProviderArgs. pulumi-aws-native's surface is narrower than the classic
// provider's, and every gap is a HARD ERROR here, never a silent degradation: a
// truncated assume-role chain would deploy as the wrong identity, and silently
// dropped retry tuning would lie about the config in effect. The error names the
// remedy (the OpenTofu engine or a classic-provider module carry the full surface).
func applyProviderBlockArgs(providerArgs *awsnative.ProviderArgs, config *awsprovider.AwsProviderConfig) error {
	if chain := config.GetAssumeRoleChain(); len(chain) > 0 {
		// Native's AssumeRole is a singular field -- there is no chain (unlike the
		// classic provider's assumeRoles list and Terraform's repeated assume_role).
		if len(chain) > 1 {
			return errors.Errorf(
				"the pulumi-aws-native provider supports a single assume_role hop, but the "+
					"provider config carries a chain of %d roles; deploy this resource with "+
					"OpenTofu or a pulumi-aws (classic) module, or shorten the chain", len(chain))
		}
		hop := chain[0]
		if hop.GetSourceIdentity() != "" {
			return errors.New(
				"the pulumi-aws-native provider has no source_identity support on assume_role; " +
					"deploy with OpenTofu or a pulumi-aws (classic) module, or drop source_identity")
		}
		hopArgs := &awsnative.ProviderAssumeRoleArgs{
			RoleArn: pulumi.String(hop.GetRoleArn()),
		}
		if hop.GetSessionName() != "" {
			hopArgs.SessionName = pulumi.String(hop.GetSessionName())
		}
		if hop.GetExternalId() != "" {
			hopArgs.ExternalId = pulumi.String(hop.GetExternalId())
		}
		if hop.GetDuration() != "" {
			// Native takes seconds where the classic provider and Terraform take a
			// duration string; convert rather than expose a second vocabulary.
			d, err := time.ParseDuration(hop.GetDuration())
			if err != nil {
				return errors.Wrapf(err, "invalid assume_role duration %q", hop.GetDuration())
			}
			hopArgs.DurationSeconds = pulumi.Int(int(d.Seconds()))
		}
		if hop.GetPolicy() != "" {
			hopArgs.Policy = pulumi.String(hop.GetPolicy())
		}
		if len(hop.GetPolicyArns()) > 0 {
			hopArgs.PolicyArns = pulumi.ToStringArray(hop.GetPolicyArns())
		}
		if len(hop.GetTags()) > 0 {
			hopArgs.Tags = pulumi.ToStringMap(hop.GetTags())
		}
		if len(hop.GetTransitiveTagKeys()) > 0 {
			hopArgs.TransitiveTagKeys = pulumi.ToStringArray(hop.GetTransitiveTagKeys())
		}
		providerArgs.AssumeRole = hopArgs
	}

	if tags := config.GetDefaultTags().GetTags(); len(tags) > 0 {
		providerArgs.DefaultTags = awsnative.ProviderDefaultTagsArgs{Tags: pulumi.ToStringMap(tags)}
	}

	if endpoints := config.GetEndpoints(); len(endpoints) > 0 {
		endpointArgs, err := buildEndpointArgs(endpoints)
		if err != nil {
			return err
		}
		providerArgs.Endpoints = awsnative.ProviderEndpointArray{endpointArgs}
	}

	if config.MaxRetries != nil {
		providerArgs.MaxRetries = pulumi.Int(int(config.GetMaxRetries()))
	}
	if config.GetRetryMode() != "" {
		// Native has no retry-mode field at all.
		return errors.Errorf(
			"the pulumi-aws-native provider has no retry_mode support (got %q); deploy with "+
				"OpenTofu or a pulumi-aws (classic) module, or drop retry_mode", config.GetRetryMode())
	}

	return nil
}

// buildEndpointArgs maps service->URL endpoint overrides onto native's
// ProviderEndpointArgs via its `pulumi:"<service>"` tags. Native's endpoint
// vocabulary is a small subset of the classic provider's (it fronts AWS Cloud
// Control), so most classic service names are unknown here -- and unknown names
// fail loudly, naming the engines that carry the full vocabulary.
func buildEndpointArgs(endpoints map[string]string) (awsnative.ProviderEndpointArgs, error) {
	endpointArgs := awsnative.ProviderEndpointArgs{}
	structValue := reflect.ValueOf(&endpointArgs).Elem()
	structType := structValue.Type()

	fieldByService := make(map[string]int, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		if tag := structType.Field(i).Tag.Get("pulumi"); tag != "" {
			fieldByService[tag] = i
		}
	}

	for service, url := range endpoints {
		fieldIndex, known := fieldByService[service]
		if !known {
			return awsnative.ProviderEndpointArgs{}, errors.Errorf(
				"the pulumi-aws-native provider has no endpoint attribute for service %q "+
					"(its endpoint vocabulary is a small subset of the classic provider's); "+
					"deploy with OpenTofu or a pulumi-aws (classic) module for full endpoint "+
					"overrides", service)
		}
		structValue.Field(fieldIndex).Set(reflect.ValueOf(pulumi.StringPtrInput(pulumi.String(url))))
	}

	return endpointArgs, nil
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
