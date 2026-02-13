---
title: "Loganalyticsworkspace"
description: "Loganalyticsworkspace deployment documentation"
icon: "package"
order: 100
componentName: "azureloganalyticsworkspace"
---

# AzureLogAnalyticsWorkspace: Research & Design Documentation

## 1. What Is Azure Log Analytics?

Azure Log Analytics is the primary log aggregation and query engine in the Azure ecosystem. It is part of Azure Monitor and provides a unified platform for collecting, analyzing, and acting on telemetry data from Azure resources, on-premises infrastructure, and multi-cloud environments.

A **Log Analytics Workspace** is the logical container where this data lives. It defines:
- Where data is stored (region)
- How long data is retained (retention policy)
- Who can access the data (RBAC scoped to the workspace)
- How much data can be ingested per day (quota)
- The pricing model (SKU/tier)

### Core Capabilities

- **Kusto Query Language (KQL)**: A powerful query language for analyzing log data
- **Data collection rules**: Configurable pipelines that control what data is collected
- **Cross-resource queries**: Query across multiple workspaces and Application Insights
- **Alerts**: Rule-based alerting on log patterns and metrics
- **Workbooks**: Interactive reports and dashboards built on log data

## 2. Deployment Landscape

### Level 0: Azure Portal

Most users start by creating workspaces through the Azure Portal. The portal auto-creates a workspace when enabling monitoring features like Container Insights.

### Level 1: Azure CLI

```bash
az monitor log-analytics workspace create \
  --resource-group my-rg \
  --workspace-name my-law \
  --location eastus \
  --sku PerGB2018 \
  --retention-time 90
```

### Level 2: ARM Templates / Bicep

```bicep
resource law 'Microsoft.OperationalInsights/workspaces@2022-10-01' = {
  name: 'my-law'
  location: 'eastus'
  properties: {
    sku: { name: 'PerGB2018' }
    retentionInDays: 90
    workspaceCapping: { dailyQuotaGb: -1 }
  }
}
```

### Level 3: Terraform

```hcl
resource "azurerm_log_analytics_workspace" "main" {
  name                = "my-law"
  location            = "eastus"
  resource_group_name = "my-rg"
  sku                 = "PerGB2018"
  retention_in_days   = 90
  daily_quota_gb      = -1
}
```

### Level 4: Pulumi

```go
workspace, _ := operationalinsights.NewAnalyticsWorkspace(ctx, "my-law",
  &operationalinsights.AnalyticsWorkspaceArgs{
    Name:              pulumi.String("my-law"),
    Location:          pulumi.String("eastus"),
    ResourceGroupName: pulumi.String("my-rg"),
    Sku:               pulumi.String("PerGB2018"),
    RetentionInDays:   pulumi.Int(90),
  })
```

### Level 5: OpenMCF (This Component)

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: my-law
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: platform-rg
      fieldPath: status.outputs.resource_group_name
  name: my-law
  retention_in_days: 90
```

OpenMCF adds composability (StringValueOrRef for resource_group) and consistency (same API pattern across all cloud providers) on top of Terraform/Pulumi.

## 3. Pricing Tiers (SKUs)

### PerGB2018 (Recommended)

- **Model**: Pay-as-you-go per GB ingested
- **Included retention**: 31 days free (for the first 5 GB/day per billing account)
- **Cost**: ~$2.76/GB ingested (varies by region)
- **Best for**: Most workloads, variable ingestion volumes
- **Why we default to this**: Simplest to reason about, no commitment required, scales linearly with usage

### CapacityReservation

- **Model**: Fixed daily commitment at a discount
- **Tiers**: 100, 200, 300, 400, 500, 1000, 2000, 5000 GB/day
- **Discount**: 15-25% vs PerGB2018
- **Best for**: High-volume environments (>100 GB/day) with predictable ingestion
- **Not in 80/20**: Requires understanding of daily volume to right-size

### Standalone / PerNode (Legacy)

- **Status**: These are legacy tiers from OMS (Operations Management Suite)
- **PerNode**: Per-server pricing model, still available for backwards compatibility
- **Standalone**: Per-GB pricing with different rate structure
- **Not recommended**: Microsoft advises migrating to PerGB2018

## 4. Retention Strategy

### Azure's Retention Model

- **Interactive retention**: 30-730 days. Data is queryable via KQL.
- **Archive tier**: Beyond interactive retention, data can be archived for up to 12 years at reduced cost ($0.02/GB/month). Archived data is not directly queryable but can be searched or restored.
- **Free retention**: First 31 days are included with PerGB2018 for the first 5 GB/day.

### Common Retention Patterns

| Use Case | Retention | Rationale |
|----------|-----------|-----------|
| Development | 30 days | Minimum cost, logs are for debugging |
| Production operations | 90 days | 3 months covers most incident investigations |
| Security/compliance | 365 days | One year satisfies most audit requirements |
| Regulated industries | 730 days | Maximum retention for PCI-DSS, HIPAA, SOX |

### Why We Default to 30 Days

The default is the minimum (30 days) rather than a longer period because:
1. Cost scales linearly with retention for large workloads
2. Users can easily increase retention after creation
3. Reducing retention requires manual configuration to avoid data loss
4. For dev/test environments, 30 days is sufficient
5. Production environments should explicitly set retention based on requirements

## 5. Daily Quota (Ingestion Cap)

### How It Works

- Set `daily_quota_gb` to a positive number to cap daily ingestion
- When the cap is reached, ingestion stops until midnight UTC
- Set to `-1` for unlimited ingestion (no cap)
- Azure resets the counter at midnight UTC daily

### When to Use Caps

- **Development/staging**: Prevent runaway costs from log storms
- **Initial deployment**: Set a reasonable cap until you understand ingestion volume
- **Cost-sensitive workloads**: Hard budget constraint on monitoring costs

### When NOT to Use Caps

- **Production**: Dropping logs during incidents is worse than the cost
- **Security monitoring**: Missing security logs creates audit gaps
- **Compliance workloads**: Regulations may require complete log collection

### Why We Default to -1 (Unlimited)

Unlimited ingestion is the safest default because:
1. Capping ingestion in production can hide critical issues
2. Users who need caps are cost-aware and will set them explicitly
3. Azure provides cost alerts as an alternative to hard caps

## 6. OpenMCF Design: The 80/20 Applied

### Fields We Include (80% of Users Need These)

| Field | Justification |
|-------|---------------|
| `region` | Every workspace needs a location |
| `resource_group` | Azure requirement, now with StringValueOrRef |
| `name` | Every workspace needs a name |
| `sku` | Controls pricing model |
| `retention_in_days` | The most impactful cost and compliance lever |
| `daily_quota_gb` | Cost protection for budget-sensitive environments |

### Fields We Exclude (20% or Less Need These)

| Azure Field | Why Excluded |
|-------------|-------------|
| `reservation_capacity_in_gb_per_day` | Only for CapacityReservation SKU (niche) |
| `internet_ingestion_enabled` | Rare security requirement |
| `internet_query_enabled` | Rare security requirement |
| `local_authentication_enabled` | Advanced security setting |
| `cmk_for_query_forced` | Niche encryption requirement |
| `identity` | Managed identity for CMK scenarios |
| `data_collection_rule_id` | Advanced data routing |
| `immediate_data_purge_on_30_days_enabled` | Niche compliance requirement |
| `allow_resource_only_permissions` | Advanced RBAC configuration |

Any of these can be added in a future iteration if demand emerges.

## 7. Integration Points

### As a Foundation Resource

Log Analytics Workspace sits at Layer 1 in Azure infra charts (just above the resource group at Layer 0). It is consumed by:

- **AzureApplicationInsights**: References `workspace_id` to store APM data
- **AzureContainerAppEnvironment**: References workspace for container logs
- **AzureAksCluster**: Container Insights addon sends node/pod/container logs
- **Any Azure resource**: Diagnostic settings can route logs to any workspace

### Output Design

We export 4 outputs:

1. **workspace_id**: The ARM resource ID. This is the primary output used by downstream resources (Container Insights, App Insights, diagnostic settings).
2. **workspace_name**: Useful for display and CLI commands.
3. **primary_shared_key**: Agent authentication key. Required by Log Analytics agents and direct ingestion APIs.
4. **secondary_shared_key**: Backup key for rotation.

### A Note on Shared Keys as Outputs

The shared keys are sensitive outputs. In a production deployment, they should be handled carefully:
- The Pulumi stack marks them as secrets
- The Terraform module marks them as `sensitive = true`
- Applications should retrieve them from the workspace at runtime, not from stored outputs
- Key rotation should use the secondary key to avoid downtime

We include them as outputs because they are needed for agent configuration in infra charts where automation provisions both the workspace and the agents that send data to it.

## 8. Scope Boundaries

### What This Component Does

- Creates an Azure Log Analytics Workspace
- Configures SKU, retention, and daily quota
- Tags the workspace with OpenMCF metadata
- Exports workspace ID, name, and shared keys

### What This Component Does NOT Do

- **Solutions/Intelligence Packs**: Azure Monitor solutions (like Container Insights, Security Center) are separate configurations applied to the workspace
- **Data collection rules**: These control what data sources send to the workspace
- **Diagnostic settings**: These are configured on individual Azure resources, not on the workspace
- **Alerts**: Log search alerts and metric alerts are separate resources
- **Workbooks/dashboards**: These are presentation-layer constructs
- **Archive tier**: Long-term archival beyond interactive retention requires separate configuration
- **Linked automation accounts**: Azure Automation integration is a separate resource
