# Azure Machine Learning Batch Endpoint

Creates a batch endpoint on an Azure Machine Learning workspace -- the stable address batch scoring jobs are submitted to, with Microsoft Entra authentication and a default-deployment pointer that routes submissions.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Batch Endpoint** -- an ARM child of the workspace (`.../workspaces/{ws}/batchEndpoints/{name}`) with its auth mode, optional managed identity, and default-deployment pointer

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureMachineLearningWorkspace** -- the endpoint is created on it.

### Azure Subscription

- **The endpoint alone runs nothing** -- job submission works once an AzureMachineLearningBatchDeployment attaches to it (and typically a compute cluster behind that).
- **Authentication is Microsoft Entra only** -- every job submission presents an Entra token; there are no endpoint keys to manage (the service rejects key auth outright).

## Deploy

### Console

Open the deployment store, find **Azure Machine Learning Batch Endpoint**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Batch Scoring Endpoint** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningBatchEndpoint
metadata:
  name: nightly-scoring
  org: acme-corp
  env: prod
spec:
  workspaceId:
    valueFrom:
      name: ml-prod
  name: nightly-scoring
  region: eastus
```

```shell
planton apply -f azure-machine-learning-batch-endpoint.yaml
```

This creates an Entra-authenticated batch endpoint on the referenced workspace, free at rest; deployments then attach to it by reference. An Infra Job tracks the provisioning in real time.

### InfraChart

In an ML-platform chart the order is: workspace → compute cluster → **batch endpoint** → batch deployment(s); the endpoint's default-deployment pointer routes submissions by deployment name. Wire the workspace by reference:

```yaml
spec:
  workspaceId:
    valueFrom:
      kind: AzureMachineLearningWorkspace
      name: ml-prod
      fieldPath: status.outputs.machine_learning_workspace_id
```

The InfraPipeline resolves the dependency graph and deploys the workspace before the endpoint.

## Key Configuration

These are the most important decisions when configuring the endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Auth mode** -- leave unset. The batch service accepts only `AADToken` (Microsoft Entra tokens); the platform applies it as the default, and the spec rejects the key modes ARM's shared enum advertises but the service refuses.

**Default deployment name** -- which deployment answers job submissions that do not name one. It usually points at a deployment that does not exist yet at endpoint creation: create the endpoint, add deployments, then set the pointer (it updates in place).

**Identity** -- optional, unlike the online endpoint sibling: batch jobs run under the SUBMITTER's Entra token plus the COMPUTE pool's managed identity, so the endpoint's own identity sits outside the batch data path. Set one only when your role-grant conventions target the endpoint object.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureMachineLearningWorkspace** | `workspaceId` | `status.outputs.machine_learning_workspace_id` |
| **AzureUserAssignedIdentity** | `identity.identityIds[]` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `batch_endpoint_id` | ARM ID of the endpoint | What AzureMachineLearningBatchDeployment wires to |
| `batch_endpoint_name` | The endpoint's name | Default-deployment routing, operational tooling |
| `scoring_uri` | The HTTPS job-submission address | Scheduler / pipeline configuration |
| `swagger_uri` | The OpenAPI document address | Client generation |
| `system_assigned_identity_principal_id` | The system identity's principal ID, when one is enabled | Role assignments |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Batch scoring endpoint** -- the everyday shape: Entra auth by default, no identity, deployments route by the default pointer. Start from the **Batch Scoring Endpoint** preset.

**Routed endpoint with identity** -- a system identity for role-grant conventions and an explicit default-deployment pointer. Start from the **Routed Endpoint with Identity** preset.

## Works With

- [**Azure Machine Learning Workspace**](/infra-catalog/azure-machine-learning-workspace) -- the parent workspace
- [**Azure Machine Learning Batch Deployment**](/infra-catalog/azure-machine-learning-batch-deployment) -- the job recipes behind the endpoint
- [**Azure Machine Learning Compute Cluster**](/infra-catalog/azure-machine-learning-compute-cluster) -- the pooled compute batch jobs run on
