# GcpPubSubTopic - Terraform Module

This Terraform/OpenTofu module provisions a Pub/Sub topic (`google_pubsub_topic`) — the named channel publishers send to. It is the Terraform-side implementation of the Planton `GcpPubSubTopic` resource kind and has feature parity with the Pulumi module.

## Overview

The module covers the full released topic surface: CMEK encryption, topic-level message retention, regional storage policy with in-transit enforcement, schema validation (via a `GcpPubSubSchema` reference), all five ingestion data sources (Kinesis, MSK, Azure Event Hubs, Cloud Storage, Confluent Cloud) with platform logging, and the ordered JavaScript UDF transform pipeline.

The module enables the Pub/Sub API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. An empty `project_id` falls back to the provider's default project, identically to the Pulumi module. User labels merge beneath the platform attribution labels (`planton-ai_*`), which win on key conflicts — identical label sets on both engines.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `provider.tf` — google provider pin (`~> 8.3`; all fields GA on the released line)
- `variables.tf` — the converter-contract `metadata`/`spec` variables
- `locals.tf` — ambient-project resolution + the merged label map
- `main.tf` — API enablement + the topic resource
- `outputs.tf` — stack outputs (must match `outputs.proto`)

## Inputs (high level)

| Variable | Description |
|----------|-------------|
| `spec.project_id` | GCP project (empty = provider default project) |
| `spec.topic_name` | Topic resource name (ForceNew) |
| `spec.kms_key_name` | CMEK key path (resolved from a GcpKmsKey reference) |
| `spec.message_retention_duration` | Topic-level retention window |
| `spec.labels` | User labels (merged beneath platform labels) |
| `spec.message_storage_policy` | Region pinning + in-transit enforcement |
| `spec.schema_settings` | Schema validation (schema resolved from a GcpPubSubSchema reference) |
| `spec.ingestion_data_source_settings` | One of five external ingestion sources |
| `spec.message_transforms` | Ordered JavaScript UDF pipeline |

## Outputs

| Name | Description |
|------|-------------|
| `topic_id` | The fully qualified topic path (`projects/{project}/topics/{name}`) — what subscriptions, event triggers, and Scheduler targets consume |
| `topic_name` | The short topic name |

## Lifecycle Notes

`topic_name` and `project` are ForceNew — renaming replaces the topic and orphans its subscriptions, so treat names as permanent. Everything else (CMEK key, retention, storage policy, schema settings, ingestion, transforms, labels) updates in place. For CMEK, grant the Pub/Sub service agent `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key before deploying, or publishes fail.
