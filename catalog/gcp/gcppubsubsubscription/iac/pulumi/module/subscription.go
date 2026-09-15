package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/pubsub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func subscription(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPubSubSubscription.Spec

	// Enable the Pub/Sub API — the control plane that owns the
	// subscription. disable_on_destroy stays false: tearing down one
	// subscription must never disable the API for everything else in the
	// project.
	pubsubApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("pubsub.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		pubsubApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdPubsubApi, err := projects.NewService(ctx,
		"gcppss-pubsub.googleapis.com", pubsubApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable pubsub.googleapis.com api")
	}

	// The topic attachment is ForceNew: repointing a subscription at a
	// different topic replaces the subscription (and its backlog).
	args := &pubsub.SubscriptionArgs{
		Name:   pulumi.String(spec.SubscriptionName),
		Topic:  pulumi.String(spec.Topic.GetValue()),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// Ack deadline: 0 accepts the API default (10s).
	if spec.AckDeadlineSeconds > 0 {
		args.AckDeadlineSeconds = pulumi.IntPtr(int(spec.AckDeadlineSeconds))
	}

	// Backlog retention window; with retain_acked_messages it also
	// bounds how far back a seek can replay.
	if spec.MessageRetentionDuration != "" {
		args.MessageRetentionDuration = pulumi.StringPtr(spec.MessageRetentionDuration)
	}

	if spec.RetainAckedMessages {
		args.RetainAckedMessages = pulumi.BoolPtr(true)
	}

	// Expiration policy: an empty ttl string means "never expires" —
	// send it as-is; the API treats empty as the never-expire sentinel.
	if spec.ExpirationPolicy != nil {
		args.ExpirationPolicy = &pubsub.SubscriptionExpirationPolicyArgs{
			Ttl: pulumi.String(spec.ExpirationPolicy.Ttl),
		}
	}

	// The attribute filter is ForceNew: changing it replaces the
	// subscription. Non-matching messages are auto-acked, never delivered.
	if spec.Filter != "" {
		args.Filter = pulumi.StringPtr(spec.Filter)
	}

	// Ordering is ForceNew: it changes how Pub/Sub stores the backlog.
	if spec.EnableMessageOrdering {
		args.EnableMessageOrdering = pulumi.BoolPtr(true)
	}

	if spec.EnableExactlyOnceDelivery {
		args.EnableExactlyOnceDelivery = pulumi.BoolPtr(true)
	}

	// Dead-letter policy: the Pub/Sub service agent needs Subscriber on
	// this subscription and Publisher on the dead-letter topic.
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

	// Retry backoff between delivery attempts after NACK/deadline events.
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

	// Delivery mode: at most one of push/bigquery/cloud-storage (spec CEL
	// enforces exclusivity); none set = pull.
	if spec.PushConfig != nil {
		pushArgs := &pubsub.SubscriptionPushConfigArgs{
			// push_endpoint is a StringValueOrRef; the resolver substitutes the
			// referenced Cloud Run service URL (or any literal HTTPS endpoint)
			// before the module runs, so only the resolved value exists here.
			PushEndpoint: pulumi.String(spec.PushConfig.PushEndpoint.GetValue()),
		}
		if len(spec.PushConfig.Attributes) > 0 {
			pushArgs.Attributes = pulumi.ToStringMap(spec.PushConfig.Attributes)
		}
		if spec.PushConfig.OidcToken != nil {
			// The service-account reference arrives resolved to a literal
			// email (the FK resolver runs before the module).
			oidcArgs := &pubsub.SubscriptionPushConfigOidcTokenArgs{
				ServiceAccountEmail: pulumi.String(spec.PushConfig.OidcToken.ServiceAccountEmail.GetValue()),
			}
			if spec.PushConfig.OidcToken.Audience != "" {
				oidcArgs.Audience = pulumi.StringPtr(spec.PushConfig.OidcToken.Audience)
			}
			pushArgs.OidcToken = oidcArgs
		}
		// The wrapper choice is a oneof: the unwrapped arm renders the provider's
		// no_wrapper block; the pubsub_wrapper arm (or no choice) is the
		// provider's default envelope and renders nothing.
		if nw := spec.PushConfig.GetNoWrapper(); nw != nil {
			pushArgs.NoWrapper = &pubsub.SubscriptionPushConfigNoWrapperArgs{
				WriteMetadata: pulumi.Bool(nw.WriteMetadata),
			}
		}
		args.PushConfig = pushArgs
	}

	if spec.BigqueryConfig != nil {
		// The table reference arrives resolved to the dotted
		// {project}.{dataset}.{table} form the Pub/Sub API expects.
		bqArgs := &pubsub.SubscriptionBigqueryConfigArgs{
			Table: pulumi.String(spec.BigqueryConfig.Table.GetValue()),
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
		if spec.BigqueryConfig.ServiceAccountEmail.GetValue() != "" {
			bqArgs.ServiceAccountEmail = pulumi.StringPtr(spec.BigqueryConfig.ServiceAccountEmail.GetValue())
		}
		args.BigqueryConfig = bqArgs
	}

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
		// The output format is a oneof: the Avro arm renders the provider's
		// avro_config block; the text_config arm (or no choice) is the provider's
		// default text output and renders nothing.
		if avro := spec.CloudStorageConfig.GetAvroConfig(); avro != nil {
			csArgs.AvroConfig = &pubsub.SubscriptionCloudStorageConfigAvroConfigArgs{
				UseTopicSchema: pulumi.BoolPtr(avro.UseTopicSchema),
				WriteMetadata:  pulumi.BoolPtr(avro.WriteMetadata),
			}
		}
		if spec.CloudStorageConfig.ServiceAccountEmail.GetValue() != "" {
			csArgs.ServiceAccountEmail = pulumi.StringPtr(spec.CloudStorageConfig.ServiceAccountEmail.GetValue())
		}
		args.CloudStorageConfig = csArgs
	}

	// Ordered transform pipeline: transforms run in list order on every
	// message before delivery to THIS subscription only; a disabled
	// transform keeps its position (the staging lever) without being
	// applied. Each step carries exactly one arm — a JavaScript UDF or
	// an AI inference call (spec-enforced).
	if len(spec.MessageTransforms) > 0 {
		transforms := pubsub.SubscriptionMessageTransformArray{}
		for _, transform := range spec.MessageTransforms {
			transformArgs := &pubsub.SubscriptionMessageTransformArgs{}
			if transform.JavascriptUdf != nil {
				transformArgs.JavascriptUdf = &pubsub.SubscriptionMessageTransformJavascriptUdfArgs{
					FunctionName: pulumi.String(transform.JavascriptUdf.FunctionName),
					Code:         pulumi.String(transform.JavascriptUdf.Code),
				}
			}
			if transform.AiInference != nil {
				// The endpoint reference is resolved to a literal
				// dedicated-endpoint / publisher-model path before the
				// module runs.
				aiArgs := &pubsub.SubscriptionMessageTransformAiInferenceArgs{
					Endpoint: pulumi.String(transform.AiInference.Endpoint.GetValue()),
				}
				if transform.AiInference.ServiceAccountEmail.GetValue() != "" {
					aiArgs.ServiceAccountEmail = pulumi.StringPtr(transform.AiInference.ServiceAccountEmail.GetValue())
				}
				if transform.AiInference.UnstructuredInference != nil && len(transform.AiInference.UnstructuredInference.Parameters) > 0 {
					aiArgs.UnstructuredInference = &pubsub.SubscriptionMessageTransformAiInferenceUnstructuredInferenceArgs{
						Parameters: pulumi.ToStringMap(transform.AiInference.UnstructuredInference.Parameters),
					}
				}
				transformArgs.AiInference = aiArgs
			}
			if transform.Disabled {
				transformArgs.Disabled = pulumi.BoolPtr(true)
			}
			transforms = append(transforms, transformArgs)
		}
		args.MessageTransforms = transforms
	}

	// Resource Manager tags bind at create time only (ForceNew): a later
	// tag change replaces the subscription and drops its backlog.
	if len(spec.ResourceManagerTags) > 0 {
		args.Tags = pulumi.ToStringMap(spec.ResourceManagerTags)
	}

	// Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
	// Sent only when set so the provider default stays in charge otherwise.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdSubscription, err := pubsub.NewSubscription(ctx, "pubsub-subscription", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdPubsubApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create pubsub subscription")
	}

	ctx.Export(OpSubscriptionId, createdSubscription.ID())
	ctx.Export(OpSubscriptionName, createdSubscription.Name)

	return nil
}
