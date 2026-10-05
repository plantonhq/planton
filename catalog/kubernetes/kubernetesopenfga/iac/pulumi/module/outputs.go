package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesOpenFgaOutputs field.
const (
	OpNamespace           = "namespace"
	OpService             = "service"
	OpApiHttpEndpoint     = "api_http_endpoint"
	OpApiGrpcEndpoint     = "api_grpc_endpoint"
	OpAuthnKeysSecretName = "authn_keys_secret_name"
	OpPortForwardCommand  = "port_forward_command"
)

// exportOutputs publishes the composition handles. The service name is
// the chart's ClusterIP Service — openfga.fullname, pinned to the
// resource name via fullnameOverride; the endpoints are built from it
// (HTTP 8080, plaintext gRPC 8081 — the chart's fixed ports). The authn
// Secret handle is the module-owned `<name>-authn-keys` Secret, or ""
// when authn is unset or rides an existing Secret.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpService, pulumi.String(locals.ServiceName))
	ctx.Export(OpApiHttpEndpoint, pulumi.String(locals.ApiHttpEndpoint))
	ctx.Export(OpApiGrpcEndpoint, pulumi.String(locals.ApiGrpcEndpoint))
	ctx.Export(OpAuthnKeysSecretName, pulumi.String(locals.AuthnKeysSecretName))
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.PortForwardCommand))
}
