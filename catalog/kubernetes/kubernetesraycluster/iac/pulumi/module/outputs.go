package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesRayClusterOutputs field
// (auth_token_secret flattens to its name/key halves, the catalog's
// convention for message-typed outputs).
const (
	OpNamespace           = "namespace"
	OpHeadService         = "head_service"
	OpClientEndpoint      = "client_endpoint"
	OpDashboardEndpoint   = "dashboard_endpoint"
	OpGcsEndpoint         = "gcs_endpoint"
	OpAuthTokenSecretName = "auth_token_secret.name"
	OpAuthTokenSecretKey  = "auth_token_secret.key"
	OpPortForwardCommand  = "port_forward_command"
)

// exportOutputs publishes the composition handles, all derived blind from
// the operator's naming contract (locals resolves them once for both
// engines). The auth-token Secret exports ONLY in token mode — no Secret
// exists when auth is disabled.
func exportOutputs(ctx *pulumi.Context, locals *Locals) error {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpHeadService, pulumi.String(locals.HeadService))
	ctx.Export(OpClientEndpoint, pulumi.String(locals.ClientEndpoint))
	ctx.Export(OpDashboardEndpoint, pulumi.String(locals.DashboardEndpoint))
	ctx.Export(OpGcsEndpoint, pulumi.String(locals.GcsEndpoint))
	if locals.TokenAuthEnabled {
		ctx.Export(OpAuthTokenSecretName, pulumi.String(locals.AuthTokenSecretName))
		ctx.Export(OpAuthTokenSecretKey, pulumi.String(vars.AuthTokenSecretKey))
	}
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
	return nil
}
