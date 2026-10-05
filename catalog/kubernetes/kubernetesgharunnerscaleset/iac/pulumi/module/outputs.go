package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesGhaRunnerScaleSetOutputs
// field.
const (
	OpNamespace          = "namespace"
	OpReleaseName        = "release_name"
	OpRunnerScaleSetName = "runner_scale_set_name"
	OpGithubConfigUrl    = "github_config_url"
)

// exportOutputs publishes the composition handles. The runner scale set
// name is the GitHub-visible fleet identity — the exact `runs-on:` value
// workflows target.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpRunnerScaleSetName, pulumi.String(locals.RunnerScaleSetName))
	ctx.Export(OpGithubConfigUrl, pulumi.String(locals.Spec.GetGithubConfigUrl()))
}
