package module

import (
	"github.com/pkg/errors"
	kubernetespriorityclassv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespriorityclass/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes PriorityClass with its
// priority value, default flag, and preemption policy.
func Resources(ctx *pulumi.Context, iacInput *kubernetespriorityclassv1alpha1.KubernetesPriorityClassIacInput) error {
	// Initialize locals with derived values
	locals := initializeLocals(ctx, iacInput)

	// Create Kubernetes provider from credentials
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		iacInput.ProviderConfig,
		"kubernetes",
	)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// Create the priority class
	if _, err := createPriorityClass(ctx, locals, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create priority class")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
