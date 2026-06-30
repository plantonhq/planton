package module

import (
	"github.com/pkg/errors"
	kubernetesjobv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesjob/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for creating all necessary Kubernetes resources
// for a Job, based on the KubernetesJob API resource definition.
func Resources(ctx *pulumi.Context, stackInput *kubernetesjobv1.KubernetesJobStackInput) error {
	// Initialize local references
	locals, err := initializeLocals(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Create a Pulumi Kubernetes provider from the given credentials
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		stackInput.ProviderConfig,
		"kubernetes",
	)
	if err != nil {
		return errors.Wrap(err, "failed to create Kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	// When create_namespace is false, createdNamespace is nil and namespaceDeps is empty.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Create the main secret resource
	if err := secret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	// Create an image pull secret if Docker credentials are provided
	if locals.ImagePullSecretData != nil {
		if err := createImagePullSecret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create image pull secret")
		}
	}

	// Create ConfigMaps
	_, err = configMaps(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create configmaps")
	}

	// Create the Job resource
	_, err = job(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create job")
	}

	// Export outputs
	exportOutputs(ctx, locals)

	return nil
}
