# AzureLogAnalyticsWorkspace

## Overview

`AzureLogAnalyticsWorkspace` provisions an Azure Log Analytics Workspace -- the central
data platform for Azure Monitor. It collects, stores, and analyzes log and performance
data from Azure resources, on-premises servers, and third-party services.

Log Analytics Workspaces are the foundation of observability in Azure. They power:

- **Diagnostic Settings** -- centralized platform logs for any Azure resource
  (`AzureMonitorDiagnosticSetting` routes into the workspace)
- **Application Insights** -- APM for web applications and functions
  (`AzureApplicationInsights` stores its telemetry here)
- **Scheduled Query Alerts** -- KQL-driven alerting
  (`AzureMonitorScheduledQueryAlert` queries the workspace)
- **Container Insights** -- monitoring for AKS clusters
- **Microsoft Sentinel** -- cloud-native SIEM

## Key Features

- **Full pricing surface** -- pay-as-you-go (PerGB2018, the default) and commitment
  tiers (CapacityReservation with the 100-50000 GB/day ladder); the
  PerGB2018 <-> CapacityReservation transition updates in place
- **Security posture** -- Entra-only mode (`local_authentication_enabled: false`),
  private-link-only ingestion and query, resource-context vs workspace-context
  query access, forced CMK for query artifacts
- **Cost guards** -- daily ingestion quota, retention 30-730 days, immediate
  post-retention purge for right-to-erasure compliance
- **Managed identity** -- system- or user-assigned (workspaces support exactly one
  model at a time), for dedicated-cluster CMK and linked-storage scenarios
- **Governance tags** -- user tags merged over the Planton-derived metadata tags

## When to Use

- As the monitoring foundation in any Azure Infra Chart -- typically one workspace
  per environment or region that everything else feeds into
- Before deploying `AzureApplicationInsights` (workspace-based APM is the only mode)
- Before wiring `AzureMonitorDiagnosticSetting` or `AzureMonitorScheduledQueryAlert`
- When centralizing logs for Microsoft Sentinel

## Spec Highlights

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `region` | string | Yes | - | Azure region |
| `resource_group` | StringValueOrRef | Yes | - | Resource group (literal or valueFrom) |
| `workspace_name` | string | Yes | - | Workspace name (4-63 letters/digits/hyphens; ForceNew) |
| `sku` | enum | No | PER_GB_2018 | PER_GB_2018 / CAPACITY_RESERVATION / PER_NODE / STANDALONE |
| `reservation_capacity_in_gb_per_day` | int32 | With CAPACITY_RESERVATION | - | Commitment tier (100...50000) |
| `retention_in_days` | int32 | No | 30 | Workspace-level retention (30-730) |
| `daily_quota_gb` | double | No | -1 | Daily ingestion cap (-1 = unlimited) |
| `identity` | message | No | - | SYSTEM_ASSIGNED or USER_ASSIGNED managed identity |
| `local_authentication_enabled` | bool | No | true | Shared keys usable in addition to Entra |
| `internet_ingestion_enabled` / `internet_query_enabled` | bool | No | true | Public paths (false = AMPLS-only) |
| `allow_resource_only_permissions` | bool | No | true | Resource-context query access |
| `cmk_for_query_forced` | bool | No | false | Customer-managed storage for query artifacts |
| `immediate_data_purge_on_30_days_enabled` | bool | No | false | No post-retention grace store |
| `data_collection_rule_id` | StringValueOrRef | No | - | Default DCR (ARM id or AzureMonitorDataCollectionRule reference) |
| `tags` | map | No | - | User tags (merged over metadata tags) |

## Outputs

| Output | Description |
|--------|-------------|
| `workspace_id` | ARM resource ID -- the FK seam App Insights, AKS, Container Apps, diagnostic settings, and query alerts reference |
| `workspace_name` | Workspace name |
| `workspace_customer_id` | The customer GUID agents authenticate against (the portal's "Workspace ID") |
| `resource_group_name` | Containing resource group |
| `primary_shared_key` / `secondary_shared_key` | Agent authentication keys (secret-bearing; unusable when local auth is disabled) |
| `identity_principal_id` | System-assigned identity principal (empty unless enabled) |

## Composition

Reference the workspace from downstream kinds:

```yaml
workspaceId:
  valueFrom:
    kind: AzureLogAnalyticsWorkspace
    name: my-platform-logs
    fieldPath: status.outputs.workspace_id
```

See `presets/` for pay-as-you-go, commitment-tier, and private-hardened starting points.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
