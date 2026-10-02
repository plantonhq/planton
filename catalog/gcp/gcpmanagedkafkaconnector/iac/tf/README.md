# GcpManagedKafkaConnector — Terraform Implementation

This directory contains the Terraform implementation for a Managed Kafka connector from the Planton spec: one `google_managed_kafka_connector`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the bare Connect cluster id, the connector id defaulted from `metadata.name` |
| `main.tf` | `google_managed_kafka_connector` |
| `outputs.tf` | `name`, `connector_id` |

## Send Posture

- **`connect_cluster`** -- the last path segment of the reference or literal.
- **`connector_id`** -- defaulting to `metadata.name`.
- **`configs`** -- sent only when set.
- **`task_restart_policy`** -- emitted when declared; each backoff sent only when set.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
