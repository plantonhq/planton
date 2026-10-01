package module

import (
	"strconv"

	"github.com/pkg/errors"
	digitaloceandatabasekafkatopicv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasekafkatopic/v1alpha1"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// kafkaTopic provisions the Kafka topic and exports its outputs. Partition
// count, replication factor, and every config leaf update in place; the
// cluster and name are create-only.
func kafkaTopic(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.DatabaseKafkaTopic, error) {
	spec := locals.DigitalOceanDatabaseKafkaTopic.Spec

	args := &digitalocean.DatabaseKafkaTopicArgs{
		// References are resolved to the literal cluster UUID before the
		// module runs.
		ClusterId: pulumi.String(spec.Cluster.GetValue()),
		Name:      pulumi.String(spec.TopicName),
	}

	if spec.PartitionCount != nil {
		args.PartitionCount = pulumi.IntPtr(int(spec.GetPartitionCount()))
	}
	if spec.ReplicationFactor != nil {
		args.ReplicationFactor = pulumi.IntPtr(int(spec.GetReplicationFactor()))
	}

	// The provider schema declares config as an unbounded list (an SDKv2
	// artifact), so the bridge exposes an array; exactly one element is
	// meaningful and this module only ever sends one.
	if spec.Config != nil {
		args.Configs = digitalocean.DatabaseKafkaTopicConfigArray{
			buildTopicConfig(spec.Config),
		}
	}

	createdTopic, err := digitalocean.NewDatabaseKafkaTopic(
		ctx,
		"topic",
		args,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean kafka topic")
	}

	// The topic's provisioning state is deliberately not exported: an
	// apply-time snapshot goes stale (creation is asynchronous), so live
	// state belongs to whoever reads the API, never to the outputs contract.
	ctx.Export(OpClusterId, createdTopic.ClusterId)
	ctx.Export(OpTopicName, createdTopic.Name)

	return createdTopic, nil
}

// buildTopicConfig maps the spec's config message onto the provider's
// config block. The provider carries the 64-bit numeric tunables as
// strings (Terraform numbers are not 64-bit safe), so present values are
// rendered to decimal strings here; absent leaves stay nil and are never
// sent, deferring to the Kafka server defaults.
func buildTopicConfig(
	config *digitaloceandatabasekafkatopicv1alpha1.DigitalOceanDatabaseKafkaTopicConfig,
) *digitalocean.DatabaseKafkaTopicConfigArgs {
	configArgs := &digitalocean.DatabaseKafkaTopicConfigArgs{}

	if config.CleanupPolicy != "" {
		configArgs.CleanupPolicy = pulumi.StringPtr(config.CleanupPolicy)
	}
	if config.CompressionType != "" {
		configArgs.CompressionType = pulumi.StringPtr(config.CompressionType)
	}
	if config.DeleteRetentionMs != nil {
		configArgs.DeleteRetentionMs = pulumi.StringPtr(strconv.FormatUint(config.GetDeleteRetentionMs(), 10))
	}
	if config.FileDeleteDelayMs != nil {
		configArgs.FileDeleteDelayMs = pulumi.StringPtr(strconv.FormatUint(config.GetFileDeleteDelayMs(), 10))
	}
	if config.FlushMessages != nil {
		configArgs.FlushMessages = pulumi.StringPtr(strconv.FormatUint(config.GetFlushMessages(), 10))
	}
	if config.FlushMs != nil {
		configArgs.FlushMs = pulumi.StringPtr(strconv.FormatUint(config.GetFlushMs(), 10))
	}
	if config.IndexIntervalBytes != nil {
		configArgs.IndexIntervalBytes = pulumi.StringPtr(strconv.FormatUint(config.GetIndexIntervalBytes(), 10))
	}
	if config.MaxCompactionLagMs != nil {
		configArgs.MaxCompactionLagMs = pulumi.StringPtr(strconv.FormatUint(config.GetMaxCompactionLagMs(), 10))
	}
	if config.MaxMessageBytes != nil {
		configArgs.MaxMessageBytes = pulumi.StringPtr(strconv.FormatUint(config.GetMaxMessageBytes(), 10))
	}
	if config.MessageDownConversionEnable != nil {
		configArgs.MessageDownConversionEnable = pulumi.BoolPtr(config.GetMessageDownConversionEnable())
	}
	if config.MessageFormatVersion != "" {
		configArgs.MessageFormatVersion = pulumi.StringPtr(config.MessageFormatVersion)
	}
	if config.MessageTimestampDifferenceMaxMs != nil {
		configArgs.MessageTimestampDifferenceMaxMs = pulumi.StringPtr(strconv.FormatInt(config.GetMessageTimestampDifferenceMaxMs(), 10))
	}
	if config.MessageTimestampType != "" {
		configArgs.MessageTimestampType = pulumi.StringPtr(config.MessageTimestampType)
	}
	if config.MinCleanableDirtyRatio != nil {
		configArgs.MinCleanableDirtyRatio = pulumi.Float64Ptr(config.GetMinCleanableDirtyRatio())
	}
	if config.MinCompactionLagMs != nil {
		configArgs.MinCompactionLagMs = pulumi.StringPtr(strconv.FormatUint(config.GetMinCompactionLagMs(), 10))
	}
	if config.MinInsyncReplicas != nil {
		configArgs.MinInsyncReplicas = pulumi.IntPtr(int(config.GetMinInsyncReplicas()))
	}
	if config.Preallocate != nil {
		configArgs.Preallocate = pulumi.BoolPtr(config.GetPreallocate())
	}
	if config.RetentionBytes != nil {
		configArgs.RetentionBytes = pulumi.StringPtr(strconv.FormatInt(config.GetRetentionBytes(), 10))
	}
	if config.RetentionMs != nil {
		configArgs.RetentionMs = pulumi.StringPtr(strconv.FormatInt(config.GetRetentionMs(), 10))
	}
	if config.SegmentBytes != nil {
		configArgs.SegmentBytes = pulumi.StringPtr(strconv.FormatUint(config.GetSegmentBytes(), 10))
	}
	if config.SegmentIndexBytes != nil {
		configArgs.SegmentIndexBytes = pulumi.StringPtr(strconv.FormatUint(config.GetSegmentIndexBytes(), 10))
	}
	if config.SegmentJitterMs != nil {
		configArgs.SegmentJitterMs = pulumi.StringPtr(strconv.FormatUint(config.GetSegmentJitterMs(), 10))
	}
	if config.SegmentMs != nil {
		configArgs.SegmentMs = pulumi.StringPtr(strconv.FormatUint(config.GetSegmentMs(), 10))
	}

	return configArgs
}
