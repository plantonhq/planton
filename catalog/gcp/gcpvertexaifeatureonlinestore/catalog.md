# GCP Vertex AI Feature Online Store

Serves your models' features in production at millisecond latency: when a request arrives for customer 42, the model asks the online store for customer 42's features and gets them back fresh. The store syncs the features it serves from Vertex AI feature groups or straight from BigQuery, on a schedule or continuously, and runs on either a managed Bigtable instance (large feature sets, autoscaling) or Google's Optimized serving (the lowest latency, optionally over Private Service Connect).

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Online store** -- a `vertex.AiFeatureOnlineStore` with Bigtable or Optimized storage
- **Feature views** -- one `vertex.AiFeatureOnlineStoreFeatureview` per declared view

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpVertexAiFeatureGroup** -- the feature groups registry-sourced views serve.
- **GcpBigQueryTable** -- the table a BigQuery-sourced view materializes.
- **GcpKmsKey** -- for customer-managed encryption, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Feature Online Store**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Bigtable Serving** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiFeatureOnlineStore
metadata:
  name: serving-store
  org: acme-corp
  env: prod
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

This creates a one-to-three-node Bigtable store serving two customer features, refreshed every six hours. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference each `GcpVertexAiFeatureGroup`'s `feature_group_id` output from a feature view, and give serving applications the store's `name` and the view names.

## Key Configuration

These are the most important decisions when configuring an online store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Bigtable or Optimized** -- Bigtable storage autoscales between node bounds and suits large feature sets served by entity key; Optimized storage serves from a dedicated endpoint at the lowest latency. The choice is fixed at creation, and both bill for capacity around the clock.

**What each view serves** -- a view either selects features from feature groups (the registry path, one definition shared with training) or materializes a BigQuery table directly. Clients query a view by name.

**How fresh** -- a cron sync copies the latest values on a schedule; `continuous: true` streams changes as the source changes. Every sync reads BigQuery as the Vertex AI Service Agent.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVertexAiFeatureGroup** | `featureViews[].featureRegistrySource.featureGroups[].featureGroupId` | `status.outputs.feature_group_id` |
| **GcpProject** | `featureViews[].featureRegistrySource.projectNumber` | `status.outputs.project_number` |
| **GcpBigQueryTable** | `featureViews[].bigQuerySource.uri` | `status.outputs.qualified_name` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The store's full resource name | Serving clients |
| `feature_online_store_id` | The store's id | SDK calls |
| `location` | The store's region | Regional clients |
| `public_endpoint_domain_name` | The dedicated endpoint's domain | Optimized-store clients |
| `service_attachment` | The PSC service attachment | A consumer's PSC forwarding rule |
| `feature_view_names` | The views' full names | FetchFeatureValues requests |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Bigtable serving** -- a small autoscaled Bigtable store serving registry features on a cron. Start from the **Bigtable Serving** preset.

**Optimized over PSC** -- an Optimized store reachable only over Private Service Connect, encrypted under your key, with a continuously synced view. Start from the **Optimized Private** preset.

## Works With

- [**GCP Vertex AI Feature Group**](/cloud-catalog/gcp-vertex-ai-feature-group) -- the registered features views serve
- [**GCP BigQuery Table**](/cloud-catalog/gcp-big-query-table) -- a view's direct source
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP Global Forwarding Rule**](/cloud-catalog/gcp-global-forwarding-rule) -- the consumer-side PSC endpoint
