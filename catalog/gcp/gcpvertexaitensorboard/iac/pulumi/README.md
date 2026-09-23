# GcpVertexAiTensorboard — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI TensorBoard and its folded experiments and runs from the Planton spec: one `gcp.projects.Service` (API enablement), one `gcp.vertex.AiTensorboard`, one `gcp.vertex.AiTensorboardExperiment` per `spec.experiments[]` entry, and one `gcp.vertex.AiTensorboardRun` per `spec.experiments[].runs[]` entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `tensorboard` |
| `module/locals.go` | The display name defaulted from `metadata.name`; the `planton-ai_*` attribution labels |
| `module/tensorboard.go` | Enables the API; maps the TensorBoard, every experiment, and every run; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `tensorboard_id`, `location`, `blob_storage_path_prefix`, `experiment_names`, `run_names`) |

## Send Posture (parity with Terraform)

- **`Region`** -- the spec's `location`; experiments and runs take it as `Location`.
- **`DisplayName`** -- `spec.display_name`, defaulting to `metadata.name`.
- **`EncryptionSpec`** -- emitted only when `kms_key_name` is set.
- **Experiments and runs** -- `Tensorboard` is the last segment of the TensorBoard's computed `Name`; a run's `Experiment` is its created experiment's id; a run's `DisplayName` defaults to its `run_id`; `Labels` are the child's own under the attribution set; `DeletionPolicy` fanned from the spec; each child is parented to its TensorBoard or experiment.
- **`experiment_names` / `run_names`** -- the created children's ids in manifest order.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
