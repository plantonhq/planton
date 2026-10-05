# AzureMachineLearningOnlineDeployment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureMachineLearningOnlineDeploymentSpec** defines a MANAGED
online deployment on an Azure Machine Learning online endpoint
(ARM: Microsoft.MachineLearningServices/workspaces/{ws}/
onlineEndpoints/{endpoint}/deployments/{name}) -- a running copy of
a model behind the endpoint's address, on Azure-managed VMs with
health probes and request settings. The endpoint's traffic map
routes scoring requests to deployments by name.

**The contract here is the ARM specification itself** (pinned
api-version 2025-06-01): azurerm carries no resource for ML
deployments, so both engines write the raw ARM shape and this
spec's validation rules carry the full contract burden -- there is
no provider-side schema behind them.

**This spec models the Managed compute type only.** The Kubernetes
and AzureMLCompute variants require attached compute whose
supported story does not exist yet -- a recorded deferral, not an
oversight. Managed deployments scale by instance_count (the
service's Default scale mode); there is no scale-to-zero.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: a registered
# model with scoring code, all three probes on ISO-8601 durations, the
# secure-egress bool -> Enabled/Disabled wire mapping, request settings,
# the data-collector block with per-collection sampling, environment
# variables, and the instance-count -> SKU-capacity contract.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningOnlineDeployment
metadata:
  name: test-ml-online-deployment
  org: test-org
  env: dev
spec:
  endpointId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace/onlineEndpoints/test-fraud-scoring
  name: blue
  region: eastus
  instanceType: Standard_DS3_v2
  instanceCount: 2
  model: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace/models/fraud-model/versions/3
  codeConfiguration:
    codeId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.MachineLearningServices/workspaces/test-ml-workspace/codes/scoring/versions/1
    scoringScript: score.py
  environmentId: azureml://registries/azureml/environments/sklearn-1.5/versions/12
  environmentVariables:
    LOG_LEVEL: info
  appInsightsEnabled: true
  egressPublicNetworkAccessEnabled: false
  livenessProbe:
    failureThreshold: 10
    period: PT10S
    timeout: PT2S
  readinessProbe:
    initialDelay: PT10S
  startupProbe:
    failureThreshold: 60
    period: PT10S
  requestSettings:
    maxConcurrentRequestsPerInstance: 2
    requestTimeout: PT30S
  dataCollector:
    collections:
      model_inputs:
        enabled: true
        samplingRate: 0.2
      model_outputs:
        enabled: true
    rollingRate: Day
    requestLogging:
      captureHeaders:
        - x-request-id
  properties:
    team: ml-platform
  description: Offline-plan deployment exercising the full surface
  tags:
    cost-center: ml-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.endpointId` | `string \| valueFrom` | yes |  | AzureMachineLearningOnlineEndpoint (`status.outputs.online_endpoint_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.instanceType` | `string` |  |  |  |
| `spec.instanceCount` | `int32` |  | `1` |  |
| `spec.model` | `string` |  |  |  |
| `spec.modelMountPath` | `string` |  |  |  |
| `spec.codeConfiguration` | `AzureMachineLearningOnlineDeploymentCodeConfiguration` |  |  |  |
| `spec.codeConfiguration.codeId` | `string` |  |  |  |
| `spec.codeConfiguration.scoringScript` | `string` | yes |  |  |
| `spec.environmentId` | `string` |  |  |  |
| `spec.environmentVariables` | `map<string, string>` |  |  |  |
| `spec.appInsightsEnabled` | `bool` |  |  |  |
| `spec.egressPublicNetworkAccessEnabled` | `bool` |  | `true` |  |
| `spec.livenessProbe` | `AzureMachineLearningOnlineDeploymentProbeSettings` |  |  |  |
| `spec.livenessProbe.failureThreshold` | `int32` |  | `30` |  |
| `spec.livenessProbe.successThreshold` | `int32` |  | `1` |  |
| `spec.livenessProbe.initialDelay` | `string` |  |  |  |
| `spec.livenessProbe.period` | `string` |  |  |  |
| `spec.livenessProbe.timeout` | `string` |  |  |  |
| `spec.readinessProbe` | `AzureMachineLearningOnlineDeploymentProbeSettings` |  |  |  |
| `spec.readinessProbe.failureThreshold` | `int32` |  | `30` |  |
| `spec.readinessProbe.successThreshold` | `int32` |  | `1` |  |
| `spec.readinessProbe.initialDelay` | `string` |  |  |  |
| `spec.readinessProbe.period` | `string` |  |  |  |
| `spec.readinessProbe.timeout` | `string` |  |  |  |
| `spec.startupProbe` | `AzureMachineLearningOnlineDeploymentProbeSettings` |  |  |  |
| `spec.startupProbe.failureThreshold` | `int32` |  | `30` |  |
| `spec.startupProbe.successThreshold` | `int32` |  | `1` |  |
| `spec.startupProbe.initialDelay` | `string` |  |  |  |
| `spec.startupProbe.period` | `string` |  |  |  |
| `spec.startupProbe.timeout` | `string` |  |  |  |
| `spec.requestSettings` | `AzureMachineLearningOnlineDeploymentRequestSettings` |  |  |  |
| `spec.requestSettings.maxConcurrentRequestsPerInstance` | `int32` |  | `1` |  |
| `spec.requestSettings.requestTimeout` | `string` |  |  |  |
| `spec.dataCollector` | `AzureMachineLearningOnlineDeploymentDataCollector` |  |  |  |
| `spec.dataCollector.collections` | `map<string, AzureMachineLearningOnlineDeploymentDataCollection>` | yes |  |  |
| `spec.dataCollector.collections.*.enabled` | `bool` |  |  |  |
| `spec.dataCollector.collections.*.dataId` | `string` |  |  |  |
| `spec.dataCollector.collections.*.clientId` | `string` |  |  |  |
| `spec.dataCollector.collections.*.samplingRate` | `double` |  | `1` |  |
| `spec.dataCollector.rollingRate` | `string` |  |  |  |
| `spec.dataCollector.requestLogging` | `AzureMachineLearningOnlineDeploymentRequestLogging` |  |  |  |
| `spec.dataCollector.requestLogging.captureHeaders` | `[]string` |  |  |  |
| `spec.properties` | `map<string, string>` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.endpointId

`string | valueFrom` · required

The online endpoint the deployment serves, by ARM ID. The
deployment's name is what the endpoint's traffic map routes to.
Fixed at creation.

- references: AzureMachineLearningOnlineEndpoint (`status.outputs.online_endpoint_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMachineLearningOnlineEndpoint, name: <that resource's name>, fieldPath: status.outputs.online_endpoint_id}} -- a bare string does not parse

### spec.name

`string` · required

The deployment's name, unique on its endpoint (ARM's own rule,
mirrored from the pinned specification: starts with a letter or
digit, then letters, digits, hyphens and underscores) -- the key
the endpoint's traffic map routes by (e.g. "blue", "green").
Changing the name replaces the deployment.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9][a-zA-Z0-9-_]{0,254}$"}}

### spec.region

`string` · required

The Azure region the deployment lives in, e.g. "eastus". Must be
its endpoint's own region. Fixed at creation.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.instanceType

`string`

The VM size each instance runs on, e.g. "Standard_DS3_v2" or
"Standard_F4s_v2". Unset applies the service's default
(Standard_F4s_v2). Managed-endpoint VM quota (separate from
regular compute quota) gates what actually provisions. Fixed at
creation.

### spec.instanceCount

`int32` · optional (explicit presence)

How many instances serve the deployment. Unspecified applies 1.
This is the scale dial -- the one setting the service updates
without touching the deployment's containers (ARM carries it as
the deployment's SKU capacity). Managed deployments do not scale
to zero.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.model

`string`

The model the deployment serves: the ARM ID of a registered
model version (.../workspaces/{ws}/models/{name}/versions/{v})
or an azureml:// asset URI. The ARM schema marks it optional
(bring-your-own-container images embed their model), but the
SERVICE rejects a create that names neither a model nor a
self-contained custom container -- synchronously, as a bare
400 BadRequest "The request is invalid." with no field detail
(live-proven). In practice: set this for every deployment
except a true BYOC shape.

### spec.modelMountPath

`string`

Where to mount the model inside a custom container image. Only
meaningful with a custom environment.

### spec.codeConfiguration

`AzureMachineLearningOnlineDeploymentCodeConfiguration`

The scoring code the deployment runs -- a registered code asset
and the script inside it that answers requests. Unset is legal
for MLflow models (the service generates scoring code) and for
custom containers with an embedded server.

### spec.codeConfiguration.codeId

`string`

The registered code asset holding the scoring script, by ARM ID
(.../workspaces/{ws}/codes/{name}/versions/{v}). Unset runs the
script from the environment image alone. Fixed at creation (the
ARM contract -- changing the code asset replaces the deployment;
ship new code as a new deployment and shift traffic).

### spec.codeConfiguration.scoringScript

`string` · required

The script that answers scoring requests, e.g. "score.py" --
required whenever the block is present (ARM's own contract).

- rule: {"required":true}

### spec.environmentId

`string`

The environment (container image + dependencies) the scoring
code runs in: the ARM ID of a registered environment version or
an azureml:// asset URI. Unset lets MLflow models infer one.

### spec.environmentVariables

`map<string, string>`

Environment variables set in the scoring container.

### spec.appInsightsEnabled

`bool`

Whether Application Insights logging is enabled for the
deployment's scoring traffic. ARM's default is false -- the zero
value passes through exactly.

### spec.egressPublicNetworkAccessEnabled

`bool` · optional (explicit presence)

Whether the deployment's OUTBOUND traffic may reach the public
internet (pulling images and models over public networks).
Unspecified applies true (ARM's default "Enabled"). Set false
for secure egress through the workspace's managed network.

- default: `true`

### spec.livenessProbe

`AzureMachineLearningOnlineDeploymentProbeSettings`

Liveness probe -- when it fails past its threshold, the service
restarts the container.

### spec.livenessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the probe reports unhealthy.
Unspecified applies 30 (the service's default).

- default: `30`
- rule: {"int32":{"gte":1}}

### spec.livenessProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes before the probe reports healthy.
Unspecified applies 1 (the service's default).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.livenessProbe.initialDelay

`string`

Delay before the first probe, as an ISO-8601 duration (e.g.
"PT10S"). Unset applies the service default.

- rule: initial_delay must be an ISO-8601 duration, e.g. PT10S

### spec.livenessProbe.period

`string`

Time between probes, as an ISO-8601 duration. Unset applies the
service default (PT10S).

- rule: period must be an ISO-8601 duration, e.g. PT10S

### spec.livenessProbe.timeout

`string`

Per-probe timeout, as an ISO-8601 duration. Unset applies the
service default (PT2S).

- rule: timeout must be an ISO-8601 duration, e.g. PT2S

### spec.readinessProbe

`AzureMachineLearningOnlineDeploymentProbeSettings`

Readiness probe -- the container receives traffic only while it
passes. Defaults mirror the liveness probe's.

### spec.readinessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the probe reports unhealthy.
Unspecified applies 30 (the service's default).

- default: `30`
- rule: {"int32":{"gte":1}}

### spec.readinessProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes before the probe reports healthy.
Unspecified applies 1 (the service's default).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.readinessProbe.initialDelay

`string`

Delay before the first probe, as an ISO-8601 duration (e.g.
"PT10S"). Unset applies the service default.

- rule: initial_delay must be an ISO-8601 duration, e.g. PT10S

### spec.readinessProbe.period

`string`

Time between probes, as an ISO-8601 duration. Unset applies the
service default (PT10S).

- rule: period must be an ISO-8601 duration, e.g. PT10S

### spec.readinessProbe.timeout

`string`

Per-probe timeout, as an ISO-8601 duration. Unset applies the
service default (PT2S).

- rule: timeout must be an ISO-8601 duration, e.g. PT2S

### spec.startupProbe

`AzureMachineLearningOnlineDeploymentProbeSettings`

Startup probe -- gates the other probes until the application
inside the container has started.

### spec.startupProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the probe reports unhealthy.
Unspecified applies 30 (the service's default).

- default: `30`
- rule: {"int32":{"gte":1}}

### spec.startupProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes before the probe reports healthy.
Unspecified applies 1 (the service's default).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.startupProbe.initialDelay

`string`

Delay before the first probe, as an ISO-8601 duration (e.g.
"PT10S"). Unset applies the service default.

- rule: initial_delay must be an ISO-8601 duration, e.g. PT10S

### spec.startupProbe.period

`string`

Time between probes, as an ISO-8601 duration. Unset applies the
service default (PT10S).

- rule: period must be an ISO-8601 duration, e.g. PT10S

### spec.startupProbe.timeout

`string`

Per-probe timeout, as an ISO-8601 duration. Unset applies the
service default (PT2S).

- rule: timeout must be an ISO-8601 duration, e.g. PT2S

### spec.requestSettings

`AzureMachineLearningOnlineDeploymentRequestSettings`

Request handling limits for the deployment.

### spec.requestSettings.maxConcurrentRequestsPerInstance

`int32` · optional (explicit presence)

Concurrent scoring requests each instance accepts. Unspecified
applies 1 (the service's default) -- raise it only when the
scoring container handles parallel requests.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.requestSettings.requestTimeout

`string`

Scoring timeout, as an ISO-8601 duration. Unset applies the
service default (PT5S -- 5000 ms).

- rule: request_timeout must be an ISO-8601 duration, e.g. PT5S

### spec.dataCollector

`AzureMachineLearningOnlineDeploymentDataCollector`

Model data collection -- capturing scoring inputs and outputs to
workspace blob storage for monitoring and drift detection.

### spec.dataCollector.collections

`map<string, AzureMachineLearningOnlineDeploymentDataCollection>` · required

The collections to capture, keyed by the service's collection
names (e.g. "model_inputs", "model_outputs", "request",
"response"). At least one entry is required whenever the block
is present (ARM's own contract).

- rule: {"map":{"minPairs":"1"}}

### spec.dataCollector.collections.*.enabled

`bool`

Whether this collection captures data. ARM's default is
disabled -- an entry usually sets this true (wire values
"Enabled"/"Disabled").

### spec.dataCollector.collections.*.dataId

`string`

The data asset the collection lands in, by ARM ID. Unset lets
the service manage the destination.

### spec.dataCollector.collections.*.clientId

`string`

The client ID of the managed identity used to write collected
data to blob storage. Unset uses the service's own path.

### spec.dataCollector.collections.*.samplingRate

`double` · optional (explicit presence)

The fraction of traffic collected, 0.0-1.0. Unspecified applies
1 (collect everything).

- default: `1`
- rule: {"double":{"lte":1,"gte":0}}

### spec.dataCollector.rollingRate

`string`

How often collected data rolls to a new blob-storage path.
Unset applies the service default (Hour); the service's
vocabulary, exactly.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Year","Month","Day","Hour","Minute"]}}

### spec.dataCollector.requestLogging

`AzureMachineLearningOnlineDeploymentRequestLogging`

Advanced payload-logging settings shared by all collections.

### spec.dataCollector.requestLogging.captureHeaders

`[]string`

Request headers to capture alongside payloads (payload-only is
the default).

### spec.properties

`map<string, string>`

The deployment's ARM property dictionary -- free-form key/value
pairs some tooling reads. ARM allows ADDING entries but never
removing or altering existing ones (the service's own contract;
removals are ignored on update).

### spec.description

`string`

What the deployment serves. Updatable in place.

### spec.tags

`map<string, string>`

Free-form tags applied to the deployment, merged over the
Planton-derived resource tags (organization, environment,
resource id); a user tag with the same key wins. Updatable in
place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMachineLearningOnlineDeployment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.online_deployment_id` | `string` | The Azure Resource Manager ID of the online deployment. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}/onlineEndpoints/{endpoint}/deployments/{name} |
| `status.outputs.online_deployment_name` | `string` | The deployment's name -- the key the endpoint's traffic map routes scoring requests by. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.endpointId` | AzureMachineLearningOnlineEndpoint | `status.outputs.online_endpoint_id` |

## See Also

- [Overview](../README.md)
