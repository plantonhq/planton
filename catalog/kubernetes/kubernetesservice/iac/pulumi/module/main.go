package module

import (
	"github.com/pkg/errors"
	kubernetesservicev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes Service with the specified configuration.
func Resources(ctx *pulumi.Context, iacInput *kubernetesservicev1alpha1.KubernetesServiceIacInput) error {
	// Initialize locals with derived values from the IaC input.
	locals := initializeLocals(ctx, iacInput)

	// Create the Kubernetes provider from credentials.
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		iacInput.ProviderConfig,
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
