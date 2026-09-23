# GCP Vertex AI Dataset

A Vertex AI managed dataset -- the registered container that Vertex AI training, AutoML, data labeling, and evaluation read their examples from. Declare its location and data type (image, text, tabular, video, or time series, fixed by Google's metadata schema), and import data items into it through the Vertex AI API, the console, or the SDK. Google assigns the dataset a numeric id; the outputs carry it for the jobs and pipelines that consume the dataset.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Dataset** -- a `vertex_ai_dataset` with the chosen metadata schema, display name, labels, and optional CMEK

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpKmsKey`** -- a key in the same region for customer-managed encryption (`kmsKeyName`). The Vertex AI Service Agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it. Immutable once set.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiDataset
metadata:
  name: product-images
spec:
  location: us-central1
  metadataSchemaUri: gs://google-cloud-aiplatform/schema/dataset/metadata/image_1.0.0.yaml
```

```shell
planton apply -f vertex-ai-dataset.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI region. Immutable. |
| `metadataSchemaUri` | `string` | One of Google's schemas under `gs://google-cloud-aiplatform/schema/dataset/metadata/` (`image_1.0.0.yaml`, `text_1.0.0.yaml`, `tabular_1.0.0.yaml`, `video_1.0.0.yaml`, `time_series_1.0.0.yaml`). Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `displayName` | `string` | `metadata.name` | Up to 128 characters; mutable. |
| `labels` | `map<string,string>` | none | Merged under the platform attribution labels; mutable. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Immutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `location` is a region (`us-central1`, never `global` or a multi-region).
- `metadataSchemaUri` is a `gs://` path to a YAML file.
- `displayName` is at most 128 characters.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project_number}/locations/{location}/datasets/{dataset_id}` |
| `dataset_id` | `string` | The numeric id Google assigned |
| `location` | `string` | The dataset's location |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The data type is permanent.** Changing `metadataSchemaUri`, `location`, or the key replaces the dataset -- and everything imported into it.
- **Data items are not part of this block.** Images, text, rows, and annotations are imported through the Vertex AI API, the console, or the SDK; the dataset only declares where they live and what type they are.
- **Destroy deletes the data.** Under `DELETE` the imported items and annotations go with the dataset; use `PREVENT` for a labeled dataset you cannot recreate.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpKmsKey** -- customer-managed encryption for the dataset
- **GcpGcsBucket** -- the bucket image, text, and video items are imported from
- **GcpBigQueryTable** -- the table tabular datasets import rows from
- **GcpVertexAiTensorboard** -- where the training jobs that read the dataset stream their metrics

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
