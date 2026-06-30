package module

import (
	"github.com/pkg/errors"
	digitaloceanfirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanfirewall/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entrypoint.
func Resources(
	ctx *pulumi.Context,
	stackInput *digitaloceanfirewallv1.DigitalOceanFirewallStackInput,
) error {
	// 1. Setup locals.
	locals := initializeLocals(ctx, stackInput)

	// 2. Instantiate provider from credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		stackInput.ProviderConfig,
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
