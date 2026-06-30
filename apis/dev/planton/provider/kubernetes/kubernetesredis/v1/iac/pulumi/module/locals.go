package module

import (
	"fmt"
	"strconv"

	kubernetesredisv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesredis/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressExternalHostname string
	KubePortForwardCommand  string
	KubeServiceFqdn         string
	KubeServiceName         string
	Namespace               string
	KubernetesRedis         *kubernetesredisv1.KubernetesRedis
	RedisPodSelectorLabels  map[string]string
	Labels                  map[string]string
	PasswordSecretName      string
	ExternalLbServiceName   string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesredisv1.KubernetesRedisStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesRedis = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesRedis.String(),
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

	locals.RedisPodSelectorLabels = map[string]string{
		"app.kubernetes.io/component": "master",
		"app.kubernetes.io/instance":  target.Metadata.Name,
		"app.kubernetes.io/name":      "redis",
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Users can prefix metadata.name with component type if needed (e.g., "redis-my-cache")
	locals.PasswordSecretName = fmt.Sprintf("%s-password", target.Metadata.Name)
	locals.ExternalLbServiceName = fmt.Sprintf("%s-external-lb", target.Metadata.Name)

	locals.KubeServiceName = fmt.Sprintf("%s-master", target.Metadata.Name)

	//export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, locals.KubeServiceName)

	//export kube-port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	//export ingress hostname
	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

	return locals
}
