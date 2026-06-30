---
title: "Container App Environment"
description: "Container App Environment deployment documentation"
icon: "package"
order: 100
componentName: "azurecontainerappenvironment"
---

# Azure Container App Environment

Deploys an Azure Container Apps Managed Environment with optional VNet injection, internal load balancing, zone redundancy, Log Analytics integration, and dedicated workload profiles for GPU or guaranteed compute. The environment is the hosting boundary for Azure Container Apps -- all apps within it share the same virtual network, logging, and Dapr infrastructure.

## What Gets Created

When you deploy an AzureContainerAppEnvironment resource, Planton provisions:

- **Managed Environment** -- a `containerapp.Environment` resource in the specified region and resource group, configured with the selected networking mode, logging destination, and optional workload profiles
- **VNet Integration** -- created only when `infrastructureSubnetId` is set, injects the environment into the specified subnet for private connectivity to databases, storage, and other VNet resources
- **Log Analytics Integration** -- created only when `logAnalyticsWorkspaceId` is set, automatically configures `log-analytics` as the logging destination for centralized log collection
- **Workload Profiles** -- dedicated compute pools (D4-D32, E4-E32, GPU) created for each entry in `workloadProfiles`, alongside the default Consumption profile
- **Azure Tags** -- resource metadata tags applied to the environment for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** where the environment will be created (can reference an AzureResourceGroup resource)
- **A subnet with /21 or larger address space** if using VNet injection -- the subnet must have at least 2048 IPs for Container Apps infrastructure
- **A Log Analytics Workspace** if persistent log collection is required (can reference an AzureLogAnalyticsWorkspace resource)

## Quick Start

Create a file `container-app-env.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerAppEnvironment
metadata:
  name: my-env
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureContainerAppEnvironment.my-env
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-env
```

Deploy:

```shell
planton apply -f container-app-env.yaml
```

This creates an external Container App Environment with Consumption-only compute, no VNet injection, and streaming-only logs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the environment (e.g., `eastus`, `westeurope`). **ForceNew**. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. **ForceNew**. | Required |
| `name` | `string` | Name of the environment. Becomes part of the default domain for apps (`{app}.{default-domain}`). **ForceNew**. | Required, 2-60 characters, pattern `^[a-z][a-z0-9-]{0,58}[a-z0-9]$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `infrastructureSubnetId` | `StringValueOrRef` | -- | Subnet ID for VNet injection. Must be /21 or larger. Enables private connectivity and unlocks `internalLoadBalancerEnabled` and `zoneRedundancyEnabled`. Can reference an AzureSubnet resource via `valueFrom`. **ForceNew**. |
| `logAnalyticsWorkspaceId` | `StringValueOrRef` | -- | Log Analytics Workspace ID for centralized log collection. When set, logging destination is automatically configured to `log-analytics`. Can reference an AzureLogAnalyticsWorkspace resource via `valueFrom`. |
| `internalLoadBalancerEnabled` | `bool` | `false` | When `true`, apps are only accessible from within the VNet. Requires `infrastructureSubnetId`. **ForceNew**. |
| `zoneRedundancyEnabled` | `bool` | `false` | Distribute infrastructure across availability zones. Requires `infrastructureSubnetId`. **ForceNew**. |
| `workloadProfiles` | `list` | `[]` | Dedicated compute profiles. Each entry has `name` (required), `workloadProfileType` (required, e.g., `D4`, `E8`, `NC24-A100`), optional `minimumCount`, and optional `maximumCount`. The Consumption profile is always available and should not be added here. |

## Examples

### Development Environment with Logging

An external environment with Log Analytics for queryable container logs:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerAppEnvironment
metadata:
  name: dev-env
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureContainerAppEnvironment.dev-env
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-env
  logAnalyticsWorkspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.OperationalInsights/workspaces/dev-logs
```

### Production VNet-Injected Environment

A VNet-injected environment with zone redundancy and logging for production workloads:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerAppEnvironment
metadata:
  name: prod-env
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureContainerAppEnvironment.prod-env
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: prod-env
  infrastructureSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/container-apps
  logAnalyticsWorkspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.OperationalInsights/workspaces/prod-logs
  zoneRedundancyEnabled: true
```

### Internal Environment with Dedicated Compute

An internal VNet-injected environment with dedicated D8 and memory-optimized E16 workload profiles for backend microservices:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerAppEnvironment
metadata:
  name: internal-env
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureContainerAppEnvironment.internal-env
spec:
  region: eastus
  resourceGroup: prod-rg
  name: internal-env
  infrastructureSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/container-apps
  logAnalyticsWorkspaceId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.OperationalInsights/workspaces/prod-logs
  internalLoadBalancerEnabled: true
  zoneRedundancyEnabled: true
  workloadProfiles:
    - name: general
      workloadProfileType: D8
      minimumCount: 1
      maximumCount: 5
    - name: high-memory
      workloadProfileType: E16
      minimumCount: 0
      maximumCount: 3
```

### Using Foreign Key References

Reference Planton-managed resources instead of hardcoding IDs:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureContainerAppEnvironment
metadata:
  name: ref-env
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureContainerAppEnvironment.ref-env
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-env
  infrastructureSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: container-apps-subnet
      field: status.outputs.subnet_id
  logAnalyticsWorkspaceId:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: my-logs
      field: status.outputs.workspace_id
  zoneRedundancyEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `environment_id` | `string` | Azure Resource Manager ID of the Container App Environment. Referenced by AzureContainerApp via `containerAppEnvironmentId`. |
| `default_domain` | `string` | Default domain for apps in this environment. Apps are accessible at `{app-name}.{default_domain}`. |
| `static_ip_address` | `string` | Static IP address of the environment. Public for external environments, private for internal. |
| `platform_reserved_cidr` | `string` | CIDR range reserved for platform infrastructure (VNet-injected environments only) |
| `platform_reserved_dns_ip_address` | `string` | DNS server IP within the platform reserved CIDR (VNet-injected environments only) |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) -- provides the resource group for environment placement
- [AzureSubnet](/docs/catalog/azure/subnet) -- provides the infrastructure subnet for VNet injection
- [AzureLogAnalyticsWorkspace](/docs/catalog/azure/log-analytics-workspace) -- provides centralized log collection
- [AzureContainerApp](/docs/catalog/azure/container-app) -- container apps hosted in this environment
- [AzureVpc](/docs/catalog/azure/vpc-virtual-network) -- provides the virtual network containing the infrastructure subnet
