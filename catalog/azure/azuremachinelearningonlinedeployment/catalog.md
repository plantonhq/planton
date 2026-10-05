# Azure Machine Learning Online Deployment

Creates a managed online deployment on an Azure Machine Learning online endpoint -- a running copy of a model behind the endpoint's address, on Azure-managed VMs with health probes, request limits, and optional model data collection.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Online Deployment** -- an ARM child of the endpoint (`.../onlineEndpoints/{endpoint}/deployments/{name}`) with its model, environment, instance fleet, and probes

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureMachineLearningOnlineEndpoint** -- the deployment attaches to it; the endpoint's identity must already hold its registry and storage grants.

### Azure Subscription

- **Managed-endpoint VM quota** -- separate from regular compute quota; check `Machine Learning managed online endpoint` quotas for the instance family and region.
- **Instances bill while the deployment lives** -- there is no scale-to-zero; the smallest honest footprint is one small instance.
- **Model and environment assets** -- a deployment referencing registered assets needs them in the workspace (or a registry) first; custom containers can embed their model instead.

## Deploy

### Console

Open the deployment store, find **Azure Machine Learning Online Deployment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Registered-Model Deployment** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningOnlineDeployment
metadata:
  name: fraud-scoring-blue
  org: acme-corp
  env: prod
spec:
  endpointId:
    valueFrom:
      name: fraud-scoring
  name: blue
  region: eastus
  instanceType: Standard_DS3_v2
  instanceCount: 2
  model: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/ml-platform-rg/providers/Microsoft.MachineLearningServices/workspaces/ml-prod/models/fraud-model/versions/3
```

```shell
planton apply -f azure-machine-learning-online-deployment.yaml
```

This creates a two-instance `blue` deployment serving registered model version 3; it provisions its instances in ten to twenty minutes, and the endpoint's traffic map then routes to it by name. An Infra Job tracks the provisioning in real time.

### InfraChart

In an ML-platform chart the order is: workspace → online endpoint → **online deployment(s)**; the endpoint's traffic map routes by the deployments' names. Wire the endpoint by reference:

```yaml
spec:
  endpointId:
    valueFrom:
      kind: AzureMachineLearningOnlineEndpoint
      name: fraud-scoring
      fieldPath: status.outputs.online_endpoint_id
```

The InfraPipeline resolves the dependency graph and deploys the endpoint before the deployment.

## Key Configuration

These are the most important decisions when configuring the deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Instance type and count** -- the fleet's size and cost. `instanceCount` is the one dial the service scales without touching containers (it rides the ARM SKU capacity); everything about the instance TYPE is fixed at creation.

**Model, environment, code** -- reference registered workspace assets (the everyday path), or bring a custom container through `environmentId` with the model embedded or mounted. MLflow models need neither code nor environment (the service infers both).

**Probes** -- the liveness/readiness/startup trio with ISO-8601 durations; the service's defaults are sensible, so set them only when your container's startup or health behavior demands it.

**Data collection** -- `dataCollector.collections` captures scoring inputs/outputs to workspace storage for drift monitoring; sampling rate controls the volume.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureMachineLearningOnlineEndpoint** | `endpointId` | `status.outputs.online_endpoint_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `online_deployment_name` | The deployment's name | The endpoint's traffic-map key |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Registered model** -- serve a model version from the workspace registry. Start from the **Registered-Model Deployment** preset.

**Hardened serving** -- secure egress, probes tuned, and data collection on. Start from the **Hardened Monitored Deployment** preset.

## Works With

- [**Azure Machine Learning Online Endpoint**](/infra-catalog/azure-machine-learning-online-endpoint) -- the parent endpoint and traffic dial
- [**Azure Machine Learning Workspace**](/infra-catalog/azure-machine-learning-workspace) -- where models, environments, and code assets live
