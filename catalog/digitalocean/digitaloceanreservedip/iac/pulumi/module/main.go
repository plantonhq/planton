package module

import (
	"github.com/pkg/errors"
	digitaloceanreservedipv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanreservedip/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceanreservedipv1alpha1.DigitalOceanReservedIpIacInput,
) error {
	// 1. Prepare locals (target handle, version switch, droplet id).
	locals, err := initializeLocals(ctx, iacInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// 2. Create a DigitalOcean provider from the supplied credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Reserve the address (and assign it, when a droplet is set).
	if locals.IsIpv6 {
		if err := reservedIpv6(ctx, locals, digitalOceanProvider); err != nil {
			return errors.Wrap(err, "failed to create reserved ipv6")
		}
		return nil
	}
	if err := reservedIpv4(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create reserved ip")
	}

	return nil
}
