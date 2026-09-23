# GcpDocumentAiProcessor Guide

The judgment this guide protects: a processor's type, location, and key are permanent, and its default version decides what every request gets. Choose the type for the documents you actually have, pin a concrete version, and treat a processor with custom-trained versions like a model you cannot rebuild cheaply.

## Choosing a type

`type` is one of the processor types Google lists for your project and location (`fetchProcessorTypes`):

- `OCR_PROCESSOR` (Enterprise OCR) -- text, layout, handwriting, and language detection from any document.
- `FORM_PARSER_PROCESSOR` -- key-value pairs, tables, checkboxes.
- `LAYOUT_PARSER_PROCESSOR` -- document structure and chunks, the input a RAG corpus wants.
- Specialized parsers -- `INVOICE_PROCESSOR`, `EXPENSE_PROCESSOR`, `ID_PROOFING_PROCESSOR`, pay slips, bank statements, and others; some need an allowlist or a specific location.
- `CUSTOM_EXTRACTION_PROCESSOR` -- a generative extractor you define and train in the console.

The type is fixed at creation. A different type is a new processor with a new id and endpoint.

## Location and data residency

Document AI's main locations are the `us` and `eu` multi-regions; documents are processed and stored only there. Pick the one your data-residency rules require. Location is permanent.

## The default version

Every processor has versions: Google's pretrained releases and, for custom extractors, the versions you train. Requests that do not name a version use the default. `defaultVersion` takes the short version id (the block composes the full path under the processor):

- **Pin a concrete version** such as `pretrained-ocr-v2.1-2024-08-07`. Results stay reproducible when Google ships a new release, and you upgrade deliberately by changing the id.
- **Never `stable` or `rc`.** Those channel aliases are answered with whatever version they currently resolve to, so a declared alias never matches what Google stores and every deploy would set the default again. The block refuses them.
- Changing the id re-points the default in place. Destroying the block leaves the last default set -- Google has no "unset".

## Encryption and destroy

`kmsKeyName` encrypts what the processor stores under your key; the Document AI service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it. It is fixed at creation. Under `DELETE`, destroy removes the processor and every version you trained; `PREVENT` makes destroy fail -- use it for custom extractors.

## Calling it

Post documents to the `process_endpoint` output (online processing, a few pages per request) or run batch processing from Cloud Storage with the `name` output. Pages processed are what bills.
