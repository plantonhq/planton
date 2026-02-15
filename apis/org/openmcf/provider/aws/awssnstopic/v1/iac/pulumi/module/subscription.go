package module

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/sns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func subscriptions(ctx *pulumi.Context, locals *Locals, createdTopic *sns.Topic, provider *aws.Provider) error {
	if len(locals.Spec.Subscriptions) == 0 {
		// No subscriptions — export empty map and return.
		ctx.Export(OpSubscriptionArns, pulumi.StringMap{})
		return nil
	}

	subscriptionArnMap := pulumi.StringMap{}

	for i, sub := range locals.Spec.Subscriptions {
		resourceName := fmt.Sprintf("%s-%s", locals.Target.Metadata.Name, sub.Name)

		args := &sns.TopicSubscriptionArgs{
			Topic:              createdTopic.Arn,
			Protocol:           pulumi.String(sub.Protocol),
			Endpoint:           pulumi.String(sub.Endpoint.GetValue()),
			RawMessageDelivery: pulumi.BoolPtr(sub.RawMessageDelivery),
		}

		// -----------------------------------------------------------
		// Filter policy
		// -----------------------------------------------------------

		if sub.FilterPolicy != nil {
			filterMap := sub.FilterPolicy.AsMap()
			filterJSON, err := json.Marshal(filterMap)
			if err != nil {
				return errors.Wrapf(err, "failed to serialize filter policy for subscription %q (index %d)", sub.Name, i)
			}
			args.FilterPolicy = pulumi.StringPtr(string(filterJSON))
		}

		if sub.FilterPolicyScope != "" {
			args.FilterPolicyScope = pulumi.StringPtr(sub.FilterPolicyScope)
		}

		// -----------------------------------------------------------
		// Redrive config (subscription DLQ)
		// -----------------------------------------------------------

		if sub.RedriveConfig != nil {
			redrivePolicy := map[string]interface{}{
				"deadLetterTargetArn": sub.RedriveConfig.DeadLetterTargetArn.GetValue(),
			}
			redriveJSON, err := json.Marshal(redrivePolicy)
			if err != nil {
				return errors.Wrapf(err, "failed to serialize redrive policy for subscription %q (index %d)", sub.Name, i)
			}
			args.RedrivePolicy = pulumi.StringPtr(string(redriveJSON))
		}

		// -----------------------------------------------------------
		// Firehose role
		// -----------------------------------------------------------

		if sub.SubscriptionRoleArn.GetValue() != "" {
			args.SubscriptionRoleArn = pulumi.StringPtr(sub.SubscriptionRoleArn.GetValue())
		}

		// -----------------------------------------------------------
		// Create subscription
		// -----------------------------------------------------------

		created, err := sns.NewTopicSubscription(ctx, resourceName, args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create SNS subscription %q (index %d)", sub.Name, i)
		}

		subscriptionArnMap[sub.Name] = created.Arn
	}

	// Export subscription ARN map matching AwsSnsTopicStackOutputs.subscription_arns.
	ctx.Export(OpSubscriptionArns, subscriptionArnMap)

	return nil
}
