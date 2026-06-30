package module

import (
	"github.com/pkg/errors"
	digitaloceanfunctionv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanfunction/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in digital_ocean_vpc.
func Resources(
	ctx *pulumi.Context,
	stackInput *digitaloceanfunctionv1.DigitalOceanFunctionStackInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a DigitalOcean provider from the supplied credential.
	digitalOceanProvider, err := pulumidigitaloceanprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup digitalocean provider")
	}

	// 3. Create the Function.
	if _, err := function(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create digitalocean function")
	}

	return nil
}
