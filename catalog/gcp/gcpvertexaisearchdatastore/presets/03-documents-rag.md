# Documents for RAG

## Use Case

An unstructured document store built for retrieval-augmented generation: contracts, manuals, and reports parsed by layout (headings, tables, figures), annotated by an LLM, chunked into passages that keep their section headings, and encrypted under your own key in the `us` multi-region. Scanned PDFs go through OCR; born-digital pages keep their native text.

## When to Use

- Grounding a chat engine or an agent on your own documents
- Documents with tables and figures whose content must be searchable
- Regulated data that must stay in one region under a customer-managed key

## What This Creates

- A `CONTENT_REQUIRED` data store in `us`, enrolled in search and chat, with `PREVENT` as its destroy policy
- Layout parsing with table and image annotation and the processed document kept for the GetProcessedDocument API; OCR for PDFs with native text preferred
- Chunking at 500 tokens with ancestor headings
- Encryption under a referenced `GcpKmsKey`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us` | `eu` for European residency, `global` when residency does not matter. |
| `documentProcessingConfig.chunkingConfig.chunkSize` | `500` | Smaller chunks (100-500) for precise citations, larger for more context per passage. |
| `documentProcessingConfig.defaultParsingConfig.layoutParsingConfig.enableLlmLayoutParsing` | off | On for complex PDF layouts the plain layout parser misreads, at higher parsing cost. |
| `kmsKeyName` | a `GcpKmsKey` reference | Remove for Google-managed encryption. |
| `deletionPolicy` | `PREVENT` | `DELETE` for disposable environments. |

The document processing config is immutable: a change replaces the store, so decide the parser and chunking before importing. The key needs the Discovery Engine service agent as an encrypter/decrypter and the location needs a CMEK registration.
