# GCP Vertex AI Model Garden Deployment

Deploys a Model Garden or Hugging Face model to a Vertex AI endpoint in one step: Vertex AI uploads the model, creates an endpoint, and deploys the model to it on the compute you declare (or on Model Garden's recommended shape when you declare none). Name a publisher model like `publishers/google/models/gemma@gemma-1.1-2b-it` or a Hugging Face id like `Qwen/Qwen3-0.6B`, accept the license, and you have a serving endpoint. Every field is immutable -- a change redeploys -- and the replicas bill from the moment they are ready.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Model, endpoint, and deployment** -- a `vertex_ai_endpoint_with_model_garden_deployment`: the uploaded Model, the Endpoint, and the DeployedModel on it

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/aiplatform.admin` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Quota and cost

Most open models deploy on a GPU; the project needs quota for the accelerator in the chosen region (`NVIDIA_L4` in `us-central1` for the presets). The replicas bill their machine and accelerator hours from ready until undeployed, whether or not requests arrive.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiModelGardenDeployment
metadata:
  name: qwen-small
spec:
  location: us-central1
  huggingFaceModelId: Qwen/Qwen3-0.6B
  modelConfig:
    acceptEula: true
```

```shell
planton apply -f model-garden-deployment.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Vertex AI location (region). Immutable. |
| `publisherModelName` XOR `huggingFaceModelId` | `string` | The model: `publishers/{publisher}/models/{model}@{version}` or a Hugging Face hub id. Exactly one. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `modelConfig.acceptEula` | `bool` | `false` | Accept the model's license; gated models refuse to deploy without it. |
| `modelConfig.huggingFaceAccessToken` | `string` (sensitive) | none | Read token for gated Hugging Face models. |
| `modelConfig.huggingFaceCacheEnabled` | `bool` | `false` | Deploy from Google's cached copy (VPC Service Controls). |
| `modelConfig.modelDisplayName` | `string` | Google-picked | Display name of the uploaded model. |
| `modelConfig.containerSpec` | `object` | model's own | Serving-container override: `imageUri`, `command`, `args`, `env`, `ports`, `grpcPorts`, `predictRoute`, `healthRoute`, `deploymentTimeout`, `sharedMemorySizeMb`, and `startupProbe` / `livenessProbe` / `healthProbe` (each exactly one of `exec`, `grpc`, `httpGet`, `tcpSocket`). |
| `deployConfig.dedicatedResources` | `object` | Model Garden's recommendation | `machineSpec { machineType, acceleratorType, acceleratorCount, tpuTopology, multihostGpuNodeCount, reservationAffinity }`, `minReplicaCount` (at least 1), `maxReplicaCount`, `requiredReplicaCount`, `spot`, `autoscalingMetricSpecs[]`. |
| `deployConfig.fastTryoutEnabled`, `deployConfig.systemLabels` | | | Model Garden's fast-tryout path; Google-managed tracking labels. |
| `endpointConfig.endpointDisplayName` | `string` | Google-picked | Display name of the created endpoint. |
| `endpointConfig.dedicatedEndpointEnabled` | `bool` | `false` | Give the endpoint its own DNS name. |
| `endpointConfig.privateServiceConnectConfig` | `object` | none | `enablePrivateServiceConnect`, `projectAllowlist[]` (`GcpProject` references), `pscAutomationConfig { projectId, network }` (one consumer network Vertex AI creates the endpoint in). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. The only mutable field. |

### Validation Rules

- Exactly one model source; model names match Google's formats.
- `acceleratorType` and `acceleratorCount` are set together; `maxReplicaCount` is at least `minReplicaCount`; `minReplicaCount` is at least 1.
- A reservation affinity carries `key` and `values` when (and only when) its type is `SPECIFIC_RESERVATION`.
- Each probe has exactly one handler and a timeout no longer than its period.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `endpoint_id` | `string` | `projects/{project}/locations/{location}/endpoints/{endpoint_name}` (the shape `GcpVertexAiEndpoint` exports) |
| `endpoint_name` | `string` | The numeric endpoint ID prediction clients pass |
| `deployed_model_id` | `string` | The deployed model's numeric ID on the endpoint |
| `deployed_model_display_name` | `string` | The display name Vertex AI gave the deployed model |
| `location` | `string` | The deployment's location |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything is immutable.** Any change undeploys the model, deletes the endpoint, and redeploys; the endpoint ID changes and clients pointed at the old one break. Roll forward with a second block and cut traffic over.
- **Replicas bill from ready.** `minReplicaCount` is the committed spend; there is no scale-to-zero.
- **Omit `deployConfig` to accept Model Garden's recommendation** -- the fastest way to a working endpoint, at whatever machine shape the model recommends.
- **The Hugging Face token lands in engine state in plain text.** Google's provider does not mark `huggingFaceAccessToken` sensitive, so Terraform state and Pulumi state carry it as written; Planton treats the spec field as a secret (it is never logged or shown), and the state backend's own encryption is what protects it at rest. Use a read-only, model-scoped token.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVertexAiEndpoint** -- an endpoint created on its own, for models deployed through the Vertex AI API
- **GcpVertexAiAgentEngine** -- an agent that calls the deployed model
- **GcpProject**, **GcpVpcNetwork** -- the consumer projects and network of a Private Service Connect endpoint
