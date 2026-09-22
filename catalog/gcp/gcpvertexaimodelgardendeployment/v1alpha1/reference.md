# GcpVertexAiModelGardenDeployment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiModelGardenDeploymentSpec deploys a Model Garden or Hugging
Face model to a Vertex AI endpoint in one step
(`google_vertex_ai_endpoint_with_model_garden_deployment`): Vertex AI
uploads the model, creates an endpoint, and deploys the model to it on
the compute declared here.

EVERYTHING in this spec is immutable. Google offers no in-place update
on this resource, so any change -- a new model version, a bigger
machine, an extra replica -- undeploys the model, deletes the endpoint,
and redeploys from scratch; the endpoint ID changes and clients pointed
at the old one break. Treat a deployment as a versioned artifact: roll
forward by declaring a second block and cutting traffic over, not by
editing this one in place. deletion_policy is the only mutable lever.

A deployed model bills its machine (and accelerator) hours from the
moment it is ready until it is undeployed, whether or not it serves a
request.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiModelGardenDeployment
metadata:
  name: qwen-small
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  # A small open Hugging Face model. Exactly one of huggingFaceModelId or
  # publisherModelName is set.
  huggingFaceModelId: Qwen/Qwen3-0.6B
  modelConfig:
    # Gated and licensed models refuse to deploy until the EULA is
    # accepted.
    acceptEula: true
  # deployConfig omitted: Model Garden picks the model's recommended
  # machine shape and one replica. Every field is immutable -- a change
  # redeploys.
  endpointConfig:
    endpointDisplayName: qwen-small
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.publisherModelName` | `string` |  |  |  |
| `spec.huggingFaceModelId` | `string` |  |  |  |
| `spec.modelConfig` | `GcpVertexAiModelGardenDeploymentModelConfig` |  |  |  |
| `spec.modelConfig.acceptEula` | `bool` |  |  |  |
| `spec.modelConfig.huggingFaceAccessToken` | `string` (sensitive) |  |  |  |
| `spec.modelConfig.huggingFaceCacheEnabled` | `bool` |  |  |  |
| `spec.modelConfig.modelDisplayName` | `string` |  |  |  |
| `spec.modelConfig.containerSpec` | `GcpVertexAiModelGardenDeploymentContainerSpec` |  |  |  |
| `spec.modelConfig.containerSpec.imageUri` | `string` | yes |  |  |
| `spec.modelConfig.containerSpec.command` | `[]string` |  |  |  |
| `spec.modelConfig.containerSpec.args` | `[]string` |  |  |  |
| `spec.modelConfig.containerSpec.env` | `[]GcpVertexAiModelGardenDeploymentEnvVar` |  |  |  |
| `spec.modelConfig.containerSpec.env[].name` | `string` | yes |  |  |
| `spec.modelConfig.containerSpec.env[].value` | `string` | yes |  |  |
| `spec.modelConfig.containerSpec.ports` | `[]GcpVertexAiModelGardenDeploymentPort` |  |  |  |
| `spec.modelConfig.containerSpec.ports[].containerPort` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.grpcPorts` | `[]GcpVertexAiModelGardenDeploymentPort` |  |  |  |
| `spec.modelConfig.containerSpec.grpcPorts[].containerPort` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.predictRoute` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.healthRoute` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.deploymentTimeout` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.sharedMemorySizeMb` | `int64` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe` | `GcpVertexAiModelGardenDeploymentProbe` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.periodSeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.successThreshold` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.failureThreshold` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.exec` | `GcpVertexAiModelGardenDeploymentExecAction` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.exec.command` | `[]string` | yes |  |  |
| `spec.modelConfig.containerSpec.startupProbe.grpc` | `GcpVertexAiModelGardenDeploymentGrpcAction` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.grpc.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.grpc.service` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet` | `GcpVertexAiModelGardenDeploymentHttpGetAction` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet.path` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet.host` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet.httpHeaders` | `[]GcpVertexAiModelGardenDeploymentHttpHeader` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.modelConfig.containerSpec.startupProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.tcpSocket` | `GcpVertexAiModelGardenDeploymentTcpSocketAction` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.tcpSocket.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.startupProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe` | `GcpVertexAiModelGardenDeploymentProbe` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.successThreshold` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.exec` | `GcpVertexAiModelGardenDeploymentExecAction` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.exec.command` | `[]string` | yes |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.grpc` | `GcpVertexAiModelGardenDeploymentGrpcAction` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.grpc.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.grpc.service` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet` | `GcpVertexAiModelGardenDeploymentHttpGetAction` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet.path` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet.host` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet.httpHeaders` | `[]GcpVertexAiModelGardenDeploymentHttpHeader` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.tcpSocket` | `GcpVertexAiModelGardenDeploymentTcpSocketAction` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.tcpSocket.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.livenessProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe` | `GcpVertexAiModelGardenDeploymentProbe` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.periodSeconds` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.successThreshold` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.failureThreshold` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.exec` | `GcpVertexAiModelGardenDeploymentExecAction` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.exec.command` | `[]string` | yes |  |  |
| `spec.modelConfig.containerSpec.healthProbe.grpc` | `GcpVertexAiModelGardenDeploymentGrpcAction` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.grpc.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.grpc.service` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet` | `GcpVertexAiModelGardenDeploymentHttpGetAction` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet.path` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet.host` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet.scheme` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet.httpHeaders` | `[]GcpVertexAiModelGardenDeploymentHttpHeader` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.modelConfig.containerSpec.healthProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.tcpSocket` | `GcpVertexAiModelGardenDeploymentTcpSocketAction` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.tcpSocket.port` | `int32` |  |  |  |
| `spec.modelConfig.containerSpec.healthProbe.tcpSocket.host` | `string` |  |  |  |
| `spec.deployConfig` | `GcpVertexAiModelGardenDeploymentDeployConfig` |  |  |  |
| `spec.deployConfig.dedicatedResources` | `GcpVertexAiModelGardenDeploymentDedicatedResources` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec` | `GcpVertexAiModelGardenDeploymentMachineSpec` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.machineType` | `string` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.acceleratorType` | `string` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.acceleratorCount` | `int32` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.tpuTopology` | `string` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.multihostGpuNodeCount` | `int32` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity` | `GcpVertexAiModelGardenDeploymentReservationAffinity` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity.reservationAffinityType` | `string` | yes |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity.key` | `string` |  |  |  |
| `spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity.values` | `[]string` |  |  |  |
| `spec.deployConfig.dedicatedResources.minReplicaCount` | `int32` |  |  |  |
| `spec.deployConfig.dedicatedResources.maxReplicaCount` | `int32` |  |  |  |
| `spec.deployConfig.dedicatedResources.requiredReplicaCount` | `int32` |  |  |  |
| `spec.deployConfig.dedicatedResources.spot` | `bool` |  |  |  |
| `spec.deployConfig.dedicatedResources.autoscalingMetricSpecs` | `[]GcpVertexAiModelGardenDeploymentAutoscalingMetricSpec` |  |  |  |
| `spec.deployConfig.dedicatedResources.autoscalingMetricSpecs[].metricName` | `string` | yes |  |  |
| `spec.deployConfig.dedicatedResources.autoscalingMetricSpecs[].target` | `int32` |  |  |  |
| `spec.deployConfig.fastTryoutEnabled` | `bool` |  |  |  |
| `spec.deployConfig.systemLabels` | `map<string, string>` |  |  |  |
| `spec.endpointConfig` | `GcpVertexAiModelGardenDeploymentEndpointConfig` |  |  |  |
| `spec.endpointConfig.endpointDisplayName` | `string` |  |  |  |
| `spec.endpointConfig.dedicatedEndpointEnabled` | `bool` |  |  |  |
| `spec.endpointConfig.privateServiceConnectConfig` | `GcpVertexAiModelGardenDeploymentPrivateServiceConnectConfig` |  |  |  |
| `spec.endpointConfig.privateServiceConnectConfig.enablePrivateServiceConnect` | `bool` |  |  |  |
| `spec.endpointConfig.privateServiceConnectConfig.projectAllowlist` | `[]string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig` | `GcpVertexAiModelGardenDeploymentPscAutomationConfig` |  |  |  |
| `spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig.projectId` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the endpoint and model live in: a literal project ID
or a GcpProject reference. If omitted, the provider's default project
is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) the model is deployed in, e.g.
"us-central1". Accelerator availability differs by region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.publisherModelName

`string`

A Model Garden model, as
publishers/{publisher}/models/{model}@{version}, e.g.
"publishers/google/models/gemma@gemma-1.1-2b-it" or
"publishers/hf-google/models/gemma-2-2b-it@001" for a Hugging Face
model Model Garden lists. Exactly one of this or
hugging_face_model_id.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^publishers/[^/]+/models/[^/@]+(@[^/]+)?$"}}

### spec.huggingFaceModelId

`string`

A Hugging Face model by its hub ID, e.g. "Qwen/Qwen3-0.6B" or
"google/gemma-2-2b-it"; gated models also need accept_eula and,
when the gate requires it, a hugging_face_access_token. Exactly one
of this or publisher_model_name.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[^/\\s]+/[^/\\s]+$"}}

### spec.modelConfig

`GcpVertexAiModelGardenDeploymentModelConfig`

License acceptance, Hugging Face access, and the serving-container
override.

### spec.modelConfig.acceptEula

`bool`

Accept the model's End User License Agreement. Gated models (Gemma,
Llama, and every Hugging Face model with a license gate) refuse to
deploy until this is true.

### spec.modelConfig.huggingFaceAccessToken

`string` · sensitive

Hugging Face read token used to pull the artifacts of a gated Hugging
Face model. Stored as a secret; never logged.

### spec.modelConfig.huggingFaceCacheEnabled

`bool`

Deploy from Google's cached copy of the Hugging Face model instead of
downloading from Hugging Face -- for VPC Service Controls perimeters
with limited internet egress.

### spec.modelConfig.modelDisplayName

`string`

Display name of the uploaded Model resource; Google picks one when
empty.

### spec.modelConfig.containerSpec

`GcpVertexAiModelGardenDeploymentContainerSpec`

Override the model's serving container.

### spec.modelConfig.containerSpec.imageUri

`string` · required

Container image URI in Artifact Registry (or Container Registry), e.g.
us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-13:latest.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.modelConfig.containerSpec.command

`[]string`

Entrypoint override (the image's ENTRYPOINT is replaced).

### spec.modelConfig.containerSpec.args

`[]string`

Arguments to the entrypoint (the image's CMD is replaced).

### spec.modelConfig.containerSpec.env

`[]GcpVertexAiModelGardenDeploymentEnvVar`

Environment variables for the container.

### spec.modelConfig.containerSpec.env[].name

`string` · required

Variable name.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_.-]*$"}}

### spec.modelConfig.containerSpec.env[].value

`string` · required

Literal value. Variables may reference each other as $(VAR_NAME);
the reference is expanded regardless of declaration order.

- rule: {"required":true}

### spec.modelConfig.containerSpec.ports

`[]GcpVertexAiModelGardenDeploymentPort`

HTTP ports the container listens on; Vertex AI sends predictions and
health checks to the FIRST one (defaults to 8080 when empty).

### spec.modelConfig.containerSpec.ports[].containerPort

`int32` · optional (explicit presence)

The port number (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.grpcPorts

`[]GcpVertexAiModelGardenDeploymentPort`

gRPC ports the container listens on, for models served over gRPC.

### spec.modelConfig.containerSpec.grpcPorts[].containerPort

`int32` · optional (explicit presence)

The port number (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.predictRoute

`string`

HTTP path Vertex AI forwards prediction requests to, e.g. "/predict";
Google's default is its AIP_PREDICT_ROUTE convention.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.modelConfig.containerSpec.healthRoute

`string`

HTTP path Vertex AI GETs to check the container's health, e.g.
"/health"; Google's default is its AIP_HEALTH_ROUTE convention.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.modelConfig.containerSpec.deploymentTimeout

`string`

How long the deployment may take before it fails, as a duration string
(e.g. "1800s"); Google caps it at two hours.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.modelConfig.containerSpec.sharedMemorySizeMb

`int64` · optional (explicit presence)

VM memory reserved as /dev/shm for the model, in megabytes -- large
models loaded across GPUs need it.

- rule: {"int64":{"gte":"1"}}

### spec.modelConfig.containerSpec.startupProbe

`GcpVertexAiModelGardenDeploymentProbe`

Probe that gates the container being considered started; liveness
and health probes wait for it.

- rule: probe timeout_seconds cannot exceed period_seconds

### spec.modelConfig.containerSpec.startupProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds after container start before the first probe (0-3600).

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.modelConfig.containerSpec.startupProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds after which one probe attempt times out (1-3600; Google's
default 1). Must not exceed period_seconds.

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.modelConfig.containerSpec.startupProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probe attempts (1-3600; Google's default 10).

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.modelConfig.containerSpec.startupProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes after a failure before the probe counts as
healthy again (Google's default 1).

- rule: {"int32":{"gte":1}}

### spec.modelConfig.containerSpec.startupProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the container counts as unhealthy
(Google's default 3).

- rule: {"int32":{"gte":1}}

### spec.modelConfig.containerSpec.startupProbe.exec

`GcpVertexAiModelGardenDeploymentExecAction`

Run a command inside the container.

### spec.modelConfig.containerSpec.startupProbe.exec.command

`[]string` · required

The command and its arguments, run without a shell (no variable
expansion, no pipes).

- rule: {"repeated":{"minItems":"1"}}

### spec.modelConfig.containerSpec.startupProbe.grpc

`GcpVertexAiModelGardenDeploymentGrpcAction`

Call the standard gRPC health service.

### spec.modelConfig.containerSpec.startupProbe.grpc.port

`int32` · optional (explicit presence)

Port of the gRPC service (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.startupProbe.grpc.service

`string`

Service name placed in the check request; empty checks the server's
overall health.

### spec.modelConfig.containerSpec.startupProbe.httpGet

`GcpVertexAiModelGardenDeploymentHttpGetAction`

HTTP GET against a path.

### spec.modelConfig.containerSpec.startupProbe.httpGet.path

`string`

Path to probe, e.g. "/health".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.modelConfig.containerSpec.startupProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.startupProbe.httpGet.host

`string`

Host header to send; defaults to the container's IP.

### spec.modelConfig.containerSpec.startupProbe.httpGet.scheme

`string`

HTTP or HTTPS (Google's default HTTP).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HTTP","HTTPS"]}}

### spec.modelConfig.containerSpec.startupProbe.httpGet.httpHeaders

`[]GcpVertexAiModelGardenDeploymentHttpHeader`

Custom headers sent with the probe request.

### spec.modelConfig.containerSpec.startupProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.modelConfig.containerSpec.startupProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.modelConfig.containerSpec.startupProbe.tcpSocket

`GcpVertexAiModelGardenDeploymentTcpSocketAction`

Open a TCP connection.

### spec.modelConfig.containerSpec.startupProbe.tcpSocket.port

`int32` · optional (explicit presence)

Port to connect to (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.startupProbe.tcpSocket.host

`string`

Host to connect to; defaults to the container's IP.

### spec.modelConfig.containerSpec.livenessProbe

`GcpVertexAiModelGardenDeploymentProbe`

Probe that restarts the container when it fails repeatedly.

- rule: probe timeout_seconds cannot exceed period_seconds

### spec.modelConfig.containerSpec.livenessProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds after container start before the first probe (0-3600).

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.modelConfig.containerSpec.livenessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds after which one probe attempt times out (1-3600; Google's
default 1). Must not exceed period_seconds.

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.modelConfig.containerSpec.livenessProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probe attempts (1-3600; Google's default 10).

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.modelConfig.containerSpec.livenessProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes after a failure before the probe counts as
healthy again (Google's default 1).

- rule: {"int32":{"gte":1}}

### spec.modelConfig.containerSpec.livenessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the container counts as unhealthy
(Google's default 3).

- rule: {"int32":{"gte":1}}

### spec.modelConfig.containerSpec.livenessProbe.exec

`GcpVertexAiModelGardenDeploymentExecAction`

Run a command inside the container.

### spec.modelConfig.containerSpec.livenessProbe.exec.command

`[]string` · required

The command and its arguments, run without a shell (no variable
expansion, no pipes).

- rule: {"repeated":{"minItems":"1"}}

### spec.modelConfig.containerSpec.livenessProbe.grpc

`GcpVertexAiModelGardenDeploymentGrpcAction`

Call the standard gRPC health service.

### spec.modelConfig.containerSpec.livenessProbe.grpc.port

`int32` · optional (explicit presence)

Port of the gRPC service (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.livenessProbe.grpc.service

`string`

Service name placed in the check request; empty checks the server's
overall health.

### spec.modelConfig.containerSpec.livenessProbe.httpGet

`GcpVertexAiModelGardenDeploymentHttpGetAction`

HTTP GET against a path.

### spec.modelConfig.containerSpec.livenessProbe.httpGet.path

`string`

Path to probe, e.g. "/health".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.modelConfig.containerSpec.livenessProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.livenessProbe.httpGet.host

`string`

Host header to send; defaults to the container's IP.

### spec.modelConfig.containerSpec.livenessProbe.httpGet.scheme

`string`

HTTP or HTTPS (Google's default HTTP).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HTTP","HTTPS"]}}

### spec.modelConfig.containerSpec.livenessProbe.httpGet.httpHeaders

`[]GcpVertexAiModelGardenDeploymentHttpHeader`

Custom headers sent with the probe request.

### spec.modelConfig.containerSpec.livenessProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.modelConfig.containerSpec.livenessProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.modelConfig.containerSpec.livenessProbe.tcpSocket

`GcpVertexAiModelGardenDeploymentTcpSocketAction`

Open a TCP connection.

### spec.modelConfig.containerSpec.livenessProbe.tcpSocket.port

`int32` · optional (explicit presence)

Port to connect to (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.livenessProbe.tcpSocket.host

`string`

Host to connect to; defaults to the container's IP.

### spec.modelConfig.containerSpec.healthProbe

`GcpVertexAiModelGardenDeploymentProbe`

Probe that decides whether the container receives traffic.

- rule: probe timeout_seconds cannot exceed period_seconds

### spec.modelConfig.containerSpec.healthProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds after container start before the first probe (0-3600).

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.modelConfig.containerSpec.healthProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds after which one probe attempt times out (1-3600; Google's
default 1). Must not exceed period_seconds.

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.modelConfig.containerSpec.healthProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probe attempts (1-3600; Google's default 10).

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.modelConfig.containerSpec.healthProbe.successThreshold

`int32` · optional (explicit presence)

Consecutive successes after a failure before the probe counts as
healthy again (Google's default 1).

- rule: {"int32":{"gte":1}}

### spec.modelConfig.containerSpec.healthProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures before the container counts as unhealthy
(Google's default 3).

- rule: {"int32":{"gte":1}}

### spec.modelConfig.containerSpec.healthProbe.exec

`GcpVertexAiModelGardenDeploymentExecAction`

Run a command inside the container.

### spec.modelConfig.containerSpec.healthProbe.exec.command

`[]string` · required

The command and its arguments, run without a shell (no variable
expansion, no pipes).

- rule: {"repeated":{"minItems":"1"}}

### spec.modelConfig.containerSpec.healthProbe.grpc

`GcpVertexAiModelGardenDeploymentGrpcAction`

Call the standard gRPC health service.

### spec.modelConfig.containerSpec.healthProbe.grpc.port

`int32` · optional (explicit presence)

Port of the gRPC service (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.healthProbe.grpc.service

`string`

Service name placed in the check request; empty checks the server's
overall health.

### spec.modelConfig.containerSpec.healthProbe.httpGet

`GcpVertexAiModelGardenDeploymentHttpGetAction`

HTTP GET against a path.

### spec.modelConfig.containerSpec.healthProbe.httpGet.path

`string`

Path to probe, e.g. "/health".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.modelConfig.containerSpec.healthProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.healthProbe.httpGet.host

`string`

Host header to send; defaults to the container's IP.

### spec.modelConfig.containerSpec.healthProbe.httpGet.scheme

`string`

HTTP or HTTPS (Google's default HTTP).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["HTTP","HTTPS"]}}

### spec.modelConfig.containerSpec.healthProbe.httpGet.httpHeaders

`[]GcpVertexAiModelGardenDeploymentHttpHeader`

Custom headers sent with the probe request.

### spec.modelConfig.containerSpec.healthProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.modelConfig.containerSpec.healthProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.modelConfig.containerSpec.healthProbe.tcpSocket

`GcpVertexAiModelGardenDeploymentTcpSocketAction`

Open a TCP connection.

### spec.modelConfig.containerSpec.healthProbe.tcpSocket.port

`int32` · optional (explicit presence)

Port to connect to (1-65535).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.modelConfig.containerSpec.healthProbe.tcpSocket.host

`string`

Host to connect to; defaults to the container's IP.

### spec.deployConfig

`GcpVertexAiModelGardenDeploymentDeployConfig`

The compute the model is deployed on. Omit to accept Model Garden's
recommended shape for the model.

### spec.deployConfig.dedicatedResources

`GcpVertexAiModelGardenDeploymentDedicatedResources`

Machine shape and replica bounds.

- rule: max_replica_count must be at least min_replica_count

### spec.deployConfig.dedicatedResources.machineSpec

`GcpVertexAiModelGardenDeploymentMachineSpec`

The machine each replica runs on.

- rule: accelerator_type and accelerator_count are set together

### spec.deployConfig.dedicatedResources.machineSpec.machineType

`string`

Compute Engine machine type, e.g. "g2-standard-12" (with an L4) or
"a2-highgpu-1g" (with an A100). Google picks a default for the model
when the whole deploy_config is omitted.

### spec.deployConfig.dedicatedResources.machineSpec.acceleratorType

`string`

Accelerator attached to each replica, e.g. "NVIDIA_L4",
"NVIDIA_TESLA_A100", "NVIDIA_H100_80GB", "TPU_V5_LITEPOD". Set together
with accelerator_count.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[A-Z0-9_]+$"}}

### spec.deployConfig.dedicatedResources.machineSpec.acceleratorCount

`int32` · optional (explicit presence)

Accelerators per replica.

- rule: {"int32":{"gte":1}}

### spec.deployConfig.dedicatedResources.machineSpec.tpuTopology

`string`

TPU topology for TPU machine types, e.g. "2x2x1".

### spec.deployConfig.dedicatedResources.machineSpec.multihostGpuNodeCount

`int32` · optional (explicit presence)

Nodes per replica for multi-host GPU deployments.

- rule: {"int32":{"gte":1}}

### spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity

`GcpVertexAiModelGardenDeploymentReservationAffinity`

Consume Compute Engine reservations for the replicas.

- rule: key and values are set when (and only when) reservation_affinity_type is SPECIFIC_RESERVATION

### spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity.reservationAffinityType

`string` · required

NO_RESERVATION never consumes a reservation; ANY_RESERVATION consumes
any matching one; SPECIFIC_RESERVATION consumes the reservation named
by key and values.

- rule: {"required":true,"string":{"in":["NO_RESERVATION","ANY_RESERVATION","SPECIFIC_RESERVATION"]}}

### spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity.key

`string`

For SPECIFIC_RESERVATION: the label key
"compute.googleapis.com/reservation-name".

### spec.deployConfig.dedicatedResources.machineSpec.reservationAffinity.values

`[]string`

For SPECIFIC_RESERVATION: the full resource name(s) of the reservation
or reservation block.

### spec.deployConfig.dedicatedResources.minReplicaCount

`int32`

Replicas always kept serving (at least 1 -- a deployed model never
scales to zero, so this is the committed spend).

- rule: {"int32":{"gte":1}}

### spec.deployConfig.dedicatedResources.maxReplicaCount

`int32` · optional (explicit presence)

Ceiling the replica count may autoscale to; Google defaults it to
min_replica_count. Sent only when set.

- rule: {"int32":{"gte":1}}

### spec.deployConfig.dedicatedResources.requiredReplicaCount

`int32` · optional (explicit presence)

Replicas that must be available for the deployment to succeed; the
rest are retried. Google defaults it to min_replica_count.

- rule: {"int32":{"gte":1}}

### spec.deployConfig.dedicatedResources.spot

`bool`

Run the replicas on Spot VMs -- much cheaper, preemptible at any
time.

### spec.deployConfig.dedicatedResources.autoscalingMetricSpecs

`[]GcpVertexAiModelGardenDeploymentAutoscalingMetricSpec`

Metrics that drive autoscaling between the replica bounds.

### spec.deployConfig.dedicatedResources.autoscalingMetricSpecs[].metricName

`string` · required

The metric:
"aiplatform.googleapis.com/prediction/online/accelerator/duty_cycle"
or "aiplatform.googleapis.com/prediction/online/cpu/utilization".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.deployConfig.dedicatedResources.autoscalingMetricSpecs[].target

`int32` · optional (explicit presence)

Target utilization percentage (1-100; Google's default 60).

- rule: {"int32":{"lte":100,"gte":1}}

### spec.deployConfig.fastTryoutEnabled

`bool`

Enable Model Garden's fast-tryout serving path for the model when it
supports one.

### spec.deployConfig.systemLabels

`map<string, string>`

Google-managed system labels for Model Garden tracking.

### spec.endpointConfig

`GcpVertexAiModelGardenDeploymentEndpointConfig`

The endpoint the model is served from.

### spec.endpointConfig.endpointDisplayName

`string`

Display name of the created endpoint; Google picks one when empty.

### spec.endpointConfig.dedicatedEndpointEnabled

`bool`

Give the endpoint its own DNS name
({endpoint}.{region}-{project}.prediction.vertexai.goog) isolated from
the shared regional host. Once on, the shared host no longer serves
it.

### spec.endpointConfig.privateServiceConnectConfig

`GcpVertexAiModelGardenDeploymentPrivateServiceConnectConfig`

Expose the endpoint through Private Service Connect.

### spec.endpointConfig.privateServiceConnectConfig.enablePrivateServiceConnect

`bool`

Publish the endpoint through Private Service Connect.

### spec.endpointConfig.privateServiceConnectConfig.projectAllowlist

`[]string | valueFrom`

Consumer projects allowed to create forwarding rules to the endpoint's
service attachment: GcpProject references or literal project IDs.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig

`GcpVertexAiModelGardenDeploymentPscAutomationConfig`

One consumer network Vertex AI creates the PSC endpoint in
automatically (Google's resource takes exactly one; more consumers
build their own forwarding rules against the service attachment).

### spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig.projectId

`string | valueFrom` · required

The consumer project the endpoint is created in: a GcpProject
reference or a literal project ID.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig.network

`string | valueFrom` · required

The consumer VPC network the endpoint is created in, as
projects/{project}/global/networks/{name}. A GcpVpcNetwork reference
resolves to its network_id output.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What happens to the deployment when this resource is destroyed:
  "" / "DELETE" -- the model is undeployed and the endpoint and model
                   are deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the deployment leaves management and keeps serving
                   (and billing)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `deployment.exactly_one_model`: set exactly one of publisher_model_name or hugging_face_model_id

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiModelGardenDeployment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.endpoint_id` | `string` | Fully qualified endpoint resource path: projects/{project}/locations/{location}/endpoints/{endpoint_name}. |
| `status.outputs.endpoint_name` | `string` | The numeric endpoint ID Vertex AI assigned -- the value prediction clients and model-deployment tooling pass as the endpoint reference. |
| `status.outputs.deployed_model_id` | `string` | The numeric ID Vertex AI assigned to the model at the time it was deployed to the endpoint; what an undeploy names. |
| `status.outputs.deployed_model_display_name` | `string` | The display name Vertex AI gave the deployed model. |
| `status.outputs.location` | `string` | The location the model is deployed in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.endpointConfig.privateServiceConnectConfig.projectAllowlist` | GcpProject | `status.outputs.project_id` |
| `spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.endpointConfig.privateServiceConnectConfig.pscAutomationConfig.network` | GcpVpcNetwork | `status.outputs.network_id` |

## See Also

- [Overview](../README.md)
