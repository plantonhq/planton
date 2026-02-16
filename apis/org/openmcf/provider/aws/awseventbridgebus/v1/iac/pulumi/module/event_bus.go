package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func eventBus(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &cloudwatch.EventBusArgs{
		Name: pulumi.StringPtr(locals.Target.Metadata.Name),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Description
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// KMS encryption
	if spec.KmsKeyIdentifier.GetValue() != "" {
		args.KmsKeyIdentifier = pulumi.StringPtr(spec.KmsKeyIdentifier.GetValue())
	}

	// Partner event source
	if spec.EventSourceName != "" {
		args.EventSourceName = pulumi.StringPtr(spec.EventSourceName)
	}

	// Dead letter config
	if spec.DeadLetterConfig != nil && spec.DeadLetterConfig.Arn.GetValue() != "" {
		args.DeadLetterConfig = &cloudwatch.EventBusDeadLetterConfigArgs{
			Arn: pulumi.StringPtr(spec.DeadLetterConfig.Arn.GetValue()),
		}
	}

	// Logging config
	if spec.LogConfig != nil && spec.LogConfig.Level != "" {
		logArgs := &cloudwatch.EventBusLogConfigArgs{
			Level: pulumi.StringPtr(spec.LogConfig.Level),
		}
		if spec.LogConfig.IncludeDetail != "" {
			logArgs.IncludeDetail = pulumi.StringPtr(spec.LogConfig.IncludeDetail)
		}
		args.LogConfig = logArgs
	}

	bus, err := cloudwatch.NewEventBus(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create EventBridge bus")
	}

	// Export outputs matching AwsEventBridgeBusStackOutputs.
	ctx.Export(OpBusName, bus.Name)
	ctx.Export(OpBusArn, bus.Arn)

	return nil
}
