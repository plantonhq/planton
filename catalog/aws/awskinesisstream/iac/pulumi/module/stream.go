package module

import (
	"encoding/json"

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

	// Warm throughput -- pre-provisioned burst capacity for ON_DEMAND
	// streams; mutually exclusive with shard_count (spec CEL mirrors the
	// provider's own rule).
	if spec.WarmThroughputMibPs > 0 {
		args.WarmThroughputMibPs = pulumi.IntPtr(int(spec.WarmThroughputMibPs))
	}

	// -------------------------------------------------------------------
	// Data retention (only set when non-zero to let AWS use defaults)
	// -------------------------------------------------------------------

	if spec.RetentionPeriodHours != 0 {
		args.RetentionPeriod = pulumi.IntPtr(int(spec.RetentionPeriodHours))
	}

	// -------------------------------------------------------------------
	// Encryption -- presence of kms_key_id implies KMS encryption.
	// encryption_type is ALWAYS sent ("KMS" or "NONE"), matching the
	// Terraform module's derived always-send rendering check-for-check.
	// -------------------------------------------------------------------

	if spec.KmsKeyId.GetValue() != "" {
		args.EncryptionType = pulumi.StringPtr("KMS")
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	} else {
		args.EncryptionType = pulumi.StringPtr("NONE")
	}

	// -------------------------------------------------------------------
	// Max record size -- Kinesis 10 MiB large-record support. In-place
	// update; needs the kinesis:UpdateMaxRecordSize IAM permission on the
	// deploying principal.
	// -------------------------------------------------------------------

	if spec.MaxRecordSizeInKib > 0 {
		args.MaxRecordSizeInKib = pulumi.IntPtr(int(spec.MaxRecordSizeInKib))
	}

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
	// Deletion behavior -- always sent (state-pinned like the Terraform
	// module), so a true -> false edit is applied instead of silently
	// leaving the old value in state.
	// -------------------------------------------------------------------

	args.EnforceConsumerDeletion = pulumi.Bool(spec.EnforceConsumerDeletion)

	// -------------------------------------------------------------------
	// Create stream
	// -------------------------------------------------------------------

	s, err := kinesis.NewStream(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Kinesis stream")
	}

	// Resource-based access policy -- AWS models this as a separate API
	// keyed by the stream ARN (one policy per stream), folded into the spec
	// because it has no identity of its own. The primary use is
	// cross-account producer/consumer grants without role assumption.
	if spec.ResourcePolicy != nil {
		policyJSON, err := json.Marshal(spec.ResourcePolicy.AsMap())
		if err != nil {
			return errors.Wrap(err, "failed to serialize resource policy")
		}
		_, err = kinesis.NewResourcePolicy(ctx, "resource-policy", &kinesis.ResourcePolicyArgs{
			ResourceArn: s.Arn,
			Policy:      pulumi.String(policyJSON),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create Kinesis resource policy")
		}
	}

	// Export outputs matching AwsKinesisStreamOutputs.
	ctx.Export(OpStreamArn, s.Arn)
	ctx.Export(OpStreamName, s.Name)

	return nil
}
