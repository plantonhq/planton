package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesFlinkDeploymentOutputs field.
const (
	OpNamespace          = "namespace"
	OpRestService        = "rest_service"
	OpRestEndpoint       = "rest_endpoint"
	OpPortForwardCommand = "port_forward_command"
)

// exportOutputs publishes the composition handles, all derived blind
// from the operator's naming contract (locals resolves them once for
// both engines): the JobManager REST Service is `<name>-rest`, serving
// the Flink REST API and web UI on 8081.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpRestService, pulumi.String(locals.RestService))
	ctx.Export(OpRestEndpoint, pulumi.String(locals.RestEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
	return nil
}
