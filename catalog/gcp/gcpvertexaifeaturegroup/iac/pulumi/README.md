# GcpVertexAiFeatureGroup — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI Feature Store feature group and its folded features from the Planton spec: one `gcp.projects.Service` (API enablement), one `gcp.vertex.AiFeatureGroup`, and one `gcp.vertex.AiFeatureGroupFeature` per `spec.features[]` entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `featureGroup` |
| `module/locals.go` | The `planton-ai_*` attribution labels |
| `module/feature_group.go` | Enables the API; maps the group, its BigQuery source, and every feature; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `feature_group_id`, `location`, `feature_names`) |

## Send Posture (parity with Terraform)

- **`Region`** -- the spec's `location`; **`Name`** -- the spec's `feature_group_id`.
- **`BigQuery`** -- emitted when the spec declares a source; `BigQuerySource.InputUri` prefixed with `bq://` when missing; `EntityIdColumns` sent only when non-empty.
- **Features** -- `FeatureGroup` from the created group; `VersionColumnName` sent only when set (Optional+Computed); `Labels` are the feature's own under the attribution set; `DeletionPolicy` fanned from the spec; each feature parented to the group.
- **`name` output** -- the group's resource ID (the SDK's `Name` is the short id); **`feature_names`** -- the features' IDs in manifest order.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
