# GcpVertexAiFeatureOnlineStore — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI Feature Store online store and its folded feature views from the Planton spec: one `gcp.projects.Service` (API enablement), one `gcp.vertex.AiFeatureOnlineStore`, and one `gcp.vertex.AiFeatureOnlineStoreFeatureview` per `spec.feature_views[]` entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `featureOnlineStore` |
| `module/locals.go` | The `planton-ai_*` attribution labels |
| `module/feature_online_store.go` | Enables the API; maps the store, its storage, endpoint, and key; exports the outputs |
| `module/feature_views.go` | Maps every feature view with its source and sync config |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`Region`** -- the spec's `location`; **`Name`** -- the spec's `feature_online_store_id`.
- **Storage** -- `Bigtable` emitted when set, with `Zone` and `CpuUtilizationTarget` only when set and `EnableDirectBigtableAccess` only when true; the spec's `optimized` bool becomes Google's empty `Optimized` block.
- **`DedicatedServingEndpoint`** -- emitted only when the spec shapes it; **`ForceDestroy`** sent as declared.
- **Feature views** -- `FeatureOnlineStore` from the created store; the BigQuery `Uri` prefixed with `bq://` when missing; `ProjectNumber` only when set; `SyncConfig.Cron` only when set and `Continuous` only when true; `Labels` are the view's own under the attribution set; `DeletionPolicy` fanned from the spec; each view parented to the store.
- **Outputs** -- `name` is the store's resource ID; the endpoint outputs read the computed block (empty on a Bigtable store); `feature_view_names` in manifest order.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
