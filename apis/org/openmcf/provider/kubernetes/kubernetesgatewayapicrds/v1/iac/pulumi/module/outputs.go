package module

import (
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the stack outputs for the Gateway API CRDs installation.
func exportOutputs(ctx *pulumi.Context, locals *Locals, crds *pulumiyaml.ConfigFile) error {
	// Export installed version
	ctx.Export("installed_version", pulumi.String(locals.Version))

	// Export installed channel
	ctx.Export("installed_channel", pulumi.String(locals.ChannelName))

	// Export the exact CRD bundle URL that was applied (encodes version + channel).
	ctx.Export("installed_manifest_url", pulumi.String(locals.ManifestURL))

	return nil
}
