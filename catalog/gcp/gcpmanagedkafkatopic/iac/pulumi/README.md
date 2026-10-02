# GcpManagedKafkaTopic — Pulumi Implementation

This directory contains the Pulumi implementation for a Managed Kafka topic from the Planton spec: one `gcp.managedkafka.Topic`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `topic` |
| `module/locals.go` | The bare cluster id and the defaulted topic name |
| `module/topic.go` | Maps the topic; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `topic_id`) |

## Send Posture (parity with Terraform)

- **`cluster`** -- the last path segment of the reference or literal (Google's resource takes the bare id).
- **`topic_id`** -- `spec.topic_id`, defaulting to `metadata.name`.
- **`partition_count`, `configs`** -- sent only when set, so Google's and the cluster's defaults apply otherwise.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
