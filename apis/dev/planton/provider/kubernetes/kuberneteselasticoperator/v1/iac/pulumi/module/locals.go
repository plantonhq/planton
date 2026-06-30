package module

import (
	"strconv"

	kuberneteselasticoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteselasticoperator/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesElasticOperator *kuberneteselasticoperatorv1.KubernetesElasticOperator
	KubeLabels                map[string]string
	Namespace                 string
	// Computed resource names to avoid conflicts when multiple instances share a namespace
	HelmReleaseName string
}

func initializeLocals(ctx *pulumi.Context, in *kuberneteselasticoperatorv1.KubernetesElasticOperatorStackInput) *Locals {
	var l Locals
	l.KubernetesElasticOperator = in.Target

	l.KubeLabels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: in.Target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesElasticOperator.String(),
	}

	if id := in.Target.Metadata.Id; id != "" {
		l.KubeLabels[kuberneteslabelkeys.ResourceId] = id
	}
	if org := in.Target.Metadata.Org; org != "" {
		l.KubeLabels[kuberneteslabelkeys.Organization] = org
	}
	if env := in.Target.Metadata.Env; env != "" {
		l.KubeLabels[kuberneteslabelkeys.Environment] = env
	}

	// Get namespace from spec, fallback to default
	l.Namespace = in.Target.Spec.Namespace.GetValue()
	if l.Namespace == "" {
		l.Namespace = vars.Namespace
	}

	// Export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(l.Namespace))

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// The Helm release name uses metadata.name to ensure uniqueness
	l.HelmReleaseName = in.Target.Metadata.Name

	return &l
}
