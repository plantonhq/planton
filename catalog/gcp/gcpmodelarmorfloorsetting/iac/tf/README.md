# GcpModelArmorFloorSetting — Terraform Implementation

This directory contains the Terraform implementation for a Model Armor floor setting from the Planton spec: an optional `google_project_service` (API enablement, project floors) and one `google_model_armor_floorsetting`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Scope selection into Google's parent, the `global` location default |
| `main.tf` | `data.google_client_config` (empty scope only), `google_project_service` (project floors), `google_model_armor_floorsetting` |
| `outputs.tf` | `name`, `parent` |

## Send Posture

- **`parent`** -- rendered from the one scope arm set; an empty scope reads the provider's project from `google_client_config` (provider configuration, no API call) -- PARITY with the Pulumi module's `GetClientConfig`.
- **`location`** -- `spec.location`, defaulting to `global`.
- **`enable_floor_setting_enforcement`** -- sent as declared.
- **`inspect_only` / `inspect_and_block`** -- exactly the flag `enforcement_type` names is sent; the other is null.
- **`floor_setting_metadata`** -- sent only when `enable_multi_language_detection` is true.
- **Destroy** -- the provider removes the floor from state only; Google keeps it.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
