package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the tenant settings as the tenant carries them after the
// apply, managed or not.
func exportOutputs(ctx *pulumi.Context, tenant *auth0.Tenant) error {
	ctx.Export("friendly_name", tenant.FriendlyName)
	ctx.Export("picture_url", tenant.PictureUrl)
	ctx.Export("support_email", tenant.SupportEmail)
	ctx.Export("support_url", tenant.SupportUrl)

	return nil
}
