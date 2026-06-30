package module

import (
	"github.com/pkg/errors"
	gcppubsubtopicv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcppubsubtopic/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/pubsub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func topic(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPubSubTopic.Spec

	args := &pubsub.TopicArgs{
		Name:    pulumi.String(spec.TopicName),
		Project: pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:  pulumi.ToStringMap(locals.GcpLabels),
	}

	// CMEK encryption.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.StringPtr(spec.KmsKeyName.GetValue())
	}

	// Message retention.
	if spec.MessageRetentionDuration != "" {
		args.MessageRetentionDuration = pulumi.StringPtr(spec.MessageRetentionDuration)
	}

	// Message storage policy.
	if spec.MessageStoragePolicy != nil && len(spec.MessageStoragePolicy.AllowedPersistenceRegions) > 0 {
		policyArgs := &pubsub.TopicMessageStoragePolicyArgs{
			AllowedPersistenceRegions: pulumi.ToStringArray(spec.MessageStoragePolicy.AllowedPersistenceRegions),
		}
		if spec.MessageStoragePolicy.EnforceInTransit {
			policyArgs.EnforceInTransit = pulumi.BoolPtr(true)
		}
		args.MessageStoragePolicy = policyArgs
	}

	// Schema settings.
	if spec.SchemaSettings != nil && spec.SchemaSettings.Schema != "" {
		schemaArgs := &pubsub.TopicSchemaSettingsArgs{
			Schema: pulumi.String(spec.SchemaSettings.Schema),
		}
		if spec.SchemaSettings.Encoding != "" {
			schemaArgs.Encoding = pulumi.StringPtr(spec.SchemaSettings.Encoding)
		}
		args.SchemaSettings = schemaArgs
	}

	// Ingestion data source settings.
	if spec.IngestionDataSourceSettings != nil {
		ingestionArgs, err := ingestionDataSourceSettings(spec.IngestionDataSourceSettings)
		if err != nil {
			return errors.Wrap(err, "failed to build ingestion data source settings")
		}
		if ingestionArgs != nil {
			args.IngestionDataSourceSettings = ingestionArgs
		}
	}

	createdTopic, err := pubsub.NewTopic(ctx, "pubsub-topic", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create pubsub topic")
	}

	ctx.Export(OpTopicId, createdTopic.ID())
	ctx.Export(OpTopicName, createdTopic.Name)

	return nil
}

func ingestionDataSourceSettings(ids *gcppubsubtopicv1.GcpPubSubTopicIngestionDataSourceSettings) (*pubsub.TopicIngestionDataSourceSettingsArgs, error) {
	result := &pubsub.TopicIngestionDataSourceSettingsArgs{}
	hasContent := false

	// AWS Kinesis.
	if ids.AwsKinesis != nil {
		result.AwsKinesis = &pubsub.TopicIngestionDataSourceSettingsAwsKinesisArgs{
			StreamArn:         pulumi.String(ids.AwsKinesis.StreamArn),
			ConsumerArn:       pulumi.String(ids.AwsKinesis.ConsumerArn),
			AwsRoleArn:        pulumi.String(ids.AwsKinesis.AwsRoleArn),
			GcpServiceAccount: pulumi.String(ids.AwsKinesis.GcpServiceAccount),
		}
		hasContent = true
	}

	// AWS MSK.
	if ids.AwsMsk != nil {
		result.AwsMsk = &pubsub.TopicIngestionDataSourceSettingsAwsMskArgs{
			ClusterArn:        pulumi.String(ids.AwsMsk.ClusterArn),
			Topic:             pulumi.String(ids.AwsMsk.Topic),
			AwsRoleArn:        pulumi.String(ids.AwsMsk.AwsRoleArn),
			GcpServiceAccount: pulumi.String(ids.AwsMsk.GcpServiceAccount),
		}
		hasContent = true
	}

	// Azure Event Hubs.
	if ids.AzureEventHubs != nil {
		azArgs := &pubsub.TopicIngestionDataSourceSettingsAzureEventHubsArgs{}
		if ids.AzureEventHubs.ResourceGroup != "" {
			azArgs.ResourceGroup = pulumi.StringPtr(ids.AzureEventHubs.ResourceGroup)
		}
		if ids.AzureEventHubs.Namespace != "" {
			azArgs.Namespace = pulumi.StringPtr(ids.AzureEventHubs.Namespace)
		}
		if ids.AzureEventHubs.EventHub != "" {
			azArgs.EventHub = pulumi.StringPtr(ids.AzureEventHubs.EventHub)
		}
		if ids.AzureEventHubs.ClientId != "" {
			azArgs.ClientId = pulumi.StringPtr(ids.AzureEventHubs.ClientId)
		}
		if ids.AzureEventHubs.TenantId != "" {
			azArgs.TenantId = pulumi.StringPtr(ids.AzureEventHubs.TenantId)
		}
		if ids.AzureEventHubs.SubscriptionId != "" {
			azArgs.SubscriptionId = pulumi.StringPtr(ids.AzureEventHubs.SubscriptionId)
		}
		if ids.AzureEventHubs.GcpServiceAccount != "" {
			azArgs.GcpServiceAccount = pulumi.StringPtr(ids.AzureEventHubs.GcpServiceAccount)
		}
		result.AzureEventHubs = azArgs
		hasContent = true
	}

	// Cloud Storage.
	if ids.CloudStorage != nil {
		csArgs := &pubsub.TopicIngestionDataSourceSettingsCloudStorageArgs{
			Bucket: pulumi.String(ids.CloudStorage.Bucket.GetValue()),
		}
		if ids.CloudStorage.MatchGlob != "" {
			csArgs.MatchGlob = pulumi.StringPtr(ids.CloudStorage.MatchGlob)
		}
		if ids.CloudStorage.MinimumObjectCreateTime != "" {
			csArgs.MinimumObjectCreateTime = pulumi.StringPtr(ids.CloudStorage.MinimumObjectCreateTime)
		}

		// Format selection: exactly one should be set.
		if ids.CloudStorage.TextFormat != nil {
			tfArgs := &pubsub.TopicIngestionDataSourceSettingsCloudStorageTextFormatArgs{}
			if ids.CloudStorage.TextFormat.Delimiter != "" {
				tfArgs.Delimiter = pulumi.StringPtr(ids.CloudStorage.TextFormat.Delimiter)
			}
			csArgs.TextFormat = tfArgs
		} else if ids.CloudStorage.AvroFormat != nil {
			csArgs.AvroFormat = &pubsub.TopicIngestionDataSourceSettingsCloudStorageAvroFormatArgs{}
		} else if ids.CloudStorage.PubsubAvroFormat != nil {
			csArgs.PubsubAvroFormat = &pubsub.TopicIngestionDataSourceSettingsCloudStoragePubsubAvroFormatArgs{}
		}

		result.CloudStorage = csArgs
		hasContent = true
	}

	// Confluent Cloud.
	if ids.ConfluentCloud != nil {
		ccArgs := &pubsub.TopicIngestionDataSourceSettingsConfluentCloudArgs{
			BootstrapServer:   pulumi.String(ids.ConfluentCloud.BootstrapServer),
			Topic:             pulumi.String(ids.ConfluentCloud.Topic),
			IdentityPoolId:    pulumi.String(ids.ConfluentCloud.IdentityPoolId),
			GcpServiceAccount: pulumi.String(ids.ConfluentCloud.GcpServiceAccount),
		}
		if ids.ConfluentCloud.ClusterId != "" {
			ccArgs.ClusterId = pulumi.StringPtr(ids.ConfluentCloud.ClusterId)
		}
		result.ConfluentCloud = ccArgs
		hasContent = true
	}

	// Platform logs settings.
	if ids.PlatformLogsSettings != nil && ids.PlatformLogsSettings.Severity != "" {
		result.PlatformLogsSettings = &pubsub.TopicIngestionDataSourceSettingsPlatformLogsSettingsArgs{
			Severity: pulumi.StringPtr(ids.PlatformLogsSettings.Severity),
		}
		hasContent = true
	}

	if !hasContent {
		return nil, nil
	}

	return result, nil
}
