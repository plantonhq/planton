package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyDefaultCustomDomain sets the tenant's default domain -- the one its email
// links and Management API notifications use -- when the spec manages it, and
// declares nothing otherwise. Auth0 has no way to unset a default, so the
// resource's delete only forgets it: destroy leaves the last-applied default in
// place.
func applyDefaultCustomDomain(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.CustomDomainDefault, error) {
	if locals.DefaultCustomDomain == nil {
		return nil, nil
	}
	defaultDomain, err := auth0.NewCustomDomainDefault(ctx, locals.ResourceName+"-default-custom-domain", &auth0.CustomDomainDefaultArgs{
		Domain: pulumi.String(*locals.DefaultCustomDomain),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make %s the default domain of the Auth0 tenant for %s", *locals.DefaultCustomDomain, locals.ResourceName)
	}
	return defaultDomain, nil
}
