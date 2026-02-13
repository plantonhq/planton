package module

import (
	scalewayblockvolumev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayblockvolume/v1"
	"github.com/pkg/errors"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/block"
	scalewayv2 "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// performanceTierToIops maps the proto enum to the integer IOPS value
// expected by the Scaleway Block Storage API.
var performanceTierToIops = map[scalewayblockvolumev1.ScalewayBlockVolumePerformanceTier]int{
	scalewayblockvolumev1.ScalewayBlockVolumePerformanceTier_sbs_5k:  5000,
	scalewayblockvolumev1.ScalewayBlockVolumePerformanceTier_sbs_15k: 15000,
}

// volume provisions the Scaleway Block Storage volume and exports
// stack outputs.
func volume(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scalewayv2.Provider,
) error {
	spec := locals.ScalewayBlockVolume.Spec

	// Resolve performance tier enum to integer IOPS.
	iops, ok := performanceTierToIops[spec.PerformanceTier]
	if !ok {
		return errors.Errorf("unsupported performance tier: %s", spec.PerformanceTier.String())
	}

	// Build volume arguments.
	volumeArgs := &block.VolumeArgs{
		Name: pulumi.StringPtr(locals.ScalewayBlockVolume.Metadata.Name),
		Iops: pulumi.Int(iops),
		Tags: pulumi.ToStringArray(locals.ScalewayTags),
	}

	// Size is required in the spec but optional in the SDK (because it
	// can be inferred from a snapshot). We always set it explicitly.
	volumeArgs.SizeInGb = pulumi.IntPtr(int(spec.SizeGb))

	// Zone (optional in SDK -- inherits from provider if omitted, but
	// we set it explicitly from the spec for clarity).
	volumeArgs.Zone = pulumi.StringPtr(spec.Zone)

	// Optional: create from snapshot.
	if spec.SnapshotId != "" {
		volumeArgs.SnapshotId = pulumi.StringPtr(spec.SnapshotId)
	}

	// Create the volume.
	createdVolume, err := block.NewVolume(
		ctx,
		"volume",
		volumeArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create scaleway block volume")
	}

	// Export stack outputs.
	ctx.Export(OpVolumeId, createdVolume.ID())
	ctx.Export(OpVolumeName, createdVolume.Name)
	ctx.Export(OpZone, pulumi.String(spec.Zone))

	return nil
}
