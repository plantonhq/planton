package module

import (
	"fmt"
	"strconv"

	kubernetesmongodbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesmongodb/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressExternalHostname  string
	KubePortForwardCommand   string
	KubeServiceFqdn          string
	KubeServiceName          string
	KubernetesMongodb        *kubernetesmongodbv1.KubernetesMongodb
	Namespace                string
	MongodbPodSelectorLabels map[string]string
	Labels                   map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	PasswordSecretName    string
	ExternalLbServiceName string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesmongodbv1.KubernetesMongodbStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesMongodb = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesMongodb.String(),
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

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.PasswordSecretName = fmt.Sprintf("%s-password", target.Metadata.Name)
	locals.ExternalLbServiceName = fmt.Sprintf("%s-external-lb", target.Metadata.Name)

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	ctx.Export(OpUsername, pulumi.String(vars.RootUsername))
	ctx.Export(OpPasswordSecretName, pulumi.String(locals.PasswordSecretName))
	ctx.Export(OpPasswordSecretKey, pulumi.String(vars.MongodbRootPasswordKey))

	locals.KubeServiceName = target.Metadata.Name

	// Percona operator uses these labels for pod selection
	// These labels are automatically applied by the operator to MongoDB pods
	locals.MongodbPodSelectorLabels = map[string]string{
		"app.kubernetes.io/name":       "percona-server-mongodb",
		"app.kubernetes.io/instance":   target.Metadata.Name,
		"app.kubernetes.io/managed-by": "percona-server-mongodb-operator",
	}

	//export kubernetes service name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, target.Metadata.Name)

	//export kube-port-forward command
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	locals.IngressExternalHostname = target.Spec.Ingress.Hostname

	ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

	return locals
}
