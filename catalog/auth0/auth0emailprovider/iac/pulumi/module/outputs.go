package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the email provider as applied -- the twin of
// iac/tf/outputs.tf.
func exportOutputs(ctx *pulumi.Context, emailProvider *auth0.EmailProvider) error {
	ctx.Export("name", emailProvider.Name)
	ctx.Export("default_from_address", emailProvider.DefaultFromAddress)

	return nil
}
