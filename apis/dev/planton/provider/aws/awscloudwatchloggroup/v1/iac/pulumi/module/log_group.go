package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type LogGroupResult struct {
	LogGroupArn  pulumi.StringOutput
	LogGroupName pulumi.StringOutput
}

func logGroup(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*LogGroupResult, error) {
	spec := locals.AwsCloudwatchLogGroup.Spec

	args := &cloudwatch.LogGroupArgs{
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Retention: 0 means never expire (the default). Only set the field when a
	// non-zero retention is configured, since Pulumi treats nil as "do not manage".
	if spec.RetentionInDays > 0 {
		args.RetentionInDays = pulumi.IntPtr(int(spec.RetentionInDays))
	}

	// KMS encryption: customer-managed key for log data at rest.
	if spec.KmsKeyId != nil {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	// Log group class: STANDARD (default), INFREQUENT_ACCESS, or DELIVERY.
	// Only set when explicitly specified; omitting lets AWS default to STANDARD.
	if spec.LogGroupClass != "" {
		args.LogGroupClass = pulumi.StringPtr(spec.LogGroupClass)
	}

	// NOTE: deletion_protection_enabled is defined in the spec but not yet
	// available in Pulumi AWS SDK v7 or TF provider 5.82.0. When the provider
	// versions are upgraded, add:
	//   if spec.DeletionProtectionEnabled {
	//       args.DeletionProtectionEnabled = pulumi.BoolPtr(true)
	//   }

	createdLogGroup, err := cloudwatch.NewLogGroup(
		ctx,
		locals.AwsCloudwatchLogGroup.Metadata.Name,
		args,
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudwatch log group")
	}

	return &LogGroupResult{
		LogGroupArn:  createdLogGroup.Arn,
		LogGroupName: createdLogGroup.Name,
	}, nil
}
