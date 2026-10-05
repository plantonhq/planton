# Azure Monitor Data Collection Rule

Deploys an Azure Monitor data collection rule (DCR) -- the routing table declaring what telemetry the Azure Monitor Agent collects, where it lands, and how the two wire together. A rule declares data sources (Linux syslog, Windows event logs, performance counters, Prometheus metrics, IIS logs, custom log files), destinations (Log Analytics workspaces, Azure Monitor metrics, Event Hubs, storage), and the data flows connecting them -- optionally reshaping records with a KQL transformation in the ingestion pipeline, before they bill. The rule alone collects nothing: machines attach through Azure Monitor Data Collection Rule Association resources, and one rule serves any number of machines.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Data collection rule** -- the collection policy with its data sources, destinations, data flows, custom-stream declarations, optional platform kind, and (optionally) a managed identity, ingesting through a Data Collection Endpoint when one is referenced
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A resource group** -- the rule lives in a referenced resource group.
- **At least one destination** -- typically an AzureLogAnalyticsWorkspace (reference its `workspace_id` output); Event Hub and storage destinations reference their own kinds' outputs.

### Azure Subscription

- **The rule alone collects nothing** -- machines must be attached with AzureMonitorDataCollectionRuleAssociation resources, and the Azure Monitor Agent must run on them.
- **Custom streams need a Data Collection Endpoint** -- stream declarations (custom log files, direct ingestion) require `dataCollectionEndpointId`; provide the literal ARM ID of a DCE (it is not yet a catalog kind).
- **Platform compatibility is enforced by Azure at deploy time** -- a `Linux` rule cannot carry Windows event logs, a `Windows` rule cannot carry syslog, and the `*Direct` destinations require kind `AgentDirectToStore`; the provider performs no early check.
- **The rule object itself is free** -- you pay for the telemetry it lands (workspace ingestion, storage, Event Hub throughput).

## Deploy

### Console

Open the deployment store, find **Azure Monitor Data Collection Rule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the data sources, the destinations, and the flows wiring them. Start from the **Linux Syslog and Performance to Workspace** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorDataCollectionRule
metadata:
  name: linux-baseline
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: prod-observability
  name: linux-baseline
  region: eastus
  dataSources:
    syslogs:
      - name: linux-syslog
        facilityNames:
          - auth
          - authpriv
          - daemon
        logLevels:
          - Warning
          - Error
          - Critical
          - Alert
          - Emergency
        streams:
          - Microsoft-Syslog
    performanceCounters:
      - name: linux-perf
        samplingFrequencyInSeconds: 60
        counterSpecifiers:
          - Processor(*)\% Processor Time
          - Memory(*)\% Used Memory
        streams:
          - Microsoft-Perf
  destinations:
    logAnalytics:
      - name: ops-workspace
        workspaceResourceId:
          value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-observability/providers/Microsoft.OperationalInsights/workspaces/acme-ops-logs
  dataFlows:
    - streams:
        - Microsoft-Syslog
        - Microsoft-Perf
      destinations:
        - ops-workspace
```

```shell
planton apply -f data-collection-rule.yaml
```

This creates a reusable Linux baseline policy -- security-relevant syslog (filtered facilities and severities, not `*`) plus a once-a-minute CPU/memory baseline, landing in one Log Analytics workspace; nothing is collected until machines associate with the rule. An Infra Job tracks the provisioning in real time.

### InfraChart

When the rule's dependencies deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-observability
      fieldPath: status.outputs.resource_group_name
  name: linux-baseline
  region: eastus
  destinations:
    logAnalytics:
      - name: ops-workspace
        workspaceResourceId:
          valueFrom:
            kind: AzureLogAnalyticsWorkspace
            name: acme-ops-logs
            fieldPath: status.outputs.workspace_id
  dataFlows:
    - streams:
        - Microsoft-Syslog
      destinations:
        - ops-workspace
```

The InfraPipeline resolves the dependency graph, deploys the resource group and the workspace first, then the rule -- and Azure Monitor Data Collection Rule Association resources downstream can reference this rule's `data_collection_rule_id` in the same chart.

## Key Configuration

These are the most important decisions when configuring a data collection rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Design rules around workloads, not machines** -- a rule is a reusable collection policy ("Linux web tier: auth syslog + CPU/memory counters to the ops workspace") and machines attach by association. Resist the one-rule-per-machine anti-pattern: it multiplies objects without adding control. A fleet-wide baseline rule plus small workload-specific rules composes better than one giant rule every team edits -- a machine can carry several associations when it genuinely needs several policies.

**Filter at the rule, not at the workspace** -- everything a flow lands in Log Analytics bills per GB at ingestion. The rule gives three progressively stronger filters that all run BEFORE billing: XPath queries on Windows event logs, facility/severity selection on syslog (auth and daemon, not `*`), and `transformKql` on the flow (drop columns, filter rows, reshape). A `"*"` facility list with an unfiltered flow is the classic surprise workspace bill.

**Names wire the rule together, and the tokens must match exactly** -- every data source and destination carries a rule-local `name`; destination names share ONE namespace across all arms (Azure rejects duplicates at deploy time), flows reference destinations by exactly those names, and a flow's `streams` must match the sources' `streams` token-for-token or the records silently go nowhere. When nothing lands, walk the chain in order: association exists, agent runs on the machine, stream names match, flow destinations name a real destination -- and give a healthy chain 3-5 minutes before debugging.

**The 60-second rule for InsightsMetrics** -- performance counters flow to two places with different contracts: `Microsoft-Perf` (the workspace's Perf table, any 1-1800s sampling) and `Microsoft-InsightsMetrics` (Azure Monitor metrics via the `azureMonitorMetrics` destination, sampling EXACTLY 60s -- Azure rejects anything else at deploy time). When a rule serves both, use two performance-counter sources with their own sampling rather than compromising one.

**Custom logs: the schema is a contract** -- a stream declaration's columns become the custom table's schema. Include a `TimeGenerated` datetime column (or produce one in `transformKql`) or the workspace stamps arrival time, which skews every time-based query. Changing a declared schema later means coordinating the rule, any transformation, and queries against the table -- version custom stream names (`Custom-MyAppLogs2`) rather than mutating a live schema. Custom streams also force the DCE decision: they only ingest through a Data Collection Endpoint.

**`kind` is a one-way door once set** -- a rule created WITHOUT a kind accepts every platform's sources and can adopt a kind later in place, but once set, ANY change to it (including clearing) replaces the rule, and every association on the old rule dies with it. Leave `kind` unset unless you need what only a kind unlocks -- the `*Direct` destinations require `AgentDirectToStore`.

**The identity flavor decides when grants can happen** -- the provider supports exactly one flavor at a time: `SYSTEM_ASSIGNED` is created with the rule (grant its `identity_principal_id` output on destinations AFTER the rule exists), while `USER_ASSIGNED` brings an AzureUserAssignedIdentity whose permissions can be composed BEFORE the rule is created -- the right choice when the deploy must work on the first pass. Azure accepts at most one user-assigned identity on a rule. Omit the block when no destination needs the rule to authenticate as itself.

**Co-locate the rule with its primary workspace** -- machines in any region can associate with the rule, but the rule performs its ingestion processing in ITS region: placing it next to the main destination workspace avoids cross-region processing latency, and Azure rejects cross-cloud bindings outright.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureLogAnalyticsWorkspace** | `destinations.logAnalytics[].workspaceResourceId` | `status.outputs.workspace_id` |
| **AzureEventHub** (Event Hub destinations) | `destinations.eventHub.eventHubId`, `destinations.eventHubDirect.eventHubId` | `status.outputs.event_hub_id` |
| **AzureStorageAccount** (storage destinations) | `destinations.storageBlobs[].storageAccountId` (and the direct arms) | `status.outputs.storage_account_id` |
| **AzureUserAssignedIdentity** (USER_ASSIGNED identity) | `identity.identityIds` | `status.outputs.identity_id` |

`dataCollectionEndpointId` and `monitorAccountId` carry no default kind -- Data Collection Endpoints and Azure Monitor workspaces are not yet catalog kinds; pass their literal ARM IDs.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `data_collection_rule_id` | The rule's ARM resource ID | An Azure Monitor Data Collection Rule Association's `dataCollectionRuleId`; an Azure Log Analytics Workspace's default `dataCollectionRuleId` |
| `immutable_id` | The rule's immutable ID -- the identifier agents and the ingestion API address the rule by, stable across moves | Direct-ingestion clients posting custom streams through a DCE |
| `identity_principal_id` | The system-assigned identity's principal ID (empty when no system identity is configured) | Role grants giving the rule access on its destinations |

`data_collection_rule_name` is also exported but only echoes the manifest's `name` back.

## Common Patterns

**Linux fleet baseline** -- filtered syslog (auth, authpriv, daemon at Warning and above) plus CPU/memory/disk counters into one ops workspace; the rule every Linux machine associates with, and the shape where rule-level filtering pays for itself. Start from the **Linux Syslog and Performance to Workspace** preset.

**Windows fleet baseline** -- the same policy for Windows: XPath-filtered System and Application event logs (Critical/Error/Warning, never Information) plus performance counters to the workspace. Start from the **Windows Events and Performance to Workspace** preset.

**Custom application logs** -- a JSON or text log file gets a declared schema, an ingestion-time KQL filter, and lands in a custom table; requires a Data Collection Endpoint, and the declaration's columns are a versioned contract. Start from the **Custom JSON Log to Workspace** preset.

**Split metric destinations** -- one rule sending `Microsoft-Perf` to the workspace for KQL analysis and `Microsoft-InsightsMetrics` (its own 60-second source) to Azure Monitor metrics for near-real-time charts and metric alerts -- two flows, two destinations, one policy.

## Works With

- [**Azure Monitor Data Collection Rule Association**](/infra-catalog/azure-monitor-data-collection-rule-association) -- attaches each machine to this rule; collection starts only through associations
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- the usual log destination, referenced by `workspace_id`; a workspace can also name this rule as its default DCR
- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- streaming destination for telemetry leaving Azure Monitor, referenced by `event_hub_id`
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- blob and table destinations for archival, referenced by `storage_account_id`
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the rule's user-assigned identity, grantable on destinations before the rule exists
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the rule lives in
