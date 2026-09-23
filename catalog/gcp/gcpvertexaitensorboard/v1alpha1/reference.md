# GcpVertexAiTensorboard

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiTensorboardSpec defines a Vertex AI TensorBoard instance
(`google_vertex_ai_tensorboard`) -- the managed, regional TensorBoard
that Vertex AI training jobs, pipelines, and Vertex AI Experiments stream
training metrics into, shared by a team and kept after the jobs end --
together with the experiments (`google_vertex_ai_tensorboard_experiment`)
and runs (`google_vertex_ai_tensorboard_run`) declared in it. Experiments
and runs are folded in because they live and die with their TensorBoard
and nothing else in the catalog refers to one.

A training job attaches by naming the TensorBoard's full resource name
(the `name` output) in its `tensorboard` field, together with a service
account and a Cloud Storage staging bucket. Google assigns the
TensorBoard a numeric id at creation.

Immutable: location and the encryption key (a change replaces the
TensorBoard and every experiment, run, and logged series in it). The
display name, description, and labels update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiTensorboard
metadata:
  name: training-metrics
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Training metrics
  description: Shared TensorBoard the team's training jobs stream into
  labels:
    team: ml
  # Experiments and runs are optional: training jobs and the Vertex AI SDK
  # create their own. Declare the ones that must exist up front.
  experiments:
    - experimentId: churn-model
      displayName: Churn model
      source: custom training job
      runs:
        - runId: baseline
          description: The run every candidate is compared against
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.experiments` | `[]GcpVertexAiTensorboardExperiment` |  |  |  |
| `spec.experiments[].experimentId` | `string` | yes |  |  |
| `spec.experiments[].displayName` | `string` |  |  |  |
| `spec.experiments[].description` | `string` |  |  |  |
| `spec.experiments[].labels` | `map<string, string>` |  |  |  |
| `spec.experiments[].source` | `string` |  |  |  |
| `spec.experiments[].runs` | `[]GcpVertexAiTensorboardRun` |  |  |  |
| `spec.experiments[].runs[].runId` | `string` | yes |  |  |
| `spec.experiments[].runs[].displayName` | `string` |  |  |  |
| `spec.experiments[].runs[].description` | `string` |  |  |  |
| `spec.experiments[].runs[].labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the TensorBoard lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) the TensorBoard lives in, e.g.
"us-central1". Training jobs that stream into it must run in the same
location. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.displayName

`string`

Human-readable name shown in the console. Defaults to metadata.name.
Mutable in place.

### spec.description

`string`

Free-text description of the TensorBoard.

### spec.labels

`map<string, string>`

Labels on the TensorBoard. The platform attribution labels are merged
in and win on key conflicts.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the TensorBoard and every
series logged to it: a GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in the same region. The Vertex AI Service Agent needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. Omit to use
Google-managed encryption. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.experiments

`[]GcpVertexAiTensorboardExperiment`

Experiments declared in the TensorBoard, each keyed by experiment_id.
Leave empty to let training jobs and the Vertex AI SDK create the
experiments they log to.

- rule: run_id must be unique within an experiment
- rule: run display names (run_id when display_name is empty) must be unique within an experiment -- Google rejects duplicates

### spec.experiments[].experimentId

`string` · required

The experiment's id -- the last segment of its resource name and the
name a training job or the Vertex AI SDK
(aiplatform.init(experiment=...)) logs under. 1-128 characters:
lowercase letters, digits, and hyphens. Unique within the TensorBoard.
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9][a-z0-9-]{0,127}$"}}

### spec.experiments[].displayName

`string`

Human-readable name shown in the TensorBoard UI.

### spec.experiments[].description

`string`

Free-text description of the experiment.

### spec.experiments[].labels

`map<string, string>`

Labels on the experiment. The platform attribution labels are merged
in and win on key conflicts.

### spec.experiments[].source

`string`

Where the experiment's data comes from, recorded as metadata -- e.g.
"custom training job" or a pipeline name. Informational; Google does
not act on it. Immutable.

### spec.experiments[].runs

`[]GcpVertexAiTensorboardRun`

Runs declared up front, each keyed by run_id. Leave empty to let
training jobs create the runs they log to.

### spec.experiments[].runs[].runId

`string` · required

The run's id -- the last segment of its resource name. 1-128
characters: lowercase letters, digits, and hyphens. Unique within its
experiment. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9][a-z0-9-]{0,127}$"}}

### spec.experiments[].runs[].displayName

`string`

Human-readable name -- Google requires one and requires it unique
among the experiment's runs. Defaults to run_id. Mutable in place.

### spec.experiments[].runs[].description

`string`

Free-text description of the run.

### spec.experiments[].runs[].labels

`map<string, string>`

Labels on the run. The platform attribution labels are merged in and
win on key conflicts.

### spec.deletionPolicy

`string`

What happens to the TensorBoard, its experiments, and its runs when
this resource is destroyed:
  "" / "DELETE" -- everything is deleted, logged series included
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.unique_experiment_ids`: experiment_id must be unique within the TensorBoard

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiTensorboard, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- what a training job's `tensorboard` field takes: projects/{project_number}/locations/{location}/tensorboards/{tensorboard_id}. |
| `status.outputs.tensorboard_id` | `string` | The numeric id Google assigned at creation (the last segment of name). |
| `status.outputs.location` | `string` | The location the TensorBoard lives in. |
| `status.outputs.blob_storage_path_prefix` | `string` | The Cloud Storage path prefix (bucket or directory, no trailing slash) in Google's tenant project where the TensorBoard stores blob data. |
| `status.outputs.experiment_names` | `[]string` | Full resource names of the declared experiments, in manifest order. |
| `status.outputs.run_names` | `[]string` | Full resource names of the declared runs, experiment by experiment in manifest order. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
