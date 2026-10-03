package module

import (
	"github.com/pkg/errors"
	digitaloceanfirewallv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanfirewall/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entrypoint.
func Resources(
	ctx *pulumi.Context,
	iacInput *digitaloceanfirewallv1alpha1.DigitalOceanFirewallIacInput,
) error {
	// 1. Setup locals.
	locals := initializeLocals(ctx, iacInput)

	// 2. Instantiate provider from credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		iacInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Provision firewall.
	if _, err := firewall(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create firewall")
	}

	return nil
}
