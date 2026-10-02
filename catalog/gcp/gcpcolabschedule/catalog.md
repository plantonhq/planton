# GCP Colab Schedule

Runs a notebook or a pipeline on a timer. A schedule launches a Colab Enterprise notebook -- executing it top to bottom on a runtime template's machine and saving the executed copy with its outputs -- or a Vertex AI Pipelines run, on a cron. Nightly reports, weekly retraining, and hourly data checks become one declared block your team can pause, bound, and review, instead of a cron job on someone's laptop.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Schedule** -- a `colab.Schedule` that launches the declared notebook or pipeline run on its cron

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project, and permission to act as the runs' service account. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **GcpColabRuntimeTemplate**, **GcpGcsBucket**, **GcpServiceAccount** -- the machine, the output bucket, and the identity for notebook runs.

## Deploy

### Console

Open the deployment store, find **GCP Colab Schedule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Nightly Notebook Report** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabSchedule
metadata:
  name: nightly-report
  org: acme-corp
  env: prod
spec:
  location: us-central1
  cron: "TZ=America/New_York 0 6 * * *"
  maxConcurrentRunCount: 1
  notebookExecutionJob:
    gcsNotebookSource:
      uri: gs://acme-notebooks/nightly.ipynb
    notebookRuntimeTemplateResourceName:
      valueFrom:
        kind: GcpColabRuntimeTemplate
        name: standard-cpu
        fieldPath: status.outputs.name
    gcsOutputUri:
      valueFrom:
        kind: GcpGcsBucket
        name: notebook-runs
        fieldPath: status.outputs.url
    serviceAccount:
      valueFrom:
        kind: GcpServiceAccount
        name: notebook-runner
        fieldPath: status.outputs.email
```

```shell
planton apply -f colab-schedule.yaml
```

This runs the nightly notebook at 06:00 New York time as a service account and saves each executed copy to the bucket. A Stack Job tracks the provisioning in real time.

### InfraChart

Wire the schedule to a `GcpColabRuntimeTemplate`, a `GcpGcsBucket`, and a `GcpServiceAccount` in the same chart; the schedule references their outputs.

## Key Configuration

These are the most important decisions when configuring a schedule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**What runs** -- a notebook (`notebookExecutionJob`) or a pipeline (`pipelineJob`), never both.

**When it runs** -- `cron` (with an optional `TZ=` prefix), bounded by `startTime`, `endTime`, and `maxRunCount`.

**How many at once** -- `maxConcurrentRunCount` and `allowQueueing` decide whether a slow run makes the next one skip or wait.

**As whom** -- runs execute as a service account (for production) or a user.

**Paused or active** -- `desiredState` pauses and resumes the schedule on every apply.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpColabRuntimeTemplate** | `notebookExecutionJob.notebookRuntimeTemplateResourceName` | `status.outputs.name` |
| **GcpGcsBucket** | `notebookExecutionJob.gcsOutputUri` | `status.outputs.url` |
| **GcpServiceAccount** | `notebookExecutionJob.serviceAccount`, `pipelineJob.serviceAccount` | `status.outputs.email` |
| **GcpVpcNetwork** | `notebookExecutionJob.customEnvironmentSpec.networkSpec.network` | `status.outputs.network_id` |
| **GcpVpcNetwork** | `pipelineJob.network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** | `notebookExecutionJob.customEnvironmentSpec.networkSpec.subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpKmsKey** | `notebookExecutionJob.kmsKeyName`, `pipelineJob.kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The schedule's full resource name | Console links, automation |
| `schedule_id` | The id Google assigned | SDK calls |
| `location` | The schedule's region | Regional clients |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Nightly notebook report** -- a notebook from Cloud Storage run nightly on a template's machine as a service account. Start from the **Nightly Notebook Report** preset.

**Weekly pipeline retraining** -- a compiled pipeline from Artifact Registry run weekly with parameters, fail-fast, on a peered network. Start from the **Weekly Pipeline Retraining** preset.

## Works With

- [**GCP Colab Runtime Template**](/cloud-catalog/gcp-colab-runtime-template) -- the machine notebook runs use
- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- notebook sources and outputs
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the run identity
