package module

import (
	"github.com/pkg/errors"
	alicloudvpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudvpc/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudvpcv1.AliCloudVpcStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudVpc.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	vpcNetwork, err := vpc.NewNetwork(ctx, spec.VpcName, &vpc.NetworkArgs{
		VpcName:         pulumi.String(spec.VpcName),
		CidrBlock:       pulumi.String(spec.CidrBlock),
		Description:     optionalString(spec.Description),
		EnableIpv6:      pulumi.Bool(spec.EnableIpv6),
		ResourceGroupId: optionalString(spec.ResourceGroupId),
		Tags:            pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create VPC %s", spec.VpcName)
	}

	ctx.Export(OpVpcId, vpcNetwork.ID())
	ctx.Export(OpVpcName, vpcNetwork.VpcName)
	ctx.Export(OpCidrBlock, vpcNetwork.CidrBlock)
	ctx.Export(OpRouterId, vpcNetwork.RouterId)
	ctx.Export(OpRouteTableId, vpcNetwork.RouteTableId)

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
