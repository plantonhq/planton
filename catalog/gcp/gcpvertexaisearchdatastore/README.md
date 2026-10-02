# GCP Vertex AI Search Data Store

A Vertex AI Search data store -- the corpus a search, chat, or recommendation engine answers from -- on the Discovery Engine API (the console calls the product AI Applications / Gemini Enterprise). Declare what the store holds (structured records, unstructured documents, or a public website), which solutions it enrolls in, how documents are parsed and chunked, and -- folded in because each belongs to exactly one store -- its schema, the URL patterns a website store crawls, and the sitemaps advanced site search reads. Documents themselves are imported through the API, BigQuery, or Cloud Storage; engines reference the store by `data_store_id`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `discoveryengine.googleapis.com` on the project (never disabled on destroy)
- **Data store** -- a `discovery_engine_data_store` with its vertical, content type, solutions, ACL flag, document processing config, and optional CMEK
- **Schema** -- a `discovery_engine_schema` when `schema` is declared (with `skipDefaultSchemaCreation`)
- **Target sites** -- one `discovery_engine_target_site` per `targetSites[]` entry on a website store
- **Sitemaps** -- one `discovery_engine_sitemap` per `sitemapUris[]` entry on an advanced site search store

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Discovery Engine admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpKmsKey`** -- a key in the store's location for customer-managed encryption (`kmsKeyName`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchDataStore
metadata:
  name: product-docs
spec:
  location: global
  industryVertical: GENERIC
  contentConfig: CONTENT_REQUIRED
  solutionTypes:
    - SOLUTION_TYPE_SEARCH
    - SOLUTION_TYPE_CHAT
  documentProcessingConfig:
    chunkingConfig:
      chunkSize: 500
      includeAncestorHeadings: true
    defaultParsingConfig:
      layoutParsingConfig:
        enableTableAnnotation: true
```

```shell
planton apply -f vertex-ai-search-data-store.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | `global`, `us`, or `eu` -- a Discovery Engine multi-region, matched by every engine over the store. Immutable. |
| `industryVertical` | `string` | `GENERIC`, `MEDIA`, or `HEALTHCARE_FHIR`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `dataStoreId` | `string` | `metadata.name` | RFC 1034 id. Immutable. |
| `displayName` | `string` | `metadata.name` | Console name; mutable. |
| `contentConfig` | `string` | `NO_CONTENT` | `NO_CONTENT` (structured records), `CONTENT_REQUIRED` (documents), `PUBLIC_WEBSITE` (a crawled site). Immutable. |
| `solutionTypes` | `[]string` | search | The solutions the store enrolls in; an engine can only use stores enrolled in its solution. Immutable. |
| `aclEnabled` | `bool` | `false` | Documents carry access lists enforced per end user. Immutable. |
| `createAdvancedSiteSearch`, `advancedSiteSearchConfig` | | `false` | Advanced site search on a website store (verified domain, sitemaps, more quota). Immutable. |
| `skipDefaultSchemaCreation`, `schema` | | inferred | Skip Google's default schema and declare your own `{ schemaId, jsonSchema }` (exactly one schema per store). |
| `documentProcessingConfig` | `object` | digital parser, whole documents | `chunkingConfig { chunkSize, includeAncestorHeadings }`, `defaultParsingConfig` (exactly one of `digitalParsing: true`, `layoutParsingConfig`, `ocrParsingConfig`), `parsingConfigOverrides[]` per file type. Immutable. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Updatable. |
| `targetSites[]` | `[]object` | none | `providedUriPattern`, `type` (`INCLUDE` / `EXCLUDE`), `exactMatch` -- website stores only. |
| `sitemapUris[]` | `[]string` | none | Sitemaps for advanced site search. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; fanned to the schema, target sites, and sitemaps. |

### Validation Rules

- `targetSites` need `contentConfig: PUBLIC_WEBSITE`; `sitemapUris` and `advancedSiteSearchConfig` need `createAdvancedSiteSearch` as well.
- A custom `schema` requires `skipDefaultSchemaCreation` (Google keeps one schema per store).
- A parsing config is at most one of digital, layout, or OCR; override file types are `pdf`, `html`, `docx`, `pptx`, `xlsm`, `xlsx`; chunk sizes are 100-500.
- Locations, verticals, content configs, solution types, and target site types are Google's values.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/collections/default_collection/dataStores/{id}` |
| `data_store_id` | `string` | The store's id -- what an engine's `dataStoreIds` lists |
| `location` | `string` | The store's location |
| `default_schema_id` | `string` | The default schema's id (empty when skipped) |
| `schema_name` | `string` | The declared schema's full name (empty when none) |
| `target_site_names` | `[]string` | The target sites' full names, in manifest order |
| `sitemap_names` | `[]string` | The sitemaps' full names, in manifest order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Almost everything is immutable.** Only the display name and the KMS key update in place; a change to the vertical, content type, solutions, ACL flag, or document processing replaces the store, documents included. Decide the parser and chunking before importing.
- **An engine holds the store.** Google refuses to delete a store an engine still uses; destroy the engine first.
- **Documents are not part of this block.** Import them through the Discovery Engine API, BigQuery, or Cloud Storage after the store exists.
- **Advanced site search needs a verified domain**; basic site search over public pages does not.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVertexAiSearchEngine** -- the search, chat, or recommendation app over one or more stores
- **GcpVertexAiSearchDataConnector** -- a collection of stores synced from Jira, Confluence, ServiceNow, and other sources
- **GcpKmsKey** -- customer-managed encryption for the store

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
