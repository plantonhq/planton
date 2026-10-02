# GcpBigQueryConnection — Pulumi Implementation

This directory contains the Pulumi implementation for a BigQuery connection from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.bigquery.Connection`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `connection` |
| `module/locals.go` | The defaulted connection id |
| `module/connection.go` | Enables the API; maps the one arm; exports the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`connection_id`** -- defaulting to `metadata.name`.
- **`cloud_resource`** -- the spec's bool emits Google's empty marker block.
- **Lifted wrappers** -- `aws.iam_role_id`, `spark.metastore_service`, `spark.history_server_dataproc_cluster`, and the configuration arm's `username_password`, `host_port`, and `network_attachment`, each block sent only when set.
- **Spanner flags** -- sent only when true.
- **Passwords** -- marked secret in Pulumi (`ToSecret`); sensitive in the provider.
- **Arm identities** -- every output always present, empty for arms not declared.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
