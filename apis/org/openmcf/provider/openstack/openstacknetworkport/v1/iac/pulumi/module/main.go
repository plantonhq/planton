package module

import (
	"github.com/pkg/errors"
	openstacknetworkportv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstacknetworkport/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/openstack/pulumiopenstackprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *openstacknetworkportv1.OpenStackNetworkPortStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi OpenStack provider from the supplied credential.
	openstackProvider, err := pulumiopenstackprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup openstack provider")
	}

	// 3. Create the Neutron port with fixed IPs and security groups.
	if err := port(ctx, locals, openstackProvider); err != nil {
		return errors.Wrap(err, "failed to create openstack network port")
	}

	return nil
}
