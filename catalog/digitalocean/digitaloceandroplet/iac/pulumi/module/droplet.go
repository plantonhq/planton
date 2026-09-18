package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// droplet provisions the DigitalOcean Droplet, modeling the complete
// digitalocean_droplet resource surface, and exports the stack outputs.
func droplet(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Droplet, error) {
	spec := locals.DigitalOceanDroplet.Spec

	// 1. Build Droplet arguments from the proto spec.
	dropletArgs := &digitalocean.DropletArgs{
		Name:             pulumi.String(spec.DropletName),
		Image:            pulumi.String(spec.Image),
		Size:             pulumi.String(spec.Size),
		Ipv6:             pulumi.Bool(spec.EnableIpv6),
		Backups:          pulumi.Bool(spec.EnableBackups),
		Monitoring:       pulumi.Bool(spec.Monitoring),
		GracefulShutdown: pulumi.Bool(spec.GracefulShutdown),
	}

	// Public networking is create-only and defaults to on. Sent only when the
	// manifest states it, so an unset field defers to DigitalOcean's default
	// and an explicit false creates a droplet with no public interface -- the
	// same presence contract as the Terraform module.
	if spec.PublicNetworking != nil {
		dropletArgs.PublicNetworking = pulumi.BoolPtr(spec.GetPublicNetworking())
	}

	// GPU partitioning is create-only and only meaningful on GPU sizes; unset
	// must arrive as null, never "" (the provider rejects it).
	if spec.GpuPartitionMode != "" {
		dropletArgs.GpuPartitionMode = pulumi.StringPtr(spec.GpuPartitionMode)
	}

	// Region is optional: unset (the zero enum value) lets DigitalOcean
	// choose a region with available capacity. The enum's value names are
	// the region slugs.
	if spec.Region != 0 {
		dropletArgs.Region = pulumi.String(spec.Region.String())
	}

	// Optional VPC placement; unset means the region's default VPC.
	if spec.GetVpc().GetValue() != "" {
		dropletArgs.VpcUuid = pulumi.String(spec.GetVpc().GetValue())
	}

	// SSH keys are create-only: the standard access path to the droplet.
	if len(spec.SshKeys) > 0 {
		var sshKeys pulumi.StringArray
		for _, key := range spec.SshKeys {
			sshKeys = append(sshKeys, pulumi.String(key))
		}
		dropletArgs.SshKeys = sshKeys
	}

	// Backup policy window (spec validation guarantees backups are enabled
	// when a policy is present). Hour 0 is a real window start (midnight),
	// so it is always sent.
	if spec.BackupPolicy != nil {
		policyArgs := &digitalocean.DropletBackupPolicyArgs{
			Hour: pulumi.Int(int(spec.BackupPolicy.Hour)),
		}
		if spec.BackupPolicy.Plan != "" {
			policyArgs.Plan = pulumi.String(spec.BackupPolicy.Plan)
		}
		if spec.BackupPolicy.Weekday != "" {
			policyArgs.Weekday = pulumi.String(spec.BackupPolicy.Weekday)
		}
		dropletArgs.BackupPolicy = policyArgs
	}

	// droplet_agent is tri-state: unset lets DigitalOcean install where the
	// image supports it; explicit values are forwarded.
	if spec.DropletAgent != nil {
		dropletArgs.DropletAgent = pulumi.Bool(spec.GetDropletAgent())
	}

	// resize_disk defaults ON provider-side; unset is never coalesced to
	// false — that would silently flip the provider default.
	if spec.ResizeDisk != nil {
		dropletArgs.ResizeDisk = pulumi.Bool(spec.GetResizeDisk())
	}

	// Block volume attachments.
	if len(spec.VolumeIds) > 0 {
		var volumeIds pulumi.StringArray
		for _, v := range spec.VolumeIds {
			if v.GetValue() != "" {
				volumeIds = append(volumeIds, pulumi.String(v.GetValue()))
			}
		}
		if len(volumeIds) > 0 {
			dropletArgs.VolumeIds = volumeIds
		}
	}

	// Cloud-init user data (create-only, hash-stored by DigitalOcean).
	if spec.UserData != "" {
		dropletArgs.UserData = pulumi.String(spec.UserData)
	}

	// User tags plus the standard Planton labels rendered as "key:value"
	// tags — the exact set the Terraform module applies.
	tagSet := map[string]bool{}
	var tagInputs pulumi.StringArray
	for _, t := range spec.Tags {
		if !tagSet[t] {
			tagSet[t] = true
			tagInputs = append(tagInputs, pulumi.String(t))
		}
	}
	for k, v := range locals.DigitalOceanLabels {
		t := k + ":" + v
		if !tagSet[t] {
			tagSet[t] = true
			tagInputs = append(tagInputs, pulumi.String(t))
		}
	}
	dropletArgs.Tags = tagInputs

	// 2. Create the Droplet.
	createdDroplet, err := digitalocean.NewDroplet(
		ctx,
		"droplet",
		dropletArgs,
		pulumi.Provider(digitalOceanProvider),
		// ssh_keys, user_data, and droplet_agent are applied at creation ONLY
		// and never read back by the API; the provider marks all three
		// ForceNew. Left unguarded, a manifest edit to any of them -- or
		// adopting an existing droplet whose manifest carries them -- would
		// plan a destroy-and-recreate of a running machine (its disk and its
		// IP with it). They have no meaning after first boot, so later changes
		// are ignored here, exactly as the Terraform module's
		// lifecycle.ignore_changes does; the spec field comments tell manifest
		// authors the same. To re-run cloud-init or change the injected keys,
		// replace the droplet deliberately.
		pulumi.IgnoreChanges([]string{"sshKeys", "userData", "dropletAgent"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean droplet")
	}

	// 3. Export stack outputs — exactly the DigitalOceanDropletStackOutputs
	// contract, from the SDK's real field names.
	ctx.Export(OpDropletId, createdDroplet.ID())
	ctx.Export(OpIpv4Address, createdDroplet.Ipv4Address)
	ctx.Export(OpIpv6Address, createdDroplet.Ipv6Address)
	ctx.Export(OpIpv4AddressPrivate, createdDroplet.Ipv4AddressPrivate)
	ctx.Export(OpUrn, createdDroplet.DropletUrn)
	ctx.Export(OpVpcUuid, createdDroplet.VpcUuid)

	return createdDroplet, nil
}
