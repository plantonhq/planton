package module

import (
	"github.com/pkg/errors"
	kubernetessecretv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes Secret with the appropriate type, data, and metadata.
func Resources(ctx *pulumi.Context, iacInput *kubernetessecretv1alpha1.KubernetesSecretIacInput) error {
	// Initialize locals with derived values
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Create Kubernetes provider from credentials
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		iacInput.ProviderConfig,
		"kubernetes",
	)
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// Create the secret
	if _, err := createSecret(ctx, locals, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
