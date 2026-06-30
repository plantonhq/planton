package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// volume provisions the Block Storage Volume itself and exports its ID.
func volume(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Volume, error) {

	// 1. Build the resource arguments straight from the proto fields.
	volumeArgs := &digitalocean.VolumeArgs{
		Name:   pulumi.String(locals.DigitalOceanVolume.Spec.VolumeName),
		Region: pulumi.String(locals.DigitalOceanVolume.Spec.Region.String()),
		Size:   pulumi.Int(int(locals.DigitalOceanVolume.Spec.SizeGib)),
	}

	// Optional fields.
	if locals.DigitalOceanVolume.Spec.Description != "" {
		volumeArgs.Description = pulumi.StringPtr(locals.DigitalOceanVolume.Spec.Description)
	}

	if locals.DigitalOceanVolume.Spec.SnapshotId != "" {
		volumeArgs.SnapshotId = pulumi.StringPtr(locals.DigitalOceanVolume.Spec.SnapshotId)
	}

	if len(locals.DigitalOceanVolume.Spec.Tags) > 0 {
		volumeArgs.Tags = pulumi.ToStringArray(locals.DigitalOceanVolume.Spec.Tags)
	}

	// Filesystem type (omit when unformatted).
	fsType := locals.DigitalOceanVolume.Spec.FilesystemType.String()
	if fsType != "unformatted" {
		volumeArgs.FilesystemType = pulumi.StringPtr(fsType)
	}

	// 2. Create the Volume.
	createdVolume, err := digitalocean.NewVolume(
		ctx,
		"volume",
		volumeArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean volume")
	}

	// 3. Export stack output.
	ctx.Export(OpVolumeId, createdVolume.ID())

	return createdVolume, nil
}
