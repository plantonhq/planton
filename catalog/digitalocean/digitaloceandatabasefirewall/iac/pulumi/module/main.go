package module

import (
	"github.com/pkg/errors"
	digitaloceandatabasefirewallv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasefirewall/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceandatabasefirewallv1alpha1.DigitalOceanDatabaseFirewallIacInput,
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

	// 3. Create the database firewall rule set.
	if _, err := databaseFirewall(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create database firewall")
	}

	return nil
}
