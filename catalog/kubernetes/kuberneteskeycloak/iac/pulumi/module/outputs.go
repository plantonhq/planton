package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesKeycloakOutputs field.
const (
	OpNamespace              = "namespace"
	OpStatefulSet            = "stateful_set"
	OpService                = "service"
	OpDiscoveryService       = "discovery_service"
	OpApiEndpoint            = "api_endpoint"
	OpManagementEndpoint     = "management_endpoint"
	OpInitialAdminSecretName = "initial_admin_secret_name"
	OpPortForwardCommand     = "port_forward_command"
)

// exportOutputs publishes the composition handles, all derived blind from
// the operator's naming contract (locals resolves them once for both
// engines). The initial-admin Secret is break-glass material: seeded at
// FIRST start only, never rotated by the operator.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpStatefulSet, pulumi.String(locals.StatefulSet))
	ctx.Export(OpService, pulumi.String(locals.ServiceName))
	ctx.Export(OpDiscoveryService, pulumi.String(locals.DiscoveryService))
	ctx.Export(OpApiEndpoint, pulumi.String(locals.ApiEndpoint))
	ctx.Export(OpManagementEndpoint, pulumi.String(locals.ManagementEndpoint))
	ctx.Export(OpInitialAdminSecretName, pulumi.String(locals.InitialAdminSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
	return nil
}
