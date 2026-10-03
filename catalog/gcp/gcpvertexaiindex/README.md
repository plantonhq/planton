# GcpVertexAiIndex

A GCP Vertex AI Vector Search index is the data structure that holds embedding vectors and answers nearest-neighbor queries -- the storage-and-retrieval half of a semantic search, recommendation, or RAG system. The index organizes vectors for fast approximate (or exact) similarity search; serving queries additionally requires a `GcpVertexAiIndexEndpoint` and a `GcpVertexAiDeployedIndex` placing the index onto that endpoint.

## When to Use

Use `GcpVertexAiIndex` when you need:

- Nearest-neighbor search over embedding vectors (semantic search, RAG retrieval, recommendations, deduplication)
- A managed vector database that scales past what in-process libraries handle
- Near-real-time vector upserts (`STREAM_UPDATE`) or bulk rebuilds from Cloud Storage (`BATCH_UPDATE`)
- Infrastructure-as-code management of the index lifecycle and its search geometry

## What This Kind Creates

This component provisions a single Vector Search index. Loading vectors (beyond the optional initial `contentsDeltaUri`), deploying the index to an endpoint, and querying it are separate steps -- deployment is modeled by `GcpVertexAiDeployedIndex`.

## Key Configuration Options

### Update Regime (immutable)

- **`BATCH_UPDATE`** (GCP's default) -- contents are replaced or amended in bulk from Cloud Storage files. Rebuilds take from tens of minutes to hours.
- **`STREAM_UPDATE`** -- individual vectors are upserted/removed via the API within seconds; costs more per GB.

### Search Geometry (`config`, immutable)

- `dimensions` (required) -- the embedding model's output size.
- Algorithm: `treeAhConfig` (approximate; the production choice) XOR `bruteForceConfig` (exact; small corpora and ground-truth evaluation). Tree-AH requires `approximateNeighborsCount`.
- `distanceMeasureType` -- match it to how the embedding model was trained; a mismatch silently degrades result quality.
- `shardSize` -- physical shard size; determines which machine types can serve the index when deployed.

### Initial Data

`contentsDeltaUri` points at a Cloud Storage DIRECTORY of vector files. It may be omitted at creation (the norm for streaming indexes) and set later. Two provider quirks are documented on the field: changes travel in their own single-field update, and GCP never reports the value back on read.

### Encryption (immutable)

`kmsKeyName` (reference a `GcpKmsKey`) turns on customer-managed encryption (CMEK) for the index data; the Vertex AI service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key. Empty means Google-managed keys.

### Destroy behavior

`deletionPolicy` is the client-side lever: empty/`DELETE` deletes the index and every vector in it, `PREVENT` makes destroy fail (a guard for a corpus that took hours to build), `ABANDON` removes it from management but keeps it standing (and billing).

### Labels

User-defined `labels` organize the index for cost attribution and ownership; they merge beneath the platform's attribution labels identically on both engines.

## Outputs

| Output | Description |
|--------|-------------|
| `index_id` | Fully qualified index path -- the deployed index's composition key |
| `index_name` | The GCP-assigned numeric index ID |
| `metadata_schema_uri` | Schema URI for index-type-specific metadata |
| `create_time` | Creation timestamp |
| `update_time` | Last-update timestamp |

## Presets

- **streaming-tree-ah** -- Near-real-time index with tree-AH search
- **batch-from-gcs** -- Bulk-built index from Cloud Storage embeddings
- **brute-force-eval** -- Exact-search index for recall evaluation

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
