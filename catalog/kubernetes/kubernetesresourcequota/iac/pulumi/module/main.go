package module

import (
	"github.com/pkg/errors"
	kubernetesresourcequotav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesresourcequota/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of the namespace-governance pair: the
// ResourceQuota, plus a companion LimitRange when limit defaults are set.
func Resources(ctx *pulumi.Context, iacInput *kubernetesresourcequotav1alpha1.KubernetesResourceQuotaIacInput) error {
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

	// Create the resource quota
	if _, err := createResourceQuota(ctx, locals, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create resource quota")
	}

	// Create the companion limit range when limit defaults are configured
	if locals.LimitRangeName != "" {
		if _, err := createLimitRange(ctx, locals, kubernetesProvider); err != nil {
			return errors.Wrap(err, "failed to create limit range")
		}
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
