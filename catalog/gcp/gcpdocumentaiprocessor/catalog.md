# GCP Document AI Processor

Turns documents into data. A Document AI processor reads PDFs, scans, and photos and returns structured results: the text and layout of any page, the fields and tables of a form, or the specific fields of an invoice, receipt, ID, or pay slip. Your applications send documents to the processor's endpoint instead of building and hosting their own extraction models.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `documentai.googleapis.com` on the project
- **Processor** -- a Document AI processor of the chosen type, location, and encryption
- **Default version** (optional) -- which processor version serves requests by default

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Document AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpKmsKey** -- for customer-managed encryption, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Document AI Processor**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Enterprise OCR** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDocumentAiProcessor
metadata:
  name: invoice-parser
  org: acme-corp
  env: prod
spec:
  location: us
  type: INVOICE_PROCESSOR
  displayName: Invoice parser
```

```shell
planton apply -f document-ai-processor.yaml
```

This creates an invoice parser in the `us` multi-region; its `process_endpoint` output is where your application posts invoices. A Stack Job tracks the provisioning in real time.

### InfraChart

Hand the processor's `process_endpoint` or `name` output to the service or pipeline that parses documents, and reference a `GcpKmsKey` from `kmsKeyName` when documents must be encrypted under your key.

## Key Configuration

These are the most important decisions when configuring a processor. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**What it extracts** -- `type` picks the model: OCR for any page, Form Parser for forms and tables, Layout Parser for chunking documents for RAG, or a specialized parser for invoices, receipts, IDs, and more. It cannot change later.

**Where documents are processed** -- `us` or `eu`; your documents never leave the location.

**Which version answers** -- `defaultVersion` pins the model version requests use unless they name one. Pinning keeps results stable when Google ships a new version.

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
| `name` | The processor's full resource name | Client libraries, batch jobs |
| `processor_id` | The id Google assigned | Console links |
| `location` | The processor's location | Regional clients |
| `process_endpoint` | Where documents are posted | Application configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Enterprise OCR** -- text, layout, and handwriting from any document, the general-purpose starting point. Start from the **Enterprise OCR** preset.

**Encrypted invoice parser** -- invoices parsed under a customer-managed key with a pinned version and destroy protection. Start from the **Encrypted Invoice Parser** preset.

## Works With

- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- batch input and output documents
- [**GCP Vertex AI Search Data Store**](/cloud-catalog/gcp-vertex-ai-search-data-store) -- search and RAG over parsed documents
