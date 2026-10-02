# GcpVectorSearchCollection — Pulumi Implementation

This directory contains the Pulumi implementation for a Vector Search
collection and its folded indexes from the Planton spec: one
`gcp.projects.Service` (API enablement), one `gcp.vectorsearch.Collection`,
and one `gcp.vectorsearch.Index` per `spec.indexes[]` entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `collection` |
| `module/locals.go` | The collection id defaulted from `metadata.name`; the `planton-ai_*` attribution labels |
| `module/collection.go` | Enables the API; maps the collection, its vector fields, and every index; exports the outputs |
| `module/outputs.go` | Output key constants (`name`, `collection_id`, `location`, `index_names`, `index_count`) |

## Send Posture (parity with Terraform)

- **`CollectionId`** -- `spec.collection_id`, defaulting to `metadata.name`.
- **`VectorSchemas[]`** -- a dense field carries `Dimensions` (when set) and `VertexEmbeddingConfig` (when set); the spec's `sparse_vector` bool becomes Google's empty `SparseVector` block.
- **`EncryptionSpec`** -- emitted only when `kms_key_name` is set.
- **Indexes** -- `CollectionId` and `Location` from the parent; `DistanceMetric`, `DenseScann.FeatureNormType`, `DedicatedInfrastructure` and its replica bounds are Optional+Computed on Google's side and sent only when set; `Labels` are the index's own under the attribution set; `DeletionPolicy` fanned from the spec.
- **`index_names`** -- the created indexes' names in manifest order; `index_count` the declared count.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
