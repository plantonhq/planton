# GcpPubSubTopic -- Terraform Module

This directory contains the Terraform implementation for the GcpPubSubTopic component.

## Module Structure

```
provider.tf    -- Google provider configuration
variables.tf   -- Input variables matching GcpPubSubTopicSpec
locals.tf      -- Derived values from spec
main.tf        -- google_pubsub_topic resource with dynamic blocks
outputs.tf     -- Outputs matching GcpPubSubTopicStackOutputs
```

## What It Creates

- `google_pubsub_topic.this` -- A Pub/Sub topic with optional CMEK encryption,
  message retention, regional storage policy, schema validation, and data
  ingestion from external sources.

## Inputs

| Variable | Type | Description |
|----------|------|-------------|
| `spec.project_id.value` | `string` | GCP project ID |
| `spec.topic_name` | `string` | Topic name (3-255 chars) |
| `spec.kms_key_name.value` | `string` | CMEK key name (optional) |
| `spec.message_retention_duration` | `string` | Retention duration, e.g. "604800s" |
| `spec.message_storage_policy` | `object` | Regional storage constraints |
| `spec.schema_settings` | `object` | Schema validation settings |
| `spec.ingestion_data_source_settings` | `object` | External source ingestion |
| `provider_config.service_account_key_base64` | `string` | Optional GCP SA key |

## Outputs

| Output | Description |
|--------|-------------|
| `topic_id` | Fully qualified topic ID (projects/{project}/topics/{name}) |
| `topic_name` | Short topic name |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Important Notes

- **Topic name is immutable.** Changing `topic_name` triggers destroy and recreate.
- **Labels are managed by Planton framework.** Framework labels are applied automatically.
- **Ingestion sources** are mutually exclusive in practice (one source per topic),
  though the API does not enforce this at the schema level.
- **CMEK requires IAM setup.** The Pub/Sub service account
  (`service-{PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com`) must have
  `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the specified key.
