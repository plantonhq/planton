package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurevpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurevpc/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurevpcv1.AzureVpcStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVpc.Spec

	// Resource group is provided externally
	resourceGroupName := pulumi.String(locals.ResourceGroup)
	location := pulumi.String(locals.Location)

	// Create Virtual Network
	vnet, err := network.NewVirtualNetwork(ctx,
		locals.VNetName,
		&network.VirtualNetworkArgs{
			Name:              pulumi.String(locals.VNetName),
			ResourceGroupName: resourceGroupName,
			Location:          location,
			AddressSpaces:     pulumi.StringArray{pulumi.String(spec.AddressSpaceCidr)},
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create virtual network")
	}

	// Create NAT Gateway if enabled
	var natGatewayId pulumi.StringOutput
	if spec.IsNatGatewayEnabled {
		// Create Public IP for NAT Gateway
		publicIp, err := network.NewPublicIp(ctx,
			fmt.Sprintf("%s-ip", locals.NatGatewayName),
			&network.PublicIpArgs{
				Name:              pulumi.String(fmt.Sprintf("%s-ip", locals.NatGatewayName)),
				ResourceGroupName: resourceGroupName,
				Location:          location,
				AllocationMethod:  pulumi.String("Static"),
				Sku:               pulumi.String("Standard"),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create public IP for NAT gateway")
		}

		// Create NAT Gateway
		natGateway, err := network.NewNatGateway(ctx,
			locals.NatGatewayName,
			&network.NatGatewayArgs{
				Name:                 pulumi.String(locals.NatGatewayName),
				ResourceGroupName:    resourceGroupName,
				Location:             location,
				SkuName:              pulumi.String("Standard"),
				IdleTimeoutInMinutes: pulumi.Int(4),
				Tags:                 pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create NAT gateway")
		}

		// Associate Public IP with NAT Gateway
		_, err = network.NewNatGatewayPublicIpAssociation(ctx,
			fmt.Sprintf("%s-ip-assoc", locals.NatGatewayName),
			&network.NatGatewayPublicIpAssociationArgs{
				NatGatewayId:      natGateway.ID(),
				PublicIpAddressId: publicIp.ID(),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(natGateway))
		if err != nil {
			return errors.Wrap(err, "failed to associate public IP with NAT gateway")
		}

		natGatewayId = natGateway.ID().ToStringOutput()
	}

	// Create Subnet (with optional NAT Gateway)
	subnetArgs := &network.SubnetArgs{
		Name:               pulumi.String(locals.SubnetName),
		ResourceGroupName:  resourceGroupName,
		VirtualNetworkName: vnet.Name,
		AddressPrefixes:    pulumi.StringArray{pulumi.String(spec.NodesSubnetCidr)},
	}

	subnet, err := network.NewSubnet(ctx,
		locals.SubnetName,
		subnetArgs,
		pulumi.Provider(azureProvider),
		pulumi.Parent(vnet))
	if err != nil {
		return errors.Wrap(err, "failed to create subnet")
	}

	// Associate NAT Gateway with Subnet (if enabled)
	if spec.IsNatGatewayEnabled {
		_, err = network.NewSubnetNatGatewayAssociation(ctx,
			fmt.Sprintf("%s-natgw-assoc", locals.SubnetName),
			&network.SubnetNatGatewayAssociationArgs{
				SubnetId:     subnet.ID(),
				NatGatewayId: natGatewayId,
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(subnet))
		if err != nil {
			return errors.Wrap(err, "failed to associate NAT gateway with subnet")
		}
	}

	// Link Private DNS Zones if specified
	for i, dnsZoneId := range spec.DnsPrivateZoneLinks {
		_, err := privatedns.NewZoneVirtualNetworkLink(ctx,
			fmt.Sprintf("dns-link-%d", i),
			&privatedns.ZoneVirtualNetworkLinkArgs{
				Name:                pulumi.String(fmt.Sprintf("%s-link-%d", locals.VNetName, i)),
				ResourceGroupName:   resourceGroupName,
				PrivateDnsZoneName:  pulumi.String(dnsZoneId), // This should be parsed
				VirtualNetworkId:    vnet.ID(),
				RegistrationEnabled: pulumi.Bool(false),
				Tags:                pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(vnet))
		if err != nil {
			return errors.Wrapf(err, "failed to link private DNS zone %s", dnsZoneId)
		}
	}

	// Export outputs
	ctx.Export(OpVnetId, vnet.ID())
	ctx.Export(OpNodesSubnetId, subnet.ID())

	return nil
}
