# GcpManagedKafkaTopic — Terraform Implementation

This directory contains the Terraform implementation for a Managed Kafka topic from the Planton spec: one `google_managed_kafka_topic`. The cluster enabled the API, so the topic enables nothing.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the bare cluster id derived from a full path or bare id, the topic name defaulted from `metadata.name`, null-for-empty optionals |
| `main.tf` | `google_managed_kafka_topic` |
| `outputs.tf` | `name`, `topic_id` |

## Send Posture

- **`cluster`** -- the last path segment of the reference or literal (Google's resource takes the bare id).
- **`topic_id`** -- `spec.topic_id`, defaulting to `metadata.name`.
- **`partition_count`, `configs`** -- sent only when set, so Google's and the cluster's defaults apply otherwise.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
