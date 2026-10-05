package module

import (
	"encoding/json"

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

	// Resource-based access policy -- AWS models this as a separate API
	// keyed by the consumer ARN (one policy per consumer), folded into the
	// spec because it has no identity of its own. The primary use is
	// cross-account enhanced fan-out: SubscribeToShard grants without role
	// assumption.
	if spec.ResourcePolicy != nil {
		policyJSON, err := json.Marshal(spec.ResourcePolicy.AsMap())
		if err != nil {
			return errors.Wrap(err, "failed to serialize resource policy")
		}
		_, err = kinesis.NewResourcePolicy(ctx, "resource-policy", &kinesis.ResourcePolicyArgs{
			ResourceArn: c.Arn,
			Policy:      pulumi.String(policyJSON),
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create Kinesis consumer resource policy")
		}
	}

	// Export outputs matching AwsKinesisStreamConsumerOutputs.
	ctx.Export(OpConsumerArn, c.Arn)
	ctx.Export(OpConsumerName, c.Name)
	ctx.Export(OpStreamArn, c.StreamArn)
	ctx.Export(OpCreationTimestamp, c.CreationTimestamp)

	return nil
}
