# GcpDocumentAiProcessor

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDocumentAiProcessorSpec defines a Document AI processor
(`google_document_ai_processor`) -- a managed model that turns documents
(PDFs, scans, photos) into structured data: text and layout (OCR), form
fields, or the fields of a specialized document such as an invoice,
receipt, pay slip, or ID. Applications send documents to the processor's
process endpoint (an output of this block) and get a structured Document
back.

The processor's type decides what it extracts and is fixed at creation.
Optionally the block also sets which processor version serves requests
by default (`google_document_ai_processor_default_version`, folded in).

Immutable: every field except default_version and deletion_policy -- a
change replaces the processor, and its id and endpoint change with it.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDocumentAiProcessor
metadata:
  name: invoice-ocr
spec:
  projectId:
    value: my-gcp-project
  # Document AI's main locations are the multi-regions us and eu.
  location: us
  type: OCR_PROCESSOR
  displayName: Invoice OCR
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.defaultVersion` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the processor lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Where documents are processed and stored: the multi-region "us" or
"eu" (Document AI's main locations), or a region Google lists for
Document AI. Your documents never leave it. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+(-[a-z]+[0-9]+)?$"}}

### spec.type

`string` · required

What the processor extracts -- one of the types Google lists through
fetchProcessorTypes for your project and location, for example:
  OCR_PROCESSOR          text, layout, and handwriting (Enterprise OCR)
  FORM_PARSER_PROCESSOR  key-value pairs, tables, and checkboxes
  LAYOUT_PARSER_PROCESSOR document structure and chunks for RAG
  INVOICE_PROCESSOR, EXPENSE_PROCESSOR, ID_PROOFING_PROCESSOR, and
  other specialized parsers; CUSTOM_EXTRACTION_PROCESSOR for a
  generative extractor you train in the console
Some types need an allowlist or a specific location. Immutable.

- rule: {"required":true,"string":{"pattern":"^[A-Z][A-Z0-9_]*$"}}

### spec.displayName

`string`

The processor's display name -- unique among the project's processors
in the location. Defaults to metadata.name. Immutable.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key for the documents and results the
processor stores: a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in a location compatible with the processor's. The Document AI service
agent needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it. Omit for
Google-managed encryption. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.defaultVersion

`string`

The processor version that serves requests which do not name one --
its short id, e.g. "pretrained-ocr-v2.1-2024-08-07" for a Google
version or the id of a version you trained. The modules compose the
full path under this processor. Omit to keep Google's default.

Use a concrete version, never the "stable" or "rc" channel alias:
Google answers with the version the alias currently resolves to, so a
declared alias would never match what Google stores and every deploy
would set the default again. Pinning a version is also what makes
extraction results reproducible. Changing it re-points the default in
place; destroying the block leaves the last default set (Google has no
"unset").

- rule: default_version must be a concrete version id (not stable or rc, and not a full path)

### spec.deletionPolicy

`string`

What happens to the processor when this resource is destroyed:
  "" / "DELETE" -- the processor, its versions, and any version you
                   trained are deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the processor leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDocumentAiProcessor, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/processors/{processor_id}. |
| `status.outputs.processor_id` | `string` | The id Google assigned the processor (the last segment of name). |
| `status.outputs.location` | `string` | The location the processor lives in. |
| `status.outputs.process_endpoint` | `string` | The REST endpoint documents are posted to: https://{location}-documentai.googleapis.com/v1/{name}:process. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
