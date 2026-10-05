package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func subscription(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*sns.TopicSubscription, error) {
	spec := locals.Spec

	args := &sns.TopicSubscriptionArgs{
		// The immutable identity trio: changing any of these replaces the
		// subscription (AWS has no repoint operation for topic/protocol/endpoint).
		Topic:    pulumi.String(spec.TopicArn.GetValue()),
		Protocol: pulumi.String(spec.Protocol),
		Endpoint: pulumi.String(spec.Endpoint.GetValue()),
	}

	// -------------------------------------------------------------------
	// Message filtering. filter_policy_scope is only sent alongside a
	// filter policy — AWS rejects a scope without a policy (CEL blocks the
	// manifest shape; the guard here keeps the module safe regardless).
	// -------------------------------------------------------------------

	if spec.FilterPolicy != nil {
		filterJSON, err := json.Marshal(spec.FilterPolicy.AsMap())
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize filter policy")
		}
		args.FilterPolicy = pulumi.StringPtr(string(filterJSON))
		if spec.FilterPolicyScope != "" {
			args.FilterPolicyScope = pulumi.StringPtr(spec.FilterPolicyScope)
		}
	}

	// -------------------------------------------------------------------
	// Delivery behavior
	// -------------------------------------------------------------------

	// Always sent (matching the Terraform module's state-pinned rendering) —
	// the provider default is false, so the explicit value never changes
	// cloud behavior.
	args.RawMessageDelivery = pulumi.BoolPtr(spec.RawMessageDelivery)

	// AWS models the subscription DLQ as a JSON redrive policy document
	// holding just the target ARN (retry exhaustion is governed by the
	// delivery policy, not a receive count).
	if spec.DeadLetterConfig != nil {
		redriveJSON, err := json.Marshal(map[string]interface{}{
			"deadLetterTargetArn": spec.DeadLetterConfig.DeadLetterTargetArn.GetValue(),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize redrive policy")
		}
		args.RedrivePolicy = pulumi.StringPtr(string(redriveJSON))
	}

	if spec.DeliveryPolicy != "" {
		args.DeliveryPolicy = pulumi.StringPtr(spec.DeliveryPolicy)
	}

	// Replay of archived messages (FIFO topics with an archive policy) — the
	// backfill mechanism for a consumer added after messages were published.
	if spec.ReplayPolicy != nil {
		replayJSON, err := json.Marshal(spec.ReplayPolicy.AsMap())
		if err != nil {
			return nil, errors.Wrap(err, "failed to serialize replay policy")
		}
		args.ReplayPolicy = pulumi.StringPtr(string(replayJSON))
	}

	// Firehose delivery requires the role SNS assumes to write to the stream.
	if spec.SubscriptionRoleArn.GetValue() != "" {
		args.SubscriptionRoleArn = pulumi.StringPtr(spec.SubscriptionRoleArn.GetValue())
	}

	// -------------------------------------------------------------------
	// HTTP/S confirmation handshake. Meaningless for auto-confirming
	// protocols (SQS/Lambda/Firehose/application); CEL keeps them off those
	// manifests.
	// -------------------------------------------------------------------

	// Always sent (matching the Terraform module's state-pinned rendering) —
	// a provider-local wait knob with default false, never transmitted to AWS.
	args.EndpointAutoConfirms = pulumi.BoolPtr(spec.EndpointAutoConfirms)
	if spec.ConfirmationTimeoutMinutes != 0 {
		args.ConfirmationTimeoutInMinutes = pulumi.IntPtr(int(spec.ConfirmationTimeoutMinutes))
	}

	// -------------------------------------------------------------------
	// Create subscription
	// -------------------------------------------------------------------

	sub, err := sns.NewTopicSubscription(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create SNS subscription")
	}

	// Export outputs matching AwsSnsSubscriptionOutputs.
	ctx.Export(OpSubscriptionArn, sub.Arn)
	ctx.Export(OpOwnerId, sub.OwnerId)
	ctx.Export(OpPendingConfirmation, sub.PendingConfirmation)
	ctx.Export(OpConfirmationWasAuthenticated, sub.ConfirmationWasAuthenticated)

	return sub, nil
}
