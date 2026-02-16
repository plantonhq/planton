package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sqs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func queue(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*sqs.Queue, error) {
	spec := locals.Spec

	args := &sqs.QueueArgs{
		Name:      pulumi.StringPtr(locals.QueueName),
		FifoQueue: pulumi.BoolPtr(spec.FifoQueue),
		Tags:      pulumi.ToStringMap(locals.AwsTags),
	}

	// -------------------------------------------------------------------
	// Delivery settings (only set when non-zero to let AWS use defaults)
	// -------------------------------------------------------------------

	if spec.VisibilityTimeoutSeconds != 0 {
		args.VisibilityTimeoutSeconds = pulumi.IntPtr(int(spec.VisibilityTimeoutSeconds))
	}
	if spec.MessageRetentionSeconds != 0 {
		args.MessageRetentionSeconds = pulumi.IntPtr(int(spec.MessageRetentionSeconds))
	}
	if spec.MaxMessageSizeBytes != 0 {
		args.MaxMessageSize = pulumi.IntPtr(int(spec.MaxMessageSizeBytes))
	}
	if spec.DelaySeconds != 0 {
		args.DelaySeconds = pulumi.IntPtr(int(spec.DelaySeconds))
	}
	if spec.ReceiveWaitTimeSeconds != 0 {
		args.ReceiveWaitTimeSeconds = pulumi.IntPtr(int(spec.ReceiveWaitTimeSeconds))
	}

	// -------------------------------------------------------------------
	// FIFO-specific settings
	// -------------------------------------------------------------------

	if spec.FifoQueue {
		if spec.ContentBasedDeduplication {
			args.ContentBasedDeduplication = pulumi.BoolPtr(true)
		}
		if spec.DeduplicationScope != "" {
			args.DeduplicationScope = pulumi.StringPtr(spec.DeduplicationScope)
		}
		if spec.FifoThroughputLimit != "" {
			args.FifoThroughputLimit = pulumi.StringPtr(spec.FifoThroughputLimit)
		}
	}

	// -------------------------------------------------------------------
	// Dead letter queue (redrive policy)
	// -------------------------------------------------------------------

	if spec.DeadLetterConfig != nil {
		redrivePolicy := map[string]interface{}{
			"deadLetterTargetArn": spec.DeadLetterConfig.TargetArn.GetValue(),
			"maxReceiveCount":     spec.DeadLetterConfig.MaxReceiveCount,
		}
		policyJSON, err := json.Marshal(redrivePolicy)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize redrive policy")
		}
		args.RedrivePolicy = pulumi.String(string(policyJSON))
	}

	// -------------------------------------------------------------------
	// Encryption
	// -------------------------------------------------------------------

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsMasterKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}
	if spec.KmsDataKeyReusePeriodSeconds != 0 {
		args.KmsDataKeyReusePeriodSeconds = pulumi.IntPtr(int(spec.KmsDataKeyReusePeriodSeconds))
	}
	if spec.SqsManagedSseEnabled {
		args.SqsManagedSseEnabled = pulumi.BoolPtr(true)
	}

	// -------------------------------------------------------------------
	// Access policy
	// -------------------------------------------------------------------

	if spec.Policy != nil {
		policyMap := spec.Policy.AsMap()
		policyJSON, err := json.Marshal(policyMap)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize access policy")
		}
		args.Policy = pulumi.String(string(policyJSON))
	}

	// -------------------------------------------------------------------
	// Create queue
	// -------------------------------------------------------------------

	q, err := sqs.NewQueue(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create SQS queue")
	}

	// Export outputs matching AwsSqsQueueStackOutputs.
	ctx.Export(OpQueueUrl, q.Url)
	ctx.Export(OpQueueArn, q.Arn)
	ctx.Export(OpQueueName, q.Name)

	return q, nil
}
