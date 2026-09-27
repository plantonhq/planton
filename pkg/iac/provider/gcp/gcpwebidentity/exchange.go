// Package gcpwebidentity exchanges a keyless (OIDC web-identity) token for a short-lived Google
// Cloud access token, for the engines whose providers read credentials only from their
// environment (OpenTofu, Terraform).
//
// Why an exchange here rather than a credentials file: the google provider's keyless form is an
// Application Default Credentials "external account" file, and it is the wrong shape for a stack
// job. It would write the minted token to disk, which the provider config promises never happens,
// and every engine process would redo the exchange with a token minted to live minutes, so a
// command starting after that token expired would fail. Exchanging once, before any engine command
// runs, hands the job an access token that lives about an hour: the same contract the AWS keyless
// path gives a job (awswebidentity).
//
// The Pulumi path never calls this. pulumi-gcp exchanges the token inside the plugin
// (pulumigoogleprovider), so an exchange here would be wasted work.
//
// The exchange is Google's Workload Identity Federation flow in two steps: STS trades the minted
// token for a federated token (the audience is the pool provider's full resource name, byte for
// byte what the token was minted for), then IAM Credentials' generateAccessToken impersonates the
// named service account. The package is issuer-agnostic: the token is an opaque JWT minted by the
// caller, and nothing here talks to an issuer or a minter.
package gcpwebidentity

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	"golang.org/x/oauth2/google/externalaccount"
)

const (
	jwtSubjectTokenType    = "urn:ietf:params:oauth:token-type:jwt"
	cloudPlatformScope     = "https://www.googleapis.com/auth/cloud-platform"
	impersonationURLFormat = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken"
)

// TokenResolver exchanges a web-identity config for a Google Cloud access token. It is an
// injectable seam so callers can unit-test their credential dispatch (the security-critical part)
// without reaching Google; production passes ResolveAccessToken.
type TokenResolver func(ctx context.Context, webIdentity *gcpprovider.GcpWebIdentityProviderConfig) (string, error)

// ResolveAccessToken performs the exchange against Google's endpoints and returns the impersonated
// service account's access token.
func ResolveAccessToken(ctx context.Context, webIdentity *gcpprovider.GcpWebIdentityProviderConfig) (string, error) {
	return exchange(ctx, webIdentity, endpoints{impersonationURLFormat: impersonationURLFormat})
}

// endpoints are Google's two exchange URLs; a test points both at a fake. An empty tokenURL is the
// library's default STS endpoint.
type endpoints struct {
	tokenURL               string
	impersonationURLFormat string
}

func exchange(ctx context.Context, webIdentity *gcpprovider.GcpWebIdentityProviderConfig, ep endpoints) (string, error) {
	if err := Validate(webIdentity); err != nil {
		return "", err
	}
	source, err := externalaccount.NewTokenSource(ctx, externalaccount.Config{
		Audience:                       webIdentity.GetAudience(),
		SubjectTokenType:               jwtSubjectTokenType,
		TokenURL:                       ep.tokenURL,
		Scopes:                         []string{cloudPlatformScope},
		SubjectTokenSupplier:           staticSubjectToken(webIdentity.GetWebIdentityToken()),
		ServiceAccountImpersonationURL: fmt.Sprintf(ep.impersonationURLFormat, webIdentity.GetServiceAccountEmail()),
	})
	if err != nil {
		return "", errors.Wrap(err, "building the Google keyless exchange")
	}
	token, err := source.Token()
	if err != nil {
		return "", errors.Wrapf(err, "Google refused the keyless exchange for service account %s through %s "+
			"(check that the pool provider trusts this token's issuer and subject, and that the service account "+
			"grants roles/iam.workloadIdentityUser to the federated principal)",
			webIdentity.GetServiceAccountEmail(), webIdentity.GetAudience())
	}
	return token.AccessToken, nil
}

// Validate checks the invariants every consumer shares before attempting an exchange: a non-nil
// web identity with its token, the pool provider audience, and the service account to impersonate.
func Validate(webIdentity *gcpprovider.GcpWebIdentityProviderConfig) error {
	if webIdentity == nil {
		return errors.New("web_identity is nil")
	}
	if webIdentity.GetWebIdentityToken() == "" || webIdentity.GetAudience() == "" || webIdentity.GetServiceAccountEmail() == "" {
		return errors.New("web_identity requires web_identity_token, audience and service_account_email")
	}
	return nil
}

// staticSubjectToken adapts an inline minted JWT to externalaccount.SubjectTokenSupplier. In memory
// only: the token never touches disk.
type staticSubjectToken string

// SubjectToken returns the minted JWT.
func (t staticSubjectToken) SubjectToken(_ context.Context, _ externalaccount.SupplierOptions) (string, error) {
	return string(t), nil
}
