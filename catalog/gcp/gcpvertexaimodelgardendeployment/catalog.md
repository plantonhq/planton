# GCP Vertex AI Model Garden Deployment

Puts an open model to work in one step. Name a model from Vertex AI Model Garden or the Hugging Face hub, accept its license, and Vertex AI uploads it, creates an endpoint, and deploys it on the compute you declare -- or on the shape Model Garden recommends when you declare none. Minutes later you have a prediction endpoint for Gemma, Llama, Qwen, or any of the hundreds of models Model Garden serves, with autoscaling, Spot capacity, a dedicated DNS name, or Private Service Connect as you choose.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Model, endpoint, and deployment** -- a `vertex.AiEndpointWithModelGardenDeployment`: the uploaded model, the endpoint, and the deployed model serving on it

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/aiplatform.admin` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Accelerator quota

Most open models serve from a GPU. The project needs quota for the accelerator in the region -- an `NVIDIA_L4` in `us-central1` for the presets -- and the replicas bill from the moment they are ready.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Model Garden Deployment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Hugging Face on the Recommended Shape** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiModelGardenDeployment
metadata:
  name: qwen-small
  org: acme-corp
  env: dev
spec:
  location: us-central1
  huggingFaceModelId: Qwen/Qwen3-0.6B
  modelConfig:
    acceptEula: true
```

```shell
planton apply -f model-garden-deployment.yaml
```

This deploys a small Hugging Face model on Model Garden's recommended shape. A Stack Job tracks the provisioning in real time; the model download is the long pole.

### InfraChart

When deploying as part of a multi-resource environment, reference `endpoint_id` or `endpoint_name` from the application or agent that calls the model, and name `GcpProject` and `GcpVpcNetwork` blocks from the Private Service Connect configuration when the endpoint must stay private.

## Key Configuration

These are the most important decisions when configuring a deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Which model** -- exactly one of `publisherModelName` (a Model Garden entry with its version) or `huggingFaceModelId` (a hub id). Gated models need `acceptEula: true`, and gated Hugging Face models a read token.

**Which compute** -- omit `deployConfig` to accept Model Garden's recommended machine and one replica, or declare `dedicatedResources`: the machine type, accelerator type and count, replica bounds, Spot, and the metrics autoscaling follows. `minReplicaCount` is the committed spend.

**Everything is immutable** -- a new model version, machine, or endpoint setting replaces the whole deployment and the endpoint ID changes. Treat a deployment as a versioned artifact and roll forward with a second block.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId`, `endpointConfig.privateServiceConnectConfig.projectAllowlist[]`, `pscAutomationConfig.projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `endpointConfig.privateServiceConnectConfig.pscAutomationConfig.network` | `status.outputs.network_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint_id` | The endpoint's full resource path | Prediction clients; IAM bindings |
| `endpoint_name` | The numeric endpoint ID | SDK `Endpoint(...)` constructors |
| `deployed_model_id` | The deployed model's ID | Traffic splits; undeploy tooling |
| `deployed_model_display_name` | The deployed model's display name | Dashboards |
| `location` | The deployment's location | Regional clients |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Hugging Face on the recommended shape** -- The fastest path from a model id to an endpoint. Start from the **Hugging Face on the Recommended Shape** preset.

**Gemma on L4 with vLLM** -- A publisher model on an explicit GPU with a declared serving container, Spot autoscaling, and a dedicated endpoint. Start from the **Gemma on L4 with vLLM** preset.

## Works With

- [**GCP Vertex AI Endpoint**](/cloud-catalog/gcp-vertex-ai-endpoint) -- an endpoint managed on its own
- [**GCP Vertex AI Agent Engine**](/cloud-catalog/gcp-vertex-ai-agent-engine) -- agents that call the deployed model
- [**GCP Project**](/cloud-catalog/gcp-project) and [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- Private Service Connect consumers
