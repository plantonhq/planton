package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the stack outputs for the Auth0 User, mapping onto
// Auth0UserStackOutputs field by field.
func exportOutputs(ctx *pulumi.Context, user *auth0.User, mintedPassword pulumi.StringOutput, minted bool) error {
	// The resource ID is the full subject, connection prefix included
	// ("auth0|..."), for every user -- whether Auth0 assigned the id or the
	// spec declared its unprefixed half. The user_id attribute reads back the
	// same value, but the ID is the one the provider guarantees.
	ctx.Export("user_id", user.ID())
	ctx.Export("email", user.Email)
	ctx.Export("username", user.Username)
	ctx.Export("name", user.Name)
	ctx.Export("nickname", user.Nickname)
	ctx.Export("picture", user.Picture)
	ctx.Export("connection_name", user.ConnectionName)

	// The minted password is reported exactly when the module generated it;
	// a declared password is never echoed back and a passwordless user has
	// none, so the output is absent in both of those cases. Marked secret so
	// the value is encrypted in the Pulumi state -- twin of the Terraform
	// output's sensitive = true.
	if minted {
		ctx.Export("password", pulumi.ToSecret(mintedPassword))
	}

	return nil
}
