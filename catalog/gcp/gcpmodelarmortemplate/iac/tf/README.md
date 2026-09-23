# GcpModelArmorTemplate — Terraform Implementation

This directory contains the Terraform implementation for a Model Armor template from the Planton spec: one `google_project_service` (API enablement) and one `google_model_armor_template`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the template id defaulted from `metadata.name`, the `planton-ai_*` labels |
| `main.tf` | `google_project_service`, `google_model_armor_template` |
| `outputs.tf` | `name`, `template_id`, `location` |

## Send Posture

- **`template_id`** -- `spec.template_id`, defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **Filter blocks** -- each rendered only when its spec message is set; every optional string inside is sent only when set.
- **`template_metadata`** -- rendered only when set; the three bools are sent as declared, codes and strings only when set.
- **`multi_language_detection`** -- the spec's lifted `enable_multi_language_detection`; the block is sent only when true.
- **`deletion_policy`** -- sent only when set.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
