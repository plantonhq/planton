package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the prompt whose partials are managed -- the twin of the
// Terraform module's outputs.tf.
func exportOutputs(ctx *pulumi.Context, partials *auth0.PromptScreenPartials) error {
	ctx.Export("prompt_type", partials.PromptType)

	return nil
}
