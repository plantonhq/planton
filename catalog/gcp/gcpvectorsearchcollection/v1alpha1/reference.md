# GcpVectorSearchCollection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVectorSearchCollectionSpec defines a Vector Search collection
(`google_vector_search_collection`) -- a schema'd store of data objects
with one or more vector fields -- together with the approximate
nearest-neighbor indexes built over those fields
(`google_vector_search_index`). Indexes are folded in because they
belong to exactly one collection, nothing else in the catalog refers to
an index on its own, and they share the collection's lifecycle.

The data itself (the objects written into the collection) is a
data-plane concern -- applications write it through the Vector Search
API or its SDKs -- and is deliberately not part of this block.

Immutable: collection_id, location, and the encryption key. The data
schema, description, display name, and labels update in place; every
index setting but labels replaces that index.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVectorSearchCollection
metadata:
  name: product-docs
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Product docs
  description: Semantic search over the product documentation
  # The non-vector fields of each data object, as a JSON Schema. Fields
  # named here can be pushed into an index and referenced from a text
  # template.
  dataSchema: '{"type":"object","properties":{"title":{"type":"string"},"body":{"type":"string"}}}'
  vectorSchemas:
    # One dense field whose embeddings Vector Search computes itself from
    # the object's title and body.
    - fieldName: text_embedding
      denseVector:
        dimensions: 768
        vertexEmbeddingConfig:
          modelId: text-embedding-005
          taskType: RETRIEVAL_DOCUMENT
          textTemplate: "Title: {title} ---- Body: {body}"
  indexes:
    # An ANN index over the dense field; every setting but labels is
    # immutable, so a change here replaces the index.
    - indexId: docs-ann
      indexField: text_embedding
      distanceMetric: DOT_PRODUCT
      storeFields:
        - title
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.collectionId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.dataSchema` | `string` |  |  |  |
| `spec.vectorSchemas` | `[]GcpVectorSearchCollectionVectorSchema` |  |  |  |
| `spec.vectorSchemas[].fieldName` | `string` | yes |  |  |
| `spec.vectorSchemas[].denseVector` | `GcpVectorSearchCollectionDenseVector` |  |  |  |
| `spec.vectorSchemas[].denseVector.dimensions` | `int32` |  |  |  |
| `spec.vectorSchemas[].denseVector.vertexEmbeddingConfig` | `GcpVectorSearchCollectionVertexEmbeddingConfig` |  |  |  |
| `spec.vectorSchemas[].denseVector.vertexEmbeddingConfig.modelId` | `string` | yes |  |  |
| `spec.vectorSchemas[].denseVector.vertexEmbeddingConfig.taskType` | `string` | yes |  |  |
| `spec.vectorSchemas[].denseVector.vertexEmbeddingConfig.textTemplate` | `string` | yes |  |  |
| `spec.vectorSchemas[].sparseVector` | `bool` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.indexes` | `[]GcpVectorSearchCollectionIndex` |  |  |  |
| `spec.indexes[].indexId` | `string` | yes |  |  |
| `spec.indexes[].indexField` | `string` | yes |  |  |
| `spec.indexes[].displayName` | `string` |  |  |  |
| `spec.indexes[].description` | `string` |  |  |  |
| `spec.indexes[].labels` | `map<string, string>` |  |  |  |
| `spec.indexes[].distanceMetric` | `string` |  |  |  |
| `spec.indexes[].featureNormType` | `string` |  |  |  |
| `spec.indexes[].filterFields` | `[]string` |  |  |  |
| `spec.indexes[].storeFields` | `[]string` |  |  |  |
| `spec.indexes[].dedicatedInfrastructure` | `GcpVectorSearchCollectionDedicatedInfrastructure` |  |  |  |
| `spec.indexes[].dedicatedInfrastructure.mode` | `string` |  |  |  |
| `spec.indexes[].dedicatedInfrastructure.autoscalingSpec` | `GcpVectorSearchCollectionAutoscalingSpec` |  |  |  |
| `spec.indexes[].dedicatedInfrastructure.autoscalingSpec.minReplicaCount` | `int32` |  |  |  |
| `spec.indexes[].dedicatedInfrastructure.autoscalingSpec.maxReplicaCount` | `int32` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the collection lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vector Search location (region), e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.collectionId

`string`

The collection's ID -- 1-63 characters, RFC 1035 (lowercase letters,
digits, hyphens; starts with a letter). Defaults to metadata.name.
Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.displayName

`string`

Human-readable name shown in the console.

### spec.description

`string`

Free-text description of the collection.

### spec.labels

`map<string, string>`

Labels on the collection.

### spec.dataSchema

`string`

JSON Schema for the non-vector fields of each data object, as a JSON
string (write it compact; Google normalizes it). Field names must be
alphanumeric characters, underscores, and hyphens. Fields named here
can be pushed into an index as filter_fields or store_fields and
referenced from a text_template.

### spec.vectorSchemas

`[]GcpVectorSearchCollectionVectorSchema`

The searchable vector fields. Only fields declared here can be
indexed and searched.

- rule: a vector field is exactly one of dense_vector or sparse_vector

### spec.vectorSchemas[].fieldName

`string` · required

The field's name -- alphanumeric characters, underscores, and hyphens.
An index names this field as its index_field.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z0-9_-]+$"}}

### spec.vectorSchemas[].denseVector

`GcpVectorSearchCollectionDenseVector`

A dense vector field, optionally with Vertex AI computing the
embeddings.

### spec.vectorSchemas[].denseVector.dimensions

`int32` · optional (explicit presence)

Number of dimensions in the vector (768 for textembedding-gecko and
text-embedding-005 at their default; 3072 for gemini-embedding-001).
Must match the embedding model when vertex_embedding_config is set.

- rule: {"int32":{"gte":1}}

### spec.vectorSchemas[].denseVector.vertexEmbeddingConfig

`GcpVectorSearchCollectionVertexEmbeddingConfig`

Have Vector Search compute this field's embeddings from the object's
text through a Vertex AI model. Omit to write precomputed vectors.

### spec.vectorSchemas[].denseVector.vertexEmbeddingConfig.modelId

`string` · required

The Vertex AI embedding model, e.g. "text-embedding-005" or
"textembedding-gecko@003". The model fixes the field's dimensionality;
dimensions on the parent must match what the model produces.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vectorSchemas[].denseVector.vertexEmbeddingConfig.taskType

`string` · required

The embedding task the model is told to optimize for. Documents stored
in the collection are embedded with RETRIEVAL_DOCUMENT; queries use
RETRIEVAL_QUERY at search time.

- rule: {"required":true,"string":{"in":["RETRIEVAL_QUERY","RETRIEVAL_DOCUMENT","SEMANTIC_SIMILARITY","CLASSIFICATION","CLUSTERING","QUESTION_ANSWERING","FACT_VERIFICATION","CODE_RETRIEVAL_QUERY"]}}

### spec.vectorSchemas[].denseVector.vertexEmbeddingConfig.textTemplate

`string` · required

The text handed to the model for each data object, with one or more
references to the object's fields in braces, e.g.
"Movie Title: {title} ---- Movie Plot: {plot}". Field names come from
the collection's data_schema.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.vectorSchemas[].sparseVector

`bool`

True declares a sparse vector field (index/value pairs over a large
vocabulary, the shape of lexical or hybrid-search embeddings such as
SPLADE or BM25). Google's sparse field carries no settings, so the
declaration is a flag.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the collection and its
indexes: a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in the same region. Omit to use Google-managed encryption. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.indexes

`[]GcpVectorSearchCollectionIndex`

The approximate-nearest-neighbor indexes over the collection's vector
fields, each keyed by its index_id. Add or remove an index by editing
this list; any other change to an index replaces it.

### spec.indexes[].indexId

`string` · required

The index's ID within the collection -- 1-63 characters, RFC 1035
(lowercase letters, digits, hyphens; starts with a letter). Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.indexes[].indexField

`string` · required

The vector field (a vector_schemas[].field_name) this index serves.
Immutable.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.indexes[].displayName

`string`

Human-readable name shown in the console.

### spec.indexes[].description

`string`

Free-text description of the index.

### spec.indexes[].labels

`map<string, string>`

Labels on the index -- the one mutable setting.

### spec.indexes[].distanceMetric

`string`

How similarity is measured: DOT_PRODUCT (Google's default; the right
choice for embeddings the model already normalizes) or COSINE_DISTANCE.
Sent only when set. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DOT_PRODUCT","COSINE_DISTANCE"]}}

### spec.indexes[].featureNormType

`string`

Feature normalization the ScaNN index applies before search: NONE, or
UNIT_L2_NORM to unit-normalize every vector (makes DOT_PRODUCT behave
as cosine similarity). Sent only when set. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NONE","UNIT_L2_NORM"]}}

### spec.indexes[].filterFields

`[]string`

Data-schema fields pushed into the index so searches can filter on
them inline, without a second lookup. Immutable.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.indexes[].storeFields

`[]string`

Data-schema fields pushed into the index so search results return
them inline, without fetching the data object. Immutable.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.indexes[].dedicatedInfrastructure

`GcpVectorSearchCollectionDedicatedInfrastructure`

Dedicated serving nodes for this index. Omit to serve from the shared
pool.

### spec.indexes[].dedicatedInfrastructure.mode

`string`

STORAGE_OPTIMIZED for large indexes where cost per vector matters more
than latency; PERFORMANCE_OPTIMIZED (Google's default) for
latency-sensitive serving. Immutable: a change replaces the index.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STORAGE_OPTIMIZED","PERFORMANCE_OPTIMIZED"]}}

### spec.indexes[].dedicatedInfrastructure.autoscalingSpec

`GcpVectorSearchCollectionAutoscalingSpec`

Replica bounds for the dedicated nodes.

### spec.indexes[].dedicatedInfrastructure.autoscalingSpec.minReplicaCount

`int32` · optional (explicit presence)

Fewest replicas kept serving (1-1000; Google defaults to 2 when unset
or 0). Sent only when set.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.indexes[].dedicatedInfrastructure.autoscalingSpec.maxReplicaCount

`int32` · optional (explicit presence)

Most replicas the index may scale to (at least min_replica_count, at
most 1000; Google defaults to the greater of min_replica_count and 2).
Sent only when set.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.deletionPolicy

`string`

What happens to the collection and its indexes when this resource is
destroyed:
  "" / "DELETE" -- the indexes and the collection are deleted, data
                   included
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVectorSearchCollection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name of the collection: projects/{project}/locations/{location}/collections/{collection_id}. |
| `status.outputs.collection_id` | `string` | The collection's ID -- the segment applications address data objects and searches by. |
| `status.outputs.location` | `string` | The collection's location, for rebuilding resource paths. |
| `status.outputs.index_names` | `[]string` | Full resource names of the indexes declared on the collection (projects/{project}/locations/{location}/collections/{collection}/indexes/{index_id}), in manifest order. |
| `status.outputs.index_count` | `int32` | Number of indexes declared on the collection. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
