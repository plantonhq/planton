package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyTenantSettings manages the presentation settings of the EXISTING tenant
// the provider's credential belongs to; the Management API cannot create or
// delete a tenant. Only the four settings are set, each only when the spec sets
// it, so the tenant's other settings (session lifetimes, flags, error pages) are
// never touched. The resource's delete is the provider's no-op: destroy leaves
// the last-applied values in place, as Auth0 has no delete for tenant settings.
func applyTenantSettings(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Tenant, error) {
	tenant, err := auth0.NewTenant(ctx, locals.ResourceName, &auth0.TenantArgs{
		FriendlyName: pulumi.StringPtrFromPtr(locals.FriendlyName),
		PictureUrl:   pulumi.StringPtrFromPtr(locals.PictureUrl),
		SupportEmail: pulumi.StringPtrFromPtr(locals.SupportEmail),
		SupportUrl:   pulumi.StringPtrFromPtr(locals.SupportUrl),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply the settings of the Auth0 tenant for %s", locals.ResourceName)
	}
	return tenant, nil
}
