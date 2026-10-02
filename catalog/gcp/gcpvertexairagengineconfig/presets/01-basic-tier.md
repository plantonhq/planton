# Basic Tier

## Use Case

Turn on Vertex AI RAG Engine's managed vector database for one location at the experimentation tier: small corpora, latency-insensitive retrieval, or a team that keeps its vectors in an external database (Vector Search, Pinecone, Weaviate) and uses RAG Engine only for ingestion and retrieval orchestration.

## When to Use

- The first RAG Engine corpus in a project
- Prototypes, evaluations, and internal tools
- RAG Engine paired with an external vector database

## What This Creates

- The RAG Engine configuration for `us-central1` set to `BASIC` (a per-location singleton Google owns; applying over an already-configured location changes the tier in place)
- `deletionPolicy: ABANDON`, so removing the block from management leaves the tier and the data alone

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The location your corpora live in; one configuration per location. |
| `tier` | `BASIC` | `SCALED` for production performance and autoscaling; `UNPROVISIONED` disables the service and DELETES its data. |
| `deletionPolicy` | `ABANDON` | `DELETE` only when unprovisioning (data loss) on destroy is what you want, as in a throwaway project. |
