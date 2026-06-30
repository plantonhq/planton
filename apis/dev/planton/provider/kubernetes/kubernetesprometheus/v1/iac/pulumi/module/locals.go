package module

import (
	"strconv"

	kubernetesprometheusv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesprometheus/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps all frequently-used, derived values in one place –
// similar to a Terraform "locals {}" block.
type Locals struct {
	Namespace            string
	Labels               map[string]string
	KubernetesPrometheus *kubernetesprometheusv1.KubernetesPrometheus
	ServiceName          string
}

// initializeLocals builds the Locals struct and immediately exports the
// values required by KubernetesPrometheusStackOutputs.
func initializeLocals(ctx *pulumi.Context,
	stackInput *kubernetesprometheusv1.KubernetesPrometheusStackInput) *Locals {

	locals := &Locals{}
	locals.KubernetesPrometheus = stackInput.Target
	target := stackInput.Target

	// ------------------------------- labels ----------------------------------
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesPrometheus.String(),
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
	ctx.Export(Namespace, pulumi.String(locals.Namespace))

	// service name
	locals.ServiceName = target.Metadata.Name + "-prometheus"
	ctx.Export(Service, pulumi.String(locals.ServiceName))

	// Export port forward command
	portForwardCmd := "kubectl port-forward -n " + locals.Namespace + " service/" + locals.ServiceName + " 9090:9090"
	ctx.Export(PortForwardCommand, pulumi.String(portForwardCmd))

	// Export Kubernetes endpoint (FQDN)
	kubeEndpoint := locals.ServiceName + "." + locals.Namespace + ".svc.cluster.local"
	ctx.Export(KubeEndpoint, pulumi.String(kubeEndpoint))

	// Export external hostname (if ingress is enabled)
	ctx.Export(ExternalHostname, pulumi.String(""))

	// Export internal hostname (if ingress is enabled)
	ctx.Export(InternalHostname, pulumi.String(""))

	return locals
}
