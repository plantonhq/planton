package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func topic(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*sns.Topic, error) {
	spec := locals.Spec

	args := &sns.TopicArgs{
		Name:      pulumi.StringPtr(locals.TopicName),
		FifoTopic: pulumi.BoolPtr(spec.FifoTopic),
		Tags:      pulumi.ToStringMap(locals.AwsTags),
	}

	// -------------------------------------------------------------------
	// FIFO-specific settings
	// -------------------------------------------------------------------

	if spec.FifoTopic {
		// Sent whenever the topic is FIFO (matching the Terraform module's
		// state-pinned rendering) — the provider default is false, so the
		// explicit value never changes cloud behavior.
		args.ContentBasedDeduplication = pulumi.BoolPtr(spec.ContentBasedDeduplication)
		if spec.FifoThroughputScope != "" {
			args.FifoThroughputScope = pulumi.StringPtr(spec.FifoThroughputScope)
		}
		// Message archiving (FIFO-only). Subscriptions opt into replay
		// individually via their own replay_policy; the topic only defines
		// the retention window.
		if spec.ArchivePolicy != nil {
			archiveJSON, err := json.Marshal(spec.ArchivePolicy.AsMap())
			if err != nil {
				return nil, errors.Wrap(err, "failed to serialize archive policy")
			}
			args.ArchivePolicy = pulumi.StringPtr(string(archiveJSON))
		}
	}

	// -------------------------------------------------------------------
	// Display name
	// -------------------------------------------------------------------

	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}

	// -------------------------------------------------------------------
	// Encryption — customer-managed KMS only (SNS has no managed-SSE option)
	// -------------------------------------------------------------------

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsMasterKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	// -------------------------------------------------------------------
	// Access policy. AWS always keeps a policy on a topic: unset here means
	// the AWS default owner-only policy applies.
	// -------------------------------------------------------------------

	if spec.Policy != nil {
		policyJSON, err := json.Marshal(spec.Policy.AsMap())
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize access policy")
		}
		args.Policy = pulumi.String(string(policyJSON))
	}

	// -------------------------------------------------------------------
	// Delivery policy (HTTP/S retry behavior)
	// -------------------------------------------------------------------

	if spec.DeliveryPolicy != "" {
		args.DeliveryPolicy = pulumi.StringPtr(spec.DeliveryPolicy)
	}

	// -------------------------------------------------------------------
	// Per-protocol delivery-status logging. Failure logging has no sample
	// rate — failures are always logged when the failure role is set.
	// -------------------------------------------------------------------

	if fb := spec.DeliveryFeedback; fb != nil {
		if p := fb.Application; p != nil {
			args.ApplicationSuccessFeedbackRoleArn = feedbackRole(p.SuccessFeedbackRole.GetValue())
			args.ApplicationFailureFeedbackRoleArn = feedbackRole(p.FailureFeedbackRole.GetValue())
			args.ApplicationSuccessFeedbackSampleRate = feedbackSampleRate(p.SuccessFeedbackSampleRate)
		}
		if p := fb.Firehose; p != nil {
			args.FirehoseSuccessFeedbackRoleArn = feedbackRole(p.SuccessFeedbackRole.GetValue())
			args.FirehoseFailureFeedbackRoleArn = feedbackRole(p.FailureFeedbackRole.GetValue())
			args.FirehoseSuccessFeedbackSampleRate = feedbackSampleRate(p.SuccessFeedbackSampleRate)
		}
		if p := fb.Http; p != nil {
			args.HttpSuccessFeedbackRoleArn = feedbackRole(p.SuccessFeedbackRole.GetValue())
			args.HttpFailureFeedbackRoleArn = feedbackRole(p.FailureFeedbackRole.GetValue())
			args.HttpSuccessFeedbackSampleRate = feedbackSampleRate(p.SuccessFeedbackSampleRate)
		}
		if p := fb.Lambda; p != nil {
			args.LambdaSuccessFeedbackRoleArn = feedbackRole(p.SuccessFeedbackRole.GetValue())
			args.LambdaFailureFeedbackRoleArn = feedbackRole(p.FailureFeedbackRole.GetValue())
			args.LambdaSuccessFeedbackSampleRate = feedbackSampleRate(p.SuccessFeedbackSampleRate)
		}
		if p := fb.Sqs; p != nil {
			args.SqsSuccessFeedbackRoleArn = feedbackRole(p.SuccessFeedbackRole.GetValue())
			args.SqsFailureFeedbackRoleArn = feedbackRole(p.FailureFeedbackRole.GetValue())
			args.SqsSuccessFeedbackSampleRate = feedbackSampleRate(p.SuccessFeedbackSampleRate)
		}
	}

	// -------------------------------------------------------------------
	// Observability
	// -------------------------------------------------------------------

	if spec.TracingConfig != "" {
		args.TracingConfig = pulumi.StringPtr(spec.TracingConfig)
	}

	if spec.SignatureVersion != 0 {
		args.SignatureVersion = pulumi.IntPtr(int(spec.SignatureVersion))
	}

	// -------------------------------------------------------------------
	// Create topic
	// -------------------------------------------------------------------

	t, err := sns.NewTopic(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create SNS topic")
	}

	// Data protection policy is a single-per-topic satellite setting (PII/PHI
	// detection with audit/mask/deny operations) keyed by the topic ARN. AWS
	// models it as its own API call, so it materializes as its own resource.
	if spec.DataProtectionPolicy != nil {
		dppJSON, err := json.Marshal(spec.DataProtectionPolicy.AsMap())
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize data protection policy")
		}
		_, err = sns.NewDataProtectionPolicy(ctx, locals.Target.Metadata.Name, &sns.DataProtectionPolicyArgs{
			Arn:    t.Arn,
			Policy: pulumi.String(string(dppJSON)),
		}, pulumi.Provider(provider), pulumi.Parent(t))
		if err != nil {
			return nil, errors.Wrap(err, "failed to create data protection policy")
		}
	}

	// Export outputs matching AwsSnsTopicOutputs.
	ctx.Export(OpTopicArn, t.Arn)
	ctx.Export(OpTopicName, t.Name)
	ctx.Export(OpOwner, t.Owner)
	ctx.Export(OpBeginningArchiveTime, t.BeginningArchiveTime)

	return t, nil
}

// feedbackRole converts an optional role ARN to a Pulumi input, leaving the
// provider attribute unset for empty values.
func feedbackRole(arn string) pulumi.StringPtrInput {
	if arn == "" {
		return nil
	}
	return pulumi.StringPtr(arn)
}

// feedbackSampleRate converts the spec's 0-means-unset sample rate to a Pulumi
// input so AWS applies its own default instead of freezing "log nothing".
func feedbackSampleRate(rate int32) pulumi.IntPtrInput {
	if rate == 0 {
		return nil
	}
	return pulumi.IntPtr(int(rate))
}
