package module

import (
	"github.com/pkg/errors"
	digitaloceandnsrecordv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandnsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point for provisioning DigitalOcean DNS records.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceandnsrecordv1alpha1.DigitalOceanDnsRecordIacInput,
) error {
	// 1. Initialize locals.
	locals := initializeLocals(ctx, iacInput)

	// 2. Create DigitalOcean provider from credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Provision the DNS record.
	if err := dnsRecord(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create dns record")
	}

	return nil
}
