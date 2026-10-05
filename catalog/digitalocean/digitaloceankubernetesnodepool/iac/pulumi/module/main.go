package module

import (
	"github.com/pkg/errors"
	digitaloceankubernetesnodepoolv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceankubernetesnodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mimics digital_ocean_vpc.Resources().
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceankubernetesnodepoolv1alpha1.DigitalOceanKubernetesNodePoolIacInput,
) error {
	// 1. Prepare locals.
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a DigitalOcean provider from the credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Provision the node‑pool.
	if _, err := nodePool(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create kubernetes node pool")
	}

	return nil
}
