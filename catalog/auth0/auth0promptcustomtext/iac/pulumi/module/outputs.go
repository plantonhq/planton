package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the prompt and language of the custom text managed, and
// its identifier (<prompt>::<language>, also its import id) -- the twin of the
// Terraform module's outputs.tf.
func exportOutputs(ctx *pulumi.Context, customText *auth0.PromptCustomText) error {
	ctx.Export("prompt", customText.Prompt)
	ctx.Export("language", customText.Language)
	ctx.Export("id", customText.ID())

	return nil
}
