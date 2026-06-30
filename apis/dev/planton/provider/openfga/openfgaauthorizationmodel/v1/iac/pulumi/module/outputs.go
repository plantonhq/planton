package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports stack outputs.
// This is a placeholder for consistency - OpenFGA has no Pulumi provider.
// No real outputs are exported since no resources are created.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	// Export placeholder values to indicate pass-through behavior
	ctx.Export("id", pulumi.String(""))

	return nil
}
