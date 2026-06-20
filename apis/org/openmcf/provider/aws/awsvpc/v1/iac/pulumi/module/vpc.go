package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func vpc(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) error {
	spec := locals.AwsVpc.Spec
	name := locals.AwsVpc.Metadata.Name

	vpcArgs := &ec2.VpcArgs{
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", name)),
		EnableDnsHostnames:               pulumi.BoolPtr(spec.EnableDnsHostnames),
		EnableNetworkAddressUsageMetrics: pulumi.BoolPtr(spec.EnableNetworkAddressUsageMetrics),
	}

	if spec.CidrBlock != "" {
		vpcArgs.CidrBlock = pulumi.StringPtr(spec.CidrBlock)
	}
	if spec.Ipv4IpamPoolId != "" {
		vpcArgs.Ipv4IpamPoolId = pulumi.StringPtr(spec.Ipv4IpamPoolId)
	}
	if spec.Ipv4NetmaskLength != 0 {
		vpcArgs.Ipv4NetmaskLength = pulumi.IntPtr(int(spec.Ipv4NetmaskLength))
	}
	if spec.InstanceTenancy != "" {
		vpcArgs.InstanceTenancy = pulumi.StringPtr(spec.InstanceTenancy)
	}
	// enable_dns_support is proto3 optional: honor an explicit value, otherwise
	// leave the argument unset so AWS applies its default (DNS support on).
	if spec.EnableDnsSupport != nil {
		vpcArgs.EnableDnsSupport = pulumi.BoolPtr(spec.GetEnableDnsSupport())
	}

	if spec.AssignGeneratedIpv6CidrBlock {
		vpcArgs.AssignGeneratedIpv6CidrBlock = pulumi.BoolPtr(true)
	}
	if spec.Ipv6CidrBlock != "" {
		vpcArgs.Ipv6CidrBlock = pulumi.StringPtr(spec.Ipv6CidrBlock)
	}
	if spec.Ipv6CidrBlockNetworkBorderGroup != "" {
		vpcArgs.Ipv6CidrBlockNetworkBorderGroup = pulumi.StringPtr(spec.Ipv6CidrBlockNetworkBorderGroup)
	}
	if spec.Ipv6IpamPoolId != "" {
		vpcArgs.Ipv6IpamPoolId = pulumi.StringPtr(spec.Ipv6IpamPoolId)
	}
	if spec.Ipv6NetmaskLength != 0 {
		vpcArgs.Ipv6NetmaskLength = pulumi.IntPtr(int(spec.Ipv6NetmaskLength))
	}

	createdVpc, err := ec2.NewVpc(ctx, name, vpcArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create vpc")
	}

	// Associate each secondary IPv4 CIDR as its own resource so it can be added
	// or removed without recreating the VPC.
	for i, secondaryCidr := range spec.SecondaryIpv4CidrBlocks {
		_, err := ec2.NewVpcIpv4CidrBlockAssociation(ctx,
			fmt.Sprintf("%s-secondary-%d", name, i),
			&ec2.VpcIpv4CidrBlockAssociationArgs{
				VpcId:     createdVpc.ID(),
				CidrBlock: pulumi.String(secondaryCidr),
			}, pulumi.Provider(provider), pulumi.Parent(createdVpc))
		if err != nil {
			return errors.Wrapf(err, "failed to associate secondary cidr %s", secondaryCidr)
		}
	}

	ctx.Export(OpVpcId, createdVpc.ID())
	ctx.Export(OpVpcArn, createdVpc.Arn)
	ctx.Export(OpCidrBlock, createdVpc.CidrBlock)
	ctx.Export(OpIpv6CidrBlock, createdVpc.Ipv6CidrBlock)
	ctx.Export(OpOwnerId, createdVpc.OwnerId)
	ctx.Export(OpMainRouteTableId, createdVpc.MainRouteTableId)
	ctx.Export(OpDefaultSecurityGroupId, createdVpc.DefaultSecurityGroupId)
	ctx.Export(OpDefaultNetworkAclId, createdVpc.DefaultNetworkAclId)
	ctx.Export(OpDefaultRouteTableId, createdVpc.DefaultRouteTableId)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
