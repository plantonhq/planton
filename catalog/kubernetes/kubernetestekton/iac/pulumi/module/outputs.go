package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesTektonOutputs field.
const (
	OpNamespace             = "namespace"
	OpProfile               = "profile"
	OpDashboardService      = "dashboard_service"
	OpDashboardKubeEndpoint = "dashboard_kube_endpoint"
	OpPortForwardCommand    = "port_forward_command"
)

// exportOutputs publishes the resolved installation handles. The
// dashboard handles are empty strings unless the profile installs the
// dashboard (`all`) — the locals resolve that once for both engines.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OpNamespace, pulumi.String(locals.TargetNamespace))
	ctx.Export(OpProfile, pulumi.String(locals.Profile))
	ctx.Export(OpDashboardService, pulumi.String(locals.DashboardService))
	ctx.Export(OpDashboardKubeEndpoint, pulumi.String(locals.DashboardKubeEndpoint))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
	return nil
}
