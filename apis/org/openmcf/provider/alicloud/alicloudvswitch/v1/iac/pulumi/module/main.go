package module

import (
	"github.com/pkg/errors"
	alicloudvswitchv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudvswitch/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudvswitchv1.AlicloudVswitchStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudVswitch.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	switchArgs := &vpc.SwitchArgs{
		VpcId:       pulumi.String(locals.VpcId),
		ZoneId:      pulumi.String(spec.ZoneId),
		CidrBlock:   pulumi.String(spec.CidrBlock),
		VswitchName: pulumi.String(spec.VswitchName),
		Description: optionalString(spec.Description),
		EnableIpv6:  pulumi.Bool(spec.EnableIpv6),
		Tags:        pulumi.ToStringMap(locals.Tags),
	}

	if spec.Ipv6CidrBlockMask != 0 {
		switchArgs.Ipv6CidrBlockMask = pulumi.Int(int(spec.Ipv6CidrBlockMask))
	}

	vswitch, err := vpc.NewSwitch(ctx, spec.VswitchName, switchArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create VSwitch %s", spec.VswitchName)
	}

	ctx.Export(OpVswitchId, vswitch.ID())
	ctx.Export(OpVswitchName, vswitch.VswitchName)
	ctx.Export(OpCidrBlock, vswitch.CidrBlock)
	ctx.Export(OpZoneId, vswitch.ZoneId)
	ctx.Export(OpIpv6CidrBlock, vswitch.Ipv6CidrBlock)

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
