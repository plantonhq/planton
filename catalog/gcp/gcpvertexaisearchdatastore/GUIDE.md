# GcpVertexAiSearchDataStore Guide

This guide explains how a Vertex AI Search data store is shaped and the decisions behind the spec. For field-by-field detail see the README; for ready-made configurations see the presets.

The product has worn three names. The API is **Discovery Engine** (`discoveryengine.googleapis.com`), the search product is **Vertex AI Search**, and the console now groups it under **AI Applications**, with **Gemini Enterprise** as the assistant built on the same stores. This kind is the data store in all three vocabularies.

## The store is the corpus; documents arrive later

A data store is a container with a shape: its vertical, what kind of content it holds, which solutions may read it, and how documents are parsed. Documents are imported afterwards -- through the Discovery Engine API, from BigQuery, from Cloud Storage, or by crawling a website -- and that import is deliberately not part of this block. Declaring the store is infrastructure; filling it is a data pipeline.

Three shapes, chosen by `contentConfig`:

- **`NO_CONTENT`** -- structured JSON records searched by their fields (a product catalog, FAQs, a directory). The schema decides which fields are indexable, searchable, retrievable, and facetable.
- **`CONTENT_REQUIRED`** -- unstructured documents (PDF, HTML, DOCX) with optional metadata. The document processing config decides how they are parsed and chunked.
- **`PUBLIC_WEBSITE`** -- a public site Google crawls. The `targetSites` say which URL patterns to include and exclude.

## Solutions and verticals decide which engines may use the store

Every engine reads only stores enrolled in its solution: `SOLUTION_TYPE_SEARCH` for a search engine, `SOLUTION_TYPE_CHAT` for a chat engine, `SOLUTION_TYPE_RECOMMENDATION` for a recommendation engine. `industryVertical` must match too: `MEDIA` for media recommendation, `HEALTHCARE_FHIR` for FHIR stores, `GENERIC` for everything else. Both are immutable, so enroll the store in every solution it might serve at creation -- a search-and-chat store costs nothing more than a search store.

## One schema per store

Google creates a default schema when the store is created and infers fields from the first import. When field types, key properties (`title`, `description`), facets, or retrievability must be exact, skip the default (`skipDefaultSchemaCreation: true`) and declare `schema` with a JSON Schema string. The spec enforces the pairing because Google keeps exactly one schema per store: a custom schema next to the default would be rejected. Write the JSON compact; Google normalizes it, and a compact string re-plans clean.

## Parsing and chunking are chosen once

`documentProcessingConfig` is immutable -- a change replaces the store, documents included -- so it is decided before the first import:

- The **digital parser** (Google's default; `digitalParsing: true` when named explicitly) extracts text from born-digital documents.
- The **layout parser** understands headings, tables, and figures, can annotate images and tables with an LLM, and can keep the processed document for the GetProcessedDocument API. It is the parser for RAG quality.
- The **OCR parser** reads scanned PDFs, preferring native text on pages that have it.

A parsing config picks at most one parser; the spec carries the digital parser as a presence bool because Google expresses it as an empty block beside the two configured ones. `parsingConfigOverrides` replace the default per file type (`pdf`, `html`, `docx`, `pptx`, `xlsm`, `xlsx`). `chunkingConfig` splits documents into passages of 100-500 tokens, optionally prefixed with the headings above them so a passage keeps its context.

## Website stores: basic or advanced

A `PUBLIC_WEBSITE` store crawls the URL patterns in `targetSites` (`INCLUDE` to crawl, `EXCLUDE` to leave out, `exactMatch` for one page). **Basic site search** indexes public pages with no proof of ownership. **Advanced site search** (`createAdvancedSiteSearch: true`) needs the domain verified in Search Console, reads `sitemapUris`, indexes with a larger quota, and unlocks Enterprise-tier features on the engines over it; `advancedSiteSearchConfig` can hold the initial crawl or automatic refresh. The spec walls target sites to website stores and sitemaps to advanced site search.

## Access control per document

`aclEnabled` makes every document carry its own access list, ingested with it and enforced per end user at search time -- the shape an intranet search needs. It requires the location's identity-provider configuration (Google's ACL config, a location singleton outside this block) and is immutable.

## Encryption and destroy

`kmsKeyName` encrypts the store under a Cloud KMS key in the store's location (a multi-region key for `global`, `us`, or `eu`); it is the one setting besides the display name that updates in place. The Discovery Engine service agent needs encrypt/decrypt on the key, and the location needs Google's CMEK registration.

`deletionPolicy` is fanned to the schema, target sites, and sitemaps: `DELETE` (default) removes everything, documents included; `PREVENT` fails any destroying plan; `ABANDON` leaves everything in Google Cloud. Google refuses to delete a store an engine still uses -- destroy the engine first -- and may reserve a deleted store's id for a while.
