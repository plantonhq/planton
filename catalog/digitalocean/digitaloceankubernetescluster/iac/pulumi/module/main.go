package module

import (
	"github.com/pkg/errors"
	digitaloceankubernetesclusterv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceankubernetescluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in digital_ocean_vpc.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceankubernetesclusterv1alpha1.DigitalOceanKubernetesClusterIacInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a DigitalOcean provider from the supplied credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Create the cluster.
	if _, err := cluster(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create kubernetes cluster")
	}

	return nil
}
