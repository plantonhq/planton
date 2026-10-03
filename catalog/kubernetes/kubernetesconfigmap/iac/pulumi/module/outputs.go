package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output keys for outputs. They match the field names in
// KubernetesConfigMapOutputs so downstream resources can compose on them.
const (
	OutputConfigMapName = "configmap_name"
	OutputNamespace     = "namespace"
)

// exportOutputs exports all outputs
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputConfigMapName, pulumi.String(locals.ConfigMapName))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))

	return nil
}
