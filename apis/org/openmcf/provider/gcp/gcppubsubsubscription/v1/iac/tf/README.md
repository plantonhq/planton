# GcpPubSubSubscription -- Terraform Module

This directory contains the Terraform implementation for the GcpPubSubSubscription component.

## Module Structure

```
provider.tf    -- Google provider configuration (~> 6.0)
variables.tf   -- Input variables matching GcpPubSubSubscriptionSpec
locals.tf      -- Derived values from spec
main.tf        -- google_pubsub_subscription resource with dynamic blocks
outputs.tf     -- Outputs matching GcpPubSubSubscriptionStackOutputs
```

## What It Creates

- `google_pubsub_subscription.this` -- A Pub/Sub subscription with support for
  pull, push, BigQuery, and Cloud Storage delivery methods, plus dead-letter
  handling, retry policy, expiration policy, filtering, message ordering, and
  exactly-once delivery.

## Inputs

| Variable | Type | Description |
|----------|------|-------------|
| `spec.project_id.value` | `string` | GCP project ID |
| `spec.subscription_name` | `string` | Subscription name (3-255 chars, immutable) |
| `spec.topic.value` | `string` | Source topic (fully qualified) |
| `spec.ack_deadline_seconds` | `number` | Ack deadline: 10-600s |
| `spec.message_retention_duration` | `string` | Backlog retention duration |
| `spec.retain_acked_messages` | `bool` | Keep acked messages for replay |
| `spec.expiration_policy` | `object` | Auto-delete inactive subscriptions |
| `spec.filter` | `string` | Attribute filter (max 256 bytes) |
| `spec.enable_message_ordering` | `bool` | FIFO by ordering key |
| `spec.enable_exactly_once_delivery` | `bool` | Exactly-once guarantees |
| `spec.dead_letter_policy` | `object` | Dead-letter topic and max attempts |
| `spec.retry_policy` | `object` | Backoff between retries |
| `spec.push_config` | `object` | Push delivery (HTTPS, OIDC, no_wrapper) |
| `spec.bigquery_config` | `object` | BigQuery delivery |
| `spec.cloud_storage_config` | `object` | Cloud Storage delivery |
| `provider_config` | `object` | Optional GCP SA key |

## Outputs

| Output | Description |
|--------|-------------|
| `subscription_id` | Fully qualified ID (projects/{project}/subscriptions/{name}) |
| `subscription_name` | Short subscription name |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Important Notes

- **Subscription name is immutable.** Changing it triggers destroy and recreate.
- **Filter is immutable.** Cannot be changed after creation.
- **Delivery methods are mutually exclusive.** Only one of push_config,
  bigquery_config, or cloud_storage_config can be set.
- **Dead-letter IAM setup required.** The Pub/Sub service account needs
  Subscriber on this subscription and Publisher on the dead-letter topic.
- **Provider version `~> 6.0`.** Cloud Storage subscription features
  (`max_messages`, `avro_config.use_topic_schema`) require Google provider v6.x.
