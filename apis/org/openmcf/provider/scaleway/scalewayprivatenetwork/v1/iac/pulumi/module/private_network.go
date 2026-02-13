package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/network"
)

// privateNetwork provisions the Scaleway Private Network and exports its ID
// and allocated IPv4 subnet CIDR.
func privateNetwork(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) (*network.PrivateNetwork, error) {
	spec := locals.ScalewayPrivateNetwork.Spec

	// 1. Build the resource arguments.
	pnArgs := &network.PrivateNetworkArgs{
		Name:   pulumi.String(locals.ScalewayPrivateNetwork.Metadata.Name),
		VpcId:  pulumi.String(locals.VpcId),
		Region: pulumi.String(spec.Region),
		Tags:   pulumi.ToStringArray(locals.ScalewayTags),
	}

	// 2. Set optional IPv4 subnet if specified.
	// If omitted, Scaleway's IPAM auto-allocates a subnet.
	if spec.Ipv4Subnet != "" {
		pnArgs.Ipv4Subnet = &network.PrivateNetworkIpv4SubnetArgs{
			Subnet: pulumi.String(spec.Ipv4Subnet),
		}
	}

	// 3. Set optional IPv6 subnets if specified.
	if len(spec.Ipv6Subnets) > 0 {
		var ipv6Subnets network.PrivateNetworkIpv6SubnetArray
		for _, cidr := range spec.Ipv6Subnets {
			ipv6Subnets = append(ipv6Subnets, &network.PrivateNetworkIpv6SubnetArgs{
				Subnet: pulumi.String(cidr),
			})
		}
		pnArgs.Ipv6Subnets = ipv6Subnets
	}

	// 4. Enable default route propagation when requested.
	if spec.EnableDefaultRoutePropagation {
		pnArgs.EnableDefaultRoutePropagation = pulumi.Bool(true)
	}

	// 5. Create the Private Network.
	createdPN, err := network.NewPrivateNetwork(
		ctx,
		"private_network",
		pnArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create scaleway private network")
	}

	// 6. Export stack outputs.
	ctx.Export(OpPrivateNetworkId, createdPN.ID())

	// The Ipv4Subnet output is always populated (either from the requested CIDR
	// or auto-allocated by IPAM). Access the nested Subnet field.
	ctx.Export(OpIpv4SubnetCidr, createdPN.Ipv4Subnet.Subnet())

	return createdPN, nil
}
