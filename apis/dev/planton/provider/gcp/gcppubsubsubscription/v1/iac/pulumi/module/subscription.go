package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/pubsub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func subscription(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPubSubSubscription.Spec

	args := &pubsub.SubscriptionArgs{
		Name:    pulumi.String(spec.SubscriptionName),
		Topic:   pulumi.String(spec.Topic.GetValue()),
		Project: pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:  pulumi.ToStringMap(locals.GcpLabels),
	}

	// Ack deadline.
	if spec.AckDeadlineSeconds > 0 {
		args.AckDeadlineSeconds = pulumi.IntPtr(int(spec.AckDeadlineSeconds))
	}

	// Message retention.
	if spec.MessageRetentionDuration != "" {
		args.MessageRetentionDuration = pulumi.StringPtr(spec.MessageRetentionDuration)
	}

	// Retain acknowledged messages.
	if spec.RetainAckedMessages {
		args.RetainAckedMessages = pulumi.BoolPtr(true)
	}

	// Expiration policy.
	if spec.ExpirationPolicy != nil {
		args.ExpirationPolicy = &pubsub.SubscriptionExpirationPolicyArgs{
			Ttl: pulumi.String(spec.ExpirationPolicy.Ttl),
		}
	}

	// Filter.
	if spec.Filter != "" {
		args.Filter = pulumi.StringPtr(spec.Filter)
	}

	// Message ordering.
	if spec.EnableMessageOrdering {
		args.EnableMessageOrdering = pulumi.BoolPtr(true)
	}

	// Exactly-once delivery.
	if spec.EnableExactlyOnceDelivery {
		args.EnableExactlyOnceDelivery = pulumi.BoolPtr(true)
	}

	// Dead letter policy.
	if spec.DeadLetterPolicy != nil {
		dlpArgs := &pubsub.SubscriptionDeadLetterPolicyArgs{}
		if spec.DeadLetterPolicy.DeadLetterTopic != nil && spec.DeadLetterPolicy.DeadLetterTopic.GetValue() != "" {
			dlpArgs.DeadLetterTopic = pulumi.StringPtr(spec.DeadLetterPolicy.DeadLetterTopic.GetValue())
		}
		if spec.DeadLetterPolicy.MaxDeliveryAttempts > 0 {
			dlpArgs.MaxDeliveryAttempts = pulumi.IntPtr(int(spec.DeadLetterPolicy.MaxDeliveryAttempts))
		}
		args.DeadLetterPolicy = dlpArgs
	}

	// Retry policy.
	if spec.RetryPolicy != nil {
		rpArgs := &pubsub.SubscriptionRetryPolicyArgs{}
		if spec.RetryPolicy.MinimumBackoff != "" {
			rpArgs.MinimumBackoff = pulumi.StringPtr(spec.RetryPolicy.MinimumBackoff)
		}
		if spec.RetryPolicy.MaximumBackoff != "" {
			rpArgs.MaximumBackoff = pulumi.StringPtr(spec.RetryPolicy.MaximumBackoff)
		}
		args.RetryPolicy = rpArgs
	}

	// Push config.
	if spec.PushConfig != nil {
		pushArgs := &pubsub.SubscriptionPushConfigArgs{
			PushEndpoint: pulumi.String(spec.PushConfig.PushEndpoint),
		}
		if len(spec.PushConfig.Attributes) > 0 {
			pushArgs.Attributes = pulumi.ToStringMap(spec.PushConfig.Attributes)
		}
		if spec.PushConfig.OidcToken != nil {
			oidcArgs := &pubsub.SubscriptionPushConfigOidcTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.PushConfig.OidcToken.ServiceAccountEmail),
			}
			if spec.PushConfig.OidcToken.Audience != "" {
				oidcArgs.Audience = pulumi.StringPtr(spec.PushConfig.OidcToken.Audience)
			}
			pushArgs.OidcToken = oidcArgs
		}
		if spec.PushConfig.NoWrapper != nil {
			pushArgs.NoWrapper = &pubsub.SubscriptionPushConfigNoWrapperArgs{
				WriteMetadata: pulumi.Bool(spec.PushConfig.NoWrapper.WriteMetadata),
			}
		}
		args.PushConfig = pushArgs
	}

	// BigQuery config.
	if spec.BigqueryConfig != nil {
		bqArgs := &pubsub.SubscriptionBigqueryConfigArgs{
			Table: pulumi.String(spec.BigqueryConfig.Table),
		}
		if spec.BigqueryConfig.UseTopicSchema {
			bqArgs.UseTopicSchema = pulumi.BoolPtr(true)
		}
		if spec.BigqueryConfig.UseTableSchema {
			bqArgs.UseTableSchema = pulumi.BoolPtr(true)
		}
		if spec.BigqueryConfig.DropUnknownFields {
			bqArgs.DropUnknownFields = pulumi.BoolPtr(true)
		}
		if spec.BigqueryConfig.WriteMetadata {
			bqArgs.WriteMetadata = pulumi.BoolPtr(true)
		}
		if spec.BigqueryConfig.ServiceAccountEmail != "" {
			bqArgs.ServiceAccountEmail = pulumi.StringPtr(spec.BigqueryConfig.ServiceAccountEmail)
		}
		args.BigqueryConfig = bqArgs
	}

	// Cloud Storage config.
	if spec.CloudStorageConfig != nil {
		csArgs := &pubsub.SubscriptionCloudStorageConfigArgs{
			Bucket: pulumi.String(spec.CloudStorageConfig.Bucket.GetValue()),
		}
		if spec.CloudStorageConfig.FilenamePrefix != "" {
			csArgs.FilenamePrefix = pulumi.StringPtr(spec.CloudStorageConfig.FilenamePrefix)
		}
		if spec.CloudStorageConfig.FilenameSuffix != "" {
			csArgs.FilenameSuffix = pulumi.StringPtr(spec.CloudStorageConfig.FilenameSuffix)
		}
		if spec.CloudStorageConfig.FilenameDatetimeFormat != "" {
			csArgs.FilenameDatetimeFormat = pulumi.StringPtr(spec.CloudStorageConfig.FilenameDatetimeFormat)
		}
		if spec.CloudStorageConfig.MaxBytes > 0 {
			csArgs.MaxBytes = pulumi.IntPtr(int(spec.CloudStorageConfig.MaxBytes))
		}
		if spec.CloudStorageConfig.MaxDuration != "" {
			csArgs.MaxDuration = pulumi.StringPtr(spec.CloudStorageConfig.MaxDuration)
		}
		if spec.CloudStorageConfig.MaxMessages > 0 {
			csArgs.MaxMessages = pulumi.IntPtr(int(spec.CloudStorageConfig.MaxMessages))
		}
		if spec.CloudStorageConfig.AvroConfig != nil {
			csArgs.AvroConfig = &pubsub.SubscriptionCloudStorageConfigAvroConfigArgs{
				UseTopicSchema: pulumi.BoolPtr(spec.CloudStorageConfig.AvroConfig.UseTopicSchema),
				WriteMetadata:  pulumi.BoolPtr(spec.CloudStorageConfig.AvroConfig.WriteMetadata),
			}
		}
		if spec.CloudStorageConfig.ServiceAccountEmail != "" {
			csArgs.ServiceAccountEmail = pulumi.StringPtr(spec.CloudStorageConfig.ServiceAccountEmail)
		}
		args.CloudStorageConfig = csArgs
	}

	createdSubscription, err := pubsub.NewSubscription(ctx, "pubsub-subscription", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create pubsub subscription")
	}

	ctx.Export(OpSubscriptionId, createdSubscription.ID())
	ctx.Export(OpSubscriptionName, createdSubscription.Name)

	return nil
}
