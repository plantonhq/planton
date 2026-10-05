package module

import (
	"github.com/pkg/errors"
	kubernetesstorageclassv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstorageclass/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes StorageClass with its
// provisioner, parameters, and volume lifecycle policies.
func Resources(ctx *pulumi.Context, iacInput *kubernetesstorageclassv1alpha1.KubernetesStorageClassIacInput) error {
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

	// Create the storage class
	if _, err := createStorageClass(ctx, locals, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create storage class")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
