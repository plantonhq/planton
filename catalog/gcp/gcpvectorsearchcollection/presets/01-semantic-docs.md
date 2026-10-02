# Semantic Docs

## Use Case

A document store for semantic search where Vector Search computes the embeddings itself: write objects with a title, a URL, and a body, and every object's `text_embedding` is filled by Vertex AI's `text-embedding-005` from the title and body. One approximate-nearest-neighbor index serves the searches and returns the title and URL inline, so a result needs no second lookup.

## When to Use

- Retrieval-augmented generation over documentation, tickets, or knowledge bases
- Any collection where the caller has text, not vectors
- The first Vector Search collection in a project

## What This Creates

- A collection in `us-central1` with a three-field data schema and one dense vector field of 768 dimensions embedded by Vertex AI
- An index `docs-ann` over that field with dot-product similarity, storing `title` and `url` inline

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `dataSchema` | title, url, body | Your object fields; names must be alphanumeric, underscores, or hyphens. |
| `vectorSchemas[].denseVector.vertexEmbeddingConfig.modelId` | `text-embedding-005` | A different Vertex embedding model; match `dimensions` to what it produces. |
| `vectorSchemas[].denseVector.vertexEmbeddingConfig.textTemplate` | title + body | Which fields feed the embedding. |
| `indexes[].storeFields` | `title`, `url` | Fields to return inline with each hit. |
| `indexes[].filterFields` | none | Fields to filter on inline (e.g. a `category`). |
| `deletionPolicy` | `DELETE` | `PREVENT` once the collection holds data you cannot regenerate. |

Every index setting but labels is immutable: change the metric, the fields, or the infrastructure and the index is replaced and rebuilt from the collection's data.
