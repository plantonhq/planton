# GcpPubSubTopic -- Pulumi Module

This directory contains the Pulumi Go implementation for the GcpPubSubTopic component.

## Module Structure

```
module/
  main.go       -- Entry point: creates GCP provider, orchestrates resources
  locals.go     -- Locals struct, GCP label computation
  topic.go      -- Creates pubsub.Topic with all field mappings + ingestion helper
  outputs.go    -- Output key constants

main.go         -- Pulumi program entrypoint (loads stack input, calls module)
Pulumi.yaml     -- Pulumi project configuration
Makefile        -- Build, preview, up, destroy targets
```

## Outputs

| Key | Description |
|-----|-------------|
| `topic_id` | Fully qualified topic ID (projects/{project}/topics/{name}) |
| `topic_name` | Short topic name |

## Local Development

```bash
make build      # Compile the Pulumi binary
make preview     # Preview changes
make up          # Apply changes
make destroy     # Destroy resources
```

## Notes

- Pub/Sub topics **support GCP labels**. Framework labels are applied automatically.
- CMEK encryption is set via `KmsKeyName` on the topic resource.
- Message storage policy maps `AllowedPersistenceRegions` to `pulumi.ToStringArray()`.
- Schema settings are conditionally set only when `schema` is non-empty.
- Ingestion data source settings are mapped via a dedicated `ingestionDataSourceSettings()`
  helper function that nil-checks each source type and builds the corresponding
  Pulumi args struct. Cloud Storage's `bucket` field uses `StringValueOrRef.GetValue()`.
- Cloud Storage format selection uses Go nil-checks on the marker messages
  (AvroFormat, PubsubAvroFormat) or the TextFormat sub-message.
