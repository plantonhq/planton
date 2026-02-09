package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/blockstorage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// volume provisions the OpenStack Cinder block storage volume and exports outputs.
func volume(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackVolume.Spec
	resourceName := locals.OpenStackVolume.Metadata.Name

	volumeArgs := &blockstorage.VolumeArgs{
		Name: pulumi.StringPtr(resourceName),
		Size: pulumi.Int(int(spec.Size)),
	}

	// Set description if provided.
	if spec.Description != "" {
		volumeArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set volume_type if provided.
	if spec.VolumeType != "" {
		volumeArgs.VolumeType = pulumi.StringPtr(spec.VolumeType)
	}

	// Set availability_zone if provided (ForceNew).
	if spec.AvailabilityZone != "" {
		volumeArgs.AvailabilityZone = pulumi.StringPtr(spec.AvailabilityZone)
	}

	// Set snapshot_id if provided (ForceNew, mutually exclusive with other sources).
	if spec.SnapshotId != "" {
		volumeArgs.SnapshotId = pulumi.StringPtr(spec.SnapshotId)
	}

	// Set source_vol_id if provided (ForceNew, mutually exclusive with other sources).
	if spec.SourceVolId != "" {
		volumeArgs.SourceVolId = pulumi.StringPtr(spec.SourceVolId)
	}

	// Set image_id if provided (ForceNew, mutually exclusive with other sources).
	// Resolved from StringValueOrRef by the FK resolver middleware.
	if locals.ImageId != "" {
		volumeArgs.ImageId = pulumi.StringPtr(locals.ImageId)
	}

	// Set metadata if provided.
	if len(spec.Metadata) > 0 {
		metadataMap := pulumi.StringMap{}
		for k, v := range spec.Metadata {
			metadataMap[k] = pulumi.String(v)
		}
		volumeArgs.Metadata = metadataMap
	}

	// Set region override if provided (ForceNew).
	if spec.Region != "" {
		volumeArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdVolume, err := blockstorage.NewVolume(
		ctx,
		strings.ToLower(resourceName),
		volumeArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack cinder volume")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpVolumeId, createdVolume.ID())
	ctx.Export(OpName, createdVolume.Name)
	ctx.Export(OpSize, createdVolume.Size)
	ctx.Export(OpVolumeType, createdVolume.VolumeType)
	ctx.Export(OpAvailabilityZone, createdVolume.AvailabilityZone)
	ctx.Export(OpRegion, createdVolume.Region)

	return nil
}
