# GCP Vector Search Collection

Stands up a Vector Search collection -- Google's managed store for objects with vector fields -- together with the approximate-nearest-neighbor indexes that make similarity search fast over them. Declare what an object looks like, which fields are vectors (and whether Vertex AI should compute the embeddings itself), and which indexes serve searches, and your applications write objects and query them through the Vector Search API. It is the vector database behind semantic search, recommendations, and retrieval-augmented generation.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `vectorsearch.googleapis.com` on the project
- **Collection** -- a `vectorsearch.Collection` with the data schema, vector fields, and optional CMEK
- **Indexes** -- one `vectorsearch.Index` per declared index

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vector Search admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpKmsKey** -- for customer-managed encryption, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vector Search Collection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Semantic Docs** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVectorSearchCollection
metadata:
  name: product-docs
  org: acme-corp
  env: prod
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

This creates a collection whose embeddings Vertex AI computes from each object's title and body, with one index that returns titles inline. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference the collection's `name` or `collection_id` output from the application or agent that searches it, and a `GcpKmsKey` from `kmsKeyName` when encryption must be customer-managed.

## Key Configuration

These are the most important decisions when configuring a collection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Who computes the vectors** -- a dense field with `vertexEmbeddingConfig` is embedded by Vertex AI from a text template over the object's fields on every write; without it, your pipeline writes the vectors. A `sparseVector: true` field holds lexical (keyword) vectors for hybrid search.

**Indexes are immutable** -- every index setting but labels replaces the index when changed, and the replacement is rebuilt from the collection's data. Pick the metric (`DOT_PRODUCT` or `COSINE_DISTANCE`), the fields to filter and store inline, and the infrastructure once.

**Pooled or dedicated** -- an index serves from Google's pooled infrastructure by default, billing per query; `dedicatedInfrastructure` gives it its own autoscaled nodes for latency-sensitive serving, billing per node-hour.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The collection's full resource name | Application configuration |
| `collection_id` | The collection's id | SDK clients |
| `location` | The collection's location | Regional clients |
| `index_names` | The indexes' full resource names, in manifest order | Search requests naming an index |
| `index_count` | Number of indexes | Dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Semantic docs** -- Vertex-computed embeddings over documents with one dot-product index. Start from the **Semantic Docs** preset.

**Hybrid dedicated** -- Precomputed dense vectors plus a sparse field, two indexes, dedicated autoscaled nodes, CMEK, and `PREVENT`. Start from the **Hybrid Dedicated** preset.

## Works With

- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP Vertex AI RAG Engine Config**](/cloud-catalog/gcp-vertex-ai-rag-engine-config) -- RAG Engine, which can use the collection as its vector database
- [**GCP Vertex AI Agent Engine**](/cloud-catalog/gcp-vertex-ai-agent-engine) -- agents whose tools search the collection
