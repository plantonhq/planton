# GcpBigQueryConnection — Terraform Implementation

This directory contains the Terraform implementation for a BigQuery connection from the Planton spec: one `google_project_service` (API enablement) and one `google_bigquery_connection`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the connection id defaulted from `metadata.name`, null-for-empty optionals |
| `main.tf` | `google_project_service`, `google_bigquery_connection` with one dynamic arm block |
| `outputs.tf` | `name`, `connection_id`, `location`, and the arm identities (empty when not declared) |

## Send Posture

- **`connection_id`** -- defaulting to `metadata.name`.
- **`cloud_resource`** -- the spec's bool emits Google's empty marker block.
- **Lifted wrappers** -- `aws.iam_role_id`, `spark.metastore_service`, `spark.history_server_dataproc_cluster`, and the configuration arm's `username_password`, `host_port`, and `network_attachment`, each block sent only when set.
- **Spanner flags** -- sent only when true.
- **Passwords** -- marked secret in Pulumi (`ToSecret`); sensitive in the provider.
- **Arm identities** -- every output always present, empty for arms not declared.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
