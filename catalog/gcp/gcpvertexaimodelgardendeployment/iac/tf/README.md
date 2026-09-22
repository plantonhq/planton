# GcpVertexAiModelGardenDeployment — Terraform Implementation

This directory contains the Terraform implementation for a one-step Model
Garden deployment from the Planton spec: one `google_project_service` (API
enablement) and one `google_vertex_ai_endpoint_with_model_garden_deployment`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback and null-for-empty optionals |
| `main.tf` | `google_project_service`, `google_vertex_ai_endpoint_with_model_garden_deployment` |
| `outputs.tf` | `endpoint_id`, `endpoint_name`, `deployed_model_id`, `deployed_model_display_name`, `location` |

## Send Posture

- **Model source** -- exactly one of `publisher_model_name` / `hugging_face_model_id`, the other null -- PARITY with the Pulumi module.
- **Optional strings and booleans** -- null when empty; `accept_eula` and `spot` sent only when true.
- **`dedicated_resources`** -- `min_replica_count` always; `max_replica_count` and `required_replica_count` (Optional+Computed) passed through as null when unset; `machine_spec` always emitted (the API requires the block) with each field null when unset.
- **`shared_memory_size_mb`** -- `tostring()` of the spec's number, the string Google's API carries.
- **`psc_automation_configs`** -- one block from the spec's single `psc_automation_config`.
- **Probes** -- the three probes render the same spec shape; each handler is a `dynamic` block emitted when its arm is set.
- **`endpoint_id`** -- rebuilt from `project`, `location`, and `endpoint`, the shape `GcpVertexAiEndpoint` exports.
- **`deletion_policy`** -- null when empty.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
