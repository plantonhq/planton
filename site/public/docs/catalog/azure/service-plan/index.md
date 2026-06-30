---
title: "Service Plan"
description: "Service Plan deployment documentation"
icon: "package"
order: 100
componentName: "azureserviceplan"
---

# Azure Service Plan

Deploys an Azure App Service Plan that defines the compute tier, VM size, instance count, and pricing for hosting Azure Web Apps, Function Apps, and Logic Apps. The plan supports Linux and Windows operating systems, zone-redundant deployments, per-site scaling, and elastic worker limits for serverless workloads.

## What Gets Created

When you deploy an AzureServicePlan resource, Planton provisions:

- **App Service Plan** -- an `appservice.ServicePlan` resource in the specified region and resource group, configured with the chosen SKU tier, OS type, instance count, and optional zone balancing
- **Azure Tags** -- resource metadata tags applied to the plan for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An Azure Resource Group** where the plan will be created (can reference an AzureResourceGroup resource)
- **SKU selection** -- determine the appropriate SKU tier before deployment: `Y1` for consumption-based Function Apps, `EP1`-`EP3` for elastic premium Functions, `B1`-`B3` for basic web apps, `P1v3`-`P3v3` for production workloads

## Quick Start

Create a file `serviceplan.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServicePlan
metadata:
  name: my-plan
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureServicePlan.my-plan
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-plan
  skuName: B1
```

Deploy:

```shell
planton apply -f serviceplan.yaml
```

This creates a Linux Basic B1 App Service Plan with a single worker instance in the `eastus` region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the plan (e.g., `eastus`, `westeurope`). **ForceNew**: changing this destroys and recreates the plan. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. **ForceNew**: changing this destroys and recreates the plan. | Required |
| `name` | `string` | Name of the Service Plan. Unique within the resource group. **ForceNew**: changing this destroys and recreates the plan. | Required, 1-60 characters, pattern `^[0-9a-zA-Z-_]{1,60}$` |
| `skuName` | `string` | SKU name determining pricing tier and compute capacity. See SKU reference below. | Required, minimum length 1 |

**SKU reference** -- common values by category:

- **Free/Shared**: `F1`, `D1` (Windows only)
- **Basic**: `B1`, `B2`, `B3` (manual scale to 3 instances)
- **Standard**: `S1`, `S2`, `S3` (autoscale to 10, staging slots)
- **Premium v3**: `P1v3`, `P2v3`, `P3v3` (30 instances, zone redundancy)
- **Consumption**: `Y1` (Function Apps pay-per-execution)
- **Elastic Premium**: `EP1`, `EP2`, `EP3` (Function Apps pre-warmed)
- **Isolated v2**: `I1v2`-`I6v2` (App Service Environment required)

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `osType` | `string` | `"Linux"` | Operating system type. Values: `Linux`, `Windows`. All apps within a plan must share the same OS type. The `D1` SKU is Windows-only. **ForceNew**: changing this destroys and recreates the plan. |
| `workerCount` | `int` | _(SKU default, typically 1)_ | Number of VM instances allocated to the plan. Maximum varies by SKU: Basic=3, Standard=10, Premium=30, Isolated=100. When `zoneBalancingEnabled` is `true`, use a multiple of the zone count (typically 3). |
| `zoneBalancingEnabled` | `bool` | `false` | Distribute instances across availability zones. Only supported on Premium (v2/v3), Elastic Premium, Isolated v2, and Workflow SKUs. |
| `perSiteScalingEnabled` | `bool` | `false` | Allow individual apps within the plan to scale independently. Supported on Standard and above. |
| `maximumElasticWorkerCount` | `int` | _(platform default, typically 20)_ | Maximum elastic workers for Elastic Premium (`EP*`) SKUs. Primary cost control lever for serverless Function App workloads. Range: 0-100. Ignored for non-EP SKUs. |

## Examples

### Development Function App Plan

A Consumption plan for pay-per-execution Azure Functions in development:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServicePlan
metadata:
  name: dev-functions
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureServicePlan.dev-functions
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-functions
  skuName: Y1
```

### Production Web App Plan with Zone Redundancy

A Premium v3 plan with zone balancing and multiple workers for production web applications:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServicePlan
metadata:
  name: prod-web
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureServicePlan.prod-web
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: prod-web
  skuName: P1v3
  workerCount: 3
  zoneBalancingEnabled: true
```

### Elastic Premium with Worker Limits

An Elastic Premium plan for serverless Function Apps with a capped elastic worker count to control scaling costs:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServicePlan
metadata:
  name: events-plan
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureServicePlan.events-plan
spec:
  region: eastus
  resourceGroup: prod-rg
  name: events-plan
  skuName: EP1
  maximumElasticWorkerCount: 50
```

### Using Foreign Key References

Reference an Planton-managed resource group instead of hardcoding the name:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServicePlan
metadata:
  name: ref-plan
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureServicePlan.ref-plan
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-plan
  skuName: P1v3
  workerCount: 3
  zoneBalancingEnabled: true
  perSiteScalingEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `plan_id` | `string` | Azure Resource Manager ID of the Service Plan. Referenced by AzureFunctionApp and AzureLinuxWebApp via `servicePlanId`. |
| `plan_name` | `string` | Name of the Service Plan |
| `os_type` | `string` | Configured operating system type (`Linux` or `Windows`) |
| `sku_name` | `string` | Configured SKU name (e.g., `P1v3`, `EP1`, `Y1`) |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) -- provides the resource group for plan placement
- [AzureFunctionApp](/docs/catalog/azure/function-app) -- serverless Function Apps hosted on this plan
- [AzureLinuxWebApp](/docs/catalog/azure/linux-web-app) -- Linux web applications hosted on this plan
