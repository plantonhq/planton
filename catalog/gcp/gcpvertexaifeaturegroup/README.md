# GCP Vertex AI Feature Group

A Vertex AI Feature Store feature group -- the registry entry that says "these features of these entities live in this BigQuery table or view" -- together with the features registered in it. Feature Store keeps the data where it is, in BigQuery; a `GcpVertexAiFeatureOnlineStore` serves the registered features at low latency through feature views that reference this group and pick features by id.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Feature group** -- a `vertex_ai_feature_group` pointing at the BigQuery source, with its entity ID columns
- **Features** -- one `vertex_ai_feature_group_feature` per `features[]` entry, keyed by `featureId`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Dependencies

- **`GcpBigQueryTable`** -- the source table or view (`bigQuery.inputUri`). It needs at least one entity ID column and a `TIMESTAMP` column named `feature_timestamp`. The Vertex AI Service Agent reads it and needs `roles/bigquery.dataViewer` on it.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiFeatureGroup
metadata:
  name: customer-features
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI region. Immutable. |
| `featureGroupId` | `string` | Up to 128 lowercase letters, digits, and underscores, not starting with a digit. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `description`, `labels` | | | Descriptive metadata; mutable. |
| `bigQuery.inputUri` | `StringValueOrRef` | none | A `GcpBigQueryTable` reference or a literal `project.dataset.table` / `bq://project.dataset.table`; the modules add `bq://` when missing. Immutable. |
| `bigQuery.entityIdColumns` | `[]string` | `entity_id` | The columns that form an entity's ID. Mutable. |
| `features[]` | `[]object` | none | `featureId` (same rule as the group id; immutable), `description`, `labels`, `versionColumnName` (the source column, when it differs from the id). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; fanned to every feature. |

### Validation Rules

- Group and feature ids follow Google's rule: `[a-z0-9_]`, first character not a digit, up to 128.
- Feature ids are unique within the group.
- A declared BigQuery source names its table.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/featureGroups/{feature_group_id}` |
| `feature_group_id` | `string` | The group's id -- what an online store's feature view references |
| `location` | `string` | The group's location |
| `feature_names` | `[]string` | Full resource names of the registered features, in manifest order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Nothing is copied.** A feature group is metadata over BigQuery; the data stays in the table, and online stores copy only what their feature views select, on their own schedule.
- **The source is permanent.** Changing the table, the group id, or the location replaces the group and its features. The provider does not mark the id as force-new, so replace the block to change it.
- **Destroy never touches BigQuery.** Deleting the group removes the registry entries only.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpBigQueryTable** -- the source table or view
- **GcpBigQueryDataset** -- the dataset holding the source
- **GcpVertexAiFeatureOnlineStore** -- serves the group's features at low latency through feature views

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
