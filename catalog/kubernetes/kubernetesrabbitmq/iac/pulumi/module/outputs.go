package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesRabbitMqOutputs field.
const (
	OpNamespace             = "namespace"
	OpClusterName           = "cluster_name"
	OpServiceName           = "service_name"
	OpHeadlessServiceName   = "headless_service_name"
	OpAmqpEndpoint          = "amqp_endpoint"
	OpManagementEndpoint    = "management_endpoint"
	OpDefaultUserSecretName = "default_user_secret_name"
	OpPortForwardCommand    = "port_forward_command"
)

// exportOutputs publishes the composition handles. The default-user Secret
// handle is empty when the Vault secret backend owns the credentials; the
// endpoints carry the effective ports (TLS ports when the plain listeners
// are closed).
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpClusterName, pulumi.String(locals.ClusterName))
	ctx.Export(OpServiceName, pulumi.String(locals.ServiceName))
	ctx.Export(OpHeadlessServiceName, pulumi.String(locals.HeadlessServiceName))
	ctx.Export(OpAmqpEndpoint, pulumi.String(locals.AmqpEndpoint))
	ctx.Export(OpManagementEndpoint, pulumi.String(locals.ManagementEndpoint))
	ctx.Export(OpDefaultUserSecretName, pulumi.String(locals.DefaultUserSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}
