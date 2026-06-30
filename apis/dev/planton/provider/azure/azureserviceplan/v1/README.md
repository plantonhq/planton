# AzureServicePlan

An Azure App Service Plan defines the compute resources (region, VM size, instance count, pricing tier) that host Azure Web Apps, Function Apps, and Logic Apps.

## Overview

The `AzureServicePlan` component provisions an `azurerm_service_plan` resource, providing the compute tier that one or more Azure app workloads run on. It is the **foundation resource** for the `function-app-environment` and `web-app-environment` infra charts.

An App Service Plan determines:
- **Region**: Where the compute resources are located
- **OS type**: Linux or Windows (immutable after creation)
- **SKU**: Pricing tier and VM size (determines max scale-out, features, SLA)
- **Instance count**: Number of VM instances allocated

## Key Features

- **Dual IaC support**: Both Pulumi and Terraform modules with feature parity
- **StringValueOrRef resource group**: Composable with `AzureResourceGroup` via `valueFrom`
- **SKU flexibility**: Supports all Azure tiers from Free (F1) through Premium v3 (P3v3) and Elastic Premium (EP3)
- **Zone redundancy**: Optional availability zone balancing for Premium and above SKUs
- **Per-site scaling**: Independent scaling for individual apps within the plan
- **Elastic worker control**: `maximum_elastic_worker_count` for EP* SKU cost control

## When to Use

- **Web applications**: Use with `AzureLinuxWebApp` for hosting web apps, APIs, or backends
- **Serverless functions**: Use with `AzureFunctionApp` for event-driven workloads
- **Shared compute**: Run multiple apps on the same plan to optimize costs
- **Infra charts**: Foundation resource in `function-app-environment` and `web-app-environment`

## SKU Selection Guide

| Use Case | Recommended SKU | Scale-Out | Zone Redundancy |
|----------|----------------|-----------|-----------------|
| Development/testing | B1 | 3 instances | No |
| Production web apps | P1v3 | 30 instances | Yes |
| High-traffic APIs | P2v3 or P3v3 | 30 instances | Yes |
| Serverless functions (pay-per-use) | Y1 | 200 (automatic) | No |
| Serverless functions (pre-warmed) | EP1 | 100 (elastic) | Yes |
| Enterprise/isolated | I1v2 | 100 instances | Yes |

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or AzureResourceGroup ref) |
| `name` | string | Yes | - | Plan name (alphanumeric, hyphens, underscores, 1-60 chars) |
| `os_type` | string | No | `"Linux"` | OS type: `"Linux"` or `"Windows"` |
| `sku_name` | string | Yes | - | SKU name (e.g., `"P1v3"`, `"B1"`, `"Y1"`, `"EP1"`) |
| `worker_count` | int32 | No | SKU default | Number of VM instances |
| `zone_balancing_enabled` | bool | No | `false` | Enable availability zone balancing |
| `per_site_scaling_enabled` | bool | No | `false` | Enable independent app scaling |
| `maximum_elastic_worker_count` | int32 | No | - | Max elastic workers (EP* SKUs) |

## Outputs

| Output | Description |
|--------|-------------|
| `plan_id` | ARM resource ID (referenced by AzureFunctionApp, AzureLinuxWebApp) |
| `plan_name` | Name of the plan |
| `os_type` | Configured OS type |
| `sku_name` | Configured SKU |

## Quick Example

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureServicePlan
metadata:
  name: my-app-plan
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  name: my-app-plan
  sku_name: P1v3
```

## Downstream Usage

The `plan_id` output is referenced by app resources:

```yaml
# AzureLinuxWebApp referencing this plan
apiVersion: azure.planton.dev/v1
kind: AzureLinuxWebApp
metadata:
  name: my-web-app
spec:
  service_plan_id:
    valueFrom:
      kind: AzureServicePlan
      name: my-app-plan
      fieldPath: status.outputs.plan_id
  name: my-web-app
  # ...
```

## What's NOT Included (80/20 Scope)

- **App Service Environment (ASE)**: `app_service_environment_id` for Isolated SKUs. Enterprise niche requiring separate ASE provisioning.
- **Premium auto-scale**: `premium_plan_auto_scale_enabled` for HTTP-based auto-scaling on Premium plans. Newer feature, can be added in v2.
- **WindowsContainer OS type**: Hyper-V containers on Premium v3 only. Very niche use case.
- **Spot instances**: Experimental `is_spot` support. Not for production foundations.

These omissions follow the 80/20 principle: the included fields cover the vast majority of production use cases while keeping the API surface clean and maintainable.
