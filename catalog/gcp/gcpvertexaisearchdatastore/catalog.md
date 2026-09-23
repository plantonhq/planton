# GCP Vertex AI Search Data Store

Stands up a Vertex AI Search data store -- the corpus a search, chat, or recommendation app answers from. Declare what the store holds (structured records, PDFs and other documents, or a public website), which apps may use it, and how documents are parsed and chunked, and import the content through the API, BigQuery, or Cloud Storage. It is the foundation of Google's enterprise search and the grounding source for chat and RAG over your own data.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `discoveryengine.googleapis.com` on the project
- **Data store** -- a `discoveryengine.DataStore` with its vertical, content type, solutions, and document processing
- **Schema, target sites, sitemaps** -- one resource each for the declared schema, every crawl pattern, and every sitemap

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Discovery Engine admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpKmsKey** -- for customer-managed encryption, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Search Data Store**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Structured Records** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchDataStore
metadata:
  name: product-docs
  org: acme-corp
  env: prod
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

This creates a document store enrolled in search and chat that parses documents by layout and chunks them into passages. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference the store's `data_store_id` output from a `GcpVertexAiSearchEngine`'s `dataStoreIds`, its `name` from an engine's controls, and a `GcpKmsKey` from `kmsKeyName` when encryption must be customer-managed.

## Key Configuration

These are the most important decisions when configuring a data store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**What the store holds** -- `contentConfig` decides everything downstream: `NO_CONTENT` for structured JSON records searched by their fields, `CONTENT_REQUIRED` for PDFs, HTML, and DOCX with metadata, `PUBLIC_WEBSITE` for a site Google crawls by the URL patterns in `targetSites`.

**Which apps may use it** -- `solutionTypes` enrolls the store in search, chat, and recommendation; an engine can only read stores enrolled in its solution, and the vertical must match too.

**How documents are parsed** -- the layout parser understands headings and tables and, with chunking, produces the passages RAG applications retrieve; the digital parser is Google's default; OCR handles scans. All of it is immutable, so choose before importing.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `data_store_id` | The store's id | A `GcpVertexAiSearchEngine`'s `dataStoreIds` |
| `name` | The store's full resource name | An engine's boost, filter, and promote controls; the widget's per-store UI |
| `location` | The store's location | The engine's location |
| `default_schema_id` | The default schema's id | Import pipelines |
| `target_site_names`, `sitemap_names` | The folded resources' names | Dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Structured records** -- A `NO_CONTENT` store with Google's inferred schema, enrolled in search and recommendation. Start from the **Structured Records** preset.

**Website advanced site search** -- A `PUBLIC_WEBSITE` store with a verified domain, include and exclude patterns, and a sitemap. Start from the **Website Advanced Site Search** preset.

**Documents for RAG** -- A `CONTENT_REQUIRED` store with layout parsing, chunking, OCR for scans, and CMEK. Start from the **Documents for RAG** preset.

## Works With

- [**GCP Vertex AI Search Engine**](/cloud-catalog/gcp-vertex-ai-search-engine) -- the app over the store
- [**GCP Vertex AI Search Data Connector**](/cloud-catalog/gcp-vertex-ai-search-data-connector) -- stores synced from third-party sources
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
