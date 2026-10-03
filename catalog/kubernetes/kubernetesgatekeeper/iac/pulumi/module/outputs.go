package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesGatekeeperOutputs field.
const (
	OpNamespace             = "namespace"
	OpReleaseName           = "release_name"
	OpWebhookServiceName    = "webhook_service_name"
	OpWebhookCertSecretName = "webhook_cert_secret_name"
)

// exportOutputs publishes the composition handles. Both name outputs are
// CHART-FIXED (hardcoded in the templates, no fullname derivation) —
// which is also why the engine is a per-cluster singleton.
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpWebhookServiceName, pulumi.String(vars.WebhookServiceName))
	ctx.Export(OpWebhookCertSecretName, pulumi.String(vars.WebhookCertSecretName))
}
