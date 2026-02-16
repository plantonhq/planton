package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// targets iterates over the spec's targets and creates an EventTarget for each.
func targets(ctx *pulumi.Context, locals *Locals, createdRule *cloudwatch.EventRule, provider *aws.Provider) error {
	if len(locals.Spec.Targets) == 0 {
		return nil
	}

	for i, target := range locals.Spec.Targets {
		resourceName := fmt.Sprintf("%s-%s", locals.Target.Metadata.Name, target.Name)

		args := &cloudwatch.EventTargetArgs{
			Rule:     createdRule.Name,
			Arn:      pulumi.String(target.Arn.GetValue()),
			TargetId: pulumi.StringPtr(target.Name),
		}

		// Event bus — must match the rule's bus.
		if locals.Spec.EventBusName.GetValue() != "" {
			args.EventBusName = pulumi.StringPtr(locals.Spec.EventBusName.GetValue())
		}

		// IAM role for target invocation
		if target.RoleArn.GetValue() != "" {
			args.RoleArn = pulumi.StringPtr(target.RoleArn.GetValue())
		}

		// -----------------------------------------------------------
		// Input transformation (mutually exclusive)
		// -----------------------------------------------------------

		if target.Input != "" {
			args.Input = pulumi.StringPtr(target.Input)
		}

		if target.InputPath != "" {
			args.InputPath = pulumi.StringPtr(target.InputPath)
		}

		if target.InputTransformer != nil {
			transformerArgs := &cloudwatch.EventTargetInputTransformerArgs{
				InputTemplate: pulumi.String(target.InputTransformer.InputTemplate),
			}
			if len(target.InputTransformer.InputPaths) > 0 {
				transformerArgs.InputPaths = pulumi.ToStringMap(target.InputTransformer.InputPaths)
			}
			args.InputTransformer = transformerArgs
		}

		// -----------------------------------------------------------
		// Dead letter config
		// -----------------------------------------------------------

		if target.DeadLetterConfig != nil && target.DeadLetterConfig.Arn.GetValue() != "" {
			args.DeadLetterConfig = &cloudwatch.EventTargetDeadLetterConfigArgs{
				Arn: pulumi.StringPtr(target.DeadLetterConfig.Arn.GetValue()),
			}
		}

		// -----------------------------------------------------------
		// Retry policy
		// -----------------------------------------------------------

		if target.RetryPolicy != nil {
			retryArgs := &cloudwatch.EventTargetRetryPolicyArgs{}
			if target.RetryPolicy.MaximumEventAgeInSeconds > 0 {
				retryArgs.MaximumEventAgeInSeconds = pulumi.IntPtr(int(target.RetryPolicy.MaximumEventAgeInSeconds))
			}
			if target.RetryPolicy.MaximumRetryAttempts > 0 {
				retryArgs.MaximumRetryAttempts = pulumi.IntPtr(int(target.RetryPolicy.MaximumRetryAttempts))
			}
			args.RetryPolicy = retryArgs
		}

		// -----------------------------------------------------------
		// SQS config (message_group_id for FIFO queues)
		// -----------------------------------------------------------

		if target.SqsConfig != nil && target.SqsConfig.MessageGroupId != "" {
			args.SqsTarget = &cloudwatch.EventTargetSqsTargetArgs{
				MessageGroupId: pulumi.StringPtr(target.SqsConfig.MessageGroupId),
			}
		}

		// -----------------------------------------------------------
		// Create target
		// -----------------------------------------------------------

		_, err := cloudwatch.NewEventTarget(ctx, resourceName, args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create EventBridge target %q (index %d)", target.Name, i)
		}
	}

	return nil
}
