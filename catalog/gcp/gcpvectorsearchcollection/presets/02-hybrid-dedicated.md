# Hybrid Dedicated

## Use Case

A production search collection that combines a dense vector field (precomputed by your own pipeline) with a sparse lexical field, so semantic and keyword recall are served side by side. The dense index runs on dedicated, autoscaled serving nodes with inline category filtering; the whole collection is encrypted with your own key.

## When to Use

- E-commerce or catalog search where exact keywords matter as much as meaning
- Latency-sensitive serving that should not share the pooled infrastructure
- Data under a customer-managed encryption requirement

## What This Creates

- A collection in `us-central1` with a 1024-dimension dense field (vectors written by the caller) and a sparse field
- Encryption under the referenced `GcpKmsKey`
- Index `dense-cosine`: cosine distance over unit-normalized vectors, inline filtering on `category`, `name` and `price` stored inline, on `PERFORMANCE_OPTIMIZED` dedicated nodes scaling between 2 and 6 replicas
- Index `lexical-ann` over the sparse field with the same filter
- `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `vectorSchemas[0].denseVector.dimensions` | `1024` | The dimensionality of the vectors your pipeline writes. |
| `indexes[0].dedicatedInfrastructure.mode` | `PERFORMANCE_OPTIMIZED` | `STORAGE_OPTIMIZED` when cost per vector matters more than latency. |
| `indexes[0].dedicatedInfrastructure.autoscalingSpec` | 2-6 replicas | Your traffic floor and ceiling; each replica bills as a node. |
| `kmsKeyName` | `search-data-key` | Your key in the same region, or omit for Google-managed encryption (immutable). |
| `deletionPolicy` | `PREVENT` | `DELETE` for a disposable environment. |

Dedicated infrastructure bills per node-hour whether or not searches arrive; the pooled default bills per query and storage. The dedicated replica bounds are the committed spend.
