package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// verifyCustomDomain asks Auth0 to verify the custom domain and waits until it
// is ready -- the provider polls until Auth0 reports "ready" and fails naming
// the last status otherwise. Verification is a one-time action with no update;
// its delete is the provider's no-op, so destroy leaves the domain verified.
func verifyCustomDomain(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.CustomDomainVerification, error) {
	verification, err := auth0.NewCustomDomainVerification(ctx, locals.ResourceName, &auth0.CustomDomainVerificationArgs{
		CustomDomainId: pulumi.String(locals.CustomDomainId),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to verify Auth0 custom domain %s", locals.CustomDomainId)
	}
	return verification, nil
}
