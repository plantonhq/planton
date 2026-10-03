// Package awswebidentity is the engine-neutral place that exchanges an OIDC web-identity
// JWT for temporary AWS credentials via STS. All three AWS keyless paths -- the pulumi-aws
// "classic" builder, the pulumi-aws-native builder, and the OpenTofu/Terraform provider-env
// path -- resolve credentials here, so the security-critical STS
// AssumeRoleWithWebIdentity exchange lives in one tested place instead of being
// copied per engine.
//
// Why a builder/runner-side exchange exists at all (rather than letting each provider plugin
// exchange the JWT natively):
//   - pulumi-aws "classic" CAN exchange a provider-level token natively, but its pre-configure
//     credential validation is currently broken for AssumeRoleWithWebIdentity (upstream
//     pulumi-aws#6228), failing provider init before STS is called -- so the classic builder
//     exchanges here and injects static credentials to take the provider's working static path.
//   - pulumi-aws-native has no web-identity field at all (upstream pulumi-aws-native#1042), so
//     it cannot exchange the JWT itself -- the caller must hand it temporary credentials.
//   - the OpenTofu AWS provider block is deliberately empty (region + credentials are
//     injected as env vars from the IaC input) -- so the runtime performs the exchange
//     and injects the resulting short-lived credentials.
//
// Each consumer documents its own switch-back trigger (when its upstream gap is fixed) in its
// own package doc. This package is issuer-agnostic: web_identity_token is an opaque OIDC JWT
// minted by the caller (e.g. the Planton runner); nothing here talks to any issuer or minter.
package awswebidentity

import (
	"context"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awsprovider "github.com/plantonhq/planton/catalog/aws"

	"github.com/pkg/errors"
)

// CredentialResolver exchanges a web-identity config for temporary AWS credentials. It is an
// injectable seam so callers can unit-test their credential dispatch (the security-critical
// part) without a live STS endpoint; production passes ResolveCredentials.
type CredentialResolver func(ctx context.Context, region string,
	webIdentity *awsprovider.AwsWebIdentityProviderConfig) (awssdk.Credentials, error)

// ResolveCredentials performs the STS exchange: AssumeRoleWithWebIdentity into the
// target role, returning the temporary credentials.
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

	provider := stscreds.NewWebIdentityRoleProvider(
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

	return awssdk.NewCredentialsCache(provider).Retrieve(ctx)
}

// Validate checks the invariants every consumer shares before attempting an exchange:
// a non-nil web identity with both a token and a role. Keeping it here means both
// engines reject malformed configs identically.
func Validate(webIdentity *awsprovider.AwsWebIdentityProviderConfig) error {
	if webIdentity == nil {
		return errors.New("web_identity is nil")
	}
	if webIdentity.GetWebIdentityToken() == "" || webIdentity.GetRoleArn() == "" {
		return errors.New("web_identity requires both web_identity_token and role_arn")
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
