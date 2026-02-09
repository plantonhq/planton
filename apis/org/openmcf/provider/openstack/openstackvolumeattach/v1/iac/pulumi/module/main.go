package module

import (
	"github.com/pkg/errors"
	openstackvolumeattachv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/openstack/openstackvolumeattach/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/openstack/pulumiopenstackprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point called by the Planton CLI.
func Resources(
	ctx *pulumi.Context,
	stackInput *openstackvolumeattachv1.OpenStackVolumeAttachStackInput,
) error {
	// 1. Gather handy references.
	locals := initializeLocals(ctx, stackInput)

	// 2. Build a Pulumi OpenStack provider from the supplied credential.
	openstackProvider, err := pulumiopenstackprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup openstack provider")
	}

	// 3. Attach the volume to the instance.
	if err := volumeAttach(ctx, locals, openstackProvider); err != nil {
		return errors.Wrap(err, "failed to attach openstack volume to instance")
	}

	return nil
}
