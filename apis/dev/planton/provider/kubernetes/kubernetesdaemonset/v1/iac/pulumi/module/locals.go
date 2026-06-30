package module

import (
	"fmt"
	"strconv"

	kubernetesdaemonsetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds local references and configurations used by the module.
type Locals struct {
	KubernetesDaemonSet *kubernetesdaemonsetv1.KubernetesDaemonSet
	Namespace           string
	Labels              map[string]string
	SelectorLabels      map[string]string
	ImagePullSecretData map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	EnvSecretName       string
	ImagePullSecretName string
}

// initializeLocals configures the Locals struct from the given stack input.
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesdaemonsetv1.KubernetesDaemonSetStackInput) (*Locals, error) {
	locals := &Locals{}
	target := stackInput.Target

	locals.KubernetesDaemonSet = target

	// Static selector labels that never change
	locals.SelectorLabels = map[string]string{
		"app": "daemonset",
	}

	// Full labels include both selector labels and metadata labels
	locals.Labels = map[string]string{
		"app":                            "daemonset",
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesDaemonSet.String(),
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

	// Get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// Export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Export daemonset name
	ctx.Export(OpDaemonSetName, pulumi.String(target.Metadata.Name))

	// Handle docker config json if specified
	if stackInput.DockerConfigJson != "" {
		locals.ImagePullSecretData = map[string]string{
			".dockerconfigjson": stackInput.DockerConfigJson,
		}
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	locals.EnvSecretName = fmt.Sprintf("%s-env-secrets", target.Metadata.Name)
	locals.ImagePullSecretName = fmt.Sprintf("%s-image-pull-secret", target.Metadata.Name)

	return locals, nil
}
