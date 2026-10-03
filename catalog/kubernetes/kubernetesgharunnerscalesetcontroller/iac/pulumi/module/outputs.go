package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per
// KubernetesGhaRunnerScaleSetControllerOutputs field.
const (
	OpNamespace          = "namespace"
	OpReleaseName        = "release_name"
	OpServiceAccountName = "service_account_name"
)

// exportOutputs publishes the composition handles. The ServiceAccount
// name equals metadata.name (fullnameOverride pins the chart fullname,
// and the chart names the created ServiceAccount exactly the fullname) —
// the handle runner scale sets reference when this controller watches a
// single namespace.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServiceAccountName, pulumi.String(locals.ServiceAccountName))
}
