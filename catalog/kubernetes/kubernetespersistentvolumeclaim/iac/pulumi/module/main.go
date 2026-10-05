package module

import (
	"github.com/pkg/errors"
	kubernetespersistentvolumeclaimv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespersistentvolumeclaim/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the main entry point for the Pulumi module.
// It orchestrates the creation of a Kubernetes PersistentVolumeClaim with its
// storage request, access modes, and provisioning class.
func Resources(ctx *pulumi.Context, iacInput *kubernetespersistentvolumeclaimv1alpha1.KubernetesPersistentVolumeClaimIacInput) error {
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

	// Create the persistent volume claim
	if _, err := createPersistentVolumeClaim(ctx, locals, kubernetesProvider); err != nil {
		return errors.Wrap(err, "failed to create persistent volume claim")
	}

	// Export outputs
	if err := exportOutputs(ctx, locals); err != nil {
		return errors.Wrap(err, "failed to export outputs")
	}

	return nil
}
