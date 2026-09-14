package module

import (
	"github.com/pkg/errors"
	gcppubsubtopicv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcppubsubtopic/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/pubsub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func topic(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPubSubTopic.Spec

	// Enable the Pub/Sub API — the control plane that owns the topic.
	// disable_on_destroy stays false: tearing down one topic must never
	// disable the API for everything else in the project.
	pubsubApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("pubsub.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		pubsubApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdPubsubApi, err := projects.NewService(ctx,
		"gcppst-pubsub.googleapis.com", pubsubApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable pubsub.googleapis.com api")
	}

	args := &pubsub.TopicArgs{
		Name:   pulumi.String(spec.TopicName),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	// CMEK encryption. The Pub/Sub service agent must hold
	// cloudkms.cryptoKeyEncrypterDecrypter on the key or publishes fail.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.StringPtr(spec.KmsKeyName.GetValue())
	}

	// Topic-level message retention: independent of subscription
	// retention, and the lever that lets any subscription seek back
	// within the window.
	if spec.MessageRetentionDuration != "" {
		args.MessageRetentionDuration = pulumi.StringPtr(spec.MessageRetentionDuration)
	}

	// Message storage policy: region pinning is enforced at publish
	// time; enforce_in_transit additionally rejects publishes from
	// non-allowed regions instead of rerouting them.
	if spec.MessageStoragePolicy != nil && len(spec.MessageStoragePolicy.AllowedPersistenceRegions) > 0 {
		policyArgs := &pubsub.TopicMessageStoragePolicyArgs{
			AllowedPersistenceRegions: pulumi.ToStringArray(spec.MessageStoragePolicy.AllowedPersistenceRegions),
		}
		if spec.MessageStoragePolicy.EnforceInTransit {
			policyArgs.EnforceInTransit = pulumi.BoolPtr(true)
		}
		args.MessageStoragePolicy = policyArgs
	}

	// Schema validation: the schema reference is resolved to the fully
	// qualified projects/{p}/schemas/{name} path before the module runs.
	// The revision bounds pin validation to a revision range (both
	// bounds equal freezes the contract exactly).
	if spec.SchemaSettings != nil && spec.SchemaSettings.Schema.GetValue() != "" {
		schemaArgs := &pubsub.TopicSchemaSettingsArgs{
			Schema: pulumi.String(spec.SchemaSettings.Schema.GetValue()),
		}
		if spec.SchemaSettings.Encoding != "" {
			schemaArgs.Encoding = pulumi.StringPtr(spec.SchemaSettings.Encoding)
		}
		if spec.SchemaSettings.FirstRevisionId.GetValue() != "" {
			schemaArgs.FirstRevisionId = pulumi.StringPtr(spec.SchemaSettings.FirstRevisionId.GetValue())
		}
		if spec.SchemaSettings.LastRevisionId.GetValue() != "" {
			schemaArgs.LastRevisionId = pulumi.StringPtr(spec.SchemaSettings.LastRevisionId.GetValue())
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

	// Ordered transform pipeline: transforms run in list order on every
	// published message; a disabled transform keeps its position (the
	// staging lever) without being applied. Each step carries exactly
	// one arm — a JavaScript UDF or an AI inference call (spec-enforced).
	if len(spec.MessageTransforms) > 0 {
		transforms := pubsub.TopicMessageTransformArray{}
		for _, transform := range spec.MessageTransforms {
			transformArgs := &pubsub.TopicMessageTransformArgs{}
			if transform.JavascriptUdf != nil {
				transformArgs.JavascriptUdf = &pubsub.TopicMessageTransformJavascriptUdfArgs{
					FunctionName: pulumi.String(transform.JavascriptUdf.FunctionName),
					Code:         pulumi.String(transform.JavascriptUdf.Code),
				}
			}
			if transform.AiInference != nil {
				// The endpoint reference is resolved to a literal
				// dedicated-endpoint / publisher-model path before the
				// module runs.
				aiArgs := &pubsub.TopicMessageTransformAiInferenceArgs{
					Endpoint: pulumi.String(transform.AiInference.Endpoint.GetValue()),
				}
				if transform.AiInference.ServiceAccountEmail.GetValue() != "" {
					aiArgs.ServiceAccountEmail = pulumi.StringPtr(transform.AiInference.ServiceAccountEmail.GetValue())
				}
				if transform.AiInference.UnstructuredInference != nil && len(transform.AiInference.UnstructuredInference.Parameters) > 0 {
					aiArgs.UnstructuredInference = &pubsub.TopicMessageTransformAiInferenceUnstructuredInferenceArgs{
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
	// tag change replaces the topic and detaches every subscription.
	if len(spec.ResourceManagerTags) > 0 {
		args.Tags = pulumi.ToStringMap(spec.ResourceManagerTags)
	}

	// Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
	// Sent only when set so the provider default stays in charge otherwise.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdTopic, err := pubsub.NewTopic(ctx, "pubsub-topic", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdPubsubApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create pubsub topic")
	}

	ctx.Export(OpTopicId, createdTopic.ID())
	ctx.Export(OpTopicName, createdTopic.Name)

	return nil
}

func ingestionDataSourceSettings(ids *gcppubsubtopicv1alpha1.GcpPubSubTopicIngestionDataSourceSettings) (*pubsub.TopicIngestionDataSourceSettingsArgs, error) {
	result := &pubsub.TopicIngestionDataSourceSettingsArgs{}
	hasContent := false

	// All gcp_service_account references arrive resolved to literal
	// emails (the FK resolver runs before the module).

	// AWS Kinesis.
	if ids.AwsKinesis != nil {
		result.AwsKinesis = &pubsub.TopicIngestionDataSourceSettingsAwsKinesisArgs{
			StreamArn:         pulumi.String(ids.AwsKinesis.StreamArn),
			ConsumerArn:       pulumi.String(ids.AwsKinesis.ConsumerArn),
			AwsRoleArn:        pulumi.String(ids.AwsKinesis.AwsRoleArn),
			GcpServiceAccount: pulumi.String(ids.AwsKinesis.GcpServiceAccount.GetValue()),
		}
		hasContent = true
	}

	// AWS MSK.
	if ids.AwsMsk != nil {
		result.AwsMsk = &pubsub.TopicIngestionDataSourceSettingsAwsMskArgs{
			ClusterArn:        pulumi.String(ids.AwsMsk.ClusterArn),
			Topic:             pulumi.String(ids.AwsMsk.Topic),
			AwsRoleArn:        pulumi.String(ids.AwsMsk.AwsRoleArn),
			GcpServiceAccount: pulumi.String(ids.AwsMsk.GcpServiceAccount.GetValue()),
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
		if ids.AzureEventHubs.GcpServiceAccount.GetValue() != "" {
			azArgs.GcpServiceAccount = pulumi.StringPtr(ids.AzureEventHubs.GcpServiceAccount.GetValue())
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

		// Format selection: the input_format oneof carries exactly one arm.
		if ids.CloudStorage.GetTextFormat() != nil {
			tfArgs := &pubsub.TopicIngestionDataSourceSettingsCloudStorageTextFormatArgs{}
			if ids.CloudStorage.GetTextFormat().Delimiter != "" {
				tfArgs.Delimiter = pulumi.StringPtr(ids.CloudStorage.GetTextFormat().Delimiter)
			}
			csArgs.TextFormat = tfArgs
		} else if ids.CloudStorage.GetAvroFormat() != nil {
			csArgs.AvroFormat = &pubsub.TopicIngestionDataSourceSettingsCloudStorageAvroFormatArgs{}
		} else if ids.CloudStorage.GetPubsubAvroFormat() != nil {
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
			GcpServiceAccount: pulumi.String(ids.ConfluentCloud.GcpServiceAccount.GetValue()),
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
