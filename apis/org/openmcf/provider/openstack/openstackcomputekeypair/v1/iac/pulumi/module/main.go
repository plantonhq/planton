package module

import (
	"github.com/pkg/errors"
	openstackcomputekeypairv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackcomputekeypair/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/openstack/pulumiopenstackprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *openstackcomputekeypairv1.OpenStackComputeKeypairStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi OpenStack provider from the supplied credential.
	openstackProvider, err := pulumiopenstackprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup openstack provider")
	}

	// 3. Create the compute keypair.
	if err := keypair(ctx, locals, openstackProvider); err != nil {
		return errors.Wrap(err, "failed to create openstack compute keypair")
	}

	return nil
}
