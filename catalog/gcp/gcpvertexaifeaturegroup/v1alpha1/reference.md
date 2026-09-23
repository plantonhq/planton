# GcpVertexAiFeatureGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiFeatureGroupSpec defines a Vertex AI Feature Store feature
group (`google_vertex_ai_feature_group`) -- the registry entry that says
"these features of these entities live in this BigQuery table" --
together with the features registered in it
(`google_vertex_ai_feature_group_feature`). Features are folded in
because they belong to exactly one group and nothing else in the catalog
references a feature as a resource: a feature view names features by id
inside a group it references.

Feature Store keeps the data in BigQuery. A GcpVertexAiFeatureOnlineStore
serves the features at low latency through feature views that reference
this group by feature_group_id and pick features by feature_id.

Immutable: feature_group_id, location, and the BigQuery source (a change
replaces the group and its features). The description, labels, entity
ID columns, and features update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiFeatureGroup
metadata:
  name: customer-features
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  # Lowercase letters, digits, and underscores -- Google's rule for
  # feature group ids.
  featureGroupId: customer_features
  description: Customer features for churn and ranking models
  labels:
    team: growth
  bigQuery:
    # The source must have an entity ID column and a TIMESTAMP column named
    # feature_timestamp. A GcpBigQueryTable reference works here too.
    inputUri:
      value: bq://my-gcp-project.features.customers
    entityIdColumns:
      - customer_id
  features:
    - featureId: age
      description: Age in years
    - featureId: lifetime_value
      description: Lifetime value in USD
      versionColumnName: ltv_usd
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.featureGroupId` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.bigQuery` | `GcpVertexAiFeatureGroupBigQuery` |  |  |  |
| `spec.bigQuery.inputUri` | `string \| valueFrom` | yes |  | GcpBigQueryTable (`status.outputs.qualified_name`) |
| `spec.bigQuery.entityIdColumns` | `[]string` |  |  |  |
| `spec.features` | `[]GcpVertexAiFeatureGroupFeature` |  |  |  |
| `spec.features[].featureId` | `string` | yes |  |  |
| `spec.features[].description` | `string` |  |  |  |
| `spec.features[].labels` | `map<string, string>` |  |  |  |
| `spec.features[].versionColumnName` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the feature group lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) the feature group lives in, e.g.
"us-central1". Online stores that serve it must be in the same
location. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.featureGroupId

`string` · required

The feature group's id -- the last segment of its resource name and
what an online store's feature view references. Up to 128 characters
of lowercase letters, digits, and underscores; the first character
cannot be a digit (hyphens are not allowed, so metadata.name cannot
stand in). Unique within the project and location. Treat it as
immutable: the provider does not mark it ForceNew, so change it by
replacing the block.

- rule: {"required":true,"string":{"pattern":"^[a-z_][a-z0-9_]{0,127}$"}}

### spec.description

`string`

Free-text description of the feature group.

### spec.labels

`map<string, string>`

Labels on the feature group. The platform attribution labels are
merged in and win on key conflicts.

### spec.bigQuery

`GcpVertexAiFeatureGroupBigQuery`

The BigQuery table or view the features come from. Google's API has no
other source type today.

### spec.bigQuery.inputUri

`string | valueFrom` · required

The BigQuery table or view holding the feature data: a GcpBigQueryTable
reference (its {project}.{dataset}.{table} name), a literal
"project.dataset.table", or the bq:// form Google stores
("bq://project.dataset.table"). The modules add the bq:// prefix when
it is missing. The source must have at least one entity ID column and
a TIMESTAMP column named `feature_timestamp` -- Google's contract for a
feature group source. Jobs read it as the Vertex AI Service Agent,
which needs roles/bigquery.dataViewer on the table. Immutable.

- references: GcpBigQueryTable (`status.outputs.qualified_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryTable, name: <that resource's name>, fieldPath: status.outputs.qualified_name}} -- a bare string does not parse

### spec.bigQuery.entityIdColumns

`[]string`

The source columns whose values together form an entity's ID (the row
key an online store serves by). Empty means the single column
`entity_id`. Mutable in place.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.features

`[]GcpVertexAiFeatureGroupFeature`

The source columns registered as features, each keyed by feature_id.
Add or remove a feature by editing this list.

### spec.features[].featureId

`string` · required

The feature's id within the group -- what a feature view's
feature_ids names. Up to 128 characters of lowercase letters, digits,
and underscores; the first character cannot be a digit. By default it
is also the source column the feature reads (see version_column_name).
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z_][a-z0-9_]{0,127}$"}}

### spec.features[].description

`string`

Free-text description of the feature.

### spec.features[].labels

`map<string, string>`

Labels on the feature. The platform attribution labels are merged in
and win on key conflicts.

### spec.features[].versionColumnName

`string`

The source column that holds this feature's values, when it differs
from feature_id. Empty means the column named like the feature. Sent
only when set (Google computes it otherwise).

### spec.deletionPolicy

`string`

What happens to the feature group and its features when this resource
is destroyed:
  "" / "DELETE" -- the features and the group are deleted (the BigQuery
                   source is never touched)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.unique_feature_ids`: feature_id must be unique within the feature group

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiFeatureGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/featureGroups/{feature_group_id}. |
| `status.outputs.feature_group_id` | `string` | The feature group's id -- what an online store's feature view references. |
| `status.outputs.location` | `string` | The location the feature group lives in. |
| `status.outputs.feature_names` | `[]string` | Full resource names of the registered features, in manifest order. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.bigQuery.inputUri` | GcpBigQueryTable | `status.outputs.qualified_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpVertexAiFeatureOnlineStore | `spec.featureViews[].featureRegistrySource.featureGroups[].featureGroupId` | `status.outputs.feature_group_id` |

## See Also

- [Overview](../README.md)
