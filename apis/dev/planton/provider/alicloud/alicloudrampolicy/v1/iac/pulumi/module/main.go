package module

import (
	"github.com/pkg/errors"
	alicloudrampolicyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudrampolicy/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ram"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudrampolicyv1.AliCloudRamPolicyStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudRamPolicy.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	policy, err := ram.NewPolicy(ctx, spec.PolicyName, &ram.PolicyArgs{
		PolicyName:     pulumi.String(spec.PolicyName),
		PolicyDocument: pulumi.String(spec.PolicyDocument),
		Description:    optionalString(spec.Description),
		RotateStrategy: optionalStringPtr(spec.RotateStrategy),
		Force:          pulumi.Bool(forceDelete(spec)),
		Tags:           pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create RAM policy %s", spec.PolicyName)
	}

	ctx.Export(OpPolicyName, policy.PolicyName)
	ctx.Export(OpPolicyType, policy.Type)

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}

func optionalStringPtr(s *string) pulumi.StringPtrInput {
	if s == nil {
		return nil
	}
	return pulumi.String(*s)
}
