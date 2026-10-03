package module

import (
	"github.com/pkg/errors"
	digitaloceanvpcpeeringv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanvpcpeering/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceanvpcpeeringv1alpha1.DigitalOceanVpcPeeringIacInput,
) error {
	// 1. Prepare locals (target handle).
	locals := initializeLocals(ctx, iacInput)

	// 2. Create a DigitalOcean provider from the supplied credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Create the VPC peering connection.
	if _, err := vpcPeering(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create vpc peering")
	}

	return nil
}
