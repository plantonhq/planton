package module

import (
	"github.com/pkg/errors"
	kubernetesnetworkpolicyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnetworkpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes NetworkPolicy with its pod
// selection and allow rules.
func Resources(ctx *pulumi.Context, iacInput *kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyIacInput) error {
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

	// Create the network policy
	if _, err := createNetworkPolicy(ctx, locals, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create network policy")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
