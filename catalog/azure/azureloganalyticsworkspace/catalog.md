# Azure Log Analytics Workspace

Deploys an Azure Log Analytics Workspace -- the central data platform for Azure Monitor. Workspaces collect, store, and query log and performance data, and they are the foundation AKS Container Insights, Application Insights, Microsoft Sentinel, diagnostic settings, and log-query alerts all build on. The kind covers the full workspace surface: pricing tier and commitment capacity, retention and daily quota, the authentication and network-access posture, the query access model, compliance switches, a managed identity, and a default Data Collection Rule.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Log Analytics Workspace** -- an `operationalinsights.AnalyticsWorkspace` in the specified Azure region and resource group, configured with the chosen pricing SKU (and commitment capacity when applicable), retention period, daily ingestion quota, access posture, and compliance switches
- **Managed Identity binding** (optional) -- a system-assigned or user-assigned identity on the workspace, for dedicated-cluster customer-managed keys and linked-storage access
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the workspace will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Log Analytics Workspace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Pay-As-You-Go Workspace** preset in the [Presets](#presets) tab to pre-populate the everyday configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureLogAnalyticsWorkspace
metadata:
  name: platform-logs
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  workspaceName: acme-log-analytics
```

```shell
planton apply -f log-analytics.yaml
```

This creates a Log Analytics Workspace on Azure's defaults: pay-as-you-go pricing (PerGB2018), 30-day retention, unlimited daily ingestion, and public ingestion/query endpoints. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the workspace to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the Log Analytics Workspace with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Log Analytics Workspace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pricing SKU** -- The `sku` field controls the billing model. Unspecified (or `PER_GB_2018`) is pay-as-you-go per GB ingested and suits nearly every workspace. Switch to `CAPACITY_RESERVATION` only at sustained 100+ GB/day -- it requires `reservationCapacityInGbPerDay` (Azure sells fixed tiers: 100 through 50000 GB/day) and carries a 31-day commitment. `PER_NODE` and `STANDALONE` are legacy pre-2018 models kept for existing estates. Pay-as-you-go and the commitment tier convert in place; any change crossing a legacy tier replaces the workspace and destroys its data.

**Data retention** -- The `retentionInDays` field controls how long logs are queryable (30-730 days; default 30). PerGB2018 includes 31 days free; beyond that, retention is billed per GB per month. Compliance workloads typically require 90-365 days. Per-table retention overrides are managed in Azure directly.

**Daily ingestion quota** -- The `dailyQuotaGb` field caps daily ingestion. `-1` (the default) means unlimited. A positive cap contains cost overruns, but it is also a data-loss dial: when the cap is reached, ingestion stops until the next UTC day.

**Access posture** -- Four optional booleans, all defaulting to Azure's open posture: `localAuthenticationEnabled` (false = Entra-only; the shared-key outputs become inert), `internetIngestionEnabled` and `internetQueryEnabled` (false = the corresponding path requires Azure Monitor Private Link Scope private endpoints), and `allowResourceOnlyPermissions` (false = every query needs workspace-level permissions).

**Managed identity** -- The optional `identity` block (`SYSTEM_ASSIGNED` or `USER_ASSIGNED` -- workspaces support exactly one model at a time, never both) serves the workspace's own access to other resources: reading a customer-managed key on a dedicated Log Analytics cluster, or querying linked storage.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** | `identity.userAssignedIdentityIds` | `status.outputs.identity_id` |
| **AzureMonitorDataCollectionRule** | `dataCollectionRuleId` | `status.outputs.data_collection_rule_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `workspace_id` | Azure Resource Manager ID of the workspace | Application Insights workspace binding, Container Apps environment logging, AKS Container Insights and Defender, diagnostic settings |
| `workspace_name` | Name of the Log Analytics Workspace | Azure CLI references, portal navigation |
| `workspace_customer_id` | The workspace GUID agents authenticate against (the portal calls it "Workspace ID" on the agents page) | Agent onboarding, direct ingestion APIs |
| `resource_group_name` | The resource group the workspace landed in | Co-locating linked resources |
| `primary_shared_key` | Primary authentication key for agent ingestion (inert when local authentication is disabled) | Log Analytics agent configuration, direct ingestion APIs |
| `secondary_shared_key` | Secondary authentication key enabling zero-downtime rotation | Backup key during key rotation |
| `identity_principal_id` | Object ID of the system-assigned identity, when one exists | Key Vault and storage role assignments for the workspace's own access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Pay-as-you-go** -- PerGB2018 pricing with 90-day retention and a daily cap as a cost guard. The everyday workspace nearly every environment starts from. Start from the **Pay-As-You-Go Workspace** preset.

**Commitment tier** -- `CAPACITY_RESERVATION` with a reserved daily capacity, for estates sustaining 100+ GB/day where commitment tiers discount ingestion. Start from the **Commitment-Tier Workspace** preset.

**Private hardened** -- Entra-only authentication with both public endpoints disabled and immediate post-retention purge, for regulated estates running Azure Monitor Private Link Scope. Start from the **Private, Hardened Workspace** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the workspace is created
- [**Azure Application Insights**](/infra-catalog/azure-application-insights) -- workspace-based Application Insights stores its telemetry here via `workspace_id`
- [**Azure Container App Environment**](/infra-catalog/azure-container-app-environment) -- persists application logs here when its logs destination is Log Analytics
- [**Azure AKS Cluster**](/infra-catalog/azure-aks-cluster) -- Container Insights and Microsoft Defender stream cluster telemetry here
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- supplies the workspace's user-assigned identity for CMK and linked-storage access
- [**Azure Monitor Data Collection Rule**](/infra-catalog/azure-monitor-data-collection-rule) -- referenced as the workspace's default DCR for data arriving without an explicit rule
- [**Azure Monitor Scheduled Query Alert**](/infra-catalog/azure-monitor-scheduled-query-alert) -- runs KQL alert rules against the logs stored here
