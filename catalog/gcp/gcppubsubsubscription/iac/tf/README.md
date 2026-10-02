# GcpPubSubSubscription - Terraform Module

This Terraform/OpenTofu module provisions a Pub/Sub subscription (`google_pubsub_subscription`) — the named message stream from one topic to one consumer. It is the Terraform-side implementation of the Planton `GcpPubSubSubscription` resource kind and has feature parity with the Pulumi module.

## Overview

The module covers all four delivery modes — pull (default), push with OIDC/no-wrapper, BigQuery (table resolved from a `GcpBigQueryTable` reference), and Cloud Storage with Avro — plus dead-lettering, retry backoff, expiration policy, exactly-once delivery, ordering, filtering, and the ordered JavaScript UDF transform pipeline.

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
- `main.tf` — API enablement + the subscription resource
- `outputs.tf` — stack outputs (must match `outputs.proto`)

## Inputs (high level)

| Variable | Description |
|----------|-------------|
| `spec.project_id` | GCP project (empty = provider default project) |
| `spec.subscription_name` | Subscription resource name (ForceNew) |
| `spec.topic` | Parent topic path (resolved from a GcpPubSubTopic reference; ForceNew) |
| `spec.push_config` / `spec.bigquery_config` / `spec.cloud_storage_config` | At most one delivery mode (spec CEL enforces exclusivity); none = pull |
| `spec.dead_letter_policy` / `spec.retry_policy` / `spec.expiration_policy` | Reliability levers |
| `spec.labels` | User labels (merged beneath platform labels) |
| `spec.message_transforms` | Ordered JavaScript UDF pipeline |

## Outputs

| Name | Description |
|------|-------------|
| `subscription_id` | The fully qualified subscription path (`projects/{project}/subscriptions/{name}`) |
| `subscription_name` | The short subscription name |

## Lifecycle Notes

`subscription_name`, `topic`, `filter`, `enable_message_ordering`, and `project` are ForceNew — changing any of them replaces the subscription and its backlog. Delivery configs, deadlines, retention, dead-lettering, retries, expiration, transforms, and labels all update in place. For dead-lettering, the Pub/Sub service agent needs Subscriber on this subscription and Publisher on the dead-letter topic.
