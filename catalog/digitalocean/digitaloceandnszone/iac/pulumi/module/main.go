package module

import (
	"github.com/pkg/errors"
	digitaloceandnszonev1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandnszone/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point, mirroring digital_ocean_vpc.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceandnszonev1alpha1.DigitalOceanDnsZoneIacInput,
) error {
	// 1. Collate locals.
	locals := initializeLocals(ctx, iacInput)

	// 2. Create DO provider from credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Provision the DNS zone (domain + records).
	if _, err := dnsZone(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create dns zone")
	}

	return nil
}
