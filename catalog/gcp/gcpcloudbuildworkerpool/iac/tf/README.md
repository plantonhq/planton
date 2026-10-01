# GcpCloudBuildWorkerPool — Terraform Implementation

This directory contains the Terraform implementation for a Cloud Build private worker pool from the Planton spec: `google_project_service` for the Cloud Build API and one `google_cloudbuild_worker_pool`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Project, pool ID default, the peered network's project-number path |
| `main.tf` | `google_project_service`, the guarded project lookup, the pool |
| `outputs.tf` | `name`, `worker_pool_id`, `state`, `uid` |

## Send Posture

- **Optional fields** -- sent only when set, so Cloud Build's defaults apply otherwise; the two worker booleans are sent whenever declared, including `false`.
- **Peered network** -- the self-link prefix stripped and a project ID resolved to its number through one `data.google_project` read (none for a numeric path) -- PARITY with the Pulumi module's `resolvePeeredNetwork`.
- **No labels** -- the pool has annotations only, so no attribution labels are applied.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
