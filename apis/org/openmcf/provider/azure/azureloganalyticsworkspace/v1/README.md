# AzureLogAnalyticsWorkspace

## Overview

`AzureLogAnalyticsWorkspace` provisions an Azure Log Analytics Workspace -- the central
data platform for Azure Monitor. It collects, stores, and analyzes log and performance
data from Azure resources, on-premises servers, and third-party services.

Log Analytics Workspaces are the foundation of observability in Azure. They power:

- **Container Insights** -- monitoring for AKS clusters
- **Application Insights** -- APM for web applications and functions
- **Microsoft Sentinel** -- cloud-native SIEM
- **VM Insights** -- performance and dependency monitoring for VMs
- **Diagnostic Settings** -- centralized logging for any Azure resource

## Key Features

- **StringValueOrRef resource_group** -- references an `AzureResourceGroup` output,
  enabling proper dependency wiring in infra charts
- **Flexible retention** -- 30 to 730 days, configurable per compliance requirements
- **Daily ingestion quota** -- prevent cost overruns with configurable daily caps
- **Pay-as-you-go pricing** -- PerGB2018 SKU (default) charges only for data ingested
- **Sensible defaults** -- PerGB2018 SKU, 30-day retention, unlimited ingestion

## When to Use

- As the monitoring foundation in any Azure infra chart
- Before deploying Application Insights (which requires a workspace)
- Before enabling Container Insights on an AKS cluster
- When centralizing logs from multiple Azure resources
- When building a SIEM solution with Microsoft Sentinel

## Spec Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or valueFrom) |
| `name` | string | Yes | - | Workspace name (4-63 chars) |
| `sku` | string | No | PerGB2018 | Pricing tier |
| `retention_in_days` | int32 | No | 30 | Data retention (30-730 days) |
| `daily_quota_gb` | double | No | -1 | Daily ingestion cap (-1 = unlimited) |

## Outputs

| Output | Description |
|--------|-------------|
| `workspace_id` | Azure Resource Manager ID (used by Container Insights, App Insights, etc.) |
| `workspace_name` | Name of the workspace |
| `primary_shared_key` | Authentication key for log ingestion agents |
| `secondary_shared_key` | Backup key for rotation without downtime |

## Quick Example

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: platform-law
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: prod-platform-law
  sku: PerGB2018
  retention_in_days: 90
```

## Downstream Resources

Resources that reference this workspace:

- **AzureApplicationInsights** -- `workspace_id` via StringValueOrRef
- **AzureContainerAppEnvironment** -- `log_analytics_workspace_id` via StringValueOrRef
- **AzureAksCluster** -- Container Insights addon references workspace_id
