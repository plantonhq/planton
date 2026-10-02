# GCP Vertex AI TensorBoard

Gives a team one managed TensorBoard that every Vertex AI training job, pipeline, and experiment streams its metrics into -- loss curves, accuracy, histograms, images -- and keeps them after the jobs finish, so a new run is compared against last month's without anyone keeping a laptop open. Declare the TensorBoard, and optionally the experiments and runs that should exist before the first job logs, and point training jobs at it by name.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **TensorBoard** -- a `vertex.AiTensorboard` with optional CMEK
- **Experiments and runs** -- one `vertex.AiTensorboardExperiment` per declared experiment and one `vertex.AiTensorboardRun` per declared run

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpKmsKey** -- for customer-managed encryption, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI TensorBoard**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team TensorBoard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiTensorboard
metadata:
  name: training-metrics
  org: acme-corp
  env: prod
spec:
  location: us-central1
  displayName: Training metrics
  experiments:
    - experimentId: churn-model
      runs:
        - runId: baseline
```

```shell
planton apply -f vertex-ai-tensorboard.yaml
```

This creates a TensorBoard in `us-central1` with one experiment and a baseline run ready for training jobs to log into. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, pass the TensorBoard's `name` output to the training job or pipeline that streams into it, and reference a `GcpKmsKey` from `kmsKeyName` when logged data must be encrypted under your key.

## Key Configuration

These are the most important decisions when configuring a TensorBoard. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One TensorBoard per team, not per job** -- training jobs attach by name and add their own experiments and runs. A shared TensorBoard is what makes runs comparable over time.

**Declare experiments only when they must exist first** -- the Vertex AI SDK and training jobs create experiments and runs as they log. Declare the ones a dashboard, an evaluation pipeline, or a baseline comparison points at by id.

**Encryption and history** -- `kmsKeyName` is fixed at creation. A destroy under `DELETE` removes every logged series, including the ones jobs created; `PREVENT` protects history the team relies on.

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
| `name` | The TensorBoard's full resource name | A training job's `tensorboard` field |
| `tensorboard_id` | The numeric id Google assigned | SDK calls and console links |
| `location` | The TensorBoard's region | Regional clients |
| `blob_storage_path_prefix` | Where Google stores blob data | Diagnostics |
| `experiment_names` | The declared experiments' full names | Dashboards, evaluation pipelines |
| `run_names` | The declared runs' full names | Baseline comparisons |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Team TensorBoard** -- one shared TensorBoard with no declared experiments; jobs create their own. Start from the **Team TensorBoard** preset.

**Encrypted with experiments** -- a CMEK TensorBoard with the team's standing experiments and baseline runs declared, protected by `PREVENT`. Start from the **Encrypted With Experiments** preset.

## Works With

- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
- [**GCP Vertex AI Persistent Resource**](/cloud-catalog/gcp-vertex-ai-persistent-resource) -- warm training capacity for the jobs that log here
- [**GCP Vertex AI Dataset**](/cloud-catalog/gcp-vertex-ai-dataset) -- the training data those jobs read
- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- the staging bucket training jobs use beside the TensorBoard
