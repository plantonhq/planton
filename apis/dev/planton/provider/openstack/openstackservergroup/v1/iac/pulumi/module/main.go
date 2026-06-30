package module

import (
	"github.com/pkg/errors"
	openstackservergroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openstack/openstackservergroup/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/openstack/pulumiopenstackprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *openstackservergroupv1.OpenStackServerGroupStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi OpenStack provider from the supplied credential.
	openstackProvider, err := pulumiopenstackprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup openstack provider")
	}

	// 3. Create the server group.
	if err := serverGroup(ctx, locals, openstackProvider); err != nil {
		return errors.Wrap(err, "failed to create openstack server group")
	}

	return nil
}
