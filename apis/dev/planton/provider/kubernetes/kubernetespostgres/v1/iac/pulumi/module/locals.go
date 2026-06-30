package module

import (
	"fmt"
	"strconv"

	kubernetespostgresv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetespostgres/v1"
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
	KubernetesPostgres      *kubernetespostgresv1.KubernetesPostgres
	Labels                  map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	ExternalLbServiceName string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetespostgresv1.KubernetesPostgresStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesPostgres = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesPostgres.String(),
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

	ctx.Export(OpUsernameSecretName,
		pulumi.Sprintf("postgres.db-%s.credentials.postgresql.acid.zalan.do",
			target.Metadata.Name))
	ctx.Export(OpUsernameSecretKey, pulumi.String("username"))

	ctx.Export(OpPasswordSecretName,
		pulumi.Sprintf("postgres.db-%s.credentials.postgresql.acid.zalan.do",
			target.Metadata.Name))
	ctx.Export(OpPasswordSecretKey, pulumi.String("password"))

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

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.ExternalLbServiceName = fmt.Sprintf("%s-external-lb", target.Metadata.Name)

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	//export external hostname
	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

	return locals
}
