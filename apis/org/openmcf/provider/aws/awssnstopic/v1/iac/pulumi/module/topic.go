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
		if spec.ContentBasedDeduplication {
			args.ContentBasedDeduplication = pulumi.BoolPtr(true)
		}
		if spec.FifoThroughputScope != "" {
			args.FifoThroughputScope = pulumi.StringPtr(spec.FifoThroughputScope)
		}
	}

	// -------------------------------------------------------------------
	// Display name
	// -------------------------------------------------------------------

	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}

	// -------------------------------------------------------------------
	// Encryption
	// -------------------------------------------------------------------

	if spec.KmsKeyId.GetValue() != "" {
		args.KmsMasterKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
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
	// Delivery policy
	// -------------------------------------------------------------------

	if spec.DeliveryPolicy != "" {
		args.DeliveryPolicy = pulumi.StringPtr(spec.DeliveryPolicy)
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

	// Export topic-level outputs.
	ctx.Export(OpTopicArn, t.Arn)
	ctx.Export(OpTopicName, t.Name)

	return t, nil
}
