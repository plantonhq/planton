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
		// Sent whenever the queue is FIFO (matching the Terraform module's
		// state-pinned rendering) — the provider default is false, so the
		// explicit value never changes cloud behavior.
		args.ContentBasedDeduplication = pulumi.BoolPtr(spec.ContentBasedDeduplication)
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

	// Redrive ALLOW policy — the permission side of the dead-letter
	// relationship: which source queues may point at THIS queue as their DLQ.
	// sourceQueueArns is only accepted alongside the byQueue mode, so it is
	// emitted conditionally (AWS rejects allowAll/denyAll documents that carry
	// the key).
	if spec.RedriveAllowPolicy != nil {
		allowPolicy := map[string]interface{}{
			"redrivePermission": spec.RedriveAllowPolicy.RedrivePermission,
		}
		if spec.RedriveAllowPolicy.RedrivePermission == "byQueue" {
			sourceArns := make([]string, 0, len(spec.RedriveAllowPolicy.SourceQueueArns))
			for _, ref := range spec.RedriveAllowPolicy.SourceQueueArns {
				sourceArns = append(sourceArns, ref.GetValue())
			}
			allowPolicy["sourceQueueArns"] = sourceArns
		}
		allowJSON, err := json.Marshal(allowPolicy)
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize redrive allow policy")
		}
		args.RedriveAllowPolicy = pulumi.String(string(allowJSON))
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
	// PRESENCE-typed: AWS enables SSE-SQS on new queues by default, so unset
	// must stay OMITTED to keep that default, while an explicit false is a
	// real configuration (an unencrypted queue) that must be SENT. The
	// provider conflicts this attribute's config PRESENCE with
	// KmsMasterKeyId — CEL blocks any manifest carrying both.
	if spec.SqsManagedSseEnabled != nil {
		args.SqsManagedSseEnabled = pulumi.BoolPtr(spec.GetSqsManagedSseEnabled())
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

	// Export outputs matching AwsSqsQueueOutputs.
	ctx.Export(OpQueueUrl, q.Url)
	ctx.Export(OpQueueArn, q.Arn)
	ctx.Export(OpQueueName, q.Name)

	return q, nil
}
