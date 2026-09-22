# GCP Vector Search Collection

A Vector Search collection -- a schema'd store of data objects with one or more vector fields -- together with the approximate-nearest-neighbor indexes built over those fields. Declare the object schema, the vector fields (dense, with Vertex AI computing the embeddings if you like, or sparse), and the indexes that serve searches, and applications write objects and run similarity queries through the Vector Search API. Indexes are part of this block because a collection owns them outright.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `vectorsearch.googleapis.com` on the project (never disabled on destroy)
- **Collection** -- a `vector_search_collection` with the data schema and vector fields
- **Indexes** -- one `vector_search_index` per `indexes[]` entry, keyed by `indexId`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vector Search admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpKmsKey`** -- a key in the same region for customer-managed encryption (`kmsKeyName`). Immutable once set.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVectorSearchCollection
metadata:
  name: product-docs
spec:
  location: us-central1
  dataSchema: '{"type":"object","properties":{"title":{"type":"string"},"body":{"type":"string"}}}'
  vectorSchemas:
    - fieldName: text_embedding
      denseVector:
        dimensions: 768
        vertexEmbeddingConfig:
          modelId: text-embedding-005
          taskType: RETRIEVAL_DOCUMENT
          textTemplate: "Title: {title} ---- Body: {body}"
  indexes:
    - indexId: docs-ann
      indexField: text_embedding
      storeFields:
        - title
```

```shell
planton apply -f vector-search-collection.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vector Search location (region). Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `collectionId` | `string` | `metadata.name` | RFC 1035 id. Immutable. |
| `displayName`, `description`, `labels` | | | Descriptive metadata; mutable. |
| `dataSchema` | `string` | none | JSON Schema of the non-vector fields, as a compact JSON string. |
| `vectorSchemas[]` | `[]object` | none | Searchable vector fields: `fieldName` plus exactly one of `denseVector` (`dimensions`, optional `vertexEmbeddingConfig`) or `sparseVector: true`. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Immutable. |
| `indexes[]` | `[]object` | none | `indexId`, `indexField`, optional `displayName`, `description`, `labels`, `distanceMetric` (`DOT_PRODUCT` / `COSINE_DISTANCE`), `featureNormType` (`NONE` / `UNIT_L2_NORM`), `filterFields`, `storeFields`, `dedicatedInfrastructure { mode, autoscalingSpec { minReplicaCount, maxReplicaCount } }`. Everything but labels is immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; fanned to every index. |

### Validation Rules

- A vector field is exactly one of dense or sparse.
- A dense field's embedding config requires `modelId`, `taskType` (one of Google's eight), and `textTemplate`.
- Collection and index ids are RFC 1035; metric, norm, and mode values are Google's; replica bounds are 1-1000.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/collections/{collection_id}` |
| `collection_id` | `string` | The collection's id |
| `location` | `string` | The collection's location |
| `index_names` | `[]string` | Full resource names of the indexes, in manifest order |
| `index_count` | `int32` | Number of indexes declared |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **An index change replaces the index.** Every index setting but labels is immutable; the replacement is rebuilt from the collection's data, and searches against it wait for the rebuild.
- **Data is not part of this block.** Objects are written by applications through the Vector Search API or its SDKs; this block declares the store and its indexes.
- **Dedicated infrastructure bills per node-hour** from creation; the pooled default bills per query and storage.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpKmsKey** -- customer-managed encryption for the collection
- **GcpVertexAiRagEngineConfig** -- RAG Engine, which can use a collection as its vector database
- **GcpVertexAiAgentEngine** -- an agent runtime whose tools can search the collection
