package module

import (
	"github.com/pkg/errors"
	kubernetesnamespacev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnamespace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a complete Kubernetes namespace with quotas, policies, and configurations.
func Resources(ctx *pulumi.Context, iacInput *kubernetesnamespacev1alpha1.KubernetesNamespaceIacInput) error {
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

	// Create the namespace with appropriate labels and annotations
	createdNamespace, err := createNamespace(ctx, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Create ResourceQuota if resource profile is configured
	if err := createResourceQuota(ctx, locals, createdNamespace, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create resource quota")
	}

	// Create LimitRange if default limits are configured
	if err := createLimitRange(ctx, locals, createdNamespace, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create limit range")
	}

	// Create NetworkPolicies if network config is specified
	if err := createNetworkPolicies(ctx, locals, createdNamespace, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create network policies")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
