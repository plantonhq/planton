package module

import (
	"github.com/pkg/errors"
	alicloudkmskeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudkmskey/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/kms"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudkmskeyv1.AliCloudKmsKeyStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudKmsKey.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	resourceName := stackInput.Target.Metadata.Name

	key, err := kms.NewKey(ctx, resourceName, &kms.KeyArgs{
		Description:                   optionalString(spec.Description),
		KeySpec:                       pulumi.String(keySpec(spec)),
		KeyUsage:                      pulumi.String(keyUsage(spec)),
		ProtectionLevel:               pulumi.String(protectionLevel(spec)),
		AutomaticRotation:             pulumi.String(automaticRotation(spec)),
		RotationInterval:              optionalString(spec.RotationInterval),
		PendingWindowInDays:           pulumi.Int(pendingWindowInDays(spec)),
		DeletionProtection:            pulumi.String(deletionProtection(spec)),
		DeletionProtectionDescription: optionalString(spec.DeletionProtectionDescription),
		Tags:                          pulumi.ToStringMap(locals.Tags),
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create KMS key %s", resourceName)
	}

	ctx.Export(OpKeyId, key.ID())
	ctx.Export(OpArn, key.Arn)

	return nil
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
