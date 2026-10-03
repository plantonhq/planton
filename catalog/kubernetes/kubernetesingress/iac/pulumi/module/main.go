package module

import (
	"github.com/pkg/errors"
	kubernetesingressv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesingress/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes Ingress with its host rules,
// TLS configuration, and backends.
func Resources(ctx *pulumi.Context, iacInput *kubernetesingressv1alpha1.KubernetesIngressIacInput) error {
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

	// Create the ingress
	createdIngress, err := createIngress(ctx, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create ingress")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals, createdIngress); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
