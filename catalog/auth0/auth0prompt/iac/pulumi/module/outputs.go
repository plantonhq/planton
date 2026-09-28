package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the login-flow settings as the tenant carries them after
// the apply, managed or not -- the twin of the Terraform module's outputs.tf.
func exportOutputs(ctx *pulumi.Context, prompt *auth0.Prompt) error {
	ctx.Export("universal_login_experience", prompt.UniversalLoginExperience)
	ctx.Export("identifier_first", prompt.IdentifierFirst)
	ctx.Export("webauthn_platform_first_factor", prompt.WebauthnPlatformFirstFactor)

	return nil
}
