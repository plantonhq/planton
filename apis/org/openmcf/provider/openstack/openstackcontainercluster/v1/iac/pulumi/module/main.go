package module

import (
	"github.com/pkg/errors"
	openstackcontainerclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackcontainercluster/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/openstack/pulumiopenstackprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *openstackcontainerclusterv1.OpenStackContainerClusterStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi OpenStack provider from the supplied credential.
	openstackProvider, err := pulumiopenstackprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup openstack provider")
	}

	// 3. Create the Magnum container cluster.
	if err := cluster(ctx, locals, openstackProvider); err != nil {
		return errors.Wrap(err, "failed to create openstack container cluster")
	}

	return nil
}
