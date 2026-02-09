package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// volumeAttach provisions the OpenStack compute volume attachment and exports outputs.
func volumeAttach(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackVolumeAttach.Spec
	resourceName := locals.OpenStackVolumeAttach.Metadata.Name

	attachArgs := &compute.VolumeAttachArgs{
		InstanceId: pulumi.String(locals.InstanceId),
		VolumeId:   pulumi.String(locals.VolumeId),
	}

	// Set device if provided (e.g., "/dev/vdb").
	// If omitted, Nova selects the next available device.
	if spec.Device != "" {
		attachArgs.Device = pulumi.StringPtr(spec.Device)
	}

	// Set region override if provided (ForceNew).
	if spec.Region != "" {
		attachArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdAttach, err := compute.NewVolumeAttach(
		ctx,
		strings.ToLower(resourceName),
		attachArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack volume attachment")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpId, createdAttach.ID())
	ctx.Export(OpInstanceId, createdAttach.InstanceId)
	ctx.Export(OpVolumeId, createdAttach.VolumeId)
	ctx.Export(OpDevice, createdAttach.Device)
	ctx.Export(OpRegion, createdAttach.Region)

	return nil
}
