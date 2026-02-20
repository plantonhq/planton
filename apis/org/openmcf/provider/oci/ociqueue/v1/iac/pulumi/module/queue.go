package module

import (
	"github.com/pkg/errors"
	ociqueuev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ociqueue/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/queue"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func queueResource(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciQueue.Spec

	args := &queue.QueueArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DisplayName:   pulumi.String(locals.QueueName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.CustomEncryptionKeyId != nil {
		args.CustomEncryptionKeyId = pulumi.String(spec.CustomEncryptionKeyId.GetValue())
	}

	if spec.DeadLetterQueueDeliveryCount != nil {
		args.DeadLetterQueueDeliveryCount = pulumi.Int(int(*spec.DeadLetterQueueDeliveryCount))
	}

	if spec.RetentionInSeconds != nil {
		args.RetentionInSeconds = pulumi.Int(int(*spec.RetentionInSeconds))
	}

	if spec.TimeoutInSeconds != nil {
		args.TimeoutInSeconds = pulumi.Int(int(*spec.TimeoutInSeconds))
	}

	if spec.VisibilityInSeconds != nil {
		args.VisibilityInSeconds = pulumi.Int(int(*spec.VisibilityInSeconds))
	}

	if spec.ChannelConsumptionLimit != nil {
		args.ChannelConsumptionLimit = pulumi.Int(int(*spec.ChannelConsumptionLimit))
	}

	caps := buildCapabilities(spec)
	if len(caps) > 0 {
		args.Capabilities = queue.QueueCapabilityArray(caps)
	}

	q, err := queue.NewQueue(ctx, locals.QueueName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create queue")
	}

	ctx.Export(OpQueueId, q.ID())
	ctx.Export(OpMessagesEndpoint, q.MessagesEndpoint)

	return nil
}

func buildCapabilities(spec *ociqueuev1.OciQueueSpec) []queue.QueueCapabilityInput {
	var caps []queue.QueueCapabilityInput

	if spec.IsLargeMessagesEnabled != nil && *spec.IsLargeMessagesEnabled {
		caps = append(caps, &queue.QueueCapabilityArgs{
			Type: pulumi.String("LARGE_MESSAGES"),
		})
	}

	if spec.ConsumerGroupConfig != nil {
		capArgs := &queue.QueueCapabilityArgs{
			Type: pulumi.String("CONSUMER_GROUPS"),
		}

		cg := spec.ConsumerGroupConfig

		if cg.IsPrimaryEnabled != nil {
			capArgs.IsPrimaryConsumerGroupEnabled = pulumi.Bool(*cg.IsPrimaryEnabled)
		}

		if cg.PrimaryDeadLetterQueueDeliveryCount != nil {
			capArgs.PrimaryConsumerGroupDeadLetterQueueDeliveryCount = pulumi.Int(int(*cg.PrimaryDeadLetterQueueDeliveryCount))
		}

		if cg.PrimaryDisplayName != "" {
			capArgs.PrimaryConsumerGroupDisplayName = pulumi.String(cg.PrimaryDisplayName)
		}

		caps = append(caps, capArgs)
	}

	return caps
}
