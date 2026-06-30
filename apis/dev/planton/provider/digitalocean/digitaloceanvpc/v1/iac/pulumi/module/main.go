package module

import (
	"github.com/pkg/errors"
	digitaloceanvpcv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanvpc/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/pulumidigitaloceanprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point—mirrors the pattern used in gcp_vpc.
func Resources(
	ctx *pulumi.Context,
	stackInput *digitaloceanvpcv1.DigitalOceanVpcStackInput,
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

	// 3. Create the VPC network.
	if _, err := vpc(ctx, locals, digitalOceanProvider); err != nil {
		return errors.Wrap(err, "failed to create vpc")
	}

	return nil
}
