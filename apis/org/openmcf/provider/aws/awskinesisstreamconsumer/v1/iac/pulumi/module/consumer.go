package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/kinesis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func consumer(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	c, err := kinesis.NewStreamConsumer(ctx, locals.Target.Metadata.Name, &kinesis.StreamConsumerArgs{
		Name:      pulumi.StringPtr(locals.ConsumerName),
		StreamArn: pulumi.String(spec.StreamArn.GetValue()),
		Tags:      pulumi.ToStringMap(locals.AwsTags),
	}, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Kinesis stream consumer")
	}

	// Export outputs matching AwsKinesisStreamConsumerStackOutputs.
	ctx.Export(OpConsumerArn, c.Arn)
	ctx.Export(OpConsumerName, c.Name)
	ctx.Export(OpStreamArn, c.StreamArn)
	ctx.Export(OpCreationTimestamp, c.CreationTimestamp)

	return nil
}
