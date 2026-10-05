# GcpFirestoreIndex

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpFirestoreIndexSpec defines a composite index on a Firestore
database — the prerequisite for any query that filters or orders on
multiple fields (Firestore rejects such queries outright until a
matching composite index exists). Declaring indexes as infrastructure
makes a deployment's query capabilities reviewable and reproducible
instead of click-created from error-message links.

Indexes are many-per-database with independent lifecycles, and every
property is immutable: changing anything replaces the index (Firestore
rebuilds it in the background; the old index serves queries until the
new one is ready when applied create-before-destroy).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreIndex
metadata:
  name: orders-by-customer
spec:
  # GCP project owning the database. Replace with your project ID.
  projectId:
    value: my-gcp-project-123

  # Firestore database — empty would fall back to "(default)".
  database:
    value: my-firestore-db

  collection: orders

  fields:
    - fieldPath: customerId
      order: ASCENDING
    - fieldPath: createdAt
      order: DESCENDING
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.database` | `string \| valueFrom` |  |  | GcpFirestoreDatabase (`status.outputs.database_name`) |
| `spec.collection` | `string` | yes |  |  |
| `spec.queryScope` | `string` |  | `COLLECTION` |  |
| `spec.apiScope` | `string` |  | `ANY_API` |  |
| `spec.density` | `string` |  |  |  |
| `spec.fields` | `[]GcpFirestoreIndexField` | yes |  |  |
| `spec.fields[].fieldPath` | `string` | yes |  |  |
| `spec.fields[].order` | `string` |  |  |  |
| `spec.fields[].arrayConfig` | `string` |  |  |  |
| `spec.fields[].vectorConfig` | `GcpFirestoreIndexVectorConfig` |  |  |  |
| `spec.fields[].vectorConfig.dimension` | `int32` | yes |  |  |
| `spec.fields[].searchConfig` | `GcpFirestoreIndexSearchConfig` |  |  |  |
| `spec.fields[].searchConfig.textSpec` | `GcpFirestoreIndexTextSpec` |  |  |  |
| `spec.fields[].searchConfig.textSpec.indexSpecs` | `[]GcpFirestoreIndexTextIndexSpec` | yes |  |  |
| `spec.fields[].searchConfig.textSpec.indexSpecs[].indexType` | `string` |  |  |  |
| `spec.fields[].searchConfig.textSpec.indexSpecs[].matchType` | `string` |  |  |  |
| `spec.fields[].searchConfig.geoSpec` | `GcpFirestoreIndexGeoSpec` |  |  |  |
| `spec.fields[].searchConfig.geoSpec.geoJsonIndexingDisabled` | `bool` |  |  |  |
| `spec.multikey` | `bool` |  |  |  |
| `spec.unique` | `bool` |  |  |  |
| `spec.skipWait` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project owning the database. Can be a literal project ID or a
reference to a GcpProject resource. If omitted, the provider's
default project is used.
Immutable: changing the project destroys and recreates the index.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.database

`string | valueFrom`

The Firestore database the index belongs to — the database name (a
GcpFirestoreDatabase reference resolves to it). Empty falls back to
the project's "(default)" database. Immutable after creation.

- references: GcpFirestoreDatabase (`status.outputs.database_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFirestoreDatabase, name: <that resource's name>, fieldPath: status.outputs.database_name}} -- a bare string does not parse

### spec.collection

`string` · required

The collection (group) ID the index applies to. Immutable after
creation.

- rule: {"required":true}

### spec.queryScope

`string` · optional (explicit presence)

Scope of the queries the index serves.
COLLECTION: queries against a single collection at this path
  (the default).
COLLECTION_GROUP: queries against every collection with this ID
  anywhere in the database.
COLLECTION_RECURSIVE: queries against the collection and everything
  beneath it (Datastore Mode only).

- default: `COLLECTION`
- rule: query_scope must be COLLECTION, COLLECTION_GROUP, or COLLECTION_RECURSIVE

### spec.apiScope

`string` · optional (explicit presence)

Which API surface the index serves.
ANY_API: Firestore Native queries (the default).
DATASTORE_MODE_API: Datastore Mode queries.
MONGODB_COMPATIBLE_API: Firestore Enterprise (MongoDB-compatible)
  queries — required for multikey and search-config indexes, and
  only valid on an ENTERPRISE-edition database. Requires
  query_scope COLLECTION_GROUP explicitly (the API rejects the
  COLLECTION default for this scope).

- default: `ANY_API`
- rule: api_scope must be ANY_API, DATASTORE_MODE_API, or MONGODB_COMPATIBLE_API

### spec.density

`string`

Index density. Leave empty for GCP's default (SPARSE_ALL).
DENSE indexes also include documents missing the indexed fields —
required for some Datastore Mode query shapes.

- rule: density must be SPARSE_ALL, SPARSE_ANY, or DENSE

### spec.fields

`[]GcpFirestoreIndexField` · required

The indexed fields, in query order (equality filters first, then
inequality/sort fields; a vector field, if any, last). At least one
field. Firestore appends __name__ automatically.

- rule: {"repeated":{"minItems":"1"}}
- rule: each field declares exactly one of order, array_config, vector_config, or search_config

### spec.fields[].fieldPath

`string` · required

Dot-separated field path in the document (e.g. "user.age").

- rule: {"required":true}

### spec.fields[].order

`string`

Sort order for a scalar field: ASCENDING or DESCENDING.

- rule: order must be ASCENDING or DESCENDING

### spec.fields[].arrayConfig

`string`

CONTAINS declares an array-membership index on the field (enables
array-contains queries).

- rule: array_config must be CONTAINS

### spec.fields[].vectorConfig

`GcpFirestoreIndexVectorConfig`

Vector index configuration for nearest-neighbor queries.

### spec.fields[].vectorConfig.dimension

`int32` · required

Dimension of the vectors indexed on this field (the embedding
model's output size, e.g. 768).

- rule: {"required":true,"int32":{"gte":1}}

### spec.fields[].searchConfig

`GcpFirestoreIndexSearchConfig`

Text or geo search index configuration for the field — the
Firestore Enterprise search surface. Requires an ENTERPRISE-edition
database (GcpFirestoreDatabase database_edition: ENTERPRISE); text
search additionally pairs with api_scope MONGODB_COMPATIBLE_API,
while geo search works under the default scope. Search and
non-search fields cannot mix in one index (API-enforced): when any
field carries search_config, every field must.

- rule: search_config must declare text_spec and/or geo_spec

### spec.fields[].searchConfig.textSpec

`GcpFirestoreIndexTextSpec`

Text search index specification for the field.

### spec.fields[].searchConfig.textSpec.indexSpecs

`[]GcpFirestoreIndexTextIndexSpec` · required

Index specifications; at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.fields[].searchConfig.textSpec.indexSpecs[].indexType

`string`

How the text field value is indexed (e.g. "TOKENIZED").

### spec.fields[].searchConfig.textSpec.indexSpecs[].matchType

`string`

How the text field value is matched (e.g. "MATCH_GLOBALLY").

### spec.fields[].searchConfig.geoSpec

`GcpFirestoreIndexGeoSpec`

Geo search index specification for the field.

### spec.fields[].searchConfig.geoSpec.geoJsonIndexingDisabled

`bool`

If true, disables GeoJSON indexing for the field (GeoJSON points are
indexed by default). Firestore GeoPoints are indexed regardless.

### spec.multikey

`bool`

Whether the index is multikey: at most one indexed path may reach or
traverse an array (MongoDB-style array indexing). Only valid with
api_scope MONGODB_COMPATIBLE_API. Immutable after creation.

### spec.unique

`bool`

Whether the index enforces uniqueness: all values of the indexed
field(s) must be unique across documents. Immutable after creation.

### spec.skipWait

`bool`

If true, the deploy returns as soon as index creation is REQUESTED
instead of waiting for the (potentially long) background build to
finish. The index serves queries only once the build completes —
use when orchestrating many indexes and the caller polls readiness
itself.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the index is deleted
  "PREVENT" -- destroy FAILS; protects indexes whose rebuild would
               be expensive on a large collection
  "ABANDON" -- the index is removed from management but kept
               serving queries in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `multikey_requires_mongodb_scope`: multikey indexes require api_scope MONGODB_COMPATIBLE_API
- `mongodb_scope_requires_collection_group`: api_scope MONGODB_COMPATIBLE_API requires query_scope COLLECTION_GROUP (set it explicitly; the COLLECTION default is rejected by the API)
- `search_and_non_search_fields_dont_mix`: an index is either a search index (every field carries search_config) or a non-search index (no field does) — the API rejects mixed indexes

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFirestoreIndex, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.index_id` | `string` | Server-defined index resource name (projects/{project}/databases/{database}/collectionGroups/{collection}/indexes/{id}). The canonical identifier for Admin API calls. |
| `status.outputs.collection` | `string` | The collection (group) the index applies to — confirms the target without chasing the reference chain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.database` | GcpFirestoreDatabase | `status.outputs.database_name` |

## See Also

- [Overview](../README.md)
