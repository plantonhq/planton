package module

import (
	pulumiyaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesRabbitMqOperatorOutputs
// field.
const (
	OpNamespace       = "namespace"
	OpDeploymentName  = "deployment_name"
	OpMetricsEndpoint = "metrics_endpoint"
	OpCrdName         = "crd_name"
)

// exportOutputs publishes the release manifest's fixed handles (the
// manifest resource is passed so the export carries its dependency).
func exportOutputs(ctx *pulumi.Context, locals *Locals, _ *pulumiyaml.ConfigFile) error {
	ctx.Export(OpNamespace, pulumi.String(vars.Namespace))
	ctx.Export(OpDeploymentName, pulumi.String(vars.DeploymentName))
	ctx.Export(OpMetricsEndpoint, pulumi.String(locals.MetricsEndpoint))
	ctx.Export(OpCrdName, pulumi.String(vars.CrdName))
	return nil
}
