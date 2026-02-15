# GcpPubSubTopic -- Pulumi Architecture Overview

## Execution Flow

```
StackInput (GcpPubSubTopicStackInput)
  |
  +-- target: GcpPubSubTopic (api.proto envelope)
  |     +-- metadata: CloudResourceMetadata
  |     +-- spec: GcpPubSubTopicSpec
  |           +-- project_id (StringValueOrRef -> GcpProject)
  |           +-- topic_name
  |           +-- kms_key_name (StringValueOrRef -> GcpKmsKey)
  |           +-- message_retention_duration
  |           +-- message_storage_policy
  |           |     +-- allowed_persistence_regions[]
  |           |     +-- enforce_in_transit
  |           +-- schema_settings
  |           |     +-- schema
  |           |     +-- encoding
  |           +-- ingestion_data_source_settings
  |                 +-- aws_kinesis { stream_arn, consumer_arn, aws_role_arn, gcp_service_account }
  |                 +-- aws_msk { cluster_arn, topic, aws_role_arn, gcp_service_account }
  |                 +-- azure_event_hubs { resource_group, namespace, event_hub, ... }
  |                 +-- cloud_storage { bucket (StringValueOrRef), match_glob, format }
  |                 +-- confluent_cloud { bootstrap_server, topic, identity_pool_id, ... }
  |                 +-- platform_logs_settings { severity }
  |
  +-- provider_config: GcpProviderConfig

  v module.Resources()

  1. initializeLocals() -> Locals { GcpLabels, spec ref }
  2. pulumigoogleprovider.Get() -> gcp.Provider
  3. topic() -> pubsub.NewTopic
       +-- Maps core fields (Name, Project, Labels, KmsKeyName, MessageRetentionDuration)
       +-- Conditionally sets MessageStoragePolicy
       +-- Conditionally sets SchemaSettings
       +-- Calls ingestionDataSourceSettings() helper for complex ingestion mapping
       +-- Exports topic_id (resource ID), topic_name
```

## Resource Mapping

| Spec Field | Pulumi Property | Notes |
|------------|-----------------|-------|
| `project_id` | `Project` | From StringValueOrRef.GetValue() |
| `topic_name` | `Name` | Required, immutable |
| `kms_key_name` | `KmsKeyName` | Optional CMEK |
| `message_retention_duration` | `MessageRetentionDuration` | Duration string |
| `message_storage_policy` | `MessageStoragePolicy` | Region constraints |
| `schema_settings` | `SchemaSettings` | Schema + encoding |
| `ingestion_data_source_settings` | `IngestionDataSourceSettings` | 5 source types |
| (framework) | `Labels` | Computed from metadata |

## Ingestion Mapping Detail

The `ingestionDataSourceSettings()` helper function maps each non-nil source
from the proto struct to the corresponding Pulumi args type:

- Nil-checks each source type (AwsKinesis, AwsMsk, AzureEventHubs, CloudStorage, ConfluentCloud)
- Cloud Storage: `bucket` uses `.GetValue()` on StringValueOrRef
- Cloud Storage format: nil-checks AvroFormat/PubsubAvroFormat/TextFormat marker messages
- Azure Event Hubs: all fields are optional, conditionally mapped with `pulumi.StringPtr()`
- Returns nil if no sources are configured (avoids sending empty args)
