# GcpVertexAiIndex

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpVertexAiIndexSpec defines a Vertex AI Vector Search index — the
data structure that holds embedding vectors and answers
nearest-neighbor queries. An index by itself stores and organizes
vectors; serving queries additionally requires a
GcpVertexAiIndexEndpoint and a GcpVertexAiDeployedIndex placing this
index onto that endpoint.

Two update regimes exist, chosen at creation and immutable:

  - **BATCH_UPDATE** (GCP's default): contents are replaced or
    amended in bulk from Cloud Storage files. Rebuilds take from tens
    of minutes to hours depending on corpus size.

  - **STREAM_UPDATE**: individual vectors are upserted/removed via
    the API within seconds. Streaming indexes cost more per GB and
    still compact in the background.

Immutable fields (ForceNew): location, index_update_method, and the
entire config block. Mutable in place: display_name, description,
labels. contents_delta_uri travels in its OWN update: when it
changes, the provider PATCHes metadata alone (no other field can
ride along in the same apply), and the imported/refreshed state never
reports it back — expect a one-field diff on the first plan after an
out-of-band data load.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiIndex
metadata:
  name: test-vector-index
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Test Vector Index
  description: Streaming index for product embedding search
  indexUpdateMethod: STREAM_UPDATE
  contentsDeltaUri: gs://my-bucket/embeddings/
  config:
    dimensions: 768
    approximateNeighborsCount: 150
    shardSize: SHARD_SIZE_SMALL
    distanceMeasureType: COSINE_DISTANCE
    featureNormType: UNIT_L2_NORM
    treeAhConfig:
      leafNodeEmbeddingCount: 1000
      leafNodesToSearchPercent: 10
  deletionPolicy: DELETE
  labels:
    team: ml-platform
    cost-center: research
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.indexUpdateMethod` | `string` |  | `BATCH_UPDATE` |  |
| `spec.contentsDeltaUri` | `string` |  |  |  |
| `spec.isCompleteOverwrite` | `bool` |  |  |  |
| `spec.config` | `GcpVertexAiIndexConfig` | yes |  |  |
| `spec.config.dimensions` | `int32` | yes |  |  |
| `spec.config.approximateNeighborsCount` | `int32` |  |  |  |
| `spec.config.shardSize` | `string` |  |  |  |
| `spec.config.distanceMeasureType` | `string` |  | `DOT_PRODUCT_DISTANCE` |  |
| `spec.config.featureNormType` | `string` |  | `NONE` |  |
| `spec.config.treeAhConfig` | `GcpVertexAiIndexTreeAhConfig` |  |  |  |
| `spec.config.treeAhConfig.leafNodeEmbeddingCount` | `int32` |  | `1000` |  |
| `spec.config.treeAhConfig.leafNodesToSearchPercent` | `int32` |  | `10` |  |
| `spec.config.bruteForceConfig` | `GcpVertexAiIndexBruteForceConfig` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the index will be created.
If omitted, the index is created in the provider's default project
(from the credential or ambient configuration).

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Region where the index will be created (e.g., "us-central1").
Deployed indexes must live in the same region. Immutable after
creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.displayName

`string` · required

Display name of the index (up to 128 UTF-8 characters). The
primary human-readable identifier; the numeric resource ID is
GCP-assigned.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.description

`string`

Description of the index.

### spec.indexUpdateMethod

`string` · optional (explicit presence)

How the index contents are updated after creation. Immutable —
migrating between regimes means recreating the index.
BATCH_UPDATE: bulk rebuilds from Cloud Storage (GCP's default).
STREAM_UPDATE: near-real-time upserts via the API.

- default: `BATCH_UPDATE`
- rule: index_update_method must be BATCH_UPDATE or STREAM_UPDATE

### spec.contentsDeltaUri

`string`

Cloud Storage DIRECTORY (not file) holding the initial or delta
vector data, e.g. "gs://my-bucket/embeddings/". File format and
layout: https://cloud.google.com/vertex-ai/docs/vector-search/setup/format-structure

May be omitted at creation (an empty index; the norm for
STREAM_UPDATE) and set later to load data. Two provider quirks,
both harmless but worth knowing: a change to this field travels in
its own single-field update, and the value is write-only — GCP
never reports it back on read, so out-of-band loads show up as a
one-field diff on the next plan.

A plain string (not a reference) because the gs:// directory URI
has no matching output shape on the GCS kinds; compose by
writing the bucket name into the URI.

- rule: contents_delta_uri must be a Cloud Storage directory URI starting with gs://

### spec.isCompleteOverwrite

`bool`

If true, an update that carries contents_delta_uri REPLACES the
whole index contents with the files at the URI; if false (the
default), the files are treated as a delta (upserts/deletes) on
top of the existing contents. Only meaningful together with
contents_delta_uri.

### spec.config

`GcpVertexAiIndexConfig` · required

Vector-search geometry: dimensions, algorithm, sharding, distance
measure. Required — an index cannot exist without its geometry.
The entire block is immutable.

- rule: {"required":true}
- rule: tree_ah_config and brute_force_config are mutually exclusive; pick the approximate (tree-AH) or the exact (brute-force) algorithm, not both
- rule: approximate_neighbors_count is required when tree_ah_config is set — it is the number of candidates fetched by approximate search before exact reordering

### spec.config.dimensions

`int32` · required

Number of dimensions of the input vectors — the embedding model's
output size (e.g. 768 for many text encoders). Immutable.

- rule: {"required":true,"int32":{"gte":1}}

### spec.config.approximateNeighborsCount

`int32`

Number of neighbors found via approximate search before exact
reordering (a more expensive distance computation over the
candidates). Required by the API when tree-AH is used; not
meaningful for brute-force. Immutable.

- rule: {"int32":{"gte":0}}

### spec.config.shardSize

`string`

Physical shard size for the index data. Determines how much data
each shard holds and which machine types can serve it when
deployed. If omitted, GCP picks a size based on the data.
SHARD_SIZE_SMALL: 2 GB per shard.
SHARD_SIZE_MEDIUM: 20 GB per shard.
SHARD_SIZE_LARGE: 50 GB per shard.
Immutable.

- rule: shard_size must be SHARD_SIZE_SMALL, SHARD_SIZE_MEDIUM, or SHARD_SIZE_LARGE

### spec.config.distanceMeasureType

`string` · optional (explicit presence)

Distance measure used in nearest-neighbor search. Match it to how
the embedding model was trained — a mismatched measure silently
degrades result quality.
SQUARED_L2_DISTANCE: Euclidean (L2) distance.
L1_DISTANCE: Manhattan (L1) distance.
COSINE_DISTANCE: 1 - cosine similarity.
DOT_PRODUCT_DISTANCE: negative dot product (GCP's default).
Immutable.

- default: `DOT_PRODUCT_DISTANCE`
- rule: distance_measure_type must be SQUARED_L2_DISTANCE, L1_DISTANCE, COSINE_DISTANCE, or DOT_PRODUCT_DISTANCE

### spec.config.featureNormType

`string` · optional (explicit presence)

Normalization applied to each vector before indexing.
UNIT_L2_NORM: unit L2 normalization (with DOT_PRODUCT_DISTANCE this
  makes ranking equivalent to cosine similarity).
NONE: no normalization (GCP's default).
Immutable.

- default: `NONE`
- rule: feature_norm_type must be UNIT_L2_NORM or NONE

### spec.config.treeAhConfig

`GcpVertexAiIndexTreeAhConfig`

Tree-AH approximate search configuration. Mutually exclusive with
brute_force_config. Immutable.

### spec.config.treeAhConfig.leafNodeEmbeddingCount

`int32` · optional (explicit presence)

Number of embeddings on each leaf node of the tree. Larger leaves
mean fewer tree levels (faster build, coarser pruning); smaller
leaves prune more aggressively per query. GCP's default is 1000.

- default: `1000`
- rule: {"int32":{"gt":0}}

### spec.config.treeAhConfig.leafNodesToSearchPercent

`int32` · optional (explicit presence)

Percentage of leaf nodes any single query may search, 1-100
inclusive. Raising it improves recall at the cost of latency.
GCP's default is 10.

- default: `10`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.config.bruteForceConfig

`GcpVertexAiIndexBruteForceConfig`

Brute-force (exact) search. Mutually exclusive with tree_ah_config.
Immutable.

### spec.labels

`map<string, string>`

User-defined labels to organize the index (cost attribution, team
ownership, environment tagging). Keys and values must follow GCP
label rules: lowercase letters, digits, underscores, and dashes,
at most 63 characters. Merged with the platform's attribution
labels; on key conflicts the platform labels win. Mutable in place.

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key for customer-managed encryption at rest (CMEK) of the
index data, as the full key resource path
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
— a GcpKmsKey reference resolves to it. The key must live in the
same region as the index, and the Vertex AI service agent needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it. If omitted, data
is encrypted with Google-managed keys. Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

Deletion policy for the index — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the index and every vector in it are deleted
  "PREVENT" -- destroy FAILS; a guard for a corpus that took hours
               of batch builds (or months of streaming upserts) to
               load
  "ABANDON" -- the index is removed from management but left
               standing (and billing for its stored vectors) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiIndex, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.index_id` | `string` | Fully qualified index resource path — the value a GcpVertexAiDeployedIndex passes as its `index` reference. Format: projects/{project}/locations/{location}/indexes/{indexId} |
| `status.outputs.index_name` | `string` | The GCP-assigned numeric index ID (the last path segment of index_id). |
| `status.outputs.metadata_schema_uri` | `string` | Cloud Storage URI of the YAML schema describing additional index-specific information GCP attaches to this index type. |
| `status.outputs.create_time` | `string` | RFC3339 timestamp of when the index was created. |
| `status.outputs.update_time` | `string` | RFC3339 timestamp of when the index was last updated (data or metadata). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpVertexAiDeployedIndex | `spec.index` | `status.outputs.index_id` |

## See Also

- [Overview](../README.md)
