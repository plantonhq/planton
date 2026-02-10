package module

import (
	"github.com/pkg/errors"
	kubernetessecretv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetessecret/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes Secret with the appropriate type, data, and metadata.
func Resources(ctx *pulumi.Context, stackInput *kubernetessecretv1.KubernetesSecretStackInput) error {
	// Initialize locals with derived values
	locals, err := initializeLocals(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Create Kubernetes provider from credentials
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx,
		stackInput.ProviderConfig,
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
