# Azure Monitor Scheduled Query Alert

Deploys an Azure Monitor scheduled query alert rule -- the log-search alert. It runs a KQL query against a Log Analytics Workspace (or an Application Insights resource) on a schedule, compares the result to a threshold, and fires action groups when the condition holds. It is the alerting half of the logging pipeline: diagnostic settings route logs INTO the workspace, and query alerts watch what arrives -- error spikes, security events, missing heartbeats, business anomalies, anything KQL can express.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Scheduled Query Rule** -- a `Microsoft.Insights/scheduledQueryRules` resource (REGIONAL -- it must live in the queried workspace's region) carrying the KQL conditions, cadence, severity, noise dials, optional managed identity, and action wiring
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the rule will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A Log Analytics Workspace** (or Application Insights resource) for the query to run against -- in the SAME region as the rule. Reference the workspace's `workspace_id` output.
- **Logs flowing into the workspace** -- a query alert on an empty workspace never fires; diagnostic settings are how resource logs arrive.
- **An action group** to notify (optional but recommended).

## Deploy

### Console

Open the deployment store, find **Azure Monitor Scheduled Query Alert**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Error Spike Alert (Row Count)** preset in the [Presets](#presets) tab to pre-populate the classic log-search pager.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorScheduledQueryAlert
metadata:
  name: error-spike-alert
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "observability-rg"
  alertName: checkout-api-error-spike
  scope:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-logs
      fieldPath: status.outputs.workspace_id
  severity: 1
  criteria:
    - query: |
        AppExceptions
        | where TimeGenerated > ago(10m)
        | where AppRoleName == "checkout-api"
      timeAggregationMethod: COUNT
      operator: GREATER_THAN
      threshold: 5
  action:
    actionGroupIds:
      - valueFrom:
          kind: AzureMonitorActionGroup
          name: platform-oncall
          fieldPath: status.outputs.action_group_id
```

```shell
planton apply -f query-alert.yaml
```

This creates a rule on the platform defaults: the query runs every 5 minutes over the last 5 minutes; each firing is its own alert (Azure's default for query rules). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the rule to the workspace it queries and the action group it fires:

```yaml
spec:
  scope:
    valueFrom:
      kind: AzureLogAnalyticsWorkspace
      name: platform-logs
      fieldPath: status.outputs.workspace_id
  action:
    actionGroupIds:
      - valueFrom:
          kind: AzureMonitorActionGroup
          name: platform-oncall
          fieldPath: status.outputs.action_group_id
```

The InfraPipeline resolves the dependency graph, deploys the workspace and group first, then provisions the rule with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a scheduled query alert. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Conditions** -- Each `criteria` entry (at least one) runs its own KQL query; Azure evaluates them independently and the rule fires when ANY holds. Two styles: `COUNT` compares the number of result rows; `AVERAGE`/`MINIMUM`/`MAXIMUM`/`TOTAL` compute over the numeric column named in `metricMeasureColumn` (required for those styles -- the query must project it). The optional `failingPeriods` pair is the flap damper. A missing-heartbeat watch inverts the comparison: COUNT LESS THAN 1 fires when expected rows stop arriving.

**Cadence and cost** -- `evaluationFrequency` (default PT5M) is a cost dial: every evaluation runs a billed query. `windowDuration` is how far back each evaluation queries; `queryTimeRangeOverride` extends the lookback for comparisons needing more history than the window.

**Noise dials, pick one** -- `autoMitigationEnabled` makes the alert stateful (fire once, self-resolve); `muteActionsAfterAlertDuration` keeps every firing but suppresses repeat notifications. They are mutually exclusive -- Azure rejects the pair.

**Query identity** -- optional managed identity (system-assigned or user-assigned; this kind has no combined model) for workspaces enforcing Entra-only query access. A system-assigned principal exists only after deploy -- grant it workspace access as a second step; user-assigned identities can be granted BEFORE the rule exists.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureLogAnalyticsWorkspace** | `scope` | `status.outputs.workspace_id` |
| **AzureUserAssignedIdentity** | `identity.userAssignedIdentityIds[]` | `status.outputs.identity_id` |
| **AzureMonitorActionGroup** | `action.actionGroupIds[]` | `status.outputs.action_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `identity_principal_id` | The system-assigned principal ID (when the identity is configured) | The workspace-access grant target -- the second step a system-assigned query identity requires |

The rule's own identifiers (`scheduled_query_alert_id`, `scheduled_query_alert_name`) are also exported for alert-history filtering and CLI reference; no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Error spike** -- COUNT of exception rows greater than a threshold over a rolling window. Start from the **Error Spike Alert (Row Count)** preset.

**Latency threshold** -- AVERAGE of a duration column above a limit -- the column-style evaluation. Start from the **Latency Threshold Alert (Metric Measurement, Per-Dimension)** preset.

**Missing heartbeat** -- COUNT LESS THAN 1: fires when expected rows STOP arriving — the absence alarm no metric can express. Start from the **Missing Heartbeat Alert (Absence of Data)** preset.

## Works With

- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- the workspace the query runs against
- [**Azure Monitor Diagnostic Setting**](/infra-catalog/azure-monitor-diagnostic-setting) -- routes resource logs INTO the workspace this rule watches
- [**Azure Monitor Action Group**](/infra-catalog/azure-monitor-action-group) -- the notification hub the rule fires into
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the pre-grantable query identity
