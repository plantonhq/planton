package module

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// Output name constants - one per KubernetesFlagdOutputs field.
const (
	OpNamespace          = "namespace"
	OpService            = "service"
	OpEvaluationEndpoint = "evaluation_endpoint"
	OpSyncEndpoint       = "sync_endpoint"
	OpOfrepEndpoint      = "ofrep_endpoint"
	OpManagementEndpoint = "management_endpoint"
	OpPortForwardCommand = "port_forward_command"
)

func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.Name))
	ctx.Export(OpEvaluationEndpoint, pulumi.String(locals.EvaluationEndpoint))
	ctx.Export(OpSyncEndpoint, pulumi.String(locals.SyncEndpoint))
	ctx.Export(OpOfrepEndpoint, pulumi.String(locals.OfrepEndpoint))
	ctx.Export(OpManagementEndpoint, pulumi.String(locals.ManagementEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}
