package module

import (
	"strconv"

	kubernetesingressnginxv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesingressnginx/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesIngressNginx *kubernetesingressnginxv1.KubernetesIngressNginx
	Namespace              string
	ReleaseName            string
	ServiceName            string
	ServiceType            string
	ChartVersion           string
	Labels                 map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesingressnginxv1.KubernetesIngressNginxStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesIngressNginx = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesIngressNginx.String(),
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

	// Extract namespace from spec or use default
	locals.Namespace = target.Spec.Namespace.GetValue()
	if locals.Namespace == "" {
		locals.Namespace = vars.Namespace
	}

	// Determine chart version
	locals.ChartVersion = target.Spec.ChartVersion
	if locals.ChartVersion == "" {
		locals.ChartVersion = vars.DefaultChartVersion
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "nginx-public")
	locals.ReleaseName = target.Metadata.Name
	locals.ServiceName = target.Metadata.Name + "-controller"
	locals.ServiceType = "LoadBalancer"

	// Export stack outputs
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(locals.ReleaseName))
	ctx.Export(OpServiceName, pulumi.String(locals.ServiceName))
	ctx.Export(OpServiceType, pulumi.String(locals.ServiceType))

	return locals
}
