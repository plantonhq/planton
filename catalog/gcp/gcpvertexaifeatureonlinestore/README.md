# GCP Vertex AI Feature Online Store

A Vertex AI Feature Store online store -- the low-latency serving layer that answers "give me this entity's features now" for models in production -- together with the feature views it serves. Storage is exactly one of two kinds: a managed Bigtable instance with autoscaling (large feature sets) or Optimized serving behind a dedicated endpoint (the lowest latency). Feature views select features from `GcpVertexAiFeatureGroup` blocks or materialize a BigQuery table directly, and sync on a schedule or continuously.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Online store** -- a `vertex_ai_feature_online_store` with Bigtable or Optimized storage, an optional Private Service Connect endpoint, and optional CMEK
- **Feature views** -- one `vertex_ai_feature_online_store_featureview` per `featureViews[]` entry, keyed by `featureViewId`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpVertexAiFeatureGroup`** -- the feature groups a registry-sourced feature view serves (`featureViews[].featureRegistrySource.featureGroups[].featureGroupId`), in the store's location.
- **`GcpBigQueryTable`** -- the table a BigQuery-sourced feature view materializes (`featureViews[].bigQuerySource.uri`).
- **`GcpKmsKey`** -- a key in the store's region for customer-managed encryption (`kmsKeyName`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiFeatureOnlineStore
metadata:
  name: serving-store
spec:
  location: us-central1
  featureOnlineStoreId: serving_store
  bigtable:
    autoScaling:
      minNodeCount: 1
      maxNodeCount: 3
  featureViews:
    - featureViewId: customer_view
      featureRegistrySource:
        featureGroups:
          - featureGroupId:
              valueFrom:
                kind: GcpVertexAiFeatureGroup
                name: customer-features
                fieldPath: status.outputs.feature_group_id
            featureIds:
              - age
              - lifetime_value
      syncConfig:
        cron: "0 */6 * * *"
```

```shell
planton apply -f vertex-ai-feature-online-store.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI region. Immutable. |
| `featureOnlineStoreId` | `string` | Up to 60 lowercase letters, digits, and underscores, not starting with a digit. Immutable. |
| `bigtable` or `optimized` | | Exactly one storage kind. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `labels` | `map<string,string>` | none | Merged under the attribution labels; mutable. |
| `bigtable.autoScaling` | `object` | -- | `minNodeCount` (at least 1), `maxNodeCount` (at least min, at most ten times it), `cpuUtilizationTarget` (10-80, Google default 50). |
| `bigtable.enableDirectBigtableAccess` | `bool` | `false` | Let clients read the managed Bigtable instance directly. |
| `bigtable.zone` | `string` | Google-chosen | The Bigtable instance's zone. |
| `optimized` | `bool` | `false` | Optimized serving behind a dedicated endpoint. |
| `dedicatedServingEndpoint.privateServiceConnectConfig` | `object` | public endpoint | `enablePrivateServiceConnect`, `projectAllowlist`. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. |
| `forceDestroy` | `bool` | `false` | Let a destroy delete a store that still holds views the block does not manage. |
| `featureViews[]` | `[]object` | none | `featureViewId` (same rule as the store id; immutable), `labels`, exactly one of `bigQuerySource { uri, entityIdColumns }` or `featureRegistrySource { featureGroups[] { featureGroupId, featureIds }, projectNumber }`, and `syncConfig { cron | continuous }`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; fanned to every view. |

### Validation Rules

- Exactly one of `bigtable` or `optimized`.
- Bigtable node bounds follow Google's rule (`1 <= min <= max <= 10 * min`); the CPU target is 10-80.
- Store and view ids follow Google's rule (`[a-z0-9_]`, first character not a digit, up to 60); view ids are unique.
- A view has exactly one source; a registry selection names at least one feature; a sync is cron or continuous, not both.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/featureOnlineStores/{feature_online_store_id}` |
| `feature_online_store_id` | `string` | The store's id |
| `location` | `string` | The store's location |
| `public_endpoint_domain_name` | `string` | The dedicated endpoint's domain (Optimized stores); empty otherwise |
| `service_attachment` | `string` | The PSC service attachment consumers target; empty unless PSC is on and a view has synced |
| `feature_view_names` | `[]string` | Full resource names of the feature views, in manifest order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Capacity bills around the clock.** Both storage kinds run serving nodes from creation, whether or not requests arrive; syncs add data-processing charges.
- **Storage kind is permanent.** Switching between Bigtable and Optimized, or changing the id or location, replaces the store and every view.
- **Syncs read BigQuery as the Vertex AI Service Agent**, which needs `roles/bigquery.dataViewer` on the sources.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVertexAiFeatureGroup** -- the registered features a view serves
- **GcpBigQueryTable** -- a view's direct BigQuery source
- **GcpKmsKey** -- customer-managed encryption
- **GcpGlobalForwardingRule** -- the consumer-side PSC endpoint targeting the store's service attachment

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
