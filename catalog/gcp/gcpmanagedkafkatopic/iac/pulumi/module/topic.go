package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/managedkafka"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// topic creates the Kafka topic on its cluster. The cluster already enabled
// the Managed Kafka API, so the topic enables nothing.
func topic(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpManagedKafkaTopic.Spec
	resourceName := locals.GcpManagedKafkaTopic.Metadata.Name

	args := &managedkafka.TopicArgs{
		Location:          pulumi.String(spec.Location),
		Cluster:           pulumi.String(locals.ClusterId),
		TopicId:           pulumi.String(locals.TopicId),
		ReplicationFactor: pulumi.Int(int(spec.ReplicationFactor)),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	// Unset optionals stay out of the payload so Google's defaults apply.
	if spec.PartitionCount > 0 {
		args.PartitionCount = pulumi.Int(int(spec.PartitionCount))
	}
	if len(spec.Configs) > 0 {
		args.Configs = pulumi.ToStringMap(spec.Configs)
	}

	// Engine-side destroy stance: PREVENT fails destroys, ABANDON removes
	// from management without deleting. Sent only when set so the provider
	// default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := managedkafka.NewTopic(ctx, resourceName, args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create managed kafka topic")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpTopicId, created.TopicId)
	return nil
}
