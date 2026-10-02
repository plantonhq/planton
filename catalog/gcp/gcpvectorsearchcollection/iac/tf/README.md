# GcpVectorSearchCollection — Terraform Implementation

This directory contains the Terraform implementation for a Vector Search
collection and its folded indexes from the Planton spec: one
`google_project_service` (API enablement), one
`google_vector_search_collection`, and one `google_vector_search_index` per
`spec.indexes[]` entry (`for_each` keyed by `index_id`).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the collection id defaulted from `metadata.name`, null-for-empty optionals, the `planton-ai_*` labels, the index map |
| `main.tf` | `google_project_service`, `google_vector_search_collection`, `google_vector_search_index` |
| `outputs.tf` | `name`, `collection_id`, `location`, `index_names`, `index_count` |

## Send Posture

- **`collection_id`** -- `spec.collection_id`, defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **`vector_schema`** -- one block per `spec.vector_schemas[]` entry; `dense_vector` and its embedding config as nested `dynamic` blocks, `sparse_vector {}` emitted from the spec's bool.
- **`encryption_spec`** -- emitted only when `kms_key_name` is set.
- **Indexes** -- `collection_id` from the created collection, `location` from the spec; `distance_metric`, `dense_scann`, `dedicated_infrastructure` and its replica bounds sent only when set (Optional+Computed); `labels` merge the index's own under the attribution set; `deletion_policy` fanned from the spec.
- **`index_names`** -- in manifest order (a comprehension over `spec.indexes`, not over the `for_each` map), so both engines export the same list.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
