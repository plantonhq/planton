# AzureResourceGroup

## Overview

`AzureResourceGroup` provisions an Azure Resource Group -- the foundational organizational
container for all Azure resources. Every Azure resource must belong to a resource group,
making this the Layer 0 dependency in any Azure infrastructure deployment.

## Why a First-Class Resource?

Resource groups are real infrastructure with their own lifecycle:

- **Lifecycle boundary** -- deleting a resource group cascades to all contained resources
- **RBAC scope** -- Azure role assignments can be scoped to a resource group
- **Cost tracking** -- Azure Cost Management reports costs per resource group
- **Deployment target** -- ARM template deployments target a resource group

By modeling resource groups as a first-class OpenMCF resource, infra charts can express
the full Azure dependency graph. Downstream resources reference the resource group via
`StringValueOrRef`, enabling the platform to build accurate topology graphs and execute
deployments in the correct topological order.

## Key Features

- **Minimal spec** -- only `name` and `region` are required
- **Tag propagation** -- automatically tags the resource group with OpenMCF metadata
  (resource kind, organization, environment)
- **Composable outputs** -- exports `resource_group_name` for downstream `StringValueOrRef`
  wiring, plus `resource_group_id` and `region`

## When to Use

- As the first resource in any Azure infra chart
- When you need explicit control over resource group naming, region, and tags
- When building enterprise Azure architectures with multiple resource groups
  (e.g., separate resource groups for networking, databases, and application tiers)

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Resource group name (1-90 characters) |
| `region` | string | Yes | Azure region (e.g., "eastus", "westeurope") |

## Outputs

| Output | Description |
|--------|-------------|
| `resource_group_id` | Azure Resource Manager ID |
| `resource_group_name` | Name of the resource group (used by downstream `StringValueOrRef`) |
| `region` | Region where the resource group was created |

## Quick Example

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureResourceGroup
metadata:
  name: platform-rg
  org: mycompany
  env: production
spec:
  name: prod-platform-rg
  region: eastus
```

## Downstream Usage

Other Azure resources reference this resource group via `StringValueOrRef`:

```yaml
# In an AzureLogAnalyticsWorkspace spec:
spec:
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  region: eastus
  name: prod-law
```
