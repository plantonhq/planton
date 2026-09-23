# GcpColabRuntime — Terraform Implementation

This directory contains the Terraform implementation for a Colab Enterprise runtime from the Planton spec: one `google_project_service` (API enablement) and one `google_colab_runtime`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the runtime id and display name defaulted from `metadata.name`, null-for-empty optionals |
| `main.tf` | `google_project_service`, `google_colab_runtime` |
| `outputs.tf` | `name`, `runtime_id`, `location` |

## Send Posture

- **`name`** -- `spec.runtime_id`, defaulting to `metadata.name`, always sent (the provider never reads it back) -- PARITY with the Pulumi module.
- **`notebook_runtime_template_ref`** -- always rendered from `spec.runtime_template`.
- **`desired_state`** -- sent only when set (the provider starts or stops to match); **`auto_upgrade`** sent as declared.
- **`description`**, **`deletion_policy`** -- sent only when set.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
