package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output key constants aligned with KubernetesPodDisruptionBudgetOutputs field names.
const (
	OutputPodDisruptionBudgetName = "pod_disruption_budget_name"
	OutputNamespace               = "namespace"
)

// exportOutputs exports the outputs from the created PodDisruptionBudget.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OutputPodDisruptionBudgetName, pulumi.String(locals.Name))
	ctx.Export(OutputNamespace, pulumi.String(locals.Namespace))

	return nil
}
