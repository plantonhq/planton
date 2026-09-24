# GcpBigQueryReservationGroup — Terraform Implementation

This directory contains the Terraform implementation for a BigQuery reservation group from the Planton spec: one `google_project_service` (API enablement) and one `google_bigquery_reservation_group`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the name defaulted from `metadata.name`, null-for-empty optionals |
| `main.tf` | `google_project_service`, `google_bigquery_reservation_group` |
| `outputs.tf` | `name` (the group's id, the full path), `reservation_group_name`, `location` |

## Send Posture

- **`name`** -- `spec.reservation_group_name`, defaulting to `metadata.name`.
- **`location`** -- sent only when set (Google's default is `US`).
- **`name` output** -- the resource's id, the full path.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
