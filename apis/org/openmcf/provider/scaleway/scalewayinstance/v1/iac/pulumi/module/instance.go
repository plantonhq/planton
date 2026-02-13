package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	scalewayinstance "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/instance"
)

// createInstance provisions the complete Instance composite:
//
//  1. An optional dedicated Flexible IP (public IPv4 address).
//  2. Optional additional local volumes (l_ssd, scratch).
//  3. The instance server with root volume configuration, optional
//     private network attachment, and optional security group.
//
// All outputs are exported for downstream resource references.
func createInstance(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) error {
	spec := locals.ScalewayInstance.Spec

	// ── 1. Create the Flexible IP (optional) ───────────────────────────────
	//
	// A dedicated public IPv4 address for the instance. Created as a separate
	// resource to give explicit lifecycle control -- the IP survives instance
	// replacement, preserving DNS records and firewall rules.
	//
	// Only created when spec.public_ip is set.

	var createdIp *scalewayinstance.Ip
	if spec.PublicIp != nil {
		ipArgs := &scalewayinstance.IpArgs{
			Tags: pulumi.ToStringArray(locals.ScalewayTags),
			Zone: pulumi.String(spec.Zone),
		}

		var err error
		createdIp, err = scalewayinstance.NewIp(
			ctx,
			"instance-ip",
			ipArgs,
			pulumi.Provider(scalewayProvider),
		)
		if err != nil {
			return errors.Wrap(err, "failed to create instance flexible ip")
		}

		// Export public IP outputs.
		ctx.Export(OpPublicIpAddress, createdIp.Address)
		ctx.Export(OpPublicIpId, createdIp.ID())
	} else {
		// No public IP -- export empty strings for consistent output schema.
		ctx.Export(OpPublicIpAddress, pulumi.String(""))
		ctx.Export(OpPublicIpId, pulumi.String(""))
	}

	// ── 2. Create additional local volumes (optional) ──────────────────────
	//
	// Local volumes (l_ssd, scratch) that are created alongside the instance
	// and attached via additional_volume_ids. These volumes share the
	// instance's lifecycle.

	additionalVolumeIds := pulumi.StringArray{}

	for i, volSpec := range spec.AdditionalVolumes {
		volName := volSpec.Name
		if volName == "" {
			volName = fmt.Sprintf("%s-vol-%d", locals.ScalewayInstance.Metadata.Name, i)
		}

		createdVolume, err := scalewayinstance.NewVolume(
			ctx,
			fmt.Sprintf("volume-%d", i),
			&scalewayinstance.VolumeArgs{
				Name:     pulumi.String(volName),
				Type:     pulumi.String(volSpec.VolumeType),
				SizeInGb: pulumi.Int(int(volSpec.SizeInGb)),
				Tags:     pulumi.ToStringArray(locals.ScalewayTags),
				Zone:     pulumi.String(spec.Zone),
			},
			pulumi.Provider(scalewayProvider),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create additional volume %d (%s)", i, volName)
		}

		additionalVolumeIds = append(additionalVolumeIds, createdVolume.ID())
	}

	// ── 3. Create the instance server ──────────────────────────────────────
	//
	// The compute instance itself. References the Flexible IP (if created),
	// additional volumes, optional security group, and optional Private
	// Network attachment.

	serverArgs := &scalewayinstance.ServerArgs{
		Name:  pulumi.String(locals.ScalewayInstance.Metadata.Name),
		Type:  pulumi.String(spec.Type),
		Image: pulumi.String(spec.Image),
		Tags:  pulumi.ToStringArray(locals.ScalewayTags),
		Zone:  pulumi.String(spec.Zone),
	}

	// Attach the dedicated Flexible IP if created.
	if spec.PublicIp != nil && createdIp != nil {
		serverArgs.IpId = createdIp.ID().ToStringOutput()
	}

	// Attach additional volumes if any were created.
	if len(additionalVolumeIds) > 0 {
		serverArgs.AdditionalVolumeIds = additionalVolumeIds
	}

	// Configure the root volume if specified.
	if spec.RootVolume != nil {
		rootVol := &scalewayinstance.ServerRootVolumeArgs{}

		if spec.RootVolume.SizeInGb > 0 {
			rootVol.SizeInGb = pulumi.Int(int(spec.RootVolume.SizeInGb))
		}

		if spec.RootVolume.VolumeType != "" {
			rootVol.VolumeType = pulumi.String(spec.RootVolume.VolumeType)
		}

		if spec.RootVolume.DeleteOnTermination {
			rootVol.DeleteOnTermination = pulumi.Bool(spec.RootVolume.DeleteOnTermination)
		}

		if spec.RootVolume.SbsIops > 0 {
			rootVol.SbsIops = pulumi.Int(int(spec.RootVolume.SbsIops))
		}

		serverArgs.RootVolume = rootVol
	}

	// Attach to security group if specified.
	if locals.SecurityGroupId != "" {
		serverArgs.SecurityGroupId = pulumi.String(locals.SecurityGroupId)
	}

	// Attach to Private Network if specified (inline block on server).
	if locals.PrivateNetworkId != "" {
		serverArgs.PrivateNetworks = scalewayinstance.ServerPrivateNetworkArray{
			&scalewayinstance.ServerPrivateNetworkArgs{
				PnId: pulumi.String(locals.PrivateNetworkId),
			},
		}
	}

	// Set cloud-init script if specified.
	if spec.CloudInit != "" {
		serverArgs.CloudInit = pulumi.String(spec.CloudInit)
	}

	// Set instance state if specified (default is "started").
	if spec.State != "" {
		serverArgs.State = pulumi.String(spec.State)
	}

	// Set deletion protection if enabled.
	if spec.Protected {
		serverArgs.Protected = pulumi.Bool(true)
	}

	createdServer, err := scalewayinstance.NewServer(
		ctx,
		"server",
		serverArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create instance server")
	}

	// ── 4. Export stack outputs ─────────────────────────────────────────────

	ctx.Export(OpServerId, createdServer.ID())

	// Export private IP address from the server's computed private_ips.
	// The first private IP is from the attached Private Network (if any).
	if locals.PrivateNetworkId != "" {
		ctx.Export(OpPrivateIpAddress, createdServer.PrivateIps.ApplyT(func(ips []scalewayinstance.ServerPrivateIp) string {
			if len(ips) > 0 && ips[0].Address != nil {
				return *ips[0].Address
			}
			return ""
		}).(pulumi.StringOutput))
	} else {
		ctx.Export(OpPrivateIpAddress, pulumi.String(""))
	}

	return nil
}
