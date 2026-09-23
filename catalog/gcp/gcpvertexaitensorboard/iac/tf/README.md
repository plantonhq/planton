# GcpVertexAiTensorboard — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI TensorBoard and its folded experiments and runs from the Planton spec: one `google_project_service` (API enablement), one `google_vertex_ai_tensorboard`, one `google_vertex_ai_tensorboard_experiment` per `spec.experiments[]` entry (`for_each` keyed by `experiment_id`), and one `google_vertex_ai_tensorboard_run` per `spec.experiments[].runs[]` entry (`for_each` keyed by `"{experiment_id}/{run_id}"`).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the display name defaulted from `metadata.name`, null-for-empty optionals, the `planton-ai_*` labels, the numeric TensorBoard id, the experiment and run maps |
| `main.tf` | `google_project_service`, `google_vertex_ai_tensorboard`, `google_vertex_ai_tensorboard_experiment`, `google_vertex_ai_tensorboard_run` |
| `outputs.tf` | `name`, `tensorboard_id`, `location`, `blob_storage_path_prefix`, `experiment_names`, `run_names` |

## Send Posture

- **`region`** -- the spec's `location` (the provider names the axis `region`); experiments and runs take it as `location`.
- **`display_name`** -- `spec.display_name`, defaulting to `metadata.name` (Google requires one) -- PARITY with the Pulumi module.
- **`encryption_spec`** -- emitted only when `kms_key_name` is set.
- **Experiments and runs** -- `tensorboard` is the last segment of the TensorBoard's computed `name`; a run's `experiment` is its experiment's id; a run's `display_name` defaults to its `run_id`; `labels` merge the child's own under the attribution set; `deletion_policy` fanned from the spec.
- **`experiment_names` / `run_names`** -- in manifest order (comprehensions over `spec.experiments`, not over the `for_each` maps), so both engines export the same lists.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
