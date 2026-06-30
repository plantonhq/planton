package module

import (
	"fmt"
	"strconv"

	kubernetesgrafanav1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgrafana/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressExternalHostname  string
	IngressInternalHostname  string
	KubePortForwardCommand   string
	KubeServiceFqdn          string
	KubeServiceName          string
	Namespace                string
	KubernetesGrafana        *kubernetesgrafanav1.KubernetesGrafana
	GrafanaPodSelectorLabels map[string]string
	Labels                   map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	ExternalIngressName string
	InternalIngressName string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesgrafanav1.KubernetesGrafanaStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesGrafana = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesGrafana.String(),
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

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	locals.GrafanaPodSelectorLabels = map[string]string{
		"app.kubernetes.io/name":     "grafana",
		"app.kubernetes.io/instance": target.Metadata.Name,
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.ExternalIngressName = fmt.Sprintf("%s-external", target.Metadata.Name)
	locals.InternalIngressName = fmt.Sprintf("%s-internal", target.Metadata.Name)

	locals.KubeServiceName = fmt.Sprintf("%s-grafana", target.Metadata.Name)

	//export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:80",
		locals.Namespace, locals.KubeServiceName)

	//export kube-port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	// Use the hostname directly from spec
	locals.IngressExternalHostname = fmt.Sprintf("https://%s", target.Spec.Ingress.Hostname)
	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

	// Internal hostname (private ingress) - prepend internal-
	locals.IngressInternalHostname = fmt.Sprintf("https://internal-%s", target.Spec.Ingress.Hostname)
	ctx.Export(OpInternalHostname, pulumi.String(locals.IngressInternalHostname))

	return locals
}
