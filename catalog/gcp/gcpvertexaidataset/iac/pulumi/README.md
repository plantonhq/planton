# GcpVertexAiDataset — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI managed dataset from the Planton spec: one `gcp.projects.Service` (API enablement) and one `gcp.vertex.AiDataset`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `dataset` |
| `module/locals.go` | The display name defaulted from `metadata.name`; the `planton-ai_*` attribution labels |
| `module/dataset.go` | Enables the API; maps the dataset; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `dataset_id`, `location`) |

## Send Posture (parity with Terraform)

- **`Region`** -- the spec's `location`.
- **`DisplayName`** -- `spec.display_name`, defaulting to `metadata.name`.
- **`EncryptionSpec`** -- emitted only when `kms_key_name` is set.
- **`DeletionPolicy`** -- sent only when set.
- **`dataset_id`** -- the last segment of the computed `Name`.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
