package module

import (
	"fmt"
	"strconv"

	kubernetesnatsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesnats/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps all frequently-used, derived values in one place –
// similar to a Terraform “locals {}” block.
type Locals struct {
	Namespace         string
	Labels            map[string]string
	KubernetesNats    *kubernetesnatsv1.KubernetesNats
	ClientURLInternal string
	ClientURLExternal string
	TlsSecretName     string
	TlsSecretKey      string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	AuthSecretName        string
	NoAuthUserSecretName  string
	ExternalLbServiceName string

	// Helm chart versions (from spec or defaults)
	NatsHelmChartVersion string
	NackHelmChartVersion string
	NackAppVersion       string
	NackCrdsUrl          string
}

// initializeLocals builds the Locals struct and immediately exports the
// values required by KubernetesNatsStackOutputs.
func initializeLocals(ctx *pulumi.Context,
	stackInput *kubernetesnatsv1.KubernetesNatsStackInput) *Locals {

	locals := &Locals{}
	locals.KubernetesNats = stackInput.Target
	target := stackInput.Target

	// ------------------------------- labels ----------------------------------
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesNats.String(),
	}
	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "nats-my-bus")
	locals.AuthSecretName = fmt.Sprintf("%s-auth", target.Metadata.Name)
	locals.NoAuthUserSecretName = fmt.Sprintf("%s-no-auth-user", target.Metadata.Name)
	locals.ExternalLbServiceName = fmt.Sprintf("%s-external-lb", target.Metadata.Name)

	// ------------------------- internal client URL ---------------------------
	// NATS Helm chart (v2.x) creates a Service named "{release-name}". Port 4222 is fixed.
	serviceName := target.Metadata.Name
	locals.ClientURLInternal = fmt.Sprintf("nats://%s.%s.svc.cluster.local:%d",
		serviceName, locals.Namespace, vars.NatsClientPort)
	ctx.Export(OpClientUrlInternal, pulumi.String(locals.ClientURLInternal))

	// ------------------------------ ingress ----------------------------------
	if target.Spec.Ingress != nil &&
		target.Spec.Ingress.Enabled &&
		target.Spec.Ingress.Hostname != "" {

		locals.ClientURLExternal = fmt.Sprintf("nats://%s:%d",
			target.Spec.Ingress.Hostname, vars.NatsClientPort)
		ctx.Export(OpClientUrlExternal, pulumi.String(locals.ClientURLExternal))
	}

	// -------------------- auth / token secret outputs ------------------------
	// Secret names are deterministic so callers / automation can pre-bake RBAC.
	ctx.Export(OpAuthSecretName, pulumi.String(locals.AuthSecretName))
	ctx.Export(OpAuthSecretKey, pulumi.String(vars.AdminAuthSecretKey))

	// ----------------------- TLS certificate secret --------------------------
	if target.Spec.TlsEnabled {
		locals.TlsSecretName = fmt.Sprintf("%s-tls", target.Metadata.Name)
		locals.TlsSecretKey = vars.TlsCertKey
		ctx.Export(OpTlsSecretName, pulumi.String(locals.TlsSecretName))
		ctx.Export(OpTlsSecretKey, pulumi.String(locals.TlsSecretKey))
	}

	// ------------------------ jet-stream domain ------------------------------
	if !target.Spec.DisableJetStream {
		localsJetDomain := fmt.Sprintf("%s", locals.Namespace) // simple default
		ctx.Export(OpJetStreamDomain, pulumi.String(localsJetDomain))
	}

	// ------------------------ helm chart versions -----------------------------
	// Defaults are guaranteed by Planton CLI from proto definitions
	locals.NatsHelmChartVersion = *target.Spec.NatsHelmChartVersion

	// NACK chart version, app version, and CRDs URL
	if target.Spec.NackController != nil {
		locals.NackHelmChartVersion = *target.Spec.NackController.HelmChartVersion
		locals.NackAppVersion = *target.Spec.NackController.AppVersion
		// CRDs URL uses app version (GitHub tag), NOT chart version
		locals.NackCrdsUrl = fmt.Sprintf(vars.NackCrdsUrlTemplate, locals.NackAppVersion)
	}

	return locals
}
