package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the template managed -- the twin of iac/tf/outputs.tf.
func exportOutputs(ctx *pulumi.Context, emailTemplate *auth0.EmailTemplate) error {
	ctx.Export("template", emailTemplate.Template)
	ctx.Export("enabled", emailTemplate.Enabled)

	return nil
}
