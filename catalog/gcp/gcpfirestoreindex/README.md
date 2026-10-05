# GCP Firestore Index

Deploys a Firestore composite index (`google_firestore_index`) on an existing Firestore database — the prerequisite for multi-field queries.

## What Gets Created

When you deploy a GcpFirestoreIndex resource, Planton provisions:

- **Composite Index** — a `google_firestore_index` on the specified collection with the declared field roles (order, array-contains, vector, or Enterprise text/geo search); the Firestore API is enabled automatically

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing Firestore database** (deploy via GcpFirestoreDatabase first)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `index.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFirestoreIndex
metadata:
  name: orders-by-customer-date
spec:
  database:
    valueFrom:
      kind: GcpFirestoreDatabase
      name: prod-firestore
      fieldPath: status.outputs.database_name
  collection: orders
  fields:
    - fieldPath: customerId
      order: ASCENDING
    - fieldPath: createdAt
      order: DESCENDING
```

Deploy:

```shell
planton apply -f index.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `collection` | `string` | Collection (group) ID the index applies to. Immutable. | Required |
| `fields[]` | `object[]` | Indexed fields in query order. At least one. Immutable. | Required, min 1 |

### Field Roles (exactly one per field)

| Field | Type | Description |
|-------|------|-------------|
| `fields[].fieldPath` | `string` | Dot-separated document field path. Required. |
| `fields[].order` | `string` | `ASCENDING` or `DESCENDING` for scalar sort/filter. |
| `fields[].arrayConfig` | `string` | `CONTAINS` for array-membership queries. |
| `fields[].vectorConfig.dimension` | `int32` | Vector dimension for nearest-neighbor queries (must be last). |
| `fields[].searchConfig.textSpec.indexSpecs[]` | `object[]` | Text search indexing (`indexType`/`matchType`, e.g. `TOKENIZED`/`MATCH_GLOBALLY`). Enterprise-edition databases only. |
| `fields[].searchConfig.geoSpec.geoJsonIndexingDisabled` | `bool` | Geo search indexing; set true to skip GeoJSON points (GeoPoints always index). Enterprise-edition databases only. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project owning the database. |
| `database` | `StringValueOrRef` | `(default)` | Database name. Immutable. |
| `queryScope` | `string` | `COLLECTION` | `COLLECTION`, `COLLECTION_GROUP`, or `COLLECTION_RECURSIVE`. |
| `apiScope` | `string` | `ANY_API` | `ANY_API`, `DATASTORE_MODE_API`, or `MONGODB_COMPATIBLE_API` (Enterprise). |
| `density` | `string` | GCP default (`SPARSE_ALL`) | `SPARSE_ALL`, `SPARSE_ANY`, or `DENSE`. |
| `multikey` | `bool` | `false` | MongoDB-style array indexing (at most one indexed path traverses an array). Requires `apiScope: MONGODB_COMPATIBLE_API`. Immutable. |
| `unique` | `bool` | `false` | Enforce uniqueness of the indexed field values across documents. Immutable. |
| `skipWait` | `bool` | `false` | Return once creation is requested instead of waiting for the background build. |
| `deletionPolicy` | `string` | `DELETE` | `PREVENT` fails destroys; `ABANDON` unmanages the index without deleting it. |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `index_id` | `string` | Fully qualified index resource name |
| `collection` | `string` | Collection (group) the index applies to |

## Important Notes

- **Every property is immutable**: changing anything replaces the index (Firestore rebuilds in the background).
- **Field order matters**: equality filters first, then inequality/sort; vector fields last.
- **Firestore appends `__name__` automatically** when needed.
- **No labels surface**: Firestore indexes do not support GCP labels — both engines skip labels identically.

## Related Kinds

- [GcpFirestoreDatabase](/docs/catalog/gcp/gcpfirestoredatabase) — the database this index belongs to
- [GcpProject](/docs/catalog/gcp/gcpproject) — the project the database lives in

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

For copy-paste ready manifests, see the [presets](presets/).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
