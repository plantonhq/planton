# GcpSccBigQueryExport — Terraform Implementation

This directory contains the Terraform implementation for a Security Command Center BigQuery export from the Planton spec: one of `google_scc_v2_project_scc_big_query_export`, `google_scc_v2_folder_scc_big_query_export`, or `google_scc_v2_organization_scc_big_query_export`, selected by the scope, plus `google_project_service` for a project export.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Scope selection (`is_project` / `is_folder` / `is_org`), the bare folder ID, the `global` location default, the trimmed dataset, the composed organization export name |
| `main.tf` | `data.google_client_config` (empty scope only), `google_project_service` (project exports), the three scope resources with `count` |
| `outputs.tf` | `name`, `principal` |

## Send Posture

- **Scope** -- exactly one resource is created; the project arm, or an empty scope (the provider's project from `google_client_config`, provider configuration, no API call), is the project resource -- PARITY with the Pulumi module's switch and `GetClientConfig`.
- **`location`** -- `spec.location`, defaulting to `global`.
- **`dataset`** -- the spec's dataset with a self link's `https://bigquery.googleapis.com/bigquery/v2/` prefix trimmed.
- **`filter`** -- sent when set.
- **`name`** (organization resource) -- always composed: `organizations/{org}/locations/{location}/bigQueryExports/{id}`.
- **`description`**, **`deletion_policy`** -- sent only when set.
- **API** -- `securitycenter.googleapis.com` on a project export's project; `disable_on_destroy = false`.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
