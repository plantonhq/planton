# GcpVertexAiDataset — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI managed dataset from the Planton spec: one `google_project_service` (API enablement) and one `google_vertex_ai_dataset`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the display name defaulted from `metadata.name`, null-for-empty optionals, the `planton-ai_*` labels |
| `main.tf` | `google_project_service`, `google_vertex_ai_dataset` |
| `outputs.tf` | `name`, `dataset_id`, `location` |

## Send Posture

- **`region`** -- the spec's `location` (the provider names the axis `region`).
- **`display_name`** -- `spec.display_name`, defaulting to `metadata.name` (Google requires one) -- PARITY with the Pulumi module.
- **`encryption_spec`** -- emitted only when `kms_key_name` is set.
- **`deletion_policy`** -- sent only when set.
- **`dataset_id`** -- the last segment of the computed `name`, the same split the Pulumi module performs.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
