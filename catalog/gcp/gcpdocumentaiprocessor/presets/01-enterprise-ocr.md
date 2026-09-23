# Enterprise OCR

## Use Case

Read the text, layout, and handwriting of any document -- scans, photos, PDFs -- as the general-purpose first step of a document pipeline or a RAG ingestion flow.

## When to Use

- Digitizing scanned archives
- Extracting text before search indexing or LLM prompting
- Any document type without a specialized parser

## What This Creates

- An Enterprise OCR processor in the `us` multi-region, displayed as "Document OCR", on Google's default version

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us` | `eu` when documents must stay in Europe. |
| `type` | `OCR_PROCESSOR` | `LAYOUT_PARSER_PROCESSOR` for RAG chunking, `FORM_PARSER_PROCESSOR` for forms. |
| `defaultVersion` | Google's default | Pin a concrete version for reproducible results. |
