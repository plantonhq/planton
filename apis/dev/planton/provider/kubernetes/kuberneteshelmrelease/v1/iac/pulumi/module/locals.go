package module

import (
	"strconv"

	kuberneteshelmreleasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteshelmrelease/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	Labels                map[string]string
	Namespace             string
	KubernetesHelmRelease *kuberneteshelmreleasev1.KubernetesHelmRelease
}

func initializeLocals(ctx *pulumi.Context, stackInput *kuberneteshelmreleasev1.KubernetesHelmReleaseStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesHelmRelease = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesHelmRelease.String(),
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

	return locals
}
