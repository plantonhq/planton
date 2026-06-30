package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurenatgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurenatgateway/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurenatgatewayv1.AzureNatGatewayStackInput) error {
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

	spec := locals.AzureNatGateway.Spec

	var publicIpIds pulumi.StringArrayInput
	var publicIpAddresses pulumi.StringArrayOutput
	var publicIpPrefixId pulumi.StringOutput

	// Create either Public IP Prefix or individual Public IP based on spec
	if spec.PublicIpPrefixLength != nil && *spec.PublicIpPrefixLength > 0 {
		// Create Public IP Prefix for scale
		publicIpPrefix, err := network.NewPublicIpPrefix(ctx,
			fmt.Sprintf("%s-prefix", locals.NatGatewayName),
			&network.PublicIpPrefixArgs{
				Name:              pulumi.String(fmt.Sprintf("%s-prefix", locals.NatGatewayName)),
				ResourceGroupName: pulumi.String(locals.ResourceGroup),
				Location:          pulumi.String(locals.Location),
				PrefixLength:      pulumi.Int(int(*spec.PublicIpPrefixLength)),
				Sku:               pulumi.String("Standard"),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create public IP prefix")
		}

		publicIpPrefixId = publicIpPrefix.ID().ToStringOutput()
		// NAT Gateway uses the prefix, no individual IPs needed
		publicIpIds = pulumi.StringArray{}
	} else {
		// Create a single Standard Public IP
		publicIp, err := network.NewPublicIp(ctx,
			fmt.Sprintf("%s-ip", locals.NatGatewayName),
			&network.PublicIpArgs{
				Name:              pulumi.String(fmt.Sprintf("%s-ip", locals.NatGatewayName)),
				ResourceGroupName: pulumi.String(locals.ResourceGroup),
				Location:          pulumi.String(locals.Location),
				AllocationMethod:  pulumi.String("Static"),
				Sku:               pulumi.String("Standard"),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrap(err, "failed to create public IP")
		}

		publicIpIds = pulumi.StringArray{publicIp.ID()}
		publicIpAddresses = pulumi.StringArray{publicIp.IpAddress}.ToStringArrayOutput()
	}

	// Create NAT Gateway
	natGateway, err := network.NewNatGateway(ctx,
		locals.NatGatewayName,
		&network.NatGatewayArgs{
			Name:                 pulumi.String(locals.NatGatewayName),
			ResourceGroupName:    pulumi.String(locals.ResourceGroup),
			Location:             pulumi.String(locals.Location),
			IdleTimeoutInMinutes: pulumi.Int(getIdleTimeoutMinutes(spec)),
			SkuName:              pulumi.String("Standard"),
			Tags:                 pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create NAT gateway")
	}

	// Associate Public IP(s) or Prefix with NAT Gateway
	if spec.PublicIpPrefixLength != nil && *spec.PublicIpPrefixLength > 0 {
		_, err = network.NewNatGatewayPublicIpPrefixAssociation(ctx,
			fmt.Sprintf("%s-prefix-assoc", locals.NatGatewayName),
			&network.NatGatewayPublicIpPrefixAssociationArgs{
				NatGatewayId:     natGateway.ID(),
				PublicIpPrefixId: publicIpPrefixId,
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(natGateway))
		if err != nil {
			return errors.Wrap(err, "failed to associate public IP prefix with NAT gateway")
		}
	} else {
		// Associate individual public IP
		_, err = network.NewNatGatewayPublicIpAssociation(ctx,
			fmt.Sprintf("%s-ip-assoc", locals.NatGatewayName),
			&network.NatGatewayPublicIpAssociationArgs{
				NatGatewayId:      natGateway.ID(),
				PublicIpAddressId: publicIpIds.ToStringArrayOutput().Index(pulumi.Int(0)),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(natGateway))
		if err != nil {
			return errors.Wrap(err, "failed to associate public IP with NAT gateway")
		}
	}

	// Associate NAT Gateway with Subnet
	_, err = network.NewSubnetNatGatewayAssociation(ctx,
		fmt.Sprintf("%s-subnet-assoc", locals.NatGatewayName),
		&network.SubnetNatGatewayAssociationArgs{
			SubnetId:     pulumi.String(locals.SubnetId),
			NatGatewayId: natGateway.ID(),
		},
		pulumi.Provider(azureProvider),
		pulumi.Parent(natGateway))
	if err != nil {
		return errors.Wrap(err, "failed to associate NAT gateway with subnet")
	}

	// Export outputs
	ctx.Export(OpNatGatewayId, natGateway.ID())

	if spec.PublicIpPrefixLength != nil && *spec.PublicIpPrefixLength > 0 {
		ctx.Export(OpPublicIpPrefixId, publicIpPrefixId)
		ctx.Export(OpPublicIpAddresses, pulumi.StringArray{})
	} else {
		ctx.Export(OpPublicIpAddresses, publicIpAddresses)
		ctx.Export(OpPublicIpPrefixId, pulumi.String(""))
	}

	return nil
}
