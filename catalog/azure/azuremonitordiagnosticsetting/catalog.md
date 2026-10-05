# Azure Monitor Diagnostic Setting

Deploys an Azure Monitor diagnostic setting -- how a resource's platform telemetry LEAVES the resource. It selects which log categories and metrics the target emits and routes them to destinations: a Log Analytics Workspace (queryable with KQL, alertable with scheduled query rules), a Storage Account (cheap archival), an Event Hub (streaming to SIEMs), or an Azure Native partner solution. Without a diagnostic setting, most Azure resources emit nothing beyond basic platform metrics -- this kind closes the logging pipeline that scheduled query alerts watch.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Diagnostic Setting** -- a `Microsoft.Insights/diagnosticSettings` EXTENSION resource living ON the target (any ARM resource -- a Key Vault, an AKS cluster, a gateway, a subscription), carrying the category selections and destination routing. A target can carry up to five settings

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A target resource** that emits diagnostics -- reference its `*_id` output or paste its ARM ID. Category names are defined per resource type: `az monitor diagnostic-settings categories list --resource <id>` lists them.
- **At least one destination** -- a Log Analytics Workspace, a Storage Account, an Event Hub namespace authorization rule (with send rights), or a partner solution.

## Deploy

### Console

Open the deployment store, find **Azure Monitor Diagnostic Setting**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **All Logs and Metrics to a Workspace** preset in the [Presets](#presets) tab to pre-populate the everyday observability wiring.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorDiagnosticSetting
metadata:
  name: vault-diagnostics
  org: acme-corp
  env: prod
spec:
  settingName: route-to-workspace
  targetResourceId:
    valueFrom:
      kind: AzureKeyVault
      name: my-app-vault
      fieldPath: status.outputs.key_vault_id
  enabledLogs:
    - categoryGroup: allLogs
  enabledMetrics:
    - category: AllMetrics
  logAnalyticsWorkspaceId:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-logs
      fieldPath: status.outputs.workspace_id
  logAnalyticsDestinationType: DEDICATED
```

```shell
planton apply -f diagnostic-setting.yaml
```

This routes every current AND future log category of the vault (the allLogs group tracks new ones automatically) plus its metrics into the workspace, landing in modern resource-specific tables. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the setting between the resource it drains and the workspace it fills:

```yaml
spec:
  targetResourceId:
    valueFrom:
      kind: AzureKeyVault
      name: my-app-vault
      fieldPath: status.outputs.key_vault_id
  logAnalyticsWorkspaceId:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-logs
      fieldPath: status.outputs.workspace_id
```

The InfraPipeline resolves the dependency graph and deploys the target and the workspace before the setting.

## Key Configuration

These are the most important decisions when configuring a diagnostic setting. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Telemetry selection** -- each `enabledLogs` entry names exactly one of a single `category` (AuditEvent on a vault, kube-audit on AKS) or a `categoryGroup` (Azure's curated allLogs/audit bundles, which track new categories automatically). At least one log or metric selection is required -- Azure rejects a setting that routes nothing. Volume note: allLogs on a chatty resource can dominate workspace ingestion costs.

**Destinations** -- at least one of: `logAnalyticsWorkspaceId` (queryable + alertable -- pair with `logAnalyticsDestinationType: DEDICATED` for modern typed tables), `storageAccountId` (cheap long-term archival), the Event Hub pair (`eventhubAuthorizationRuleId` -- a NAMESPACE-scoped send rule -- with the optional `eventhubName` picking one hub), or `partnerSolutionId` (Elastic/Datadog via Marketplace). One setting can feed several at once.

**The extension model** -- `settingName` and `targetResourceId` are fixed at creation (the setting lives ON the target); selections and destinations update in place -- re-routing telemetry is the designed operation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **Any diagnostics-emitting kind** | `targetResourceId` | the target's `*_id` output (explicit valueFrom) |
| **AzureLogAnalyticsWorkspace** | `logAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| **AzureStorageAccount** | `storageAccountId` | `status.outputs.storage_account_id` |
| **AzureEventHubAuthorizationRule** | `eventhubAuthorizationRuleId` | `status.outputs.authorization_rule_id` |
| **AzureEventHub** | `eventhubName` | `status.outputs.event_hub_name` |

### What This Kind Provides

The setting is a leaf in the dependency graph: `status.outputs` carries only its own identifiers (`diagnostic_setting_id`, `diagnostic_setting_name`) and the as-deployed `target_resource_id` as a routing audit trail -- no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Logs to workspace** -- allLogs + AllMetrics into the platform workspace: the everyday observability wiring, and what makes scheduled query alerts see the resource. Start from the **All Logs and Metrics to a Workspace** preset.

**Archive to storage** -- audit categories into blob storage for long-horizon compliance retention. Start from the **Audit Trail to Storage Archival** preset.

**Stream to SIEM** -- categories into an Event Hub an external pipeline consumes. Start from the **Security Stream to an External SIEM** preset.

## Works With

- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- the queryable, alertable destination
- [**Azure Monitor Scheduled Query Alert**](/infra-catalog/azure-monitor-scheduled-query-alert) -- watches the workspace this setting fills
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- the archival destination
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the classic audit-log target
