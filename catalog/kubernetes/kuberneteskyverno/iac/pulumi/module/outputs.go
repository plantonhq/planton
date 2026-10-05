package module

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Output name constants — one per KubernetesKyvernoOutputs field.
const (
	OpNamespace            = "namespace"
	OpReleaseName          = "release_name"
	OpAdmissionServiceName = "admission_service_name"
	OpConfigMapName        = "config_map_name"
)

// exportOutputs publishes the composition handles. Both name outputs
// derive from the fullnameOverride pin: the admission webhook Service is
// "<name>-svc" and the runtime ConfigMap is the fullname itself
// (chart-truth: the admission-controller serviceName and config
// configMapName helpers).
func exportOutputs(ctx *pulumi.Context, locals *Locals) {
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpAdmissionServiceName, pulumi.String(locals.AdmissionServiceName))
	ctx.Export(OpConfigMapName, pulumi.String(locals.ConfigMapName))
}
