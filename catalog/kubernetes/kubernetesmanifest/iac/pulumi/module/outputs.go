package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output constants define the keys for outputs exported by this
// module, mirroring KubernetesManifestOutputs. The Terraform module
// exports the identical set.
const (
	// OpNamespace is the anchor namespace: where namespaced documents
	// without an explicit metadata.namespace were applied.
	OpNamespace = "namespace"
	// OpAppliedResources is the applied-resource inventory, one
	// "apiVersion/Kind/name" entry per document in manifest order.
	OpAppliedResources = "applied_resources"
)

// exportOutputs exports the manifest's observable handles. The inventory is
// derived from the input YAML (see locals), never from engine-side child
// tracking, so both engines export identical values.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpAppliedResources, pulumi.ToStringArray(locals.AppliedResources))
}
