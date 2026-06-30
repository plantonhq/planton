package module

import (
	"github.com/pkg/errors"
	digitaloceankubernetesnodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceankubernetesnodepool/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mimics digital_ocean_vpc.Resources().
func Resources(
	ctx *pulumi.Context,
	stackInput *digitaloceankubernetesnodepoolv1.DigitalOceanKubernetesNodePoolStackInput,
) error {
	// 1. Prepare locals.
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a DigitalOcean provider from the credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Provision the node‑pool.
	if _, err := nodePool(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create kubernetes node pool")
	}

	return nil
}
