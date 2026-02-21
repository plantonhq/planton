package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func primaryIp(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudPrimaryIp.Spec

	createdPrimaryIp, err := hcloud.NewPrimaryIp(
		ctx,
		"primary-ip",
		&hcloud.PrimaryIpArgs{
			Name:             pulumi.String(locals.HetznerCloudPrimaryIp.Metadata.Name),
			Type:             pulumi.String(spec.Type.String()),
			Location:         pulumi.StringPtr(spec.Location),
			AssigneeType:     pulumi.String("server"),
			AutoDelete:       pulumi.Bool(false),
			Labels:           pulumi.ToStringMap(locals.Labels),
			DeleteProtection: pulumi.Bool(spec.DeleteProtection),
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud primary ip")
	}

	if spec.DnsPtr != "" {
		primaryIpIdInt := createdPrimaryIp.ID().ApplyT(func(id pulumi.ID) (int, error) {
			return strconv.Atoi(string(id))
		}).(pulumi.IntOutput)

		if _, err := hcloud.NewRdns(
			ctx,
			"rdns",
			&hcloud.RdnsArgs{
				PrimaryIpId: primaryIpIdInt,
				IpAddress:   createdPrimaryIp.IpAddress,
				DnsPtr:      pulumi.String(spec.DnsPtr),
			},
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create reverse dns record")
		}
	}

	ctx.Export(OpPrimaryIpId, createdPrimaryIp.ID())
	ctx.Export(OpIpAddress, createdPrimaryIp.IpAddress)
	ctx.Export(OpIpNetwork, createdPrimaryIp.IpNetwork)

	return nil
}
