# GcpVertexAiSearchDataStore — Terraform Implementation

This directory contains the Terraform implementation for a Vertex AI Search
data store and its folded schema, target sites, and sitemaps from the
Planton spec: one `google_project_service` (API enablement), one
`google_discovery_engine_data_store`, at most one
`google_discovery_engine_schema` (`count`), one
`google_discovery_engine_target_site` per `spec.target_sites[]` entry
(`for_each` keyed by pattern), and one `google_discovery_engine_sitemap`
per `spec.sitemap_uris[]` entry (`for_each` keyed by URI).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials
  or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the store id and display name defaulted from `metadata.name`, null-for-empty optionals, the target-site and sitemap maps |
| `main.tf` | `google_project_service`, `google_discovery_engine_data_store` with the document-processing tree, `google_discovery_engine_schema`, `google_discovery_engine_target_site`, `google_discovery_engine_sitemap` |
| `outputs.tf` | `name`, `data_store_id`, `location`, `default_schema_id`, `schema_name`, `target_site_names`, `sitemap_names` |

## Send Posture

- **`data_store_id` / `display_name`** -- `spec.data_store_id` / `spec.display_name`, each defaulting to `metadata.name` -- PARITY with the Pulumi module.
- **Optional strings and lists** become `null` when empty so the provider omits them (`content_config`, `solution_types`, `kms_key_name`, the layout parser's lists).
- **`digital_parsing_config {}`** -- emitted from the spec's `digital_parsing` bool through a `dynamic` block (the marker-block idiom); the layout and OCR arms are `dynamic` blocks with their settings.
- **Companions** -- `data_store_id` from the created store, `location` from the spec, `deletion_policy` fanned; `type` on a target site becomes `null` when empty so Google defaults to INCLUDE.
- **List outputs** -- in manifest order (comprehensions over the spec lists, not over the `for_each` maps), so both engines export the same lists.
- **No labels** -- Discovery Engine resources carry none.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
