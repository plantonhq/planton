# GCP Document AI Processor

A Document AI processor -- a managed model that turns documents (PDFs, scans, photos) into structured data. The processor's type decides what it extracts: text, layout, and handwriting (Enterprise OCR), form fields and tables (Form Parser), document structure for RAG (Layout Parser), or the fields of a specialized document such as an invoice, receipt, pay slip, or ID. Applications post documents to the processor's process endpoint and get a structured Document back. The block can also pin which processor version serves requests by default.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `documentai.googleapis.com` on the project (never disabled on destroy)
- **Processor** -- a `document_ai_processor` of the chosen type, location, and encryption
- **Default version** (when `defaultVersion` is set) -- a `document_ai_processor_default_version` binding

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Document AI admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpKmsKey`** -- customer-managed encryption (`kmsKeyName`). The Document AI service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDocumentAiProcessor
metadata:
  name: invoice-ocr
spec:
  location: us
  type: OCR_PROCESSOR
```

```shell
planton apply -f document-ai-processor.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | `us` or `eu` (Document AI's main locations), or a region Google lists for Document AI. Immutable. |
| `type` | `string` | The processor type, e.g. `OCR_PROCESSOR`, `FORM_PARSER_PROCESSOR`, `LAYOUT_PARSER_PROCESSOR`, `INVOICE_PROCESSOR`, `CUSTOM_EXTRACTION_PROCESSOR`. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `displayName` | `string` | `metadata.name` | Unique among the project's processors in the location. Immutable. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Immutable. |
| `defaultVersion` | `string` | Google's default | The short id of the version that serves requests naming none, e.g. `pretrained-ocr-v2.1-2024-08-07`. Never `stable` or `rc`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `type` is an upper-case identifier.
- `defaultVersion` is a concrete version id: not `stable`, not `rc`, not a full path.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/processors/{processor_id}` |
| `processor_id` | `string` | The id Google assigned |
| `location` | `string` | The processor's location |
| `process_endpoint` | `string` | `https://{location}-documentai.googleapis.com/v1/{name}:process` -- where documents are posted |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything but the default version is permanent.** Changing the type, location, display name, or key replaces the processor -- and its id and endpoint change with it.
- **Pin a concrete default version.** The `stable` and `rc` channel aliases never settle (Google answers with the version they resolve to), so the block refuses them; a pinned version also makes extraction results reproducible.
- **Destroy keeps nothing.** Under `DELETE` the processor and any version you trained are deleted; `PREVENT` guards a processor with custom-trained versions.
- **Some types need an allowlist or a specific location.** Check `fetchProcessorTypes` for your project before choosing one.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpKmsKey** -- customer-managed encryption for the processor
- **GcpGcsBucket** -- where batch-processing input and output documents live
- **GcpVertexAiSearchDataStore** -- a RAG corpus built from parsed documents

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
