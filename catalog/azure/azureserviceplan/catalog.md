# Azure Service Plan

Deploys an Azure App Service Plan -- the compute tier that hosts Azure Web Apps, Function Apps, and Logic Apps. The plan determines the region, OS type, VM size, instance count, pricing tier, and scaling behavior. A single plan can host multiple apps sharing the same compute resources. The resource group dependency is wired through ValueFromRef for InfraChart composition.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Azure App Service Plan** -- a compute resource in the specified region and resource group, configured with the selected SKU tier, OS type, worker count, and scaling settings
- **Zone-Balanced Instances** -- created only when `zoneBalancingEnabled` is `true`; distributes workers across availability zones for higher resilience (requires Premium, Elastic Premium, or Isolated v2 SKUs)
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the plan for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the Service Plan will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Service Plan**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Linux Standard Plan** preset for a production-ready S1 configuration in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServicePlan
metadata:
  name: app-plan
  org: acme-corp
  env: prod
spec:
  region: "eastus"
  resourceGroup:
    value: "acme-prod-rg"
  servicePlanName: "acme-prod-plan"
  osType: LINUX
  skuName: STANDARD_S1
```

```shell
planton apply -f service-plan.yaml
```

This creates a Linux Standard S1 plan with a single worker instance -- zone balancing, per-site scaling, and elastic worker limits stay at their defaults. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Service Plan to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the Service Plan with the resolved value.

## Key Configuration

These are the most important decisions when configuring a Service Plan. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SKU tier** (`skuName`) -- One enum value picks both the tier's capabilities and the VM size. Free (`FREE_F1`) and Shared (`SHARED_D1`, `SHARED`) are dev/test only. Basic (`BASIC_B1`-`BASIC_B3`) is dedicated compute with manual scaling. Standard (`STANDARD_S1`-`STANDARD_S3`) adds auto-scaling, staging slots, and backups. Premium v3/v4 (`PREMIUM_P0V3`-`PREMIUM_P5MV4`) is the price-performance sweet spot with zone redundancy and premium plan auto-scale. Consumption (`CONSUMPTION_Y1`), Elastic Premium (`ELASTIC_PREMIUM_EP1`-`EP3`), and Flex Consumption (`FLEX_CONSUMPTION_FC1`) are the serverless Functions tiers. Isolated (`ISOLATED_I1`-`ISOLATED_I5MV2`) runs single-tenant inside an App Service Environment. Workflow (`WORKFLOW_WS1`-`WS3`) hosts Logic Apps Standard. The SKU is NOT fixed -- plans re-tier in place.

**OS type** (`osType`) -- `LINUX`, `WINDOWS`, or `WINDOWS_CONTAINER`; leaving it unset deploys Linux. Immutable after creation, and all apps within a plan share it. The Shared-tier SKUs are Windows-only.

**Worker count and zone balancing** -- Set `workerCount` to control the number of VM instances (leave it unset for Consumption, Flex, and Elastic Premium -- Azure manages those automatically). For zone redundancy, enable `zoneBalancingEnabled` with `workerCount` at a multiple of 3 and never below 2 -- enabling it later on a plan below 2 workers forces a destroy-and-recreate. Zone balancing needs Premium or higher tiers; Free, Shared, Basic, and Standard do not support it.

**Per-site scaling** (`perSiteScalingEnabled`) -- When enabled, individual apps within the plan can scale to fewer instances than the plan runs.

**Premium plan auto-scale** (`premiumPlanAutoScaleEnabled`) -- Premium SKUs only: Azure adds and removes instances based on HTTP load without a separate autoscale rule resource, up to the elastic worker cap.

**Elastic worker limit** (`maximumElasticWorkerCount`) -- The upper bound on event-driven scale-out: applies natively on Elastic Premium (default 20, max 100) and Workflow SKUs, and on Premium SKUs when premium plan auto-scale is enabled. The primary cost-control lever for serverless workloads.

**App Service Environment** (`appServiceEnvironmentId`) -- The ASEv3 ARM ID for single-tenant, network-isolated hosting. Required by (and only legal with) the Isolated SKUs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_plan_id` | Azure Resource Manager ID of the Service Plan | AzureLinuxWebApp and AzureFunctionApp `servicePlanId` field -- the reference that places an app on this plan's compute |

The remaining outputs (`service_plan_name`, `os_type`, `sku_name`, `kind`, `reserved`) are readbacks of the deployed configuration for debugging, cost tracking, and audit trails; they are not wired into other Infra Components.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Linux Standard plan** -- S1 tier with a single Linux worker. Provides auto-scaling up to 10 instances, staging slots, daily backups, and a 99.95% SLA. The entry-level production configuration for web apps and APIs. Start from the **Linux Standard Plan** preset.

**Linux Premium with zone redundancy** -- P1v3 tier with 3 Linux workers distributed across availability zones. Faster processors, SSD storage, and auto-scaling up to 30 instances. The standard configuration for high-availability production workloads. Start from the **Linux Premium Plan with Zone Redundancy** preset.

**Consumption (serverless)** -- Y1 tier that scales to zero and bills per execution. The cheapest option for event-driven Azure Functions with sporadic traffic. No worker count or scaling configuration needed. Start from the **Consumption (Serverless) Service Plan** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the Service Plan is created
- [**Azure Linux Web App**](/infra-catalog/azure-linux-web-app) -- deploys onto this plan by referencing `service_plan_id`
- [**Azure Function App**](/infra-catalog/azure-function-app) -- runs on this plan's compute (Consumption, Elastic Premium, or dedicated tiers) via `service_plan_id`