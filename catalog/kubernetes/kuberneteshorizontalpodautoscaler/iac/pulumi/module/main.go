package module

import (
	"github.com/pkg/errors"
	kuberneteshorizontalpodautoscalerv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteshorizontalpodautoscaler/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes HorizontalPodAutoscaler with
// its scale target, metrics, and scaling behavior.
func Resources(ctx *pulumi.Context, iacInput *kuberneteshorizontalpodautoscalerv1alpha1.KubernetesHorizontalPodAutoscalerIacInput) error {
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

	// Create the horizontal pod autoscaler
	if _, err := createHorizontalPodAutoscaler(ctx, locals, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create horizontal pod autoscaler")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
