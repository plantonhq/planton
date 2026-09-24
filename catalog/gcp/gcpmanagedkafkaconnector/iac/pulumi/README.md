# GcpManagedKafkaConnector — Pulumi Implementation

This directory contains the Pulumi implementation for a Managed Kafka connector from the Planton spec: one `gcp.managedkafka.Connector`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `connector` |
| `module/locals.go` | The bare Connect cluster id and the defaulted connector id |
| `module/connector.go` | Maps the connector; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`connect_cluster`** -- the last path segment of the reference or literal.
- **`connector_id`** -- defaulting to `metadata.name`.
- **`configs`** -- sent only when set.
- **`task_restart_policy`** -- emitted when declared; each backoff sent only when set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
