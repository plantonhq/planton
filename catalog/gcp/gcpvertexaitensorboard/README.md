# GCP Vertex AI TensorBoard

A managed Vertex AI TensorBoard -- the regional, shared TensorBoard that Vertex AI training jobs, pipelines, and Vertex AI Experiments stream their training metrics into and that the team keeps after the jobs end -- together with the experiments and runs declared in it up front. A training job attaches by naming the TensorBoard's full resource name; Google assigns the TensorBoard a numeric id at creation.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **TensorBoard** -- a `vertex_ai_tensorboard` with its display name, description, labels, and optional CMEK
- **Experiments** -- one `vertex_ai_tensorboard_experiment` per `experiments[]` entry, keyed by `experimentId`
- **Runs** -- one `vertex_ai_tensorboard_run` per `experiments[].runs[]` entry, keyed by `runId` within its experiment

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
kind: GcpVertexAiTensorboard
metadata:
  name: training-metrics
spec:
  location: us-central1
  experiments:
    - experimentId: churn-model
      runs:
        - runId: baseline
```

```shell
planton apply -f vertex-ai-tensorboard.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI region. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `displayName` | `string` | `metadata.name` | Mutable. |
| `description`, `labels` | | | Descriptive metadata; mutable. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Immutable. |
| `experiments[]` | `[]object` | none | `experimentId` (lowercase letters, digits, hyphens; immutable), `displayName`, `description`, `labels`, `source` (immutable), `runs[]` (`runId`, `displayName` defaulting to `runId`, `description`, `labels`). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`; fanned to every experiment and run. |

### Validation Rules

- `location` is a region.
- Experiment ids are unique within the TensorBoard; run ids and run display names are unique within their experiment (Google requires unique run display names).
- Experiment and run ids are 1-128 lowercase letters, digits, and hyphens.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project_number}/locations/{location}/tensorboards/{tensorboard_id}` -- what a training job's `tensorboard` field takes |
| `tensorboard_id` | `string` | The numeric id Google assigned |
| `location` | `string` | The TensorBoard's location |
| `blob_storage_path_prefix` | `string` | Where Google stores the TensorBoard's blob data |
| `experiment_names` | `[]string` | Full resource names of the declared experiments, in manifest order |
| `run_names` | `[]string` | Full resource names of the declared runs, in manifest order |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Declare only what must exist up front.** Training jobs and the Vertex AI SDK create their own experiments and runs as they log; those never conflict with the declared ones, and the block never deletes them -- until the TensorBoard itself is destroyed.
- **Destroy deletes every logged series.** Under `DELETE` the TensorBoard goes with every experiment and run in it, including the ones jobs created. Use `PREVENT` for a TensorBoard holding history the team compares against.
- **Location and key are permanent.** Changing either replaces the TensorBoard.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpKmsKey** -- customer-managed encryption
- **GcpVertexAiPersistentResource** -- a warm cluster for the training jobs that stream into the TensorBoard
- **GcpVertexAiDataset** -- the data those training jobs read
- **GcpGcsBucket** -- the staging bucket training jobs use beside the TensorBoard

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
