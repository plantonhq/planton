# GcpVertexAiFeatureOnlineStore

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiFeatureOnlineStoreSpec defines a Vertex AI Feature Store
online store (`google_vertex_ai_feature_online_store`) -- the low-latency
serving layer that answers "give me this entity's features now" for
models in production -- together with the feature views it serves
(`google_vertex_ai_feature_online_store_featureview`). Feature views are
folded in because they belong to exactly one store and nothing else in
the catalog references a view.

Storage is exactly one of two kinds: `bigtable` (a managed Bigtable
instance with autoscaling, for large feature sets) or `optimized` (Google's
serving infrastructure behind a dedicated endpoint, for the lowest
latency). Both bill for capacity around the clock.

Immutable: feature_online_store_id, location, and the storage kind (a
change replaces the store and every view). Bigtable scaling, the
dedicated endpoint, the encryption key, labels, and the views update in
place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiFeatureOnlineStore
metadata:
  name: serving-store
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  # Lowercase letters, digits, and underscores -- Google's rule for online
  # store and feature view ids.
  featureOnlineStoreId: serving_store
  labels:
    team: ml
  # Exactly one storage kind: a managed Bigtable instance (large feature
  # sets, autoscaling) or optimized: true (lowest latency).
  bigtable:
    autoScaling:
      minNodeCount: 1
      maxNodeCount: 3
  featureViews:
    - featureViewId: customer_view
      featureRegistrySource:
        featureGroups:
          - featureGroupId:
              value: customer_features
            featureIds:
              - age
              - lifetime_value
      syncConfig:
        cron: "0 */6 * * *"
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.featureOnlineStoreId` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.bigtable` | `GcpVertexAiFeatureOnlineStoreBigtable` |  |  |  |
| `spec.bigtable.autoScaling` | `GcpVertexAiFeatureOnlineStoreBigtableAutoScaling` | yes |  |  |
| `spec.bigtable.autoScaling.minNodeCount` | `int32` |  |  |  |
| `spec.bigtable.autoScaling.maxNodeCount` | `int32` |  |  |  |
| `spec.bigtable.autoScaling.cpuUtilizationTarget` | `int32` |  |  |  |
| `spec.bigtable.enableDirectBigtableAccess` | `bool` |  |  |  |
| `spec.bigtable.zone` | `string` |  |  |  |
| `spec.optimized` | `bool` |  |  |  |
| `spec.dedicatedServingEndpoint` | `GcpVertexAiFeatureOnlineStoreDedicatedServingEndpoint` |  |  |  |
| `spec.dedicatedServingEndpoint.privateServiceConnectConfig` | `GcpVertexAiFeatureOnlineStorePrivateServiceConnectConfig` |  |  |  |
| `spec.dedicatedServingEndpoint.privateServiceConnectConfig.enablePrivateServiceConnect` | `bool` |  |  |  |
| `spec.dedicatedServingEndpoint.privateServiceConnectConfig.projectAllowlist` | `[]string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.featureViews` | `[]GcpVertexAiFeatureOnlineStoreFeatureView` |  |  |  |
| `spec.featureViews[].featureViewId` | `string` | yes |  |  |
| `spec.featureViews[].labels` | `map<string, string>` |  |  |  |
| `spec.featureViews[].bigQuerySource` | `GcpVertexAiFeatureOnlineStoreBigQuerySource` |  |  |  |
| `spec.featureViews[].bigQuerySource.uri` | `string \| valueFrom` | yes |  | GcpBigQueryTable (`status.outputs.qualified_name`) |
| `spec.featureViews[].bigQuerySource.entityIdColumns` | `[]string` | yes |  |  |
| `spec.featureViews[].featureRegistrySource` | `GcpVertexAiFeatureOnlineStoreFeatureRegistrySource` |  |  |  |
| `spec.featureViews[].featureRegistrySource.featureGroups` | `[]GcpVertexAiFeatureOnlineStoreFeatureGroupSelection` | yes |  |  |
| `spec.featureViews[].featureRegistrySource.featureGroups[].featureGroupId` | `string \| valueFrom` | yes |  | GcpVertexAiFeatureGroup (`status.outputs.feature_group_id`) |
| `spec.featureViews[].featureRegistrySource.featureGroups[].featureIds` | `[]string` | yes |  |  |
| `spec.featureViews[].featureRegistrySource.projectNumber` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_number`) |
| `spec.featureViews[].syncConfig` | `GcpVertexAiFeatureOnlineStoreSyncConfig` |  |  |  |
| `spec.featureViews[].syncConfig.cron` | `string` |  |  |  |
| `spec.featureViews[].syncConfig.continuous` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the store lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) the store lives in, e.g.
"us-central1". Feature groups its views serve must be in the same
location. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.featureOnlineStoreId

`string` · required

The store's id -- the last segment of its resource name. Up to 60
characters of lowercase letters, digits, and underscores; the first
character cannot be a digit (hyphens are not allowed, so
metadata.name cannot stand in). Unique within the project and
location. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z_][a-z0-9_]{0,59}$"}}

### spec.labels

`map<string, string>`

Labels on the store. The platform attribution labels are merged in and
win on key conflicts.

### spec.bigtable

`GcpVertexAiFeatureOnlineStoreBigtable`

Serve from a managed Bigtable instance. Exactly one of bigtable or
optimized.

### spec.bigtable.autoScaling

`GcpVertexAiFeatureOnlineStoreBigtableAutoScaling` · required

Node bounds for the managed Bigtable instance.

- rule: {"required":true}
- rule: max_node_count must be at least min_node_count and at most 10 times min_node_count

### spec.bigtable.autoScaling.minNodeCount

`int32`

Fewest nodes kept running -- at least 1. Each node bills around the
clock.

- rule: {"int32":{"gte":1}}

### spec.bigtable.autoScaling.maxNodeCount

`int32`

Most nodes Google may scale to -- at least min_node_count and at most
ten times it.

- rule: {"int32":{"gte":1}}

### spec.bigtable.autoScaling.cpuUtilizationTarget

`int32` · optional (explicit presence)

The CPU utilization (10-80 percent) Bigtable scales to hold: above it
nodes are added, well below it nodes are removed. Google defaults to
50. Sent only when set.

- rule: {"int32":{"lte":80,"gte":10}}

### spec.bigtable.enableDirectBigtableAccess

`bool`

True lets clients read the managed Bigtable instance directly, beside
the Feature Store serving API.

### spec.bigtable.zone

`string`

The zone the Bigtable instance is created in, e.g. "us-central1-a".
Google picks one in the store's region when empty. Sent only when set.

### spec.optimized

`bool`

True serves from Google's Optimized online serving (a dedicated
endpoint, the lowest latency). Google's optimized block carries no
settings, so the choice is a flag. Exactly one of bigtable or
optimized.

### spec.dedicatedServingEndpoint

`GcpVertexAiFeatureOnlineStoreDedicatedServingEndpoint`

The store's dedicated serving endpoint. Set it to serve over Private
Service Connect; omit to keep Google's default (a public dedicated
endpoint on Optimized stores). Sent only when set.

### spec.dedicatedServingEndpoint.privateServiceConnectConfig

`GcpVertexAiFeatureOnlineStorePrivateServiceConnectConfig`

Private Service Connect for the dedicated endpoint.

### spec.dedicatedServingEndpoint.privateServiceConnectConfig.enablePrivateServiceConnect

`bool`

True serves the dedicated endpoint only through a PSC service
attachment (the `service_attachment` output) that consumers target
with a forwarding rule; false keeps the public endpoint.

### spec.dedicatedServingEndpoint.privateServiceConnectConfig.projectAllowlist

`[]string`

Projects (IDs or numbers) allowed to create forwarding rules that
target the service attachment.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting both the online and the
offline data: a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in the store's region. Omit to use Google-managed encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.forceDestroy

`bool`

True lets a destroy delete the store even when it still holds feature
views and features the block does not manage (views created outside
this block, for example). Without it Google refuses to delete a
non-empty store. The views declared below are deleted first either
way.

### spec.featureViews

`[]GcpVertexAiFeatureOnlineStoreFeatureView`

The feature views the store serves, each keyed by feature_view_id.

- rule: a feature view has exactly one of big_query_source or feature_registry_source

### spec.featureViews[].featureViewId

`string` · required

The view's id -- what serving clients name. Up to 60 characters of
lowercase letters, digits, and underscores; the first character cannot
be a digit. Unique within the store. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z_][a-z0-9_]{0,59}$"}}

### spec.featureViews[].labels

`map<string, string>`

Labels on the view. The platform attribution labels are merged in and
win on key conflicts.

### spec.featureViews[].bigQuerySource

`GcpVertexAiFeatureOnlineStoreBigQuerySource`

Materialize a BigQuery table or view directly.

### spec.featureViews[].bigQuerySource.uri

`string | valueFrom` · required

The BigQuery table or view materialized on each sync: a
GcpBigQueryTable reference (its {project}.{dataset}.{table} name), a
literal "project.dataset.table", or the bq:// form Google stores. The
modules add the bq:// prefix when it is missing.

- references: GcpBigQueryTable (`status.outputs.qualified_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryTable, name: <that resource's name>, fieldPath: status.outputs.qualified_name}} -- a bare string does not parse

### spec.featureViews[].bigQuerySource.entityIdColumns

`[]string` · required

The columns that form the entity ID the view serves by. Google
supports exactly one today.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.featureViews[].featureRegistrySource

`GcpVertexAiFeatureOnlineStoreFeatureRegistrySource`

Serve features registered in feature groups.

### spec.featureViews[].featureRegistrySource.featureGroups

`[]GcpVertexAiFeatureOnlineStoreFeatureGroupSelection` · required

The feature groups and features the view serves.

- rule: {"repeated":{"minItems":"1"}}

### spec.featureViews[].featureRegistrySource.featureGroups[].featureGroupId

`string | valueFrom` · required

The feature group: a GcpVertexAiFeatureGroup reference or its literal
id. The group must be in the store's location.

- references: GcpVertexAiFeatureGroup (`status.outputs.feature_group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiFeatureGroup, name: <that resource's name>, fieldPath: status.outputs.feature_group_id}} -- a bare string does not parse

### spec.featureViews[].featureRegistrySource.featureGroups[].featureIds

`[]string` · required

The features to serve, by feature_id within the group.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.featureViews[].featureRegistrySource.projectNumber

`string | valueFrom`

The number of the project that owns the feature groups, when it is not
the store's project: a GcpProject reference (its project number) or a
literal number.

- references: GcpProject (`status.outputs.project_number`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_number}} -- a bare string does not parse

### spec.featureViews[].syncConfig

`GcpVertexAiFeatureOnlineStoreSyncConfig`

When the view copies fresh values in. Omit for Google's default.

- rule: a feature view syncs on a cron schedule or continuously, not both

### spec.featureViews[].syncConfig.cron

`string`

A cron schedule for batch syncs, e.g. "0 */6 * * *"; prefix
"CRON_TZ=America/New_York " (or "TZ=") to pin a time zone. Sent only
when set.

### spec.featureViews[].syncConfig.continuous

`bool`

True syncs continuously as the source changes instead of on a
schedule.

### spec.deletionPolicy

`string`

What happens to the store and its feature views when this resource is
destroyed:
  "" / "DELETE" -- the views and the store are deleted (the sources are
                   never touched)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and keeps serving (and
                   billing)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_storage`: an online store has exactly one storage kind: bigtable or optimized
- `spec.unique_feature_view_ids`: feature_view_id must be unique within the online store

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiFeatureOnlineStore, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/featureOnlineStores/{feature_online_store_id}. |
| `status.outputs.feature_online_store_id` | `string` | The store's id. |
| `status.outputs.location` | `string` | The location the store lives in. |
| `status.outputs.public_endpoint_domain_name` | `string` | The dedicated endpoint's public domain name (Optimized stores); empty on Bigtable stores and behind Private Service Connect. |
| `status.outputs.service_attachment` | `string` | The Private Service Connect service attachment consumers target, once PSC is enabled and a view has synced; empty otherwise. |
| `status.outputs.feature_view_names` | `[]string` | Full resource names of the declared feature views, in manifest order. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.featureViews[].bigQuerySource.uri` | GcpBigQueryTable | `status.outputs.qualified_name` |
| `spec.featureViews[].featureRegistrySource.featureGroups[].featureGroupId` | GcpVertexAiFeatureGroup | `status.outputs.feature_group_id` |
| `spec.featureViews[].featureRegistrySource.projectNumber` | GcpProject | `status.outputs.project_number` |

## See Also

- [Overview](../README.md)
