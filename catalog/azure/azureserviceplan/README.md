# AzureServicePlan

An Azure App Service Plan defines the compute resources (region, VM size, instance count, pricing tier) that host Azure Web Apps, Function Apps, and Logic Apps Standard workflows.

## Overview

The `AzureServicePlan` component provisions an `azurerm_service_plan` resource, providing the compute tier that one or more Azure app workloads run on. It is the **foundation resource** for the `function-app-environment` and `web-app-environment` Infra Charts.

An App Service Plan determines:
- **Region**: Where the compute resources are located
- **OS type**: Linux, Windows, or Windows Container (immutable after creation)
- **SKU**: Pricing tier and VM size (determines max scale-out, features, SLA)
- **Instance count**: Number of VM instances allocated
- **Placement**: Multi-tenant App Service, or single-tenant inside an App Service Environment v3 (Isolated SKUs)

## Key Features

- **Dual IaC support**: Both Pulumi and Terraform modules with feature parity
- **StringValueOrRef resource group**: Composable with `AzureResourceGroup` via `valueFrom`
- **The full SKU vocabulary as a closed enum**: every tier Azure offers -- Free (F1) and Shared (D1) through Basic, Standard, Premium v2/v3/v4 (including the memory-optimized variants), Consumption (Y1), Elastic Premium (EP*), Flex Consumption (FC1), Isolated (ASEv2/ASEv3), and Workflow (WS*) -- validated at manifest time, deployed in Azure's exact spelling
- **SKU-gated rules front-loaded**: premium-only auto-scale, the elastic worker-count pairing, zone-balancing tier support, and the ASE-requires-Isolated rule are all rejected at validation time, mirroring the provider's own apply-time checks
- **Zone redundancy**: Optional availability zone balancing for Premium and above SKUs
- **Premium plan auto-scale**: HTTP-load-based automatic scaling for Premium plans without an autoscale rule resource
- **Per-site scaling**: Independent scaling for individual apps within the plan
- **Elastic worker control**: `maximum_elastic_worker_count` for Elastic Premium / Workflow cost control
- **App Service Environment placement**: `app_service_environment_id` for Isolated (single-tenant) plans
- **User tags**: Free-form Azure resource tags merged over the platform's metadata-derived tags

## When to Use

- **Web applications**: Use with `AzureLinuxWebApp` for hosting web apps, APIs, or backends
- **Serverless functions**: Use with `AzureFunctionApp` for event-driven workloads
- **Shared compute**: Run multiple apps on the same plan to optimize costs
- **Infra Charts**: Foundation resource in `function-app-environment` and `web-app-environment`

## SKU Selection Guide

| Use Case | Recommended SKU | Scale-Out | Zone Redundancy |
|----------|----------------|-----------|-----------------|
| Development/testing | `BASIC_B1` | 3 instances | No |
| Production web apps | `PREMIUM_P1V3` | 30 instances | Yes |
| High-traffic APIs | `PREMIUM_P2V3` or `PREMIUM_P3V3` | 30 instances | Yes |
| Serverless functions (pay-per-use) | `CONSUMPTION_Y1` | 200 (automatic) | Yes |
| Serverless functions (pre-warmed) | `ELASTIC_PREMIUM_EP1` | 100 (elastic) | Yes |
| Enterprise/isolated | `ISOLATED_I1V2` | 100 instances | Yes |

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or AzureResourceGroup ref) |
| `service_plan_name` | string | Yes | - | Plan name (alphanumeric, hyphens, underscores, 1-60 chars) |
| `os_type` | enum | No | `LINUX` | OS type: `LINUX`, `WINDOWS`, or `WINDOWS_CONTAINER` |
| `sku_name` | enum | Yes | - | SKU (e.g. `PREMIUM_P1V3`, `BASIC_B1`, `CONSUMPTION_Y1`, `ELASTIC_PREMIUM_EP1`) |
| `app_service_environment_id` | string | No | - | ARM ID of the App Service Environment v3 (Isolated SKUs only) |
| `worker_count` | int32 | No | SKU default | Number of VM instances |
| `premium_plan_auto_scale_enabled` | bool | No | `false` | HTTP-load auto-scaling (Premium SKUs only) |
| `maximum_elastic_worker_count` | int32 | No | - | Scale-out ceiling (Elastic Premium / Workflow; Premium with auto-scale) |
| `zone_balancing_enabled` | bool | No | `false` | Availability zone balancing (Premium and above) |
| `per_site_scaling_enabled` | bool | No | `false` | Independent app scaling |
| `tags` | map | No | - | Free-form Azure resource tags (merged over metadata-derived tags) |

## Outputs

| Output | Description |
|--------|-------------|
| `service_plan_id` | ARM resource ID (referenced by AzureFunctionApp, AzureLinuxWebApp) |
| `service_plan_name` | Name of the plan |
| `os_type` | Configured OS type in Azure's spelling |
| `sku_name` | Configured SKU in Azure's spelling (e.g. `P1v3`) |
| `kind` | Azure's computed plan classification (e.g. `linux`, `elastic`) |
| `reserved` | Whether the plan runs Linux workers |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureServicePlan
metadata:
  name: my-app-plan
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  servicePlanName: my-app-plan
  skuName: PREMIUM_P1V3
```

## Downstream Usage

The `service_plan_id` output is referenced by app resources:

```yaml
# AzureLinuxWebApp referencing this plan
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLinuxWebApp
metadata:
  name: my-web-app
spec:
  servicePlanId:
    valueFrom:
      kind: AzureServicePlan
      name: my-app-plan
      fieldPath: status.outputs.service_plan_id
  name: my-web-app
  # ...
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
