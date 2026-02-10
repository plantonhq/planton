package module

import (
	"github.com/pkg/errors"
	kubernetesservicev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesservice/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes Service with the specified configuration.
func Resources(ctx *pulumi.Context, stackInput *kubernetesservicev1.KubernetesServiceStackInput) error {
	// Initialize locals with derived values from the stack input.
	locals := initializeLocals(ctx, stackInput)

	// Create the Kubernetes provider from credentials.
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		stackInput.ProviderConfig,
		"kubernetes",
	)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// Create the Kubernetes Service resource.
	createdService, err := createService(ctx, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes service")
	}

	// Export outputs.
	if err := exportOutputs(ctx, locals, createdService); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
