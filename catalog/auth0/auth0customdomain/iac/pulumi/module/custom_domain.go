package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createCustomDomain creates the custom domain. Auth0 answers with the DNS record
// that proves control of the name (exported as dns_record_*); the domain serves
// nothing until an Auth0CustomDomainVerification has confirmed that record.
// Changing the domain or its type replaces it. Destroy deletes the domain.
func createCustomDomain(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.CustomDomain, error) {
	args := &auth0.CustomDomainArgs{
		Domain:                 pulumi.String(locals.Domain),
		Type:                   pulumi.String(locals.Type),
		CustomClientIpHeader:   pulumi.StringPtrFromPtr(locals.CustomClientIpHeader),
		TlsPolicy:              pulumi.StringPtrFromPtr(locals.TlsPolicy),
		RelyingPartyIdentifier: pulumi.StringPtrFromPtr(locals.RelyingPartyIdentifier),
	}
	if len(locals.DomainMetadata) > 0 {
		args.DomainMetadata = pulumi.ToStringMap(locals.DomainMetadata)
	}

	customDomain, err := auth0.NewCustomDomain(ctx, locals.ResourceName, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 custom domain %s", locals.Domain)
	}
	return customDomain, nil
}
