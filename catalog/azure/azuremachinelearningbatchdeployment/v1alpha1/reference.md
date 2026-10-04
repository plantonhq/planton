# AzureMachineLearningBatchDeployment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningBatchDeploymentSpec** defines a batch
deployment on an Azure Machine Learning batch endpoint (ARM:
Microsoft.MachineLearningServices/workspaces/{ws}/
batchEndpoints/{endpoint}/deployments/{name}) -- the job RECIPE
behind the endpoint's address: which model to run, on which compute
pool, and how to split, retry, and collect the work. Nothing runs
or bills at create time; each endpoint invocation materializes a
job from this recipe.

**The contract here is the ARM specification itself** (pinned
api-version 2025-06-01): azurerm carries no resource for ML
deployments, so both engines write the raw ARM shape and this
spec's validation rules carry the full contract burden -- there is
no provider-side schema behind them.

**Everything in the recipe updates in place** via full PUT (ARM
flags nothing immutable on this surface -- unlike the online
deployment's ForceNew instance type); name, region and endpoint
replace the deployment.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: the id arm of
# the model reference union (the discriminator built from the block),
# a compute-cluster reference, scoring code, per-job resources, every
# batching dial including the -1-legal error threshold, retry settings
# on an ISO-8601 duration, and the ARM property dictionary. The
# pipeline-component arm is deliberately absent here (it is an
# either/or shape with the model recipe -- the presets carry it).
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningBatchDeployment
metadata:
  name: test-ml-batch-deployment
  org: test-org
  env: dev
spec:
  endpointId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace/batchEndpoints/test-nightly-scoring
  name: production
  region: eastus
  computeId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace/computes/test-cpu-pool
  model:
    id:
      assetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace/models/churn/versions/3
  codeConfiguration:
    codeId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace/codes/scoring/versions/1
    scoringScript: score.py
  environmentId: azureml://registries/azureml/environments/sklearn-1.5/versions/12
  environmentVariables:
    LOG_LEVEL: info
  resources:
    instanceCount: 4
    instanceType: Standard_DS3_v2
  miniBatchSize: 50
  maxConcurrencyPerInstance: 2
  errorThreshold: 100
  retrySettings:
    maxRetries: 5
    timeout: PT1M
  outputAction: AppendRow
  outputFileName: scores.csv
  loggingLevel: Debug
  properties:
    team: ml-platform
  description: Offline-plan batch deployment exercising the full surface
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.endpointId` | `string \| valueFrom` | yes |  | AzureMachineLearningBatchEndpoint (`status.outputs.batch_endpoint_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.computeId` | `string \| valueFrom` |  |  | AzureMachineLearningComputeCluster (`status.outputs.machine_learning_compute_cluster_id`) |
| `spec.resources` | `AzureMachineLearningBatchDeploymentResources` |  |  |  |
| `spec.resources.instanceCount` | `int32` |  | `1` |  |
| `spec.resources.instanceType` | `string` |  |  |  |
| `spec.model` | `AzureMachineLearningBatchDeploymentModel` |  |  |  |
| `spec.model.id` | `AzureMachineLearningBatchDeploymentModelIdReference` |  |  |  |
| `spec.model.id.assetId` | `string` | yes |  |  |
| `spec.model.dataPath` | `AzureMachineLearningBatchDeploymentModelDataPathReference` |  |  |  |
| `spec.model.dataPath.datastoreId` | `string` |  |  |  |
| `spec.model.dataPath.path` | `string` |  |  |  |
| `spec.model.outputPath` | `AzureMachineLearningBatchDeploymentModelOutputPathReference` |  |  |  |
| `spec.model.outputPath.jobId` | `string` |  |  |  |
| `spec.model.outputPath.path` | `string` |  |  |  |
| `spec.codeConfiguration` | `AzureMachineLearningBatchDeploymentCodeConfiguration` |  |  |  |
| `spec.codeConfiguration.codeId` | `string` |  |  |  |
| `spec.codeConfiguration.scoringScript` | `string` | yes |  |  |
| `spec.environmentId` | `string` |  |  |  |
| `spec.environmentVariables` | `map<string, string>` |  |  |  |
| `spec.miniBatchSize` | `int64` |  | `10` |  |
| `spec.maxConcurrencyPerInstance` | `int32` |  | `1` |  |
| `spec.errorThreshold` | `int32` |  | `-1` |  |
| `spec.retrySettings` | `AzureMachineLearningBatchDeploymentRetrySettings` |  |  |  |
| `spec.retrySettings.maxRetries` | `int32` |  | `3` |  |
| `spec.retrySettings.timeout` | `string` |  |  |  |
| `spec.outputAction` | `string` |  |  |  |
| `spec.outputFileName` | `string` |  |  |  |
| `spec.loggingLevel` | `string` |  |  |  |
| `spec.pipelineComponent` | `AzureMachineLearningBatchDeploymentPipelineComponent` |  |  |  |
| `spec.pipelineComponent.componentId` | `string` | yes |  |  |
| `spec.pipelineComponent.settings` | `map<string, string>` |  |  |  |
| `spec.pipelineComponent.jobTags` | `map<string, string>` |  |  |  |
| `spec.pipelineComponent.jobDescription` | `string` |  |  |  |
| `spec.properties` | `map<string, string>` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.endpointId

`string | valueFrom` · required

The batch endpoint the deployment serves, by ARM ID. The
deployment's name is what the endpoint's default-deployment
pointer routes submissions to. Fixed at creation.

- references: AzureMachineLearningBatchEndpoint (`status.outputs.batch_endpoint_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMachineLearningBatchEndpoint, name: <that resource's name>, fieldPath: status.outputs.batch_endpoint_id}} -- a bare string does not parse

### spec.name

`string` · required

The deployment's name, unique on its endpoint (ARM's own rule,
mirrored from the pinned specification: starts with a letter or
digit, then letters, digits, hyphens and underscores) -- the key
the endpoint's default-deployment pointer routes by. Changing
the name replaces the deployment.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9-_]{0,254}$"}}

### spec.region

`string` · required

The Azure region the deployment lives in, e.g. "eastus". Must be
its endpoint's own region. Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.computeId

`string | valueFrom`

The compute pool jobs run on, by ARM ID
(.../workspaces/{ws}/computes/{name}) -- typically an
AzureMachineLearningComputeCluster that scales to zero between
jobs. REQUIRED for a Model-type recipe despite the ARM schema
marking it optional: the service validates the create
synchronously and rejects an empty compute with 400
Code="UserError", "'Compute Id' must not be empty."
(ArgumentNullOrEmpty; live-proven). A pipeline-component recipe
carries its compute in the component's own settings instead.

- references: AzureMachineLearningComputeCluster (`status.outputs.machine_learning_compute_cluster_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMachineLearningComputeCluster, name: <that resource's name>, fieldPath: status.outputs.machine_learning_compute_cluster_id}} -- a bare string does not parse

### spec.resources

`AzureMachineLearningBatchDeploymentResources`

Per-job compute sizing -- how many nodes a job spreads across
and what VM size they run (within what the compute pool offers).

### spec.resources.instanceCount

`int32` · optional (explicit presence)

Nodes a job spreads across. Unspecified applies 1 (the
service's default). The compute pool's max node count is the
real ceiling -- ARM accepts any positive count here.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.resources.instanceType

`string`

The VM size jobs request, e.g. "Standard_DS3_v2". Unset uses
the compute pool's own size. Meaningful mainly for serverless
compute -- a cluster pool already fixes its VM size.

### spec.model

`AzureMachineLearningBatchDeploymentModel`

The model the recipe runs -- a registered model asset referenced
one of three ways (exactly one when the block is present).
REQUIRED for a Model-type recipe despite the ARM schema marking
it optional: the service validates the create synchronously and
rejects an empty model with 400 Code="UserError", "'Model
Reference' must not be empty." (ArgumentNullOrEmpty;
live-proven -- an environment image cannot stand in for it).
Only a pipeline-component recipe omits it: the component
carries its own steps.

- rule: exactly one of id, data_path or output_path must be set

### spec.model.id

`AzureMachineLearningBatchDeploymentModelIdReference`

Reference by asset ID -- the standard arm: a registered model
version's ARM resource ID or azureml:// asset URI.

### spec.model.id.assetId

`string` · required

The registered model version's ARM resource ID
(.../workspaces/{ws}/models/{name}/versions/{v}) or azureml://
asset URI -- required whenever this arm is used (ARM's own
contract).

- rule: {"required":true}

### spec.model.dataPath

`AzureMachineLearningBatchDeploymentModelDataPathReference`

Reference by datastore path -- a model stored at a path in a
workspace datastore (the legacy addressing arm).

### spec.model.dataPath.datastoreId

`string`

The workspace datastore holding the model, by ARM resource ID
(.../workspaces/{ws}/datastores/{name}).

### spec.model.dataPath.path

`string`

The file or directory path inside the datastore.

### spec.model.outputPath

`AzureMachineLearningBatchDeploymentModelOutputPathReference`

Reference by job output -- a model produced by a training job's
output (the lineage arm).

### spec.model.outputPath.jobId

`string`

The training job that produced the model, by ARM resource ID.

### spec.model.outputPath.path

`string`

The file or directory path inside the job's output.

### spec.codeConfiguration

`AzureMachineLearningBatchDeploymentCodeConfiguration`

The scoring code the recipe runs -- a registered code asset and
the script inside it that processes each mini-batch. Unset is
legal for MLflow models (the service generates scoring code).

### spec.codeConfiguration.codeId

`string`

The registered code asset holding the scoring script, by ARM ID
(.../workspaces/{ws}/codes/{name}/versions/{v}). Unset runs the
script from the environment image alone.

### spec.codeConfiguration.scoringScript

`string` · required

The script that processes each mini-batch, e.g. "score.py" --
required whenever the block is present (ARM's own contract).

- rule: {"required":true}

### spec.environmentId

`string`

The environment (container image + dependencies) the scoring
code runs in: the ARM ID of a registered environment version or
an azureml:// asset URI. Unset lets MLflow models infer one.

### spec.environmentVariables

`map<string, string>`

Environment variables set in the scoring containers.

### spec.miniBatchSize

`int64` · optional (explicit presence)

How many files (file data) or bytes (tabular data) each scoring
invocation receives. Unspecified applies 10 (the service's
default).

- default: `10`
- rule: {"int64":{"gte":"1"}}

### spec.maxConcurrencyPerInstance

`int32` · optional (explicit presence)

Parallel scoring invocations per node. Unspecified applies 1
(the service's default) -- raise it when the scoring code is
I/O-bound.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.errorThreshold

`int32` · optional (explicit presence)

Failures tolerated before the whole job aborts: file failures
for file data, record failures for tabular data. Unspecified
applies -1 (the service's default: ignore all failures and
finish the job).

- default: `-1`
- rule: {"int32":{"gte":-1}}

### spec.retrySettings

`AzureMachineLearningBatchDeploymentRetrySettings`

Per-mini-batch retry behavior.

### spec.retrySettings.maxRetries

`int32` · optional (explicit presence)

Retries per failed mini-batch. Unspecified applies 3 (the
service's default).

- default: `3`
- rule: {"int32":{"gte":0}}

### spec.retrySettings.timeout

`string`

Per-mini-batch invocation timeout, as an ISO-8601 duration
(e.g. "PT30S"). Unset applies the service default (PT30S).

- rule: timeout must be an ISO-8601 duration, e.g. PT30S

### spec.outputAction

`string`

How job output is organized: "AppendRow" concatenates every
invocation's returned rows into one output file (the service's
default -- predictions-style output); "SummaryOnly" writes no
aggregated file (the scoring code persists its own outputs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AppendRow","SummaryOnly"]}}

### spec.outputFileName

`string`

The aggregated output file's name -- meaningful only with the
AppendRow output action. Unset applies the service's default
("predictions.csv").

### spec.loggingLevel

`string`

Log verbosity for job runs, the service's vocabulary exactly
(increasing verbosity: Warning, Info, Debug). Unset applies the
service's default (Info).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Warning","Info","Debug"]}}

### spec.pipelineComponent

`AzureMachineLearningBatchDeploymentPipelineComponent`

Runs a registered PIPELINE COMPONENT per job instead of a model
scoring recipe (ARM's PipelineComponent deployment type; absent
means the default Model type). A pipeline-component recipe
carries its own steps -- the model / code / environment fields
above describe the Model type and do not feed a pipeline run.

### spec.pipelineComponent.componentId

`string` · required

The registered pipeline component to run, by ARM resource ID
(.../workspaces/{ws}/components/{name}/versions/{v}) --
required whenever this block is present: ARM marks it optional
but a pipeline recipe with no component runs nothing (a
recorded tightening).

- rule: {"required":true}

### spec.pipelineComponent.settings

`map<string, string>`

Run-time settings applied to each pipeline job, string-valued
(e.g. "default_compute": the compute pool pipeline steps run on,
"continue_on_step_failure": "false"). String values are the
typed engine SDK's own shape for this bag; both engines send
them verbatim.

### spec.pipelineComponent.jobTags

`map<string, string>`

Tags applied to each pipeline JOB the recipe creates (distinct
from the deployment's own ARM tags).

### spec.pipelineComponent.jobDescription

`string`

A description applied to each pipeline job.

### spec.properties

`map<string, string>`

The deployment's ARM property dictionary -- free-form key/value
pairs some tooling reads. ARM allows ADDING entries but never
removing or altering existing ones (the service's own contract;
removals are ignored on update).

### spec.description

`string`

What the recipe does. Updatable in place.

### spec.tags

`map<string, string>`

Free-form tags applied to the deployment, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. Updatable in
place.

## Validation Rules

- `ml_batch_deployment_model_type_requires_model_and_compute`: a model-type recipe (no pipeline_component) requires both model and compute_id -- the batch service rejects a create missing either

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningBatchDeployment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.batch_deployment_id` | `string` | The Azure Resource Manager ID of the batch deployment. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}/batchEndpoints/{endpoint}/deployments/{name} |
| `status.outputs.batch_deployment_name` | `string` | The deployment's name -- the key the endpoint's default-deployment pointer routes job submissions by. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.endpointId` | AzureMachineLearningBatchEndpoint | `status.outputs.batch_endpoint_id` |
| `spec.computeId` | AzureMachineLearningComputeCluster | `status.outputs.machine_learning_compute_cluster_id` |

## See Also

- [Overview](../README.md)
