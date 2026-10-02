# GcpColabSchedule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpColabScheduleSpec defines a Vertex AI schedule
(`google_colab_schedule`) -- a cron that launches a run on a timer. It
launches one of two things, set by exactly one arm:

  notebook_execution_job  a Colab Enterprise notebook run: a notebook
                          from Cloud Storage or a Dataform repository,
                          executed on a runtime template's machine (or a
                          custom machine), with the executed notebook
                          written to Cloud Storage -- nightly reports,
                          scheduled retraining, data-quality checks
  pipeline_job            a Vertex AI Pipelines run: a compiled Kubeflow
                          pipeline (inline or from a template registry)
                          with its runtime parameters

Both are Google's Schedules API; the provider files it under Colab.

A schedule can be paused and resumed (desired_state) and bounded by a
time window and a run count. Mutable in place: everything except
location, the project, and the notebook arm (a change to the notebook
run REPLACES the schedule; the pipeline arm updates in place).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabSchedule
metadata:
  name: nightly-report
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Nightly report
  cron: "TZ=America/New_York 0 6 * * *"
  maxConcurrentRunCount: 1
  notebookExecutionJob:
    gcsNotebookSource:
      uri: gs://acme-notebooks/reports/nightly.ipynb
    notebookRuntimeTemplateResourceName:
      valueFrom:
        kind: GcpColabRuntimeTemplate
        name: standard-runtime
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
    executionTimeout: 3600s
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.cron` | `string` | yes |  |  |
| `spec.maxConcurrentRunCount` | `int64` |  |  |  |
| `spec.allowQueueing` | `bool` |  |  |  |
| `spec.startTime` | `string` |  |  |  |
| `spec.endTime` | `string` |  |  |  |
| `spec.maxRunCount` | `int64` |  |  |  |
| `spec.maxConcurrentActiveRunCount` | `int64` |  |  |  |
| `spec.desiredState` | `string` |  |  |  |
| `spec.notebookExecutionJob` | `GcpColabScheduleNotebookExecutionJob` |  |  |  |
| `spec.notebookExecutionJob.displayName` | `string` |  |  |  |
| `spec.notebookExecutionJob.gcsNotebookSource` | `GcpColabScheduleGcsNotebookSource` |  |  |  |
| `spec.notebookExecutionJob.gcsNotebookSource.uri` | `string` | yes |  |  |
| `spec.notebookExecutionJob.gcsNotebookSource.generation` | `string` |  |  |  |
| `spec.notebookExecutionJob.dataformRepositorySource` | `GcpColabScheduleDataformRepositorySource` |  |  |  |
| `spec.notebookExecutionJob.dataformRepositorySource.dataformRepositoryResourceName` | `string` | yes |  |  |
| `spec.notebookExecutionJob.dataformRepositorySource.commitSha` | `string` |  |  |  |
| `spec.notebookExecutionJob.notebookRuntimeTemplateResourceName` | `string \| valueFrom` |  |  | GcpColabRuntimeTemplate (`status.outputs.name`) |
| `spec.notebookExecutionJob.customEnvironmentSpec` | `GcpColabScheduleCustomEnvironmentSpec` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec` | `GcpColabScheduleMachineSpec` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.machineType` | `string` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.acceleratorType` | `string` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.acceleratorCount` | `int32` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.gpuPartitionSize` | `string` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.tpuTopology` | `string` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity` | `GcpColabScheduleReservationAffinity` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.reservationAffinityType` | `string` | yes |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.key` | `string` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.values` | `[]string` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.useReservationPool` | `bool` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.networkSpec` | `GcpColabScheduleNetworkSpec` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.enableInternetAccess` | `bool` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.notebookExecutionJob.customEnvironmentSpec.persistentDiskSpec` | `GcpColabSchedulePersistentDiskSpec` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.persistentDiskSpec.diskType` | `string` |  |  |  |
| `spec.notebookExecutionJob.customEnvironmentSpec.persistentDiskSpec.diskSizeGb` | `int64` |  |  |  |
| `spec.notebookExecutionJob.workbenchRuntime` | `bool` |  |  |  |
| `spec.notebookExecutionJob.gcsOutputUri` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.url`) |
| `spec.notebookExecutionJob.executionUser` | `string` |  |  |  |
| `spec.notebookExecutionJob.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.notebookExecutionJob.executionTimeout` | `string` |  |  |  |
| `spec.notebookExecutionJob.kernelName` | `string` |  |  |  |
| `spec.notebookExecutionJob.labels` | `map<string, string>` |  |  |  |
| `spec.notebookExecutionJob.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.pipelineJob` | `GcpColabSchedulePipelineJob` |  |  |  |
| `spec.pipelineJob.displayName` | `string` |  |  |  |
| `spec.pipelineJob.pipelineSpec` | `string` |  |  |  |
| `spec.pipelineJob.templateUri` | `string` |  |  |  |
| `spec.pipelineJob.runtimeConfig` | `GcpColabSchedulePipelineRuntimeConfig` |  |  |  |
| `spec.pipelineJob.runtimeConfig.gcsOutputDirectory` | `string` | yes |  |  |
| `spec.pipelineJob.runtimeConfig.failurePolicy` | `string` |  |  |  |
| `spec.pipelineJob.runtimeConfig.parameterValues` | `map<string, string>` |  |  |  |
| `spec.pipelineJob.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.pipelineJob.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.pipelineJob.reservedIpRanges` | `[]string` |  |  |  |
| `spec.pipelineJob.preflightValidations` | `bool` |  |  |  |
| `spec.pipelineJob.labels` | `map<string, string>` |  |  |  |
| `spec.pipelineJob.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.pipelineJob.pscInterfaceConfig` | `GcpColabSchedulePscInterfaceConfig` |  |  |  |
| `spec.pipelineJob.pscInterfaceConfig.networkAttachment` | `string` |  |  |  |
| `spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs` | `[]GcpColabScheduleDnsPeeringConfig` |  |  |  |
| `spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].domain` | `string` | yes |  |  |
| `spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the schedule lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI region the schedule and its runs live in, e.g.
"us-central1". Changing it needs a new schedule.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.displayName

`string`

The schedule's name in the console -- up to 128 characters. Defaults
to metadata.name.

- rule: {"string":{"maxLen":"128"}}

### spec.cron

`string` · required

When runs launch, in cron syntax, optionally prefixed with a time zone:
"0 6 * * *" (06:00 UTC daily) or "TZ=America/New_York 0 9 * * 1-5"
(09:00 New York time on weekdays).

- rule: {"required":true,"string":{"minLen":"9"}}

### spec.maxConcurrentRunCount

`int64`

How many runs may be STARTED at the same time. A run that would exceed
it is skipped (or queued, with allow_queueing). This limits launching,
not how long a run executes.

- rule: {"int64":{"gte":"1"}}

### spec.allowQueueing

`bool`

Queue a run that hits max_concurrent_run_count instead of skipping it.

### spec.startTime

`string`

The earliest a run may launch, RFC 3339 (e.g. "2026-10-01T00:00:00Z").
Unset: the schedule's creation time. Sent only when set.

- rule: start_time must be an RFC 3339 timestamp

### spec.endTime

`string`

After this time no new runs launch and the schedule completes, RFC
3339. Unset: no end.

- rule: end_time must be an RFC 3339 timestamp

### spec.maxRunCount

`int64`

Complete the schedule after this many runs have launched. Unset (0):
runs keep launching until the schedule is paused, ended, or deleted.

- rule: {"int64":{"gte":"0"}}

### spec.maxConcurrentActiveRunCount

`int64`

How many runs may be in a non-terminal state at once -- Google applies
it to pipeline schedules only. Unset (0): no limit beyond
max_concurrent_run_count.

- rule: {"int64":{"gte":"0"}}

### spec.desiredState

`string`

Whether the schedule launches runs:
  "ACTIVE" -- runs launch on the cron (the default)
  "PAUSED" -- nothing launches; resume by setting ACTIVE again
The block pauses or resumes the schedule to match on every apply.

- rule: desired_state must be ACTIVE or PAUSED

### spec.notebookExecutionJob

`GcpColabScheduleNotebookExecutionJob`

The Colab Enterprise notebook run each tick launches. Exactly one of
notebook_execution_job or pipeline_job. A change replaces the
schedule.

- rule: set exactly one of gcs_notebook_source or dataform_repository_source
- rule: set exactly one of notebook_runtime_template_resource_name or custom_environment_spec
- rule: set exactly one of execution_user or service_account

### spec.notebookExecutionJob.displayName

`string`

The run's name in the console -- up to 128 characters. Defaults to the
schedule's display name.

- rule: {"string":{"maxLen":"128"}}

### spec.notebookExecutionJob.gcsNotebookSource

`GcpColabScheduleGcsNotebookSource`

Run a notebook stored in Cloud Storage. Exactly one source.

### spec.notebookExecutionJob.gcsNotebookSource.uri

`string` · required

The notebook: gs://bucket/path/notebook.ipynb.

- rule: {"required":true,"string":{"pattern":"^gs://[^/]+/.+$"}}

### spec.notebookExecutionJob.gcsNotebookSource.generation

`string`

Pin an object generation (version) of the notebook. Unset: the current
version at each run.

- rule: generation must be a Cloud Storage object generation number

### spec.notebookExecutionJob.dataformRepositorySource

`GcpColabScheduleDataformRepositorySource`

Run a notebook from a Dataform repository (notebooks under version
control). Exactly one source.

### spec.notebookExecutionJob.dataformRepositorySource.dataformRepositoryResourceName

`string` · required

The repository:
projects/{project}/locations/{location}/repositories/{repository}.

- rule: {"required":true,"string":{"pattern":"^projects/[^/]+/locations/[^/]+/repositories/[^/]+$"}}

### spec.notebookExecutionJob.dataformRepositorySource.commitSha

`string`

Pin a commit. Unset: the repository's HEAD at each run.

### spec.notebookExecutionJob.notebookRuntimeTemplateResourceName

`string | valueFrom`

Run on a runtime template's machine, network, and image: a
GcpColabRuntimeTemplate reference or a literal
projects/{project}/locations/{location}/notebookRuntimeTemplates/{id}.
Exactly one of this or custom_environment_spec.

- references: GcpColabRuntimeTemplate (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpColabRuntimeTemplate, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.notebookExecutionJob.customEnvironmentSpec

`GcpColabScheduleCustomEnvironmentSpec`

Run on a machine described here instead of a template. Exactly one of
this or notebook_runtime_template_resource_name.

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec

`GcpColabScheduleMachineSpec`

The machine and accelerators.

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.machineType

`string`

The Compute Engine machine type, e.g. "n1-standard-4". Immutable.

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.acceleratorType

`string`

The accelerator, e.g. "NVIDIA_TESLA_T4", "NVIDIA_L4",
"NVIDIA_H100_80GB", "TPU_V5_LITEPOD".

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.acceleratorCount

`int32`

How many accelerators the machine gets.

- rule: {"int32":{"gte":0}}

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.gpuPartitionSize

`string`

Split each GPU into NVIDIA multi-instance partitions of this size
(e.g. "1g.10gb"); then accelerator_count should be 1. Immutable.

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.tpuTopology

`string`

The TPU topology for a TPU accelerator, e.g. "2x2x1". Immutable.

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity

`GcpColabScheduleReservationAffinity`

Draw the machine from a Compute Engine reservation.

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.reservationAffinityType

`string` · required

"NO_RESERVATION", "ANY_RESERVATION", "SPECIFIC_RESERVATION",
"SPECIFIC_THEN_ANY_RESERVATION", or "SPECIFIC_THEN_NO_RESERVATION".

- rule: {"required":true,"string":{"in":["NO_RESERVATION","ANY_RESERVATION","SPECIFIC_RESERVATION","SPECIFIC_THEN_ANY_RESERVATION","SPECIFIC_THEN_NO_RESERVATION"]}}

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.key

`string`

The reservation label key; for a reservation by name,
"compute.googleapis.com/reservation-name".

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.values

`[]string`

The reservation label values -- for a named reservation, its full
resource name.

### spec.notebookExecutionJob.customEnvironmentSpec.machineSpec.reservationAffinity.useReservationPool

`bool`

Draw from Google's shared Vertex AI capacity pool.

### spec.notebookExecutionJob.customEnvironmentSpec.networkSpec

`GcpColabScheduleNetworkSpec`

The network the run joins.

### spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.enableInternetAccess

`bool`

Give the run a public internet path. Default false.

### spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.network

`string | valueFrom`

The VPC network: a GcpVpcNetwork reference or a literal
projects/{project}/global/networks/{name}.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.subnetwork

`string | valueFrom`

The subnetwork: a GcpSubnetwork reference or a literal
projects/{project}/regions/{region}/subnetworks/{name}.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.notebookExecutionJob.customEnvironmentSpec.persistentDiskSpec

`GcpColabSchedulePersistentDiskSpec`

The run's persistent disk.

### spec.notebookExecutionJob.customEnvironmentSpec.persistentDiskSpec.diskType

`string`

"pd-standard" (Google's default), "pd-balanced", "pd-ssd", or
"pd-extreme".

- rule: disk_type must be pd-standard, pd-balanced, pd-ssd, or pd-extreme

### spec.notebookExecutionJob.customEnvironmentSpec.persistentDiskSpec.diskSizeGb

`int64`

Disk size in GB. Unset (0): Google's default of 100.

- rule: {"int64":{"gte":"0"}}

### spec.notebookExecutionJob.workbenchRuntime

`bool`

Run in a Vertex AI Workbench instance-based environment (Google's empty
workbench_runtime marker); still needs a template or a custom
environment for the machine. Sent only when true.

### spec.notebookExecutionJob.gcsOutputUri

`string | valueFrom` · required

Where the executed notebook (with its outputs) is written:
a GcpGcsBucket reference (its gs:// URL) or a literal gs://bucket or
gs://bucket/prefix. The run's identity needs write access.

- references: GcpGcsBucket (`status.outputs.url`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.url}} -- a bare string does not parse

### spec.notebookExecutionJob.executionUser

`string`

Run as this user (their email) -- their access applies. Exactly one of
execution_user or service_account.

- rule: execution_user must be an email address

### spec.notebookExecutionJob.serviceAccount

`string | valueFrom`

Run as this service account: a GcpServiceAccount reference or a
literal email. The schedule's caller needs iam.serviceAccounts.actAs
on it. Exactly one of execution_user or service_account; the choice
for unattended production runs.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.notebookExecutionJob.executionTimeout

`string`

The longest a run may execute, in seconds ending in "s" (e.g.
"3600s"). Unset: Google's default of 24 hours.

- rule: execution_timeout must be a duration in seconds such as 3600s

### spec.notebookExecutionJob.kernelName

`string`

The Jupyter kernel to run the notebook with. Unset: the notebook's
default kernel.

### spec.notebookExecutionJob.labels

`map<string, string>`

Labels on every run the schedule launches.

### spec.notebookExecutionJob.kmsKeyName

`string | valueFrom`

Customer-managed encryption key for the run: a GcpKmsKey reference or
a literal key path in the schedule's region.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.pipelineJob

`GcpColabSchedulePipelineJob`

The Vertex AI Pipelines run each tick launches. Exactly one of
notebook_execution_job or pipeline_job. Updates in place.

### spec.pipelineJob.displayName

`string`

The run's name in the console -- up to 128 characters.

- rule: {"string":{"maxLen":"128"}}

### spec.pipelineJob.pipelineSpec

`string`

The compiled pipeline definition as a JSON string -- the output of the
Kubeflow Pipelines SDK compiler. Set this or template_uri.

### spec.pipelineJob.templateUri

`string`

Where the compiled pipeline is downloaded from when pipeline_spec is
empty -- an Artifact Registry pipeline template URI
(https://{region}-kfp.pkg.dev/{project}/{repository}/{template}/{tag}).

### spec.pipelineJob.runtimeConfig

`GcpColabSchedulePipelineRuntimeConfig`

The run's parameters and output location.

### spec.pipelineJob.runtimeConfig.gcsOutputDirectory

`string` · required

The Cloud Storage root the run's artifacts are written under
({job_id}/{task_id}/{output_key}), e.g. gs://bucket/pipeline-root. The
pipeline's service account needs storage.objects.get and
storage.objects.create on it.

- rule: {"required":true,"string":{"pattern":"^gs://[^/]+(/.*)?$"}}

### spec.pipelineJob.runtimeConfig.failurePolicy

`string`

How the run reacts to a failed task:
  "PIPELINE_FAILURE_POLICY_FAIL_SLOW" -- running tasks finish, nothing
                                         new is scheduled
  "PIPELINE_FAILURE_POLICY_FAIL_FAST" -- the run stops at once

- rule: failure_policy must be PIPELINE_FAILURE_POLICY_FAIL_SLOW or PIPELINE_FAILURE_POLICY_FAIL_FAST

### spec.pipelineJob.runtimeConfig.parameterValues

`map<string, string>`

Runtime parameters substituted into the pipeline's placeholders
(pipelines compiled with KFP SDK 1.9+ / the v2 DSL). Values are
strings.

### spec.pipelineJob.serviceAccount

`string | valueFrom`

The service account the pipeline's steps run as: a GcpServiceAccount
reference or a literal email. Unset: the Compute Engine default
service account. The schedule's caller needs
iam.serviceAccounts.actAs on it.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.pipelineJob.network

`string | valueFrom`

A VPC network the pipeline's workloads peer with (private services
access must already be configured on it): a GcpVpcNetwork reference or
a literal network path. Google wants
projects/{project_NUMBER}/global/networks/{name}; the modules resolve a
project ID to its number. Unset: no peering.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.pipelineJob.reservedIpRanges

`[]string`

Names of reserved IP ranges on the peered network the workloads may
use (e.g. "vertex-ai-ip-range"). Unset: any range on the network.

### spec.pipelineJob.preflightValidations

`bool`

Validate every component before the run starts.

### spec.pipelineJob.labels

`map<string, string>`

Labels on every pipeline run. Google overrides the reserved
vertex-ai-pipelines-run-billing-id key.

### spec.pipelineJob.kmsKeyName

`string | valueFrom`

Customer-managed encryption key for the run: a GcpKmsKey reference or
a literal key path in the schedule's region.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.pipelineJob.pscInterfaceConfig

`GcpColabSchedulePscInterfaceConfig`

Private Service Connect interface: reach private services through a
network attachment instead of peering.

### spec.pipelineJob.pscInterfaceConfig.networkAttachment

`string`

The Compute Engine network attachment the run attaches to, in the
schedule's region and project (created beforehand).

### spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs

`[]GcpColabScheduleDnsPeeringConfig`

DNS peering so the run resolves private domains through another
project's Cloud DNS. The Vertex AI service agent needs roles/dns.peer
on each target project.

### spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].domain

`string` · required

The DNS suffix to peer, ending with a dot, e.g.
"internal.example.com.".

- rule: {"required":true,"string":{"pattern":"^.+\\.$"}}

### spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork

`string | valueFrom` · required

The VPC network in target_project where the zone is visible: a
GcpVpcNetwork reference (its name) or a literal network name.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].targetProject

`string | valueFrom` · required

The project hosting the Cloud DNS zone: a GcpProject reference or a
literal project ID.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What happens to the schedule when this resource is destroyed:
  "" / "DELETE" -- the schedule is deleted (runs it already launched
                   and their outputs stay)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the schedule leaves management and keeps launching
                   runs

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_run`: set exactly one of notebook_execution_job or pipeline_job

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpColabSchedule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/schedules/{schedule_id}. |
| `status.outputs.schedule_id` | `string` | The id Google assigned the schedule (the last segment of name). |
| `status.outputs.location` | `string` | The region the schedule lives in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.notebookExecutionJob.notebookRuntimeTemplateResourceName` | GcpColabRuntimeTemplate | `status.outputs.name` |
| `spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.notebookExecutionJob.customEnvironmentSpec.networkSpec.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.notebookExecutionJob.gcsOutputUri` | GcpGcsBucket | `status.outputs.url` |
| `spec.notebookExecutionJob.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.notebookExecutionJob.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.pipelineJob.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.pipelineJob.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.pipelineJob.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.pipelineJob.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
