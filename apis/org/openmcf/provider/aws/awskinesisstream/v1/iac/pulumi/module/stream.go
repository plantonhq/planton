package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func stream(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &kinesis.StreamArgs{
		Name: pulumi.StringPtr(locals.StreamName),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// -------------------------------------------------------------------
	// Capacity mode
	// -------------------------------------------------------------------

	args.StreamModeDetails = &kinesis.StreamStreamModeDetailsArgs{
		StreamMode: pulumi.String(spec.StreamMode),
	}

	if spec.StreamMode == "PROVISIONED" && spec.ShardCount > 0 {
		args.ShardCount = pulumi.IntPtr(int(spec.ShardCount))
	}

	// -------------------------------------------------------------------
	// Data retention (only set when non-zero to let AWS use defaults)
	// -------------------------------------------------------------------

	if spec.RetentionPeriodHours != 0 {
		args.RetentionPeriod = pulumi.IntPtr(int(spec.RetentionPeriodHours))
	}

	// -------------------------------------------------------------------
	// Encryption -- presence of kms_key_id implies KMS encryption
	// -------------------------------------------------------------------

	if spec.KmsKeyId.GetValue() != "" {
		args.EncryptionType = pulumi.StringPtr("KMS")
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	// -------------------------------------------------------------------
	// Max record size
	//
	// DEFERRED: max_record_size_in_kib is defined in the spec but is not
	// available in the pinned Pulumi AWS SDK v7.3.0. This field was added
	// in a newer SDK version. The spec retains the field for forward
	// compatibility; it will be wired when the SDK dependency is upgraded.
	// -------------------------------------------------------------------

	// -------------------------------------------------------------------
	// Enhanced shard-level monitoring
	// -------------------------------------------------------------------

	if len(spec.ShardLevelMetrics) > 0 {
		metrics := make(pulumi.StringArray, len(spec.ShardLevelMetrics))
		for i, m := range spec.ShardLevelMetrics {
			metrics[i] = pulumi.String(m)
		}
		args.ShardLevelMetrics = metrics
	}

	// -------------------------------------------------------------------
	// Deletion behavior
	// -------------------------------------------------------------------

	if spec.EnforceConsumerDeletion {
		args.EnforceConsumerDeletion = pulumi.BoolPtr(true)
	}

	// -------------------------------------------------------------------
	// Create stream
	// -------------------------------------------------------------------

	s, err := kinesis.NewStream(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Kinesis stream")
	}

	// Export outputs matching AwsKinesisStreamStackOutputs.
	ctx.Export(OpStreamArn, s.Arn)
	ctx.Export(OpStreamName, s.Name)

	return nil
}
