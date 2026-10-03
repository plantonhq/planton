# Azure Resource Group

Deploys an Azure Resource Group -- the foundational organizational container for all Azure resources. Every Azure resource must belong to a resource group, making this the typical first step in any Azure environment setup. The resource group provides lifecycle management, RBAC boundaries, and cost tracking for all contained resources.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Azure Resource Group** -- a named container in the specified Azure region that holds and organizes related Azure resources for unified lifecycle management, access control, and cost allocation. Planton-derived resource tags (organization, environment, resource kind, resource ID) are applied by the IaC module; the spec itself carries only name and region

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An active Azure subscription** where the resource group will be created. No other prerequisites are required -- resource groups are top-level containers.

## Deploy

### Console

Open the deployment store, find **Azure Resource Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the two spec fields: name and region. Start from the **Standard Resource Group** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureResourceGroup
metadata:
  name: platform-rg
  org: acme-corp
  env: prod
spec:
  name: "acme-prod-rg"
  region: "eastus"
```

```shell
planton apply -f resource-group.yaml
```

This creates a resource group named `acme-prod-rg` in the `eastus` region, ready to hold downstream deployments. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a resource group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Name** -- Must be unique within the Azure subscription. Use a descriptive naming convention like `{project}-{env}-rg` to make resource groups identifiable in the portal and CLI. Maximum 90 characters, supports alphanumeric, underscores, hyphens, periods, and parentheses.

**Region** -- Determines where the resource group metadata is stored. Resources inside the group can be in different regions. Choose a region close to your primary deployment target for consistency, though this has no performance impact on contained resources.

**Lifecycle implications** -- Deleting a resource group deletes all resources within it. Use separate resource groups for resources with different lifecycles (e.g., shared networking in one group, application workloads in another).

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `resource_group_id` | Azure Resource Manager ID of the resource group (`/subscriptions/{id}/resourceGroups/{name}`) | Scope for Azure Role Assignments granting access to everything in the group |
| `resource_group_name` | Name of the resource group | Nearly every Azure Infra Component references this via its `resourceGroup` field with `valueFrom` |

## Common Patterns

**Foundation of every Azure InfraChart** -- Virtually every Azure deployment begins by creating one or more resource groups, then wiring downstream components to `status.outputs.resource_group_name` via `valueFrom`. Start from the **Standard Resource Group** preset.

**One group per lifecycle, not per resource** -- Deleting a resource group deletes everything inside it, so group resources that live and die together: shared networking in one group, each application's workloads in another. Splitting by lifecycle means a teardown of one application can never take the VNet with it; lumping everything into one group turns every cleanup into a risk assessment.

**Environment boundaries** -- One resource group per environment (`acme-dev-rg`, `acme-prod-rg`) gives you clean RBAC scoping and per-environment cost tracking with no extra tooling -- Azure cost analysis and role assignments both operate naturally at the resource group boundary.

## Works With

- [**Azure Virtual Network**](/infra-catalog/azure-virtual-network) -- networking foundations deploy into a resource group and are typically the next resource created after it
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- references the resource group name for placement of storage resources
- [**Azure Service Plan**](/infra-catalog/azure-service-plan) -- compute plans for web apps and functions deploy into a resource group
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- grants roles scoped to the resource group using its ARM ID as the scope