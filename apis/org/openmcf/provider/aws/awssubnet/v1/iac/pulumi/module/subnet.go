package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func subnet(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) (*ec2.Subnet, error) {
	spec := locals.AwsSubnet.Spec
	name := locals.AwsSubnet.Metadata.Name

	subnetArgs := &ec2.SubnetArgs{
		VpcId:                                   pulumi.String(spec.VpcId.GetValue()),
		CidrBlock:                               pulumi.StringPtr(spec.CidrBlock),
		AvailabilityZone:                        pulumi.StringPtr(spec.AvailabilityZone),
		MapPublicIpOnLaunch:                     pulumi.BoolPtr(spec.MapPublicIpOnLaunch),
		AssignIpv6AddressOnCreation:             pulumi.BoolPtr(spec.AssignIpv6AddressOnCreation),
		EnableDns64:                             pulumi.BoolPtr(spec.EnableDns64),
		EnableResourceNameDnsARecordOnLaunch:    pulumi.BoolPtr(spec.EnableResourceNameDnsARecordOnLaunch),
		EnableResourceNameDnsAaaaRecordOnLaunch: pulumi.BoolPtr(spec.EnableResourceNameDnsAaaaRecordOnLaunch),
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", name)),
	}

	if spec.Ipv6CidrBlock != "" {
		subnetArgs.Ipv6CidrBlock = pulumi.StringPtr(spec.Ipv6CidrBlock)
	}
	if spec.GetPrivateDnsHostnameTypeOnLaunch() != "" {
		subnetArgs.PrivateDnsHostnameTypeOnLaunch = pulumi.StringPtr(spec.GetPrivateDnsHostnameTypeOnLaunch())
	}

	createdSubnet, err := ec2.NewSubnet(ctx, name, subnetArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create subnet")
	}

	ctx.Export(OpSubnetId, createdSubnet.ID())
	ctx.Export(OpSubnetArn, createdSubnet.Arn)
	ctx.Export(OpAvailabilityZone, createdSubnet.AvailabilityZone)
	ctx.Export(OpCidrBlock, createdSubnet.CidrBlock)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return createdSubnet, nil
}
