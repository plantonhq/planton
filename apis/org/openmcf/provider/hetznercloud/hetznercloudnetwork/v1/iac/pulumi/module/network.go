package module

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func network(
	ctx *pulumi.Context,
	locals *Locals,
	hcloudProvider *hcloud.Provider,
) error {
	spec := locals.HetznerCloudNetwork.Spec

	createdNetwork, err := hcloud.NewNetwork(
		ctx,
		"network",
		&hcloud.NetworkArgs{
			Name:                  pulumi.String(locals.HetznerCloudNetwork.Metadata.Name),
			IpRange:               pulumi.String(spec.IpRange),
			Labels:                pulumi.ToStringMap(locals.Labels),
			DeleteProtection:      pulumi.Bool(spec.DeleteProtection),
			ExposeRoutesToVswitch: pulumi.Bool(spec.ExposeRoutesToVswitch),
		},
		pulumi.Provider(hcloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create hetzner cloud network")
	}

	// The Pulumi hcloud SDK expects IntInput for NetworkId in subnet and route
	// args, but Network.ID() returns IDOutput (string). Convert via ApplyT.
	networkIdInt := createdNetwork.ID().ApplyT(func(id pulumi.ID) (int, error) {
		return strconv.Atoi(string(id))
	}).(pulumi.IntOutput)

	for _, subnet := range spec.Subnets {
		subnetArgs := &hcloud.NetworkSubnetArgs{
			NetworkId:   networkIdInt,
			Type:        pulumi.String(subnet.Type.String()),
			NetworkZone: pulumi.String(subnet.NetworkZone),
			IpRange:     pulumi.String(subnet.IpRange),
		}

		if subnet.VswitchId != 0 {
			subnetArgs.VswitchId = pulumi.Int(int(subnet.VswitchId))
		}

		resourceName := fmt.Sprintf("subnet-%s", sanitizeCidr(subnet.IpRange))

		if _, err := hcloud.NewNetworkSubnet(
			ctx,
			resourceName,
			subnetArgs,
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrapf(err, "failed to create subnet %s", subnet.IpRange)
		}
	}

	for _, route := range spec.Routes {
		resourceName := fmt.Sprintf("route-%s", sanitizeCidr(route.Destination))

		if _, err := hcloud.NewNetworkRoute(
			ctx,
			resourceName,
			&hcloud.NetworkRouteArgs{
				NetworkId:   networkIdInt,
				Destination: pulumi.String(route.Destination),
				Gateway:     pulumi.String(route.Gateway),
			},
			pulumi.Provider(hcloudProvider),
		); err != nil {
			return errors.Wrapf(err, "failed to create route to %s", route.Destination)
		}
	}

	ctx.Export(OpNetworkId, createdNetwork.ID())

	return nil
}

// sanitizeCidr converts a CIDR or IP string into a Pulumi-safe resource name
// component by replacing dots and slashes with hyphens.
// Example: "10.0.1.0/24" -> "10-0-1-0-24"
func sanitizeCidr(cidr string) string {
	r := strings.NewReplacer(".", "-", "/", "-", ":", "-")
	return r.Replace(cidr)
}
