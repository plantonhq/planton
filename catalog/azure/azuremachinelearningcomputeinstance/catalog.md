# Azure Machine Learning Compute Instance

Creates a compute instance on an Azure Machine Learning workspace -- a single always-on VM serving as one data scientist's cloud workstation for notebooks, interactive debugging, and small jobs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Instance** -- an ARM child of the workspace (`.../workspaces/{ws}/computes/{name}`): one VM with its size, ownership, identity, and networking

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **An AzureMachineLearningWorkspace** -- the instance is created on it and always runs in its region.

### Azure Subscription

- **A running VM bills around the clock** -- the instance does not scale to zero; stop it (portal/CLI) outside working hours or delete it when its owner moves on.
- **No update path** -- EVERY change (tags included) replaces the instance; its OS disk and local files go with it. Keep work in datastores and git.
- **Names are region-scoped** -- instance names are unique per Azure region per subscription, not per workspace.

## Deploy

### Console

Open the deployment store, find **Azure Machine Learning Compute Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Personal Dev Instance** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMachineLearningComputeInstance
metadata:
  name: alice-dev
  org: acme-corp
  env: prod
spec:
  workspaceId:
    valueFrom:
      name: ml-prod
  name: alice-dev
  virtualMachineSize: STANDARD_DS3_V2
  authorizationType: personal
  assignToUser:
    tenantId: 00000000-0000-0000-0000-000000000000
    objectId: 11111111-1111-1111-1111-111111111111
  identity:
    type: SYSTEM_ASSIGNED
```

```shell
planton apply -f azure-machine-learning-compute-instance.yaml
```

This creates a DS3v2 workstation owned by the assigned user, with a system identity and SSH disabled; it provisions in roughly five to ten minutes. An Infra Job tracks the provisioning in real time.

### InfraChart

In an ML-platform chart the order is: workspace → **compute instance** (one per team member). Wire the workspace by reference:

```yaml
spec:
  workspaceId:
    valueFrom:
      kind: AzureMachineLearningWorkspace
      name: ml-prod
      fieldPath: status.outputs.machine_learning_workspace_id
```

The InfraPipeline resolves the dependency graph and deploys the workspace before the instance.

## Key Configuration

These are the most important decisions when configuring the instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Ownership** -- `authorizationType: personal` with `assignToUser` is the platform-team pattern: provision the instance FOR a colleague (their tenant and object IDs); without `assignToUser` the deploying principal owns it.

**VM size** -- this VM runs around the clock; size it for interactive work (general-purpose DS-series) and push heavy training to a compute cluster instead.

**Identity** -- give the instance a managed identity and grant it data access; its owner's notebooks then reach storage credential-free. The whole block is fixed at creation here (unlike the cluster's).

**SSH** -- absent means the SSH port is disabled (the right default). When present, the service assigns the username and port -- read them from the outputs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureMachineLearningWorkspace** | `workspaceId` | `status.outputs.machine_learning_workspace_id` |
| **AzureSubnet** | `subnetId` | `status.outputs.subnet_id` |
| **AzureUserAssignedIdentity** | `identity.identityIds[]` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `machine_learning_compute_instance_name` | The instance's name | What its owner selects as their compute |
| `system_assigned_identity_principal_id` | The system identity's principal ID | Storage / Key Vault role assignments |
| `ssh_username`, `ssh_port` | Service-assigned SSH coordinates | Remote development tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Personal dev instance** -- a general-purpose workstation assigned to one user. Start from the **Personal Dev Instance** preset.

**Private workstation** -- a VNet-placed instance with no public IP. Start from the **Private Instance** preset.

## Works With

- [**Azure Machine Learning Workspace**](/infra-catalog/azure-machine-learning-workspace) -- the parent workspace
- [**Azure Machine Learning Compute Cluster**](/infra-catalog/azure-machine-learning-compute-cluster) -- where heavy training belongs
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- VNet placement for the instance
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- bring-your-own identity for data grants
