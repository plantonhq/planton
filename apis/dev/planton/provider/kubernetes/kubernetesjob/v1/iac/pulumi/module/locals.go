package module

import (
	"fmt"
	"strconv"

	kubernetesjobv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesjob/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds local references and configurations used by the module.
type Locals struct {
	KubernetesJob       *kubernetesjobv1.KubernetesJob
	Namespace           string
	Labels              map[string]string
	ImagePullSecretData map[string]string

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	EnvSecretsSecretName string
	ImagePullSecretName  string
}

// initializeLocals configures the Locals struct from the given stack input.
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesjobv1.KubernetesJobStackInput) (*Locals, error) {
	locals := &Locals{}
	target := stackInput.Target

	locals.KubernetesJob = target
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesJob.String(),
	}

	// add resource id if present
	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	// add organization if present
	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	// add environment if present
	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// handle docker config json if specified
	if stackInput.DockerConfigJson != "" {
		locals.ImagePullSecretData = map[string]string{
			".dockerconfigjson": stackInput.DockerConfigJson,
		}
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	locals.EnvSecretsSecretName = fmt.Sprintf("%s-env-secrets", target.Metadata.Name)
	locals.ImagePullSecretName = fmt.Sprintf("%s-image-pull-secret", target.Metadata.Name)

	return locals, nil
}
