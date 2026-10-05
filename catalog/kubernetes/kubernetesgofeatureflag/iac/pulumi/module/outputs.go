package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants - one per KubernetesGoFeatureFlagOutputs field.
const (
	OpNamespace          = "namespace"
	OpService            = "service"
	OpApiEndpoint        = "api_endpoint"
	OpMonitoringEndpoint = "monitoring_endpoint"
	OpPortForwardCommand = "port_forward_command"
)

// exportOutputs publishes the composition handles. The Service is the
// chart's (its fullname, pinned to metadata.name); the endpoints are built
// from it and the resolved ports.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.ServiceName))
	ctx.Export(OpApiEndpoint, pulumi.String(locals.ApiEndpoint))
	ctx.Export(OpMonitoringEndpoint, pulumi.String(locals.MonitoringEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}
