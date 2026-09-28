package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the tenant settings as the tenant carries them after the
// apply, managed or not, and the default domain when this resource sets it.
func exportOutputs(ctx *pulumi.Context, tenant *auth0.Tenant, defaultDomain *auth0.CustomDomainDefault) error {
	ctx.Export("friendly_name", tenant.FriendlyName)
	ctx.Export("picture_url", tenant.PictureUrl)
	ctx.Export("support_email", tenant.SupportEmail)
	ctx.Export("support_url", tenant.SupportUrl)

	if defaultDomain != nil {
		ctx.Export("default_custom_domain", defaultDomain.Domain)
	} else {
		ctx.Export("default_custom_domain", pulumi.String(""))
	}

	return nil
}
