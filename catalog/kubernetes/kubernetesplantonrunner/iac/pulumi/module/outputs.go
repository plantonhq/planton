package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesPlantonRunnerOutputs
// field.
const (
	OpNamespace       = "namespace"
	OpReleaseName     = "release_name"
	OpTokenSecretName = "token_secret_name"
	OpRunnerName      = "runner_name"
)

// exportOutputs publishes the component's outputs — every value is
// deterministic from the resolved inputs, so they export as plain strings.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpTokenSecretName, pulumi.String(locals.TokenSecretName))
	ctx.Export(OpRunnerName, pulumi.String(locals.RunnerName))
}
