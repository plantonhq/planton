package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func floatingIp(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudFloatingIp.Spec

	floatingIpArgs := &hcloud.FloatingIpArgs{
		Name:             pulumi.String(locals.HetznerCloudFloatingIp.Metadata.Name),
		Type:             pulumi.String(spec.Type.String()),
		HomeLocation:     pulumi.StringPtr(spec.HomeLocation),
		Labels:           pulumi.ToStringMap(locals.Labels),
		DeleteProtection: pulumi.Bool(spec.DeleteProtection),
	}

	if spec.Description != "" {
		floatingIpArgs.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.ServerId != nil && spec.ServerId.GetValue() != "" {
		serverIdInt, err := strconv.Atoi(spec.ServerId.GetValue())
		if err != nil {
			return errors.Wrapf(err, "failed to parse server_id %q as integer", spec.ServerId.GetValue())
		}
		floatingIpArgs.ServerId = pulumi.IntPtr(serverIdInt)
	}

	createdFloatingIp, err := hcloud.NewFloatingIp(
		ctx,
		"floating-ip",
		floatingIpArgs,
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud floating ip")
	}

	if spec.DnsPtr != "" {
		floatingIpIdInt := createdFloatingIp.ID().ApplyT(func(id pulumi.ID) (int, error) {
			return strconv.Atoi(string(id))
		}).(pulumi.IntOutput)

		if _, err := hcloud.NewRdns(
			ctx,
			"rdns",
			&hcloud.RdnsArgs{
				FloatingIpId: floatingIpIdInt,
				IpAddress:    createdFloatingIp.IpAddress,
				DnsPtr:       pulumi.String(spec.DnsPtr),
			},
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create reverse dns record")
		}
	}

	ctx.Export(OpFloatingIpId, createdFloatingIp.ID())
	ctx.Export(OpIpAddress, createdFloatingIp.IpAddress)
	ctx.Export(OpIpNetwork, createdFloatingIp.IpNetwork)

	return nil
}
