# GcpVertexAiFeatureOnlineStore — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI Feature Store online store and its folded feature views from the Planton spec: one `google_project_service` (API enablement), one `google_vertex_ai_feature_online_store`, and one `google_vertex_ai_feature_online_store_featureview` per `spec.feature_views[]` entry (`for_each` keyed by `feature_view_id`).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, null-for-empty optionals, the `planton-ai_*` labels, the feature view map |
| `main.tf` | `google_project_service`, `google_vertex_ai_feature_online_store`, `google_vertex_ai_feature_online_store_featureview` |
| `outputs.tf` | `name`, `feature_online_store_id`, `location`, `public_endpoint_domain_name`, `service_attachment`, `feature_view_names` |

## Send Posture

- **`region`** -- the spec's `location`; **`name`** -- the spec's `feature_online_store_id`.
- **Storage** -- `bigtable` emitted when set, with `zone` and `cpu_utilization_target` sent only when set (Optional+Computed) and `enable_direct_bigtable_access` only when true; `optimized {}` emitted from the spec's bool -- PARITY with the Pulumi module.
- **`dedicated_serving_endpoint`** -- Optional+Computed; emitted only when the spec shapes it.
- **`encryption_spec`** -- emitted only when `kms_key_name` is set; **`force_destroy`** sent as declared.
- **Feature views** -- `feature_online_store` from the created store; the BigQuery `uri` prefixed with `bq://` when missing; `project_number` sent only when set; `sync_config.cron` sent only when set and `continuous` only when true; `labels` merge the view's own under the attribution set; `deletion_policy` fanned from the spec.
- **Outputs** -- `name` is the store's `id` (the provider's `name` attribute is the short id); the endpoint outputs read the computed block with `try(...)` so a Bigtable store exports empty strings; `feature_view_names` in manifest order.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
