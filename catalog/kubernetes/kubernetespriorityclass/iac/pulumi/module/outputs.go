package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesPriorityClassOutputs field names.
const (
	OutputPriorityClassName = "priority_class_name"
	OutputValue             = "value"
)

// exportOutputs exports the outputs from the created PriorityClass.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputPriorityClassName, pulumi.String(locals.Name))
	ctx.Export(OutputValue, pulumi.Int(int(locals.Spec.GetValue())))

	return nil
}
