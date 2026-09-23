# GCP Vertex AI Dataset

Registers a Vertex AI managed dataset -- the container Vertex AI training, AutoML, data labeling, and evaluation read their examples from. Pick the location and the data type (image, text, tabular, video, or time series), and your team imports the examples into it through the console, the SDK, or a pipeline. Every job that trains on those examples then names this one dataset instead of a pile of files, so the training data has an owner, labels, and encryption of its own.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Dataset** -- a `vertex.AiDataset` with the chosen data type, display name, labels, and optional CMEK

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpKmsKey** -- for customer-managed encryption, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Dataset**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Image Dataset** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiDataset
metadata:
  name: product-images
  org: acme-corp
  env: prod
spec:
  location: us-central1
  displayName: Product images
  metadataSchemaUri: gs://google-cloud-aiplatform/schema/dataset/metadata/image_1.0.0.yaml
```

```shell
planton apply -f vertex-ai-dataset.yaml
```

This registers an empty image dataset in `us-central1`, ready for imports and labeling. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference the dataset's `name` output from the pipeline or training configuration that reads it, and a `GcpKmsKey` from `kmsKeyName` when the examples must be encrypted under your key.

## Key Configuration

These are the most important decisions when configuring a dataset. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The data type** -- `metadataSchemaUri` names one of Google's five schemas and decides which training objectives and labeling tasks the dataset supports. It cannot change: a different type is a new dataset.

**Where it lives** -- the dataset, the training jobs that read it, and the models they produce live in one region. Pick the region your training runs in.

**Encryption** -- `kmsKeyName` encrypts the dataset and everything imported into it under your key. It is fixed at creation.

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
| `name` | The dataset's full resource name | Training jobs, pipelines, labeling tasks |
| `dataset_id` | The numeric id Google assigned | SDK calls and console links |
| `location` | The dataset's region | Regional clients |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Image dataset** -- an image dataset for classification or object detection, labelled by team. Start from the **Image Dataset** preset.

**Encrypted tabular dataset** -- a tabular dataset under a customer-managed key with `PREVENT`, for regulated training data. Start from the **Encrypted Tabular** preset.

## Works With

- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- where image, text, and video items are imported from
- [**GCP BigQuery Table**](/cloud-catalog/gcp-big-query-table) -- where tabular rows are imported from
- [**GCP Vertex AI TensorBoard**](/cloud-catalog/gcp-vertex-ai-tensorboard) -- where the training runs over the dataset stream their metrics
