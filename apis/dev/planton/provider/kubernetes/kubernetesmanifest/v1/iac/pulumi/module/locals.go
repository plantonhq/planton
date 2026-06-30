package module

import (
	"strconv"

	kubernetesmanifestv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesmanifest/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	Namespace          string
	KubernetesManifest *kubernetesmanifestv1.KubernetesManifest
	Labels             map[string]string
	ManifestYAML       string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesmanifestv1.KubernetesManifestStackInput) (*Locals, error) {
	locals := &Locals{}

	locals.KubernetesManifest = stackInput.Target

	target := stackInput.Target

	// Labels for namespace and resources
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesManifest.String(),
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

	// Get namespace from spec, it is a required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// Export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Store manifest YAML
	locals.ManifestYAML = target.Spec.ManifestYaml

	return locals, nil
}
