package module

import (
	"github.com/pkg/errors"
	awssubnetv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awssubnet/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// configureRouting attaches a route table to the subnet. Inline routes create a
// dedicated, subnet-owned table; route_table_id adopts an external one. When
// neither is set, the subnet stays on the VPC main route table and route_table_id
// is exported empty.
func configureRouting(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, createdSubnet *ec2.Subnet) error {
	spec := locals.AwsSubnet.Spec

	if spec.RouteTableId == nil && len(spec.Routes) == 0 {
		ctx.Export(OpRouteTableId, pulumi.String(""))
		return nil
	}

	var routeTableId pulumi.StringInput
	if len(spec.Routes) > 0 {
		createdRouteTable, err := routeTable(ctx, locals, provider)
		if err != nil {
			return err
		}
		routeTableId = createdRouteTable.ID().ToStringOutput()
	} else {
		routeTableId = pulumi.String(spec.RouteTableId.GetValue())
	}

	_, err := ec2.NewRouteTableAssociation(ctx, locals.AwsSubnet.Metadata.Name, &ec2.RouteTableAssociationArgs{
		SubnetId:     createdSubnet.ID(),
		RouteTableId: routeTableId,
	}, pulumi.Provider(provider), pulumi.Parent(createdSubnet))
	if err != nil {
		return errors.Wrap(err, "failed to associate route table with subnet")
	}

	ctx.Export(OpRouteTableId, routeTableId)
	return nil
}

func routeTable(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*ec2.RouteTable, error) {
	spec := locals.AwsSubnet.Spec
	name := locals.AwsSubnet.Metadata.Name

	routes := make(ec2.RouteTableRouteArray, len(spec.Routes))
	for i, route := range spec.Routes {
		routes[i] = routeArgs(route)
	}

	createdRouteTable, err := ec2.NewRouteTable(ctx, name, &ec2.RouteTableArgs{
		VpcId:  pulumi.String(spec.VpcId.GetValue()),
		Routes: routes,
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", name)),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create route table")
	}

	return createdRouteTable, nil
}

// routeArgs maps one spec route onto the AWS route attributes: the destination
// fills the matching CIDR/prefix-list field, and target_type selects which target
// attribute target_id is assigned to.
func routeArgs(route *awssubnetv1.AwsSubnetSpec_AwsSubnetRoute) ec2.RouteTableRouteArgs {
	args := ec2.RouteTableRouteArgs{}

	switch {
	case route.DestinationCidrBlock != "":
		args.CidrBlock = pulumi.StringPtr(route.DestinationCidrBlock)
	case route.DestinationIpv6CidrBlock != "":
		args.Ipv6CidrBlock = pulumi.StringPtr(route.DestinationIpv6CidrBlock)
	case route.DestinationPrefixListId != "":
		args.DestinationPrefixListId = pulumi.StringPtr(route.DestinationPrefixListId)
	}

	targetID := route.TargetId.GetValue()
	switch route.TargetType {
	case awssubnetv1.AwsSubnetSpec_AwsSubnetRoute_internet_gateway:
		args.GatewayId = pulumi.StringPtr(targetID)
	case awssubnetv1.AwsSubnetSpec_AwsSubnetRoute_nat_gateway:
		args.NatGatewayId = pulumi.StringPtr(targetID)
	case awssubnetv1.AwsSubnetSpec_AwsSubnetRoute_transit_gateway:
		args.TransitGatewayId = pulumi.StringPtr(targetID)
	case awssubnetv1.AwsSubnetSpec_AwsSubnetRoute_vpc_peering_connection:
		args.VpcPeeringConnectionId = pulumi.StringPtr(targetID)
	case awssubnetv1.AwsSubnetSpec_AwsSubnetRoute_vpc_endpoint:
		args.VpcEndpointId = pulumi.StringPtr(targetID)
	case awssubnetv1.AwsSubnetSpec_AwsSubnetRoute_network_interface:
		args.NetworkInterfaceId = pulumi.StringPtr(targetID)
	case awssubnetv1.AwsSubnetSpec_AwsSubnetRoute_egress_only_internet_gateway:
		args.EgressOnlyGatewayId = pulumi.StringPtr(targetID)
	}

	return args
}
