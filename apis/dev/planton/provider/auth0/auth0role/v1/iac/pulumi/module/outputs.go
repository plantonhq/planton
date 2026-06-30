package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports stack outputs for the Auth0 Role.
func exportOutputs(ctx *pulumi.Context, role *auth0.Role, locals *Locals) error {
	ctx.Export("id", role.ID())
	ctx.Export("name", role.Name)
	ctx.Export("description", role.Description)

	return nil
}
