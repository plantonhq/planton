package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudramrolev1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudramrole/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ram"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudramrolev1.AliCloudRamRoleStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudRamRole.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	role, err := ram.NewRole(ctx, spec.RoleName, &ram.RoleArgs{
		RoleName:                 pulumi.String(spec.RoleName),
		AssumeRolePolicyDocument: pulumi.String(spec.AssumeRolePolicyDocument),
		Description:              optionalString(spec.Description),
		MaxSessionDuration:       pulumi.Int(maxSessionDuration(spec)),
		Force:                    pulumi.Bool(forceDelete(spec)),
		Tags:                     pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create RAM role %s", spec.RoleName)
	}

	for _, pa := range spec.PolicyAttachments {
		if err := policyAttachment(ctx, alicloudProvider, role, spec.RoleName, pa); err != nil {
			return err
		}
	}

	ctx.Export(OpRoleId, role.RoleId)
	ctx.Export(OpRoleName, role.RoleName)
	ctx.Export(OpArn, role.Arn)

	return nil
}

func policyAttachment(
	ctx *pulumi.Context,
	provider *alicloud.Provider,
	role *ram.Role,
	roleName string,
	pa *alicloudramrolev1.AliCloudRamRolePolicyAttachment,
) error {
	pt := policyType(pa)
	resourceName := fmt.Sprintf("%s-%s-%s", roleName, pa.PolicyName, pt)

	_, err := ram.NewRolePolicyAttachment(ctx, resourceName, &ram.RolePolicyAttachmentArgs{
		RoleName:   pulumi.String(roleName),
		PolicyName: pulumi.String(pa.PolicyName),
		PolicyType: pulumi.String(pt),
	}, pulumi.Provider(provider), pulumi.Parent(role))
	if err != nil {
		return errors.Wrapf(err, "failed to attach policy %s to role %s", pa.PolicyName, roleName)
	}

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
