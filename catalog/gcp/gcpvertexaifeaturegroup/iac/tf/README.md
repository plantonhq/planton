# GcpVertexAiFeatureGroup — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI Feature Store feature group and its folded features from the Planton spec: one `google_project_service` (API enablement), one `google_vertex_ai_feature_group`, and one `google_vertex_ai_feature_group_feature` per `spec.features[]` entry (`for_each` keyed by `feature_id`).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, null-for-empty optionals, the `bq://`-prefixed source URI, the `planton-ai_*` labels, the feature map |
| `main.tf` | `google_project_service`, `google_vertex_ai_feature_group`, `google_vertex_ai_feature_group_feature` |
| `outputs.tf` | `name`, `feature_group_id`, `location`, `feature_names` |

## Send Posture

- **`region`** -- the spec's `location`; **`name`** -- the spec's `feature_group_id`.
- **`big_query`** -- emitted when the spec declares a source; the spec's lifted `input_uri` goes back under `big_query_source`, prefixed with `bq://` when missing -- PARITY with the Pulumi module; `entity_id_columns` sent only when non-empty.
- **Features** -- `feature_group` from the created group, `region` from the spec; `version_column_name` sent only when set (Optional+Computed); `labels` merge the feature's own under the attribution set; `deletion_policy` fanned from the spec.
- **`name` output** -- the group's `id` (the provider's `name` attribute is the short id); **`feature_names`** -- the features' `id`s in manifest order.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
