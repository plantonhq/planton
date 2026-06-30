package module

import (
	"github.com/pkg/errors"
	alicloudsecuritygroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudsecuritygroup/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudsecuritygroupv1.AliCloudSecurityGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudSecurityGroup.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	sg, err := ecs.NewSecurityGroup(ctx, spec.SecurityGroupName, &ecs.SecurityGroupArgs{
		SecurityGroupName: pulumi.String(spec.SecurityGroupName),
		Description:       optionalString(spec.Description),
		VpcId:             pulumi.String(spec.VpcId.GetValue()),
		InnerAccessPolicy: pulumi.String(innerAccessPolicy(spec)),
		ResourceGroupId:   optionalString(spec.ResourceGroupId),
		Tags:              pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create security group %s", spec.SecurityGroupName)
	}

	for i, rule := range spec.Rules {
		if err := securityGroupRule(ctx, alicloudProvider, sg, spec.SecurityGroupName, i, rule); err != nil {
			return err
		}
	}

	ctx.Export(OpSecurityGroupId, sg.ID())
	ctx.Export(OpSecurityGroupName, sg.SecurityGroupName)

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
