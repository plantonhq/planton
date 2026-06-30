package module

import (
	"strconv"

	"github.com/pkg/errors"
	hetznercloudvolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudvolume/v1"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func volume(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudVolume.Spec

	volumeArgs := &hcloud.VolumeArgs{
		Name:             pulumi.String(locals.HetznerCloudVolume.Metadata.Name),
		Size:             pulumi.Int(spec.Size),
		Location:         pulumi.StringPtr(spec.Location),
		Labels:           pulumi.ToStringMap(locals.Labels),
		DeleteProtection: pulumi.Bool(spec.DeleteProtection),
	}

	if spec.Format != hetznercloudvolumev1.HetznerCloudVolumeSpec_format_unspecified {
		volumeArgs.Format = pulumi.StringPtr(spec.Format.String())
	}

	createdVolume, err := hcloud.NewVolume(
		ctx,
		"volume",
		volumeArgs,
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud volume")
	}

	if spec.ServerId != nil && spec.ServerId.GetValue() != "" {
		volumeIdInt := createdVolume.ID().ApplyT(func(id pulumi.ID) (int, error) {
			return strconv.Atoi(string(id))
		}).(pulumi.IntOutput)

		serverIdInt, err := strconv.Atoi(spec.ServerId.GetValue())
		if err != nil {
			return errors.Wrapf(err, "failed to parse server_id %q as integer",
				spec.ServerId.GetValue())
		}

		attachmentArgs := &hcloud.VolumeAttachmentArgs{
			VolumeId: volumeIdInt,
			ServerId: pulumi.Int(serverIdInt),
		}

		if spec.Automount {
			attachmentArgs.Automount = pulumi.BoolPtr(true)
		}

		if _, err := hcloud.NewVolumeAttachment(
			ctx,
			"volume-attachment",
			attachmentArgs,
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create volume attachment")
		}
	}

	ctx.Export(OpVolumeId, createdVolume.ID())
	ctx.Export(OpLinuxDevice, createdVolume.LinuxDevice)

	return nil
}
