# AzureServicePlan: Research & Design Documentation

## Executive Summary

Azure App Service Plan (`Microsoft.Web/serverfarms`) is the compute abstraction that underpins Azure Web Apps, Function Apps, and Logic Apps Standard. It defines the region, OS type, VM SKU, instance count, and pricing tier for the hosting environment. One or more apps share the same plan's compute resources.

This document captures the research, design rationale, and 80/20 scoping decisions behind the `AzureServicePlan` OpenMCF component (enum 442, id_prefix `azsp`).

## Azure Deployment Landscape

### Service Plan Architecture

Azure App Service runs on a multi-tenant platform. A Service Plan maps to a **server farm** -- a set of VMs in a specific region, at a specific pricing tier. The key architectural points:

1. **Region-locked**: Plans are created in a specific Azure region. All apps in the plan run in that region.
2. **OS-locked**: Plans are either Linux (`reserved = true`) or Windows (`reserved = false`). This is immutable after creation.
3. **Shared compute**: Multiple apps can run on the same plan, sharing CPU, memory, and instances.
4. **Scale unit**: Scaling the plan (changing `worker_count`) scales all apps in the plan simultaneously, unless per-site scaling is enabled.

### SKU Tier Comparison

| Tier | SKUs | Instances | Auto-Scale | Zones | SLA | Use Case |
|------|------|-----------|------------|-------|-----|----------|
| Free | F1 | 1 | No | No | None | Exploration |
| Shared | D1 | 1 | No | No | None | Low-traffic sites |
| Basic | B1-B3 | 1-3 | No | No | 99.95% | Dev/test |
| Standard | S1-S3 | 1-10 | Rule-based | No | 99.95% | Production entry |
| Premium v2 | P1v2-P3v2 | 1-30 | Yes | Yes | 99.95% | Production |
| Premium v3 | P0v3-P5mv3 | 1-30 | Yes | Yes | 99.95% | Production (recommended) |
| Consumption | Y1 | 0-200 | Automatic | No | 99.95% | Serverless Functions |
| Elastic Premium | EP1-EP3 | 1-100 | Elastic | Yes | 99.95% | Functions (pre-warmed) |
| Isolated v2 | I1v2-I6v2 | 1-100 | Yes | Yes | 99.95% | Enterprise (ASE v3) |

### Pricing Model

- **Free/Shared**: CPU minutes per day (shared VMs)
- **Basic-Premium**: Per-instance per-hour (dedicated VMs)
- **Consumption (Y1)**: Per-execution + per-GB-second (no idle cost)
- **Elastic Premium (EP*)**: Per-instance per-hour for pre-warmed + per-execution overage

### Premium v3 vs Premium v2

Premium v3 is the recommended tier for production workloads:
- **2x memory**: Compared to equivalent v2 SKUs
- **Better CPU**: Dv3 series VMs (Intel Xeon Platinum 8370C)
- **Memory-optimized variants**: P1mv3-P5mv3 with doubled RAM per core
- **Same price**: P1v3 costs the same as P1v2 in most regions
- **P0v3**: New smallest Premium (1 vCPU, 4 GB RAM) -- cost-effective for light production

## Design Decisions

### 1. String+CEL for os_type (not proto enum)

The T02 spec proposed `OsType os_type (enum: LINUX, WINDOWS)`. We changed to `string` with CEL validation `this in ['Linux', 'Windows']` for consistency with the established pattern (since R02 AzureApplicationInsights). Benefits:
- Provider-authentic casing ("Linux"/"Windows" not "LINUX"/"WINDOWS")
- IaC modules pass the value directly without enum-to-string mapping
- Consistent with all other string+CEL fields in the Azure resource family

### 2. No SKU name validation in proto

The Terraform provider validates against 50+ known SKU names. We chose `required + min_len = 1` instead of a CEL whitelist because:
- The list changes with Azure GA releases (Premium v4 added Sept 2025)
- A stale whitelist creates false negatives
- Azure API provides clear error messages for invalid SKUs
- Documenting SKU categories in proto comments provides sufficient guidance

### 3. maximum_elastic_worker_count added (not in T02)

This field was absent from the T02 spec but is essential for EP* SKUs:
- EP* plans auto-scale to 20 workers by default
- Without this cap, costs can spike unexpectedly
- It's the primary cost control lever for serverless Function App workloads
- Terraform provider: `maximum_elastic_worker_count`, validated `IntAtLeast(0)`

### 4. Omitted WindowsContainer os_type

The Terraform provider accepts `"WindowsContainer"` which sets `hyperV = true` in the Azure API. We omitted this because:
- Only supported on Premium v3 SKUs
- Very niche use case (Windows containers in Hyper-V isolation)
- Can be added as a third `os_type` value in v2 if demand exists
- Follows 80/20 principle

### 5. Omitted app_service_environment_id

ASE v3 (Isolated SKUs) requires a separate App Service Environment resource. We omitted this because:
- Isolated SKUs are enterprise-niche (dedicated network, higher cost)
- ASE provisioning is not yet modeled in OpenMCF
- Without an `AzureAppServiceEnvironment` component, the field would require raw ARM IDs
- Can be added when ASE support is implemented

### 6. Omitted premium_plan_auto_scale_enabled

This newer feature enables HTTP traffic-based auto-scaling on Premium plans:
- Requires explicit opt-in (`elasticScaleEnabled = true` in ARM)
- Interacts with `maximum_elastic_worker_count` (must be set together)
- Adds complexity to the spec without broad demand yet
- Can be added in v2 alongside more sophisticated auto-scaling controls

## Terraform Provider Analysis

### Source Files

- `internal/services/appservice/service_plan_resource.go` -- Resource implementation
- `internal/services/appservice/helpers/service_plan.go` -- SKU helper functions
- `internal/services/appservice/validate/service_plan_name.go` -- Name validation

### Key Behaviors

1. **Name validation**: `^[0-9a-zA-Z-_]{1,60}$` (alphanumeric + hyphens + underscores, max 60)
2. **SKU case-insensitivity**: `DiffSuppressFunc` handles case-insensitive SKU comparison
3. **worker_count computed**: If not specified, Azure sets it to the SKU's default capacity
4. **ForceNew fields**: `name`, `resource_group_name`, `location`, `os_type`
5. **Zone balancing ForceNew**: Enabling zone balancing with `worker_count < 2` forces recreation
6. **CustomizeDiff**: Validates SKU-feature compatibility at plan time (before Azure API call)

### API Version

- Azure API: `Microsoft.Web` version `2023-12-01`
- Resource ID: `/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Web/serverFarms/{name}`

## Pulumi Provider Analysis

### Package

- `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice`
- Resource: `appservice.NewServicePlan`
- All spec fields map directly to `ServicePlanArgs` properties

### Field Mapping

| Spec Field | Pulumi Property |
|------------|----------------|
| `name` | `Name` |
| `region` | `Location` |
| `resource_group` | `ResourceGroupName` |
| `os_type` | `OsType` |
| `sku_name` | `SkuName` |
| `worker_count` | `WorkerCount` |
| `zone_balancing_enabled` | `ZoneBalancingEnabled` |
| `per_site_scaling_enabled` | `PerSiteScalingEnabled` |
| `maximum_elastic_worker_count` | `MaximumElasticWorkerCount` |

## Downstream Dependencies

### Resources that reference AzureServicePlan

| Resource | Field | Reference Path |
|----------|-------|---------------|
| AzureFunctionApp | `service_plan_id` | `status.outputs.plan_id` |
| AzureLinuxWebApp | `service_plan_id` | `status.outputs.plan_id` |

### Infra Charts

| Chart | Role |
|-------|------|
| `function-app-environment` | Foundation resource (Service Plan -> Function App) |
| `web-app-environment` | Foundation resource (Service Plan -> Linux Web App) |

## References

- [Azure App Service Plan documentation](https://learn.microsoft.com/en-us/azure/app-service/overview-hosting-plans)
- [Terraform azurerm_service_plan](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/service_plan)
- [Azure App Service pricing](https://azure.microsoft.com/en-us/pricing/details/app-service/linux/)
- [App Service Plan SKU comparison](https://learn.microsoft.com/en-us/azure/app-service/overview-compare)
