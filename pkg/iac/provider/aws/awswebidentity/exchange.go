// Package awswebidentity is the engine-neutral place that exchanges an OIDC web-identity
// JWT for temporary AWS credentials via STS. Both the pulumi-aws-native builder and the
// OpenTofu/Terraform provider-env path resolve credentials here, so the security-critical
// STS dance -- AssumeRoleWithWebIdentity (single hop for `oidc`) plus any chained
// AssumeRole hops (the `cross_account_trust` second hop) -- lives in one tested place
// instead of being copied per engine.
//
// Why a builder/runner-side exchange exists at all:
//   - pulumi-aws-native has no web-identity field (upstream #1042), so it cannot exchange
//     the JWT itself -- the caller must hand it temporary credentials.
//   - the OpenTofu AWS provider block is deliberately empty (region + credentials are
//     injected as env vars from the stack input), and a two-hop chained assume-role is not
//     expressible as a single set of SDK env vars -- so the runtime performs the exchange
//     and injects the resulting short-lived credentials.
//
// The pulumi-aws "classic" builder is the exception: it hands the inline token to the
// provider plugin and lets the plugin exchange it, so it does NOT use this package.
//
// This package is issuer-agnostic: web_identity_token is an opaque OIDC JWT minted by the
// caller (e.g. the Planton runner); nothing here talks to any issuer or minter.
package awswebidentity

import (
	"context"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws"

	"github.com/pkg/errors"
)

// CredentialResolver exchanges a web-identity config for temporary AWS credentials. It is an
// injectable seam so callers can unit-test their credential dispatch (the security-critical
// part) without a live STS endpoint; production passes ResolveCredentials.
type CredentialResolver func(ctx context.Context, region string,
	webIdentity *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error)

// ResolveCredentials performs the STS exchange: AssumeRoleWithWebIdentity into the first-hop
// role (oidc single hop), then any chained AssumeRole hops in order (cross_account_trust
// second hop into the customer role), returning the final temporary credentials.
//
// AssumeRoleWithWebIdentity needs no ambient credentials -- the JWT itself is the credential
// -- so the base config supplies only the region and the SDK's HTTP client; it does not read
// the runner's (non-existent) ambient AWS chain.
func ResolveCredentials(ctx context.Context, region string,
	webIdentity *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error) {

	baseCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
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

	// Each chained hop assumes the next role using the previous hop's credentials. Role
	// chaining caps the session at 1h (enforced by AWS, encoded by callers via Duration).
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

	return awssdk.NewCredentialsCache(provider).Retrieve(ctx)
}

// Validate checks the invariants every consumer shares before attempting an exchange:
// a non-nil web identity with both a token and a first-hop role, and a role on every
// chained hop. Keeping it here means both engines reject malformed configs identically.
func Validate(webIdentity *awsprovider.AwsWebIdentityProviderConfig) error {
	if webIdentity == nil {
		return errors.New("web_identity is nil")
	}
	// The builder-side exchange (this package, used by aws-native + tofu) is one-shot: it
	// resolves credentials once at build time and cannot refresh. A token *file* only adds value
	// where the provider itself re-reads it (the pulumi-aws classic provider), so reject it here
	// with an explanation instead of silently exchanging it once like an inline token.
	if webIdentity.GetWebIdentityTokenFile() != "" {
		return errors.New("web_identity_token_file is honored only by the pulumi-aws classic provider, " +
			"which re-reads it to refresh; the builder-side exchange (aws-native, tofu) is one-shot and " +
			"requires the inline web_identity_token")
	}
	if webIdentity.GetWebIdentityToken() == "" || webIdentity.GetRoleArn() == "" {
		return errors.New("web_identity requires both web_identity_token and role_arn")
	}
	for i, hop := range webIdentity.GetChainedAssumeRoles() {
		if hop.GetRoleArn() == "" {
			return errors.Errorf("chained_assume_roles[%d] requires role_arn", i)
		}
	}
	return nil
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
