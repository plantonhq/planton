# GcpVertexAiSearchDataStore — Pulumi Implementation

This directory contains the Pulumi implementation for a Vertex AI Search
data store and its folded schema, target sites, and sitemaps from the
Planton spec: one `gcp.projects.Service` (API enablement), one
`gcp.discoveryengine.DataStore`, at most one `gcp.discoveryengine.Schema`,
one `gcp.discoveryengine.TargetSite` per `spec.target_sites[]` entry, and
one `gcp.discoveryengine.Sitemap` per `spec.sitemap_uris[]` entry.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, the store, and its companions, then exports the outputs |
| `module/locals.go` | The store id and display name defaulted from `metadata.name` |
| `module/data_store.go` | Enables the API; maps the store and the document-processing tree (the `digital_parsing` bool becomes Google's empty marker block) |
| `module/schema.go` | The custom schema when declared |
| `module/target_sites.go` | One target site per pattern, names in manifest order |
| `module/sitemaps.go` | One sitemap per URI, names in manifest order |
| `module/outputs.go` | Output key constants (`name`, `data_store_id`, `location`, `default_schema_id`, `schema_name`, `target_site_names`, `sitemap_names`) |

## Send Posture (parity with Terraform)

- **`DataStoreId` / `DisplayName`** -- `spec.data_store_id` / `spec.display_name`, each defaulting to `metadata.name` (Google requires a display name).
- **Optional strings and lists** (`ContentConfig`, `SolutionTypes`, `KmsKeyName`, the layout parser's lists) are sent only when set so Google's defaults stay in charge.
- **Bools** (`AclEnabled`, `CreateAdvancedSiteSearch`, `SkipDefaultSchemaCreation`, the parser flags) are sent as declared; the two virtual inputs are never read back.
- **`DocumentProcessingConfig`** -- the chunking levers under `LayoutBasedChunkingConfig`; the default parser and each per-file-type override share the spec's parsing-config shape and map to the provider's two distinct block types.
- **Companions** -- `DataStoreId` from the created store, `Location` from the spec, `DeletionPolicy` fanned from the spec; parented to the store so destroy order is companions -> store.
- **No labels** -- Discovery Engine resources carry none.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
