package module

import (
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the stack outputs for the Istio CRDs installation.
func exportOutputs(ctx *pulumi.Context, locals *Locals, _ *pulumiyaml.ConfigFile) error {
	// Istio release the CRDs were installed from.
	ctx.Export("installed_release", pulumi.String(locals.Release))

	// The exact CRD bundle URL that was applied.
	ctx.Export("installed_manifest_url", pulumi.String(locals.ManifestURL))

	return nil
}
