package module

import (
	"fmt"

	"github.com/pkg/errors"
	civo "github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// volume provisions the Civo Block Storage Volume itself and exports its ID.
func volume(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.Volume, error) {

	// 1. Build the resource arguments from the proto fields.
	volumeArgs := &civo.VolumeArgs{
		Name:   pulumi.String(locals.CivoVolume.Spec.VolumeName),
		Region: pulumi.String(locals.CivoVolume.Spec.Region.String()),
		SizeGb: pulumi.Int(int(locals.CivoVolume.Spec.SizeGib)),
	}

	// 2. Handle filesystem_type if specified.
	// Note: The Civo Volume API supports filesystem formatting, but the Pulumi/Terraform provider
	// doesn't currently expose this parameter. Users must format the volume manually after creation
	// or use cloud-init/configuration management to automate formatting.
	filesystemName := locals.CivoVolume.Spec.FilesystemType.String()
	if filesystemName != "unformatted" {
		ctx.Log.Info(fmt.Sprintf(
			"Filesystem type '%s' requested for volume '%s'. "+
				"Note: The Civo provider doesn't expose filesystem formatting. "+
				"The volume will be created unformatted. Use cloud-init or configuration management "+
				"to format the volume as %s after attachment.",
			filesystemName,
			locals.CivoVolume.Spec.VolumeName,
			filesystemName,
		), nil)
	}

	// 3. Handle snapshot_id if specified.
	// Note: Civo Volumes support creation from snapshots, but snapshot functionality
	// is only available in CivoStack (private cloud), not on public Civo cloud.
	// The provider doesn't expose snapshot_id as a parameter.
	if locals.CivoVolume.Spec.SnapshotId != "" {
		ctx.Log.Warn(fmt.Sprintf(
			"Snapshot ID '%s' specified for volume '%s'. "+
				"Note: Civo Volume snapshots are not currently supported on public Civo cloud. "+
				"This parameter is reserved for future use or CivoStack deployments. "+
				"The volume will be created empty.",
			locals.CivoVolume.Spec.SnapshotId,
			locals.CivoVolume.Spec.VolumeName,
		), nil)
	}

	// 4. Handle tags if specified.
	// Note: The Civo Volume provider doesn't currently support tags.
	// Tags in the spec are available for logical organization and metadata
	// but aren't applied to the Civo resource.
	if len(locals.CivoVolume.Spec.Tags) > 0 {
		ctx.Log.Info(fmt.Sprintf(
			"Tags specified for volume '%s': %v. "+
				"Note: The Civo Volume provider doesn't currently support tags. "+
				"Tags are recorded in metadata but not applied to the Civo resource. "+
				"Use the Civo labels (applied automatically by Planton) for resource organization.",
			locals.CivoVolume.Spec.VolumeName,
			locals.CivoVolume.Spec.Tags,
		), nil)
	}

	// 5. Create the Volume.
	createdVolume, err := civo.NewVolume(
		ctx,
		"volume",
		volumeArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create civo volume")
	}

	// 6. Export stack outputs.
	ctx.Export(OpVolumeId, createdVolume.ID())
	// Note: attached_instance_id and device_path are only available after attachment,
	// which is handled separately (either via civo_volume_attachment resource or
	// dynamically by Kubernetes CSI driver). These outputs remain empty for now.

	return createdVolume, nil
}
