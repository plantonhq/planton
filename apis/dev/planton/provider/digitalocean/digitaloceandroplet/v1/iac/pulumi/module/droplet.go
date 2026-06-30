package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// droplet provisions the DigitalOcean Droplet and exports stack outputs.
func droplet(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Droplet, error) {

	// 1. Build Droplet arguments directly from the proto spec.
	dropletArgs := &digitalocean.DropletArgs{
		Name:    pulumi.String(locals.DigitalOceanDroplet.Metadata.Name),
		Region:  pulumi.String(locals.DigitalOceanDroplet.Spec.Region.String()),
		Size:    pulumi.String(locals.DigitalOceanDroplet.Spec.Size),
		Image:   pulumi.String(locals.DigitalOceanDroplet.Spec.Image),
		Ipv6:    pulumi.Bool(locals.DigitalOceanDroplet.Spec.EnableIpv6),
		Backups: pulumi.Bool(locals.DigitalOceanDroplet.Spec.EnableBackups),
		Monitoring: pulumi.Bool(
			!locals.DigitalOceanDroplet.Spec.DisableMonitoring),
	}

	// Optional: user‑provided cloud‑init script (≤32 .KiB).
	if locals.DigitalOceanDroplet.Spec.UserData != "" {
		dropletArgs.UserData = pulumi.String(
			locals.DigitalOceanDroplet.Spec.UserData)
	}

	// Optional: VPC UUID.
	if locals.DigitalOceanDroplet.Spec.Vpc != nil &&
		locals.DigitalOceanDroplet.Spec.Vpc.GetValue() != "" {
		dropletArgs.VpcUuid = pulumi.String(
			locals.DigitalOceanDroplet.Spec.Vpc.GetValue())
	}

	// Optional: existing volume attachments.
	if len(locals.DigitalOceanDroplet.Spec.VolumeIds) > 0 {
		var volumeIds pulumi.StringArray
		for _, v := range locals.DigitalOceanDroplet.Spec.VolumeIds {
			if v.GetValue() != "" {
				volumeIds = append(volumeIds, pulumi.String(v.GetValue()))
			}
		}
		if len(volumeIds) > 0 {
			dropletArgs.VolumeIds = volumeIds
		}
	}

	// Optional: tags supplied by the user.
	if len(locals.DigitalOceanDroplet.Spec.Tags) > 0 {
		var tags pulumi.StringArray
		for _, t := range locals.DigitalOceanDroplet.Spec.Tags {
			tags = append(tags, pulumi.String(t))
		}
		dropletArgs.Tags = tags
	}

	// 2. Create the Droplet.
	createdDroplet, err := digitalocean.NewDroplet(
		ctx,
		"droplet",
		dropletArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(
			err, "failed to create digitalocean droplet")
	}

	// 3. Export stack outputs.
	ctx.Export(OpDropletId, createdDroplet.ID())
	ctx.Export(OpIpv4Address, createdDroplet.Ipv4Address)
	ctx.Export(OpIpv6Address, createdDroplet.Ipv6Address)
	ctx.Export(OpImageId, createdDroplet.Image)
	ctx.Export(OpVpcUuid, createdDroplet.VpcUuid)

	return createdDroplet, nil
}
