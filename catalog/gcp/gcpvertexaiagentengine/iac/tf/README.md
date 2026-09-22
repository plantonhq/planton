# GcpVertexAiAgentEngine — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI Agent
Engine instance from the Planton spec: one `google_project_service` (API
enablement) and one `google_vertex_ai_reasoning_engine`.

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the display name defaulted from `metadata.name`, null-for-empty optionals, the `planton-ai_*` labels |
| `main.tf` | `google_project_service`, `google_vertex_ai_reasoning_engine` (the agent `spec` and the memory bank as nested `dynamic` blocks) |
| `outputs.tf` | `name`, `reasoning_engine_id`, `location`, `create_time`, `update_time` |

## Send Posture

- **`region`** -- the spec's `location` -- PARITY with the Pulumi module.
- **`display_name`** -- `spec.display_name`, defaulting to `metadata.name`.
- **Optional strings, booleans, maps** -- null when empty; Optional+Computed numbers passed through as null when unset.
- **`source_code_spec`** -- exactly one source and one build recipe (proto-enforced), each a `dynamic` block emitted when set.
- **`build_spec`** -- `worker_pool` only; `service_account` is an SDK gap held out on both engines.
- **Memory-bank example parts** -- every payload but `audio_transcription` (an SDK gap held out on both engines).
- **`name`** -- the resource id (the full path); `reasoning_engine_id` the provider's numeric `name` attribute.
- **`deletion_policy`** -- null when empty.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
