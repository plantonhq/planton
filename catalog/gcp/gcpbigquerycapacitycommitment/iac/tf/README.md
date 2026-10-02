# GcpBigQueryCapacityCommitment — Terraform Implementation

This directory contains the Terraform implementation for a BigQuery capacity commitment from the Planton spec: one `google_project_service` (API enablement) and one `google_bigquery_capacity_commitment`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the id defaulted from `metadata.name`, null-for-empty optionals, the single-admin-project guard mapped to the provider's string |
| `main.tf` | `google_project_service`, `google_bigquery_capacity_commitment` |
| `outputs.tf` | `name`, `state`, `commitment_start_time`, `commitment_end_time` |

## Send Posture

- **`capacity_commitment_id`** -- defaulting to `metadata.name` rather than a Google-generated id.
- **`enforce_single_admin_project_per_org`** -- the spec's bool sent as the provider's string `"true"` when set, omitted otherwise.
- **`renewal_plan`, `edition`, `location`** -- sent only when set.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
