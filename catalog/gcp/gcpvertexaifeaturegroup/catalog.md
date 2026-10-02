# GCP Vertex AI Feature Group

Registers the features your models use -- a customer's age, a product's click rate, a merchant's fraud score -- in Vertex AI Feature Store, straight from the BigQuery table they already live in. Nothing is copied: the group tells Feature Store which table holds which entities' features and which columns are features, so training pipelines and online stores all read the same definitions. An online store then serves the features to models in production through feature views that point at this group.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Feature group** -- a `vertex.AiFeatureGroup` over the BigQuery source
- **Features** -- one `vertex.AiFeatureGroupFeature` per registered feature

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Dependencies

- **GcpBigQueryTable** -- the source table or view, with an entity ID column and a `TIMESTAMP` column named `feature_timestamp`, referenced by `bigQuery.inputUri`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Feature Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **BigQuery Features** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiFeatureGroup
metadata:
  name: customer-features
  org: acme-corp
  env: prod
spec:
  location: us-central1
  featureGroupId: customer_features
  bigQuery:
    inputUri:
      valueFrom:
        kind: GcpBigQueryTable
        name: customer-features-source
        fieldPath: status.outputs.qualified_name
    entityIdColumns:
      - customer_id
  features:
    - featureId: age
    - featureId: lifetime_value
```

```shell
planton apply -f vertex-ai-feature-group.yaml
```

This registers two customer features from a BigQuery table, keyed by `customer_id`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference a `GcpBigQueryTable` from `bigQuery.inputUri`, and have a `GcpVertexAiFeatureOnlineStore` feature view reference this group's `feature_group_id` output.

## Key Configuration

These are the most important decisions when configuring a feature group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The source table's shape** -- Google requires at least one entity ID column and a `TIMESTAMP` column named `feature_timestamp`; every row is one entity's values at one moment. Name the entity ID columns in `entityIdColumns` when they are not the single column `entity_id`.

**Which columns are features** -- each `features[]` entry registers one column; `versionColumnName` points a feature at a differently named column. Only registered features can be served by an online store.

**IDs use underscores** -- group and feature ids allow lowercase letters, digits, and underscores only, so they are set explicitly rather than derived from the resource name.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpBigQueryTable** | `bigQuery.inputUri` | `status.outputs.qualified_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The group's full resource name | SDK clients |
| `feature_group_id` | The group's id | An online store's feature view |
| `location` | The group's region | Regional clients |
| `feature_names` | The registered features' full names | Dashboards, lineage |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**BigQuery features** -- a group over one table with a handful of features. Start from the **BigQuery Features** preset.

**Composite entity key** -- a group whose entities are keyed by two columns, with a feature reading a differently named column and `PREVENT`. Start from the **Composite Entity Key** preset.

## Works With

- [**GCP BigQuery Table**](/cloud-catalog/gcp-big-query-table) -- the source table or view
- [**GCP BigQuery Dataset**](/cloud-catalog/gcp-big-query-dataset) -- the dataset holding the source
- [**GCP Vertex AI Feature Online Store**](/cloud-catalog/gcp-vertex-ai-feature-online-store) -- serves the features in production
