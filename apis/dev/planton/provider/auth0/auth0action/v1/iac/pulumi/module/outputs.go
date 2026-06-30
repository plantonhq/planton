package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func exportOutputs(ctx *pulumi.Context, action *auth0.Action, locals *Locals) error {
	ctx.Export("id", action.ID())
	ctx.Export("name", action.Name)
	ctx.Export("version_id", action.VersionId)
	ctx.Export("runtime", action.Runtime)

	ctx.Export("metadata_name", pulumi.String(locals.ActionName))

	return nil
}
