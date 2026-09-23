# GCP Colab Schedule

A Vertex AI schedule -- a cron that launches a run on a timer, either a **Colab Enterprise notebook run** (a notebook from Cloud Storage or a Dataform repository, executed on a runtime template's machine or a custom machine, its executed copy written to Cloud Storage) or a **Vertex AI Pipelines run** (a compiled Kubeflow pipeline with its parameters). Nightly reports, scheduled retraining, and data-quality checks become one declared block that can be paused, bounded by a time window and a run count, and reviewed like code.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Schedule** -- a `colab_schedule` launching the declared notebook or pipeline run, paused or active per `desiredState`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project, and `iam.serviceAccounts.actAs` on the identity the runs execute as.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpColabRuntimeTemplate`** -- the machine notebook runs use (`notebookExecutionJob.notebookRuntimeTemplateResourceName`).
- **`GcpGcsBucket`** -- where executed notebooks land (`notebookExecutionJob.gcsOutputUri`).
- **`GcpServiceAccount`** -- the identity runs execute as.
- **`GcpVpcNetwork`**, **`GcpKmsKey`** -- private networking and encryption for runs.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabSchedule
metadata:
  name: nightly-report
spec:
  location: us-central1
  cron: "0 6 * * *"
  maxConcurrentRunCount: 1
  notebookExecutionJob:
    gcsNotebookSource:
      uri: gs://acme-notebooks/nightly.ipynb
    notebookRuntimeTemplateResourceName:
      value: projects/acme/locations/us-central1/notebookRuntimeTemplates/standard-cpu
    gcsOutputUri:
      value: gs://acme-notebook-runs
    serviceAccount:
      value: notebooks@acme.iam.gserviceaccount.com
```

```shell
planton apply -f colab-schedule.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI region. |
| `cron` | `string` | Cron syntax, optionally prefixed `TZ=Area/City`. |
| `maxConcurrentRunCount` | `int64` | How many runs may start at the same time (at least 1). |
| exactly one run | `object` | `notebookExecutionJob` or `pipelineJob`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `displayName` | `string` | `metadata.name` | Up to 128 characters. |
| `allowQueueing` | `bool` | `false` | Queue a run over the concurrency limit instead of skipping it. |
| `startTime` / `endTime` | `string` | now / none | RFC 3339 window. |
| `maxRunCount` | `int64` | unlimited | Complete after this many runs. |
| `maxConcurrentActiveRunCount` | `int64` | none | Pipeline schedules only: runs in flight at once. |
| `desiredState` | `string` | `ACTIVE` | `ACTIVE` or `PAUSED`; enforced on every apply. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

**`notebookExecutionJob`** (a change replaces the schedule): exactly one source (`gcsNotebookSource` or `dataformRepositorySource`), exactly one machine (`notebookRuntimeTemplateResourceName` or `customEnvironmentSpec`), exactly one identity (`executionUser` or `serviceAccount`), `gcsOutputUri` (required), plus `displayName`, `executionTimeout`, `kernelName`, `labels`, `kmsKeyName`, `workbenchRuntime`.

**`pipelineJob`** (updates in place): `pipelineSpec` (compiled JSON) or `templateUri`, `runtimeConfig` (`gcsOutputDirectory`, `failurePolicy`, `parameterValues`), `serviceAccount`, `network` (peering; the modules resolve the project number Google requires), `reservedIpRanges`, `pscInterfaceConfig`, `preflightValidations`, `labels`, `kmsKeyName`, `displayName`.

### Validation Rules

- Exactly one of `notebookExecutionJob` and `pipelineJob`, and the notebook arm's three exactly-one pairs.
- Timestamps are RFC 3339; durations are seconds ending in `s`; notebook sources are `gs://` paths.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/schedules/{schedule_id}` |
| `schedule_id` | `string` | The id Google assigned |
| `location` | `string` | The schedule's region |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Unattended runs should use a service account.** `executionUser` runs as a person and stops working when they leave.
- **The notebook run is immutable.** Changing it replaces the schedule (history restarts); pipeline runs update in place.
- **`desiredState` is enforced on every apply.** A schedule paused in the console resumes on the next apply if the manifest says `ACTIVE`.
- **Destroy stops future runs only.** Runs already launched and their outputs stay; `ABANDON` keeps the schedule launching runs outside management.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpColabRuntimeTemplate** -- the machine notebook runs use
- **GcpGcsBucket** -- notebook sources and executed outputs
- **GcpServiceAccount** -- the identity runs execute as
- **GcpColabRuntime** -- an interactive runtime instead of scheduled runs

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
