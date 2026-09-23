# GcpVertexAiSearchDataStore

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiSearchDataStoreSpec defines a Vertex AI Search data store
(`google_discovery_engine_data_store`, the Discovery Engine API behind
the console's AI Applications / Gemini Enterprise): the corpus a search,
chat, or recommendation engine answers from. Folded in because each
belongs to exactly one store and shares its life: the store's schema
(`google_discovery_engine_schema`), the URL patterns a website store
crawls (`google_discovery_engine_target_site`), and the sitemaps advanced
site search reads (`google_discovery_engine_sitemap`).

Three kinds of store, chosen by content_config: NO_CONTENT holds
structured records (JSON documents; the schema says which fields are
searchable), CONTENT_REQUIRED holds unstructured documents (PDF, HTML,
DOCX with optional metadata), PUBLIC_WEBSITE indexes a public site by
URL pattern. Documents themselves are data-plane -- imported through the
Discovery Engine API, BigQuery, or Cloud Storage -- and are not part of
this block. Engines that search the store reference it by data_store_id
(GcpVertexAiSearchEngine).

Immutable: data_store_id, location, industry_vertical, content_config,
solution_types, acl_enabled, the advanced site search and document
processing configs. display_name and kms_key_name update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchDataStore
metadata:
  name: product-docs
spec:
  projectId:
    value: my-gcp-project
  # Discovery Engine multi-regions: global, us, or eu. Engines must match.
  location: global
  displayName: Product docs
  industryVertical: GENERIC
  # Unstructured documents (PDF, HTML, DOCX) imported later through the
  # API, BigQuery, or Cloud Storage; the store enrolls in search and chat.
  contentConfig: CONTENT_REQUIRED
  solutionTypes:
    - SOLUTION_TYPE_SEARCH
    - SOLUTION_TYPE_CHAT
  # Parse documents by layout (headings, tables) and chunk them into
  # passages -- the shape RAG applications want. PDFs that are scans go
  # through OCR instead.
  documentProcessingConfig:
    chunkingConfig:
      chunkSize: 500
      includeAncestorHeadings: true
    defaultParsingConfig:
      layoutParsingConfig:
        enableTableAnnotation: true
        excludeHtmlElements:
          - nav
          - footer
    parsingConfigOverrides:
      - fileType: pdf
        parsingConfig:
          ocrParsingConfig:
            useNativeText: true
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.dataStoreId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.industryVertical` | `string` | yes |  |  |
| `spec.contentConfig` | `string` |  |  |  |
| `spec.solutionTypes` | `[]string` |  |  |  |
| `spec.aclEnabled` | `bool` |  |  |  |
| `spec.createAdvancedSiteSearch` | `bool` |  |  |  |
| `spec.advancedSiteSearchConfig` | `GcpVertexAiSearchDataStoreAdvancedSiteSearchConfig` |  |  |  |
| `spec.advancedSiteSearchConfig.disableInitialIndex` | `bool` |  |  |  |
| `spec.advancedSiteSearchConfig.disableAutomaticRefresh` | `bool` |  |  |  |
| `spec.skipDefaultSchemaCreation` | `bool` |  |  |  |
| `spec.schema` | `GcpVertexAiSearchDataStoreSchema` |  |  |  |
| `spec.schema.schemaId` | `string` | yes |  |  |
| `spec.schema.jsonSchema` | `string` | yes |  |  |
| `spec.documentProcessingConfig` | `GcpVertexAiSearchDataStoreDocumentProcessingConfig` |  |  |  |
| `spec.documentProcessingConfig.chunkingConfig` | `GcpVertexAiSearchDataStoreChunkingConfig` |  |  |  |
| `spec.documentProcessingConfig.chunkingConfig.chunkSize` | `int32` |  |  |  |
| `spec.documentProcessingConfig.chunkingConfig.includeAncestorHeadings` | `bool` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig` | `GcpVertexAiSearchDataStoreParsingConfig` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.digitalParsing` | `bool` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig` | `GcpVertexAiSearchDataStoreLayoutParsingConfig` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableGetProcessedDocument` | `bool` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableImageAnnotation` | `bool` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableLlmLayoutParsing` | `bool` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableTableAnnotation` | `bool` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.excludeHtmlClasses` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.excludeHtmlElements` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.excludeHtmlIds` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.structuredContentTypes` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.ocrParsingConfig` | `GcpVertexAiSearchDataStoreOcrParsingConfig` |  |  |  |
| `spec.documentProcessingConfig.defaultParsingConfig.ocrParsingConfig.useNativeText` | `bool` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides` | `[]GcpVertexAiSearchDataStoreParsingConfigOverride` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].fileType` | `string` | yes |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig` | `GcpVertexAiSearchDataStoreParsingConfig` | yes |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.digitalParsing` | `bool` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig` | `GcpVertexAiSearchDataStoreLayoutParsingConfig` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableGetProcessedDocument` | `bool` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableImageAnnotation` | `bool` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableLlmLayoutParsing` | `bool` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableTableAnnotation` | `bool` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.excludeHtmlClasses` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.excludeHtmlElements` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.excludeHtmlIds` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.structuredContentTypes` | `[]string` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.ocrParsingConfig` | `GcpVertexAiSearchDataStoreOcrParsingConfig` |  |  |  |
| `spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.ocrParsingConfig.useNativeText` | `bool` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.targetSites` | `[]GcpVertexAiSearchDataStoreTargetSite` |  |  |  |
| `spec.targetSites[].providedUriPattern` | `string` | yes |  |  |
| `spec.targetSites[].type` | `string` |  |  |  |
| `spec.targetSites[].exactMatch` | `bool` |  |  |  |
| `spec.sitemapUris` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the store lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Where the store and its engines live: "global", "us", or "eu"
(Discovery Engine multi-regions, not Compute regions). An engine must
be in the same location as its stores. Immutable.

- rule: {"required":true,"string":{"in":["global","us","eu"]}}

### spec.dataStoreId

`string`

The store's id -- 1-63 characters, RFC 1034 (lowercase letters, digits,
hyphens; starts with a letter). Defaults to metadata.name. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.displayName

`string`

Human-readable name shown in the console (up to 128 characters).
Defaults to metadata.name. Mutable.

- rule: {"string":{"maxLen":"128"}}

### spec.industryVertical

`string` · required

The industry the corpus belongs to: GENERIC for most content, MEDIA for
videos/articles/music with media recommendation, HEALTHCARE_FHIR for
FHIR stores. Fixes which engines can use the store. Immutable.

- rule: {"required":true,"string":{"in":["GENERIC","MEDIA","HEALTHCARE_FHIR"]}}

### spec.contentConfig

`string`

What the store holds: NO_CONTENT for structured records (JSON with a
schema), CONTENT_REQUIRED for unstructured documents (PDF, HTML, DOCX
and their metadata), PUBLIC_WEBSITE for a public website crawled by
URL pattern. Empty lets Google default (NO_CONTENT). Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NO_CONTENT","CONTENT_REQUIRED","PUBLIC_WEBSITE"]}}

### spec.solutionTypes

`[]string`

The solutions the store enrolls in -- an engine can only use stores
enrolled in its solution: SOLUTION_TYPE_SEARCH, SOLUTION_TYPE_CHAT,
SOLUTION_TYPE_RECOMMENDATION, SOLUTION_TYPE_GENERATIVE_CHAT. Empty lets
Google default (search). Immutable.

- rule: {"repeated":{"items":{"string":{"in":["SOLUTION_TYPE_RECOMMENDATION","SOLUTION_TYPE_SEARCH","SOLUTION_TYPE_CHAT","SOLUTION_TYPE_GENERATIVE_CHAT"]}}}}

### spec.aclEnabled

`bool`

True means every document carries access-control information that is
ingested with it and enforced at search time (needs the location's
ACL config / identity provider). Documents in an ACL-enabled store
cannot be read back through GetDocument. Immutable.

### spec.createAdvancedSiteSearch

`bool`

Create an ADVANCED site search store (PUBLIC_WEBSITE only; ignored
otherwise): domain-verified crawling, sitemaps, and more index quota
versus basic site search over public pages. Input-only; Google does
not read it back. Immutable.

### spec.advancedSiteSearchConfig

`GcpVertexAiSearchDataStoreAdvancedSiteSearchConfig`

Advanced site search behavior; applies only with
create_advanced_site_search. Immutable.

### spec.advancedSiteSearchConfig.disableInitialIndex

`bool`

Do not index the site when the store is created; indexing starts on
demand (the console's "Recrawl" or the API).

### spec.advancedSiteSearchConfig.disableAutomaticRefresh

`bool`

Do not refresh the index automatically; content is recrawled only on
demand.

### spec.skipDefaultSchemaCreation

`bool`

Do not create Google's default schema. Required when declaring a
custom schema below, and only then: a store with no schema accepts no
documents until one exists. Input-only.

### spec.schema

`GcpVertexAiSearchDataStoreSchema`

The store's schema (with skip_default_schema_creation). Omit to let
Google create the default schema and infer fields from the data.

### spec.schema.schemaId

`string` · required

The schema's id within the store, e.g. "default_schema" or your own.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9](?:[-a-z0-9_]{0,61}[a-z0-9])?$"}}

### spec.schema.jsonSchema

`string` · required

The schema as a compact JSON string (Google normalizes it): a JSON
Schema object whose properties may carry Vertex AI Search's
keyPropertyMapping, indexable, searchable, retrievable, dynamicFacetable
annotations; "datetime_detection" and "geolocation_detection" at the
top level ask Google to infer those types.

- rule: {"required":true,"string":{"minLen":"2"}}

### spec.documentProcessingConfig

`GcpVertexAiSearchDataStoreDocumentProcessingConfig`

How documents are parsed and chunked. Omit for Google's default: the
digital parser, whole documents. Immutable.

### spec.documentProcessingConfig.chunkingConfig

`GcpVertexAiSearchDataStoreChunkingConfig`

Layout-based chunking. Omit to index whole documents.

### spec.documentProcessingConfig.chunkingConfig.chunkSize

`int32` · optional (explicit presence)

Token limit per chunk, 100-500 (Google defaults to 500). Sent only
when set.

- rule: {"int32":{"lte":500,"gte":100}}

### spec.documentProcessingConfig.chunkingConfig.includeAncestorHeadings

`bool`

Prepend the headings above a chunk (section, subsection) to chunks cut
from the middle of a document, so a passage keeps its context.

### spec.documentProcessingConfig.defaultParsingConfig

`GcpVertexAiSearchDataStoreParsingConfig`

The parser applied to every file type not overridden below. Omit for
Google's default (the digital parser).

- rule: a parsing config is at most one of digital_parsing, layout_parsing_config, or ocr_parsing_config

### spec.documentProcessingConfig.defaultParsingConfig.digitalParsing

`bool`

True picks the digital parser: plain text extraction from digital
(born-electronic) documents, Google's default. It has no settings.

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig

`GcpVertexAiSearchDataStoreLayoutParsingConfig`

The layout parser: structure-aware parsing for headings, tables, and
figures -- the parser to pick for RAG quality.

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableGetProcessedDocument

`bool`

Keep the processed (layout-parsed) document available through the
GetProcessedDocument API, so an application can fetch the parsed
structure instead of the raw file.

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableImageAnnotation

`bool`

Have an LLM describe each image during parsing and add the description
to the indexed text, so image content becomes searchable.

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableLlmLayoutParsing

`bool`

Refine the detected PDF layout with an LLM (better reading order and
section boundaries on complex pages, at higher parsing cost).

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableTableAnnotation

`bool`

Have an LLM describe each table during parsing and add the description
to the indexed text.

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.excludeHtmlClasses

`[]string`

HTML classes whose elements are dropped from the parsed content (menus,
footers, cookie banners).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.excludeHtmlElements

`[]string`

HTML element names dropped from the parsed content, e.g. "nav",
"footer", "script".

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.excludeHtmlIds

`[]string`

HTML element ids dropped from the parsed content.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.structuredContentTypes

`[]string`

Structured content types the parser must extract from the document.
Google supports "shareholder-structure" today.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.defaultParsingConfig.ocrParsingConfig

`GcpVertexAiSearchDataStoreOcrParsingConfig`

The OCR parser for scanned PDFs.

### spec.documentProcessingConfig.defaultParsingConfig.ocrParsingConfig.useNativeText

`bool`

On pages that already carry a text layer, use that native text instead
of the OCR output -- faster and more accurate for mixed documents.

### spec.documentProcessingConfig.parsingConfigOverrides

`[]GcpVertexAiSearchDataStoreParsingConfigOverride`

Per-file-type parser overrides, one per file type.

### spec.documentProcessingConfig.parsingConfigOverrides[].fileType

`string` · required

The file type this override applies to: "pdf" (digital, OCR, or layout
parsing), "html" (digital or layout), "docx", "pptx", "xlsm", "xlsx"
(digital or layout). One override per file type.

- rule: {"required":true,"string":{"in":["pdf","html","docx","pptx","xlsm","xlsx"]}}

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig

`GcpVertexAiSearchDataStoreParsingConfig` · required

The parser for this file type.

- rule: {"required":true}
- rule: a parsing config is at most one of digital_parsing, layout_parsing_config, or ocr_parsing_config

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.digitalParsing

`bool`

True picks the digital parser: plain text extraction from digital
(born-electronic) documents, Google's default. It has no settings.

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig

`GcpVertexAiSearchDataStoreLayoutParsingConfig`

The layout parser: structure-aware parsing for headings, tables, and
figures -- the parser to pick for RAG quality.

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableGetProcessedDocument

`bool`

Keep the processed (layout-parsed) document available through the
GetProcessedDocument API, so an application can fetch the parsed
structure instead of the raw file.

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableImageAnnotation

`bool`

Have an LLM describe each image during parsing and add the description
to the indexed text, so image content becomes searchable.

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableLlmLayoutParsing

`bool`

Refine the detected PDF layout with an LLM (better reading order and
section boundaries on complex pages, at higher parsing cost).

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.enableTableAnnotation

`bool`

Have an LLM describe each table during parsing and add the description
to the indexed text.

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.excludeHtmlClasses

`[]string`

HTML classes whose elements are dropped from the parsed content (menus,
footers, cookie banners).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.excludeHtmlElements

`[]string`

HTML element names dropped from the parsed content, e.g. "nav",
"footer", "script".

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.excludeHtmlIds

`[]string`

HTML element ids dropped from the parsed content.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.layoutParsingConfig.structuredContentTypes

`[]string`

Structured content types the parser must extract from the document.
Google supports "shareholder-structure" today.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.ocrParsingConfig

`GcpVertexAiSearchDataStoreOcrParsingConfig`

The OCR parser for scanned PDFs.

### spec.documentProcessingConfig.parsingConfigOverrides[].parsingConfig.ocrParsingConfig.useNativeText

`bool`

On pages that already carry a text layer, use that native text instead
of the OCR output -- faster and more accurate for mixed documents.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the store's data: a
GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
Omit for Google-managed encryption. Updatable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.targetSites

`[]GcpVertexAiSearchDataStoreTargetSite`

The URL patterns a PUBLIC_WEBSITE store crawls (INCLUDE) or leaves out
(EXCLUDE). Basic site search needs no domain verification for public
pages; advanced site search needs the domain verified in Search
Console. Add or remove a pattern by editing the list; any other change
replaces that target site.

### spec.targetSites[].providedUriPattern

`string` · required

The URI pattern, e.g. "cloud.google.com/docs/*" or "www.example.com".
Without a wildcard and with exact_match false, every page whose address
contains the pattern is included.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.targetSites[].type

`string`

INCLUDE (Google's default when empty) crawls pages matching the
pattern; EXCLUDE removes matching pages from an included site.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["INCLUDE","EXCLUDE"]}}

### spec.targetSites[].exactMatch

`bool`

True matches the pattern exactly (or the one specific page it names);
false (Google's default) matches every page containing the pattern.

### spec.sitemapUris

`[]string`

Public sitemap URIs an advanced site search store reads, e.g.
"https://www.example.com/sitemap.xml". Each is one immutable sitemap
resource.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.deletionPolicy

`string`

What happens to the store, its schema, target sites, and sitemaps when
this resource is destroyed:
  "" / "DELETE" -- everything is deleted, documents included
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in GCP
A store an engine still uses cannot be deleted; destroy the engine
first.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `target_sites.website_only`: target_sites apply only to a PUBLIC_WEBSITE store
- `sitemap_uris.advanced_site_search_only`: sitemap_uris apply only to a PUBLIC_WEBSITE store with create_advanced_site_search
- `schema.requires_skip_default`: a custom schema requires skip_default_schema_creation, because a store holds exactly one schema
- `advanced_site_search_config.requires_flag`: advanced_site_search_config applies only with create_advanced_site_search

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiSearchDataStore, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name of the store: projects/{project}/locations/{location}/collections/default_collection/dataStores/{data_store_id}. Control actions and widget configs on an engine name a store this way. |
| `status.outputs.data_store_id` | `string` | The store's id -- what an engine's data_store_ids lists. |
| `status.outputs.location` | `string` | The store's location (global, us, or eu); an engine must match it. |
| `status.outputs.default_schema_id` | `string` | The id of the default schema Google created (empty when the default schema was skipped). |
| `status.outputs.schema_name` | `string` | Full resource name of the declared schema (empty when none was declared). |
| `status.outputs.target_site_names` | `[]string` | Full resource names of the target sites declared on a website store, in manifest order. |
| `status.outputs.sitemap_names` | `[]string` | Full resource names of the sitemaps declared on an advanced site search store, in manifest order. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpVertexAiSearchEngine | `spec.dataStoreIds` | `status.outputs.data_store_id` |
| GcpVertexAiSearchEngine | `spec.controls[].boostAction.dataStore` | `status.outputs.name` |
| GcpVertexAiSearchEngine | `spec.controls[].filterAction.dataStore` | `status.outputs.name` |
| GcpVertexAiSearchEngine | `spec.controls[].promoteAction.dataStore` | `status.outputs.name` |
| GcpVertexAiSearchEngine | `spec.widgetConfig.uiSettings.dataStoreUiConfigs[].name` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
