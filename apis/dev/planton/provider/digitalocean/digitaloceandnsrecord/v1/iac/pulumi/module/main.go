package module

import (
	"github.com/pkg/errors"
	digitaloceandnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceandnsrecord/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point for provisioning DigitalOcean DNS records.
func Resources(
	ctx *pulumi.Context,
	stackInput *digitaloceandnsrecordv1.DigitalOceanDnsRecordStackInput,
) error {
	// 1. Initialize locals.
	locals := initializeLocals(ctx, stackInput)

	// 2. Create DigitalOcean provider from credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		stackInput.ProviderConfig,
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
