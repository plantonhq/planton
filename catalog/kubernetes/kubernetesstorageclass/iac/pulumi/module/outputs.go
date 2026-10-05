package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesStorageClassOutputs field names.
const (
	OutputStorageClassName = "storage_class_name"
	OutputProvisioner      = "provisioner"
	OutputIsDefaultClass   = "is_default_class"
)

// exportOutputs exports the outputs from the created StorageClass.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputStorageClassName, pulumi.String(locals.Name))
	ctx.Export(OutputProvisioner, pulumi.String(locals.Spec.GetProvisioner()))
	ctx.Export(OutputIsDefaultClass, pulumi.Bool(locals.Spec.GetIsDefaultClass()))

	return nil
}
