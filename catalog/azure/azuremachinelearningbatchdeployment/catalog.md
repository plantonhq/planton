# Azure Machine Learning Batch Deployment

Creates a batch deployment on an Azure Machine Learning batch endpoint -- the job recipe (model, compute, batching behavior) the endpoint's default-deployment pointer routes submissions to.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Batch Deployment** -- an ARM child of the endpoint (`.../batchEndpoints/{endpoint}/deployments/{name}`) carrying the scoring recipe: model reference, compute target, mini-batch sizing, retry policy, and output shape

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureMachineLearningBatchEndpoint** -- the deployment attaches to it.
- **Typically an AzureMachineLearningComputeCluster** -- the pooled compute jobs run on (unset runs serverless where the workspace supports it).

### Azure Subscription

- **Nothing provisions at create time** -- the deployment is a recipe; compute provisions per job when the endpoint is invoked and scales back down after.
- **Model and code are REGISTERED ASSETS** -- the recipe references models/code/environments already registered in the workspace (`az ml model create`, etc.); registering assets is a data-science workflow step, not an infrastructure one.

## Deploy

### Console

Open the deployment store, find **Azure Machine Learning Batch Deployment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Model Scoring Recipe** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningBatchDeployment
metadata:
  name: churn-scoring
  org: acme-corp
  env: prod
spec:
  endpointId:
    valueFrom:
      name: nightly-scoring
  name: production
  region: eastus
  computeId:
    valueFrom:
      name: cpu-pool
  model:
    id:
      assetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/ml-rg/providers/Microsoft.MachineLearningServices/workspaces/ml-prod/models/churn/versions/3
  resources:
    instanceCount: 4
```

```shell
planton apply -f azure-machine-learning-batch-deployment.yaml
```

This registers a scoring recipe named `production` behind the endpoint -- registered model version 3, four cluster nodes per job; invoke the endpoint to run a job from it. An Infra Job tracks the provisioning in real time.

### InfraChart

In an ML-platform chart the order is: workspace → compute cluster → batch endpoint → **batch deployment**; the endpoint's default-deployment pointer then routes submissions to this recipe by name. Wire the parents by reference:

```yaml
spec:
  endpointId:
    valueFrom:
      kind: AzureMachineLearningBatchEndpoint
      name: nightly-scoring
      fieldPath: status.outputs.batch_endpoint_id
  computeId:
    valueFrom:
      kind: AzureMachineLearningComputeCluster
      name: cpu-pool
      fieldPath: status.outputs.machine_learning_compute_cluster_id
```

The InfraPipeline resolves the dependency graph and deploys the endpoint and cluster before the deployment.

## Key Configuration

These are the most important decisions when configuring the deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Model reference** -- one of three arms: `id` (a registered model version's ARM ID -- the standard arm), `dataPath` (a datastore path -- legacy addressing), or `outputPath` (a training job's output -- lineage addressing). Exactly one arm when the block is present; unset is legal for recipes whose environment embeds the model.

**Compute** -- reference an AzureMachineLearningComputeCluster (typically min 0 nodes, so idle cost is zero); unset runs serverless where supported. `resources.instanceCount` decides how many nodes each JOB spreads across.

**Batching dials** -- `miniBatchSize` (how much input each scoring invocation receives; service default 10), `maxConcurrencyPerInstance` (parallel invocations per node; default 1), `errorThreshold` (failures tolerated before the job aborts; default -1 = ignore all), `retrySettings` (per-mini-batch retries and ISO-8601 timeout).

**Pipeline component** -- set the `pipelineComponent` block to run a registered pipeline per job instead of a model recipe; its string-valued `settings` map carries run-time settings like `default_compute`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureMachineLearningBatchEndpoint** | `endpointId` | `status.outputs.batch_endpoint_id` |
| **AzureMachineLearningComputeCluster** | `computeId` | `status.outputs.machine_learning_compute_cluster_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `batch_deployment_id` | ARM ID of the deployment | Operational tooling, imports |
| `batch_deployment_name` | The deployment's name | The endpoint's default-deployment pointer |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Model scoring recipe** -- registered model, compute-cluster reference, tuned batching dials. Start from the **Model Scoring Recipe** preset.

**Pipeline-component recipe** -- a registered pipeline behind the endpoint's address. Start from the **Pipeline Component Recipe** preset.

## Works With

- [**Azure Machine Learning Batch Endpoint**](/infra-catalog/azure-machine-learning-batch-endpoint) -- the parent endpoint
- [**Azure Machine Learning Compute Cluster**](/infra-catalog/azure-machine-learning-compute-cluster) -- the pooled compute jobs run on
- [**Azure Machine Learning Workspace**](/infra-catalog/azure-machine-learning-workspace) -- the workspace everything lives in
