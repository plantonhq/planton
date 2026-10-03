# GCP Firestore Index

Declares a composite index on a Cloud Firestore database — the prerequisite for any query that filters or orders on multiple fields (Firestore rejects such queries outright until a matching index exists). Declaring indexes as infrastructure makes a deployment's query capabilities reviewable and reproducible instead of click-created from error-message links. Supports single-collection and collection-group scopes, Datastore Mode API surfaces, density tuning, and vector fields for nearest-neighbor (embedding similarity) queries.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Firestore Composite Index** -- an index on the specified collection (or collection group) with the declared fields in order; Firestore builds it in the background and appends `__name__` automatically
- **Vector Index** -- created when a field carries `vectorConfig`; enables `find_nearest` queries against embeddings of the declared dimension (flat index type)
- **Search Index** -- created when a field carries `searchConfig` (text and/or geo); the Firestore Enterprise search surface, requiring an ENTERPRISE-edition database
- **Firestore API enablement** -- `firestore.googleapis.com` enabled in the target project (never disabled on destroy)

Single-field indexes are built by Firestore automatically and never need this resource. Every index property is immutable at the API — changing the definition replaces the index with a background rebuild, and the old index keeps serving queries until the new one is ready (create-before-destroy).

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Firestore database** in the target project. Leave `database` empty to target the project's `"(default)"` database, or reference a `GcpFirestoreDatabase` Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **GCP Firestore Index**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Composite Filter-and-Sort Index** preset in the [Presets](#presets) tab for the classic equality-filter-plus-sort query shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreIndex
metadata:
  name: orders-by-customer
  org: acme-corp
  env: prod
spec:
  collection: orders
  fields:
    - fieldPath: customerId
      order: ASCENDING
    - fieldPath: createdAt
      order: DESCENDING
```

```shell
planton apply -f firestore-index.yaml
```

This creates a composite index on the default database's `orders` collection serving `WHERE customerId == X ORDER BY createdAt DESC` queries. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the index to a Firestore database deployed in the same InfraPipeline:

```yaml
spec:
  database:
    valueFrom:
      kind: GcpFirestoreDatabase
      name: orders-database
      fieldPath: status.outputs.database_name
  collection: orders
```

The InfraPipeline resolves the dependency graph, deploys the database first, then declares its indexes.

## Key Configuration

These are the most important decisions when configuring a Firestore index. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Fields and order** -- Declare the fields in QUERY order: equality filters first, then the inequality or sort field, then array/vector fields. Each field plays exactly one role — `order` (`ASCENDING`/`DESCENDING`), `arrayConfig` (`CONTAINS` for array-contains queries), `vectorConfig` (nearest-neighbor), or `searchConfig` (Enterprise text/geo search). Sort direction matters: an index on `(customerId ASC, createdAt DESC)` does not serve the reverse ordering.

**Query scope** -- `COLLECTION` (the default) serves queries against a single collection at one path. `COLLECTION_GROUP` serves fan-out queries across every collection with this ID anywhere in the database — those need their own indexes.

**Vector fields** -- A field with `vectorConfig.dimension` enables Firestore's `find_nearest` embedding queries. The dimension must match the embedding model's output size exactly (768 for text-embedding-004, 1536 for OpenAI ada-002), and the vector field must be the LAST field of the index.

**Density** -- Leave empty for GCP's default (`SPARSE_ALL`). `DENSE` also indexes documents missing the indexed fields — required for some Datastore Mode query shapes, at higher storage and write cost.

**Enterprise surface** -- On an ENTERPRISE-edition database (`GcpFirestoreDatabase` `databaseEdition: ENTERPRISE`), `searchConfig` builds text (`indexType: TOKENIZED`, `matchType: MATCH_GLOBALLY`) or geo search indexes, `apiScope: MONGODB_COMPATIBLE_API` serves MongoDB-compatible queries, `multikey` allows one indexed path to traverse arrays, and `unique` enforces uniqueness across documents.

**Lifecycle controls** -- `skipWait` returns as soon as index creation is requested (useful when orchestrating many indexes; the background build continues). `deletionPolicy: PREVENT` protects indexes whose rebuild would be expensive on large collections; `ABANDON` unmanages without deleting.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpFirestoreDatabase** (optional) | `database` | `status.outputs.database_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `index_id` | Full index resource name (`projects/{project}/databases/{database}/collectionGroups/{collection}/indexes/{id}`) | Firestore Admin API calls addressing the index |
| `collection` | The collection (group) ID the index serves | Confirms the target without dereferencing the spec |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Composite Filter Sort** -- The classic multi-field index: equality filter on one field, sort on another (`WHERE customerId == X ORDER BY createdAt DESC`). The most common index shape for list views. Start from the **Composite Filter-and-Sort Index** preset.

**Vector Neighbors** -- A filter field plus a vector field last, enabling filtered nearest-neighbor queries over embeddings — the Firestore-native RAG/semantic-search building block. Start from the **Vector Nearest-Neighbor Index** preset.

## Works With

- [**GCP Firestore Database**](/infra-catalog/gcp-firestore-database) -- provides the database the index attaches to
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project when it differs from the connection default
- [**GCP Firestore Backup Schedule**](/infra-catalog/gcp-firestore-backup-schedule) -- the same database's protection layer, declared side by side
