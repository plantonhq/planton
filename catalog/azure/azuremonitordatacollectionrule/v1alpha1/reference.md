# AzureMonitorDataCollectionRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureMonitorDataCollectionRuleSpec** defines an Azure Monitor data
collection rule (DCR) -- the routing table of Azure Monitor. A rule
declares WHAT telemetry to collect (data sources: Linux syslog,
Windows event logs, performance counters, Prometheus metrics, IIS
logs, custom log files, ...), WHERE it goes (destinations: Log
Analytics workspaces, Azure Monitor metrics, Event Hubs, storage),
and HOW the two wire together (data flows, optionally with a KQL
transformation applied in the ingestion pipeline).

A rule does nothing on its own: machines are attached to it with
AzureMonitorDataCollectionRuleAssociation resources (one association
per machine), and the Azure Monitor Agent on each associated machine
downloads the rule and starts collecting. One rule serves any number
of machines; one machine can carry many associations.

**Names wire the rule together**: every data source and every
destination carries a `name` that is local to this rule; data flows
reference destinations by exactly those names, and Azure rejects a
deploy whose flow references a name that no destination carries.
Destination names must be unique across ALL destination arms (they
share one namespace -- Azure enforces this at deploy time).

## Example

```yaml
# Deep-shape example for docs and offline validation: a Linux rule
# collecting filtered syslog and performance counters, landing logs in
# a Log Analytics workspace and metrics in Azure Monitor metrics --
# with an ingestion-time KQL filter on the syslog flow. References are
# literal ARM ids so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorDataCollectionRule
metadata:
  name: test-dcr
  id: test-dcr
  org: test-org
  env: test
spec:
  resourceGroup:
    value: test-rg
  name: linux-baseline
  region: eastus
  kind: Linux
  description: Linux fleet baseline - filtered syslog and perf counters
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
      - name: linux-insights-metrics
        # Streams targeting Microsoft-InsightsMetrics require EXACTLY
        # 60-second sampling (Azure enforces at deploy time).
        samplingFrequencyInSeconds: 60
        counterSpecifiers:
          - Processor(*)\% Processor Time
        streams:
          - Microsoft-InsightsMetrics
  destinations:
    logAnalytics:
      - name: ops-workspace
        workspaceResourceId:
          value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.OperationalInsights/workspaces/test-law
    azureMonitorMetrics:
      name: azure-metrics
  dataFlows:
    - streams:
        - Microsoft-Syslog
      destinations:
        - ops-workspace
      # Drop noisy repeated daemon chatter at ingestion, before it bills.
      transformKql: source | where SyslogMessage !has 'systemd' or SeverityLevel != 'warning'
      outputStream: Microsoft-Syslog
    - streams:
        - Microsoft-Perf
      destinations:
        - ops-workspace
    - streams:
        - Microsoft-InsightsMetrics
      destinations:
        - azure-metrics
  tags:
    cost-center: platform-observability
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.kind` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.dataCollectionEndpointId` | `string \| valueFrom` |  |  |  |
| `spec.identity` | `AzureMonitorDataCollectionRuleIdentity` |  |  |  |
| `spec.identity.type` | `enum` | yes |  |  |
| `spec.identity.identityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.dataSources` | `AzureMonitorDataCollectionRuleDataSources` |  |  |  |
| `spec.dataSources.syslogs` | `[]AzureMonitorDataCollectionRuleSyslog` |  |  |  |
| `spec.dataSources.syslogs[].name` | `string` | yes |  |  |
| `spec.dataSources.syslogs[].facilityNames` | `[]string` | yes |  |  |
| `spec.dataSources.syslogs[].logLevels` | `[]string` | yes |  |  |
| `spec.dataSources.syslogs[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.performanceCounters` | `[]AzureMonitorDataCollectionRulePerformanceCounter` |  |  |  |
| `spec.dataSources.performanceCounters[].name` | `string` | yes |  |  |
| `spec.dataSources.performanceCounters[].samplingFrequencyInSeconds` | `int32` | yes |  |  |
| `spec.dataSources.performanceCounters[].counterSpecifiers` | `[]string` | yes |  |  |
| `spec.dataSources.performanceCounters[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.windowsEventLogs` | `[]AzureMonitorDataCollectionRuleWindowsEventLog` |  |  |  |
| `spec.dataSources.windowsEventLogs[].name` | `string` | yes |  |  |
| `spec.dataSources.windowsEventLogs[].xPathQueries` | `[]string` | yes |  |  |
| `spec.dataSources.windowsEventLogs[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.extensions` | `[]AzureMonitorDataCollectionRuleExtension` |  |  |  |
| `spec.dataSources.extensions[].name` | `string` | yes |  |  |
| `spec.dataSources.extensions[].extensionName` | `string` | yes |  |  |
| `spec.dataSources.extensions[].extensionJson` | `string` |  |  |  |
| `spec.dataSources.extensions[].inputDataSources` | `[]string` |  |  |  |
| `spec.dataSources.extensions[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.iisLogs` | `[]AzureMonitorDataCollectionRuleIisLog` |  |  |  |
| `spec.dataSources.iisLogs[].name` | `string` | yes |  |  |
| `spec.dataSources.iisLogs[].logDirectories` | `[]string` |  |  |  |
| `spec.dataSources.iisLogs[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.logFiles` | `[]AzureMonitorDataCollectionRuleLogFile` |  |  |  |
| `spec.dataSources.logFiles[].name` | `string` | yes |  |  |
| `spec.dataSources.logFiles[].filePatterns` | `[]string` | yes |  |  |
| `spec.dataSources.logFiles[].format` | `string` | yes |  |  |
| `spec.dataSources.logFiles[].settings` | `AzureMonitorDataCollectionRuleLogFileSettings` |  |  |  |
| `spec.dataSources.logFiles[].settings.text` | `AzureMonitorDataCollectionRuleLogFileSettingsText` | yes |  |  |
| `spec.dataSources.logFiles[].settings.text.recordStartTimestampFormat` | `string` | yes |  |  |
| `spec.dataSources.logFiles[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.prometheusForwarders` | `[]AzureMonitorDataCollectionRulePrometheusForwarder` |  |  |  |
| `spec.dataSources.prometheusForwarders[].name` | `string` | yes |  |  |
| `spec.dataSources.prometheusForwarders[].labelIncludeFilters` | `[]AzureMonitorDataCollectionRuleLabelIncludeFilter` |  |  |  |
| `spec.dataSources.prometheusForwarders[].labelIncludeFilters[].label` | `string` | yes |  |  |
| `spec.dataSources.prometheusForwarders[].labelIncludeFilters[].value` | `string` | yes |  |  |
| `spec.dataSources.prometheusForwarders[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.windowsFirewallLogs` | `[]AzureMonitorDataCollectionRuleWindowsFirewallLog` |  |  |  |
| `spec.dataSources.windowsFirewallLogs[].name` | `string` | yes |  |  |
| `spec.dataSources.windowsFirewallLogs[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.platformTelemetries` | `[]AzureMonitorDataCollectionRulePlatformTelemetry` |  |  |  |
| `spec.dataSources.platformTelemetries[].name` | `string` | yes |  |  |
| `spec.dataSources.platformTelemetries[].streams` | `[]string` | yes |  |  |
| `spec.dataSources.dataImport` | `AzureMonitorDataCollectionRuleDataImport` |  |  |  |
| `spec.dataSources.dataImport.eventHubDataSource` | `AzureMonitorDataCollectionRuleDataImportEventHub` | yes |  |  |
| `spec.dataSources.dataImport.eventHubDataSource.name` | `string` | yes |  |  |
| `spec.dataSources.dataImport.eventHubDataSource.stream` | `string` | yes |  |  |
| `spec.dataSources.dataImport.eventHubDataSource.consumerGroup` | `string` |  |  |  |
| `spec.destinations` | `AzureMonitorDataCollectionRuleDestinations` | yes |  |  |
| `spec.destinations.logAnalytics` | `[]AzureMonitorDataCollectionRuleLogAnalytics` |  |  |  |
| `spec.destinations.logAnalytics[].name` | `string` | yes |  |  |
| `spec.destinations.logAnalytics[].workspaceResourceId` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`) |
| `spec.destinations.azureMonitorMetrics` | `AzureMonitorDataCollectionRuleAzureMonitorMetrics` |  |  |  |
| `spec.destinations.azureMonitorMetrics.name` | `string` | yes |  |  |
| `spec.destinations.eventHub` | `AzureMonitorDataCollectionRuleEventHubDestination` |  |  |  |
| `spec.destinations.eventHub.name` | `string` | yes |  |  |
| `spec.destinations.eventHub.eventHubId` | `string \| valueFrom` | yes |  | AzureEventHub (`status.outputs.event_hub_id`) |
| `spec.destinations.eventHubDirect` | `AzureMonitorDataCollectionRuleEventHubDestination` |  |  |  |
| `spec.destinations.eventHubDirect.name` | `string` | yes |  |  |
| `spec.destinations.eventHubDirect.eventHubId` | `string \| valueFrom` | yes |  | AzureEventHub (`status.outputs.event_hub_id`) |
| `spec.destinations.monitorAccounts` | `[]AzureMonitorDataCollectionRuleMonitorAccount` |  |  |  |
| `spec.destinations.monitorAccounts[].name` | `string` | yes |  |  |
| `spec.destinations.monitorAccounts[].monitorAccountId` | `string \| valueFrom` | yes |  |  |
| `spec.destinations.storageBlobs` | `[]AzureMonitorDataCollectionRuleStorageBlobDestination` |  |  |  |
| `spec.destinations.storageBlobs[].name` | `string` | yes |  |  |
| `spec.destinations.storageBlobs[].containerName` | `string` | yes |  |  |
| `spec.destinations.storageBlobs[].storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.destinations.storageBlobDirects` | `[]AzureMonitorDataCollectionRuleStorageBlobDestination` |  |  |  |
| `spec.destinations.storageBlobDirects[].name` | `string` | yes |  |  |
| `spec.destinations.storageBlobDirects[].containerName` | `string` | yes |  |  |
| `spec.destinations.storageBlobDirects[].storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.destinations.storageTableDirects` | `[]AzureMonitorDataCollectionRuleStorageTableDirect` |  |  |  |
| `spec.destinations.storageTableDirects[].name` | `string` | yes |  |  |
| `spec.destinations.storageTableDirects[].tableName` | `string` | yes |  |  |
| `spec.destinations.storageTableDirects[].storageAccountId` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.dataFlows` | `[]AzureMonitorDataCollectionRuleDataFlow` | yes |  |  |
| `spec.dataFlows[].streams` | `[]string` | yes |  |  |
| `spec.dataFlows[].destinations` | `[]string` | yes |  |  |
| `spec.dataFlows[].builtInTransform` | `string` |  |  |  |
| `spec.dataFlows[].outputStream` | `string` |  |  |  |
| `spec.dataFlows[].transformKql` | `string` |  |  |  |
| `spec.streamDeclarations` | `[]AzureMonitorDataCollectionRuleStreamDeclaration` |  |  |  |
| `spec.streamDeclarations[].streamName` | `string` | yes |  |  |
| `spec.streamDeclarations[].columns` | `[]AzureMonitorDataCollectionRuleStreamDeclarationColumn` | yes |  |  |
| `spec.streamDeclarations[].columns[].name` | `string` | yes |  |  |
| `spec.streamDeclarations[].columns[].type` | `string` | yes |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group the data collection rule lives in. Can be
a literal string or a reference to an AzureResourceGroup output.

**ForceNew**: changing this destroys and recreates the rule.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The name of the data collection rule, unique within the resource
group.

**ForceNew**: changing this destroys and recreates the rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.region

`string` · required

The Azure region the rule is created in, e.g. "eastus". Machines in
any region can associate with the rule; the region only places the
rule object (and its ingestion processing) itself.

**ForceNew**: changing this destroys and recreates the rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.kind

`string` · optional (explicit presence)

The rule's platform kind. OMIT for the default rule kind, which
accepts data sources of every platform. "Linux" and "Windows"
restrict the rule to that platform's sources (Azure rejects a
Linux rule carrying windows_event_log sources, and a Windows rule
carrying syslog sources, at deploy time -- the provider performs no
early check). "AgentDirectToStore" is required for the *_direct
destinations (agent writes straight to storage/Event Hubs), and
"WorkspaceTransforms" is for workspace-scoped transformation rules
that carry no agent data sources.

**Replace-on-change once set**: a rule created WITHOUT a kind can
adopt one in place, but changing (or clearing) a kind that has been
set destroys and recreates the rule -- the provider enforces this
lifecycle.

- rule: {"string":{"in":["Linux","Windows","AgentDirectToStore","WorkspaceTransforms"]}}

### spec.description

`string`

A free-text description of the rule, shown in the portal.

### spec.dataCollectionEndpointId

`string | valueFrom`

The ARM ID of the Azure Monitor Data Collection Endpoint (DCE) the
rule ingests through. Required by Azure when the rule declares
custom streams (stream_declarations / the log_file and
prometheus_forwarder sources); optional otherwise. Format:
/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Insights/dataCollectionEndpoints/{name}
Provide the literal ARM ID of an endpoint managed outside this
catalog; leave unset to use Azure's default ingestion endpoints.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.identity

`AzureMonitorDataCollectionRuleIdentity`

The rule's managed identity -- required when a destination needs
the rule to authenticate as itself (for example storage
destinations on AgentDirectToStore rules, or transformations that
query other resources). Omit when no destination requires one.

- rule: identity_ids is required for USER_ASSIGNED and must be empty for SYSTEM_ASSIGNED

### spec.identity.type

`enum` · required

Identity flavor. SYSTEM_ASSIGNED is created and rotated by Azure
with the rule; USER_ASSIGNED brings an identity you manage
(grantable on destinations BEFORE the rule exists). The provider
supports exactly one flavor at a time on this resource -- there is
no combined mode.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_monitor_data_collection_rule_identity_type_unspecified` -- Not specified: rejected -- an identity block requires a flavor.
- `SYSTEM_ASSIGNED` -- Azure-managed identity created with the rule.
- `USER_ASSIGNED` -- An identity you create and manage (AzureUserAssignedIdentity).

### spec.identity.identityIds

`[]string | valueFrom`

For USER_ASSIGNED: the user-assigned identity attached to the
rule, by ARM ID. Reference an AzureUserAssignedIdentity resource so
destination grants can be composed before the rule is created.
Azure currently supports at most ONE user-assigned identity on a
data collection rule (enforced at deploy time).

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.dataSources

`AzureMonitorDataCollectionRuleDataSources`

The data sources to collect. Omit entirely for rules that carry no
agent sources (for example WorkspaceTransforms rules, or rules used
purely for direct ingestion through a DCE).

### spec.dataSources.syslogs

`[]AzureMonitorDataCollectionRuleSyslog`

Linux syslog sources (requires rule kind "Linux" or no kind).

### spec.dataSources.syslogs[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.syslogs[].facilityNames

`[]string` · required

The syslog facilities to collect ("*" for all). Values are the
provider's exact lowercase tokens.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["*","alert","audit","auth","authpriv","clock","cron","daemon","ftp","kern","local0","local1","local2","local3","local4","local5","local6","local7","lpr","mail","mark","news","nopri","ntp","syslog","user","uucp"]}}}}

### spec.dataSources.syslogs[].logLevels

`[]string` · required

The minimum severities to collect ("*" for all).

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["*","Alert","Critical","Debug","Emergency","Error","Info","Notice","Warning"]}}}}

### spec.dataSources.syslogs[].streams

`[]string` · required

The streams the collected records are tagged with -- syslog
records flow on "Microsoft-Syslog".

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.performanceCounters

`[]AzureMonitorDataCollectionRulePerformanceCounter`

Performance counter sources (both platforms).

### spec.dataSources.performanceCounters[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.performanceCounters[].samplingFrequencyInSeconds

`int32` · required · optional (explicit presence)

How often the counters are sampled, in seconds (1-1800). Streams
targeting "Microsoft-InsightsMetrics" require exactly 60 (Azure
enforces this at deploy time).

- rule: {"required":true,"int32":{"lte":1800,"gte":1}}

### spec.dataSources.performanceCounters[].counterSpecifiers

`[]string` · required

The counter specifiers to sample, e.g.
"\\Processor(_Total)\\% Processor Time" on Windows or
"Processor(*)\\% Processor Time" on Linux ("\\*" collects all
counters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.performanceCounters[].streams

`[]string` · required

The streams the samples are tagged with -- "Microsoft-Perf" for
Log Analytics, "Microsoft-InsightsMetrics" for Azure Monitor
metrics (the latter requires 60-second sampling).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.windowsEventLogs

`[]AzureMonitorDataCollectionRuleWindowsEventLog`

Windows event log sources (requires rule kind "Windows" or no
kind).

### spec.dataSources.windowsEventLogs[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.windowsEventLogs[].xPathQueries

`[]string` · required

XPath queries selecting the events to collect, e.g.
"System!*[System[(Level=1 or Level=2 or Level=3)]]". The part
before "!" is the channel; the part after filters its events.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.windowsEventLogs[].streams

`[]string` · required

The streams the events are tagged with -- Windows events flow on
"Microsoft-Event".

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.extensions

`[]AzureMonitorDataCollectionRuleExtension`

Azure Monitor Agent extension sources -- telemetry produced by an
agent extension (for example the workload insights extensions).

### spec.dataSources.extensions[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.extensions[].extensionName

`string` · required

The agent extension that produces the telemetry (for example
"DependencyAgent").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dataSources.extensions[].extensionJson

`string`

Extension-specific settings as a JSON object string. Both engines
validate the JSON syntax at plan time; the KEYS inside are defined
by each extension.

### spec.dataSources.extensions[].inputDataSources

`[]string`

The names of other data sources in this rule the extension
consumes as its inputs.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.dataSources.extensions[].streams

`[]string` · required

The streams the extension's records are tagged with.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.iisLogs

`[]AzureMonitorDataCollectionRuleIisLog`

IIS access-log sources (Windows).

### spec.dataSources.iisLogs[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.iisLogs[].logDirectories

`[]string`

Absolute paths of the IIS log directories to collect from. Omit to
collect from the server's configured IIS log location.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.dataSources.iisLogs[].streams

`[]string` · required

The streams the log records are tagged with -- IIS logs flow on
"Microsoft-W3CIISLog".

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.logFiles

`[]AzureMonitorDataCollectionRuleLogFile`

Custom text/JSON log-file sources. Each references a CUSTOM stream
that must be declared in stream_declarations, and custom streams
require the rule to ingest through a Data Collection Endpoint.

### spec.dataSources.logFiles[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.logFiles[].filePatterns

`[]string` · required

Glob patterns of the files to collect, e.g. "/var/log/myapp/*.log"
(Linux) or "C:\\logs\\*.log" (Windows).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.logFiles[].format

`string` · required

The file format: "text" or "json".

- rule: {"required":true,"string":{"in":["json","text"]}}

### spec.dataSources.logFiles[].settings

`AzureMonitorDataCollectionRuleLogFileSettings`

Text-format settings; omit for JSON files.

### spec.dataSources.logFiles[].settings.text

`AzureMonitorDataCollectionRuleLogFileSettingsText` · required

How a new record is recognized in the text file.

- rule: {"required":true}

### spec.dataSources.logFiles[].settings.text.recordStartTimestampFormat

`string` · required

The timestamp format that marks the start of a record, from
Azure's fixed vocabulary.

- rule: {"required":true,"string":{"in":["ISO 8601","YYYY-MM-DD HH:MM:SS","M/D/YYYY HH:MM:SS AM/PM","Mon DD, YYYY HH:MM:SS","yyMMdd HH:mm:ss","ddMMyy HH:mm:ss","MMM d hh:mm:ss","dd/MMM/yyyy:HH:mm:ss zzz","yyyy-MM-ddTHH:mm:ssK"]}}

### spec.dataSources.logFiles[].streams

`[]string` · required

The CUSTOM stream (declared in stream_declarations) the file's
records are tagged with, e.g. "Custom-MyAppLogs".

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.prometheusForwarders

`[]AzureMonitorDataCollectionRulePrometheusForwarder`

Prometheus metrics forwarder sources (Kubernetes / Managed
Prometheus).

### spec.dataSources.prometheusForwarders[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.prometheusForwarders[].labelIncludeFilters

`[]AzureMonitorDataCollectionRuleLabelIncludeFilter`

Label filters: only time series carrying one of these label/value
pairs are forwarded. Omit to forward everything.

### spec.dataSources.prometheusForwarders[].labelIncludeFilters[].label

`string` · required

The filter label -- Azure currently supports exactly
"microsoft_metrics_include_label".

- rule: {"required":true,"string":{"in":["microsoft_metrics_include_label"]}}

### spec.dataSources.prometheusForwarders[].labelIncludeFilters[].value

`string` · required

The label value that must be present for a time series to be
forwarded.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dataSources.prometheusForwarders[].streams

`[]string` · required

The stream the metrics are tagged with -- Prometheus metrics flow
only on "Microsoft-PrometheusMetrics".

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["Microsoft-PrometheusMetrics"]}}}}

### spec.dataSources.windowsFirewallLogs

`[]AzureMonitorDataCollectionRuleWindowsFirewallLog`

Windows Firewall log sources (Windows).

### spec.dataSources.windowsFirewallLogs[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.windowsFirewallLogs[].streams

`[]string` · required

The streams the log records are tagged with.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.platformTelemetries

`[]AzureMonitorDataCollectionRulePlatformTelemetry`

Platform telemetry (resource-level telemetry collected without an
agent).

### spec.dataSources.platformTelemetries[].name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.platformTelemetries[].streams

`[]string` · required

The streams the telemetry is tagged with.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataSources.dataImport

`AzureMonitorDataCollectionRuleDataImport`

Event Hub data import -- the rule pulls events FROM an Event Hub
as a source.

### spec.dataSources.dataImport.eventHubDataSource

`AzureMonitorDataCollectionRuleDataImportEventHub` · required

The Event Hub source. Azure's rule model carries exactly ONE
event-hub import per rule -- this is a single block by design (the
provider accepts a list but silently uses only the first entry).

- rule: {"required":true}

### spec.dataSources.dataImport.eventHubDataSource.name

`string` · required

The source's rule-local name.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"32"}}

### spec.dataSources.dataImport.eventHubDataSource.stream

`string` · required

The stream the imported events are tagged with.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dataSources.dataImport.eventHubDataSource.consumerGroup

`string`

The Event Hub consumer group to read from. Omit for the default
consumer group.

### spec.destinations

`AzureMonitorDataCollectionRuleDestinations` · required

The destinations telemetry can be routed to. At least one
destination is required; each carries a rule-local `name` that
data flows reference. Destination names share ONE namespace across
all arms -- Azure rejects duplicates at deploy time.

- rule: {"required":true}
- rule: configure at least one destination (log_analytics, azure_monitor_metrics, event_hub, event_hub_direct, monitor_accounts, storage_blobs, storage_blob_directs or storage_table_directs)

### spec.destinations.logAnalytics

`[]AzureMonitorDataCollectionRuleLogAnalytics`

Log Analytics workspace destinations.

### spec.destinations.logAnalytics[].name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.logAnalytics[].workspaceResourceId

`string | valueFrom` · required

The destination workspace, by ARM resource ID. Can be a literal
ARM ID or a reference to an AzureLogAnalyticsWorkspace output.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.destinations.azureMonitorMetrics

`AzureMonitorDataCollectionRuleAzureMonitorMetrics`

The Azure Monitor metrics destination (at most one per rule) --
where "Microsoft-InsightsMetrics" performance counters land.

### spec.destinations.azureMonitorMetrics.name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.eventHub

`AzureMonitorDataCollectionRuleEventHubDestination`

An Event Hub destination (at most one per rule).

### spec.destinations.eventHub.name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.eventHub.eventHubId

`string | valueFrom` · required

The destination Event Hub, by ARM resource ID. Can be a literal
ARM ID or a reference to an AzureEventHub output.

Collected data is DELIVERED into the hub; the rule lives in its own
resource group, so on a diagram the reference is access, not
placement -- the diagnostic setting's rule.

- references: AzureEventHub (`status.outputs.event_hub_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHub, name: <that resource's name>, fieldPath: status.outputs.event_hub_id}} -- a bare string does not parse

### spec.destinations.eventHubDirect

`AzureMonitorDataCollectionRuleEventHubDestination`

An Event Hub DIRECT destination (at most one per rule; requires
rule kind "AgentDirectToStore" -- the agent writes straight to the
hub).

### spec.destinations.eventHubDirect.name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.eventHubDirect.eventHubId

`string | valueFrom` · required

The destination Event Hub, by ARM resource ID. Can be a literal
ARM ID or a reference to an AzureEventHub output.

Collected data is DELIVERED into the hub; the rule lives in its own
resource group, so on a diagram the reference is access, not
placement -- the diagnostic setting's rule.

- references: AzureEventHub (`status.outputs.event_hub_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHub, name: <that resource's name>, fieldPath: status.outputs.event_hub_id}} -- a bare string does not parse

### spec.destinations.monitorAccounts

`[]AzureMonitorDataCollectionRuleMonitorAccount`

Azure Monitor workspace (managed Prometheus) destinations -- where
"Microsoft-PrometheusMetrics" streams land.

### spec.destinations.monitorAccounts[].name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.monitorAccounts[].monitorAccountId

`string | valueFrom` · required

The ARM ID of the Azure Monitor workspace (managed Prometheus
account). Format:
/subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Monitor/accounts/{name}
Provide the literal ARM ID of a workspace managed outside this
catalog.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.destinations.storageBlobs

`[]AzureMonitorDataCollectionRuleStorageBlobDestination`

Storage blob destinations.

### spec.destinations.storageBlobs[].name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.storageBlobs[].containerName

`string` · required

The blob container the telemetry is written to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.storageBlobs[].storageAccountId

`string | valueFrom` · required

The destination storage account, by ARM resource ID. Can be a
literal ARM ID or a reference to an AzureStorageAccount output.

Collected data is DELIVERED into this account; the rule lives in its
own resource group, so on a diagram the reference is access, not
placement -- the diagnostic setting's rule.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.destinations.storageBlobDirects

`[]AzureMonitorDataCollectionRuleStorageBlobDestination`

Storage blob DIRECT destinations (require rule kind
"AgentDirectToStore").

### spec.destinations.storageBlobDirects[].name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.storageBlobDirects[].containerName

`string` · required

The blob container the telemetry is written to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.storageBlobDirects[].storageAccountId

`string | valueFrom` · required

The destination storage account, by ARM resource ID. Can be a
literal ARM ID or a reference to an AzureStorageAccount output.

Collected data is DELIVERED into this account; the rule lives in its
own resource group, so on a diagram the reference is access, not
placement -- the diagnostic setting's rule.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.destinations.storageTableDirects

`[]AzureMonitorDataCollectionRuleStorageTableDirect`

Storage table DIRECT destinations (require rule kind
"AgentDirectToStore").

### spec.destinations.storageTableDirects[].name

`string` · required

The destination's rule-local name (referenced by data flows).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.storageTableDirects[].tableName

`string` · required

The storage table the telemetry is written to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.destinations.storageTableDirects[].storageAccountId

`string | valueFrom` · required

The destination storage account, by ARM resource ID. Can be a
literal ARM ID or a reference to an AzureStorageAccount output.

Collected data is DELIVERED into this account; the rule lives in its
own resource group, so on a diagram the reference is access, not
placement -- the diagnostic setting's rule.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.dataFlows

`[]AzureMonitorDataCollectionRuleDataFlow` · required

The flows wiring streams to destinations (at least one). Each flow
takes the records of the listed streams and delivers them to the
listed destinations, optionally transforming them with KQL on the
way in.

- rule: {"repeated":{"minItems":"1"}}

### spec.dataFlows[].streams

`[]string` · required

The streams this flow picks up: platform streams
("Microsoft-Syslog", "Microsoft-Perf", "Microsoft-InsightsMetrics",
"Microsoft-Event", ...) or custom streams declared in
stream_declarations ("Custom-...").

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataFlows[].destinations

`[]string` · required

The destination NAMES (from the destinations block) this flow
delivers to. Azure rejects a flow referencing a name no
destination carries.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.dataFlows[].builtInTransform

`string`

A built-in transformation applied in the ingestion pipeline.
Mutually relevant with transform_kql -- most flows set neither.

### spec.dataFlows[].outputStream

`string`

The stream the transformed records continue on when the
transformation changes the schema (e.g. "Microsoft-Syslog" records
transformed into a custom table's shape). Requires transform_kql.

### spec.dataFlows[].transformKql

`string`

A KQL query applied to every record in the ingestion pipeline --
filter, project, or reshape records before they land (for example
"source | where SeverityLevel != 'verbose'").

### spec.streamDeclarations

`[]AzureMonitorDataCollectionRuleStreamDeclaration`

Declarations of CUSTOM streams (names starting "Custom-") used for
text/JSON log files and direct ingestion: each declares the schema
(column names and types) of one custom stream. Platform streams
(Microsoft-*) are never declared here. Custom streams require the
rule to ingest through a Data Collection Endpoint
(data_collection_endpoint_id).

### spec.streamDeclarations[].streamName

`string` · required

The custom stream's name. Azure requires custom stream names to
start with "Custom-" (enforced at deploy time), e.g.
"Custom-MyAppLogs".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.streamDeclarations[].columns

`[]AzureMonitorDataCollectionRuleStreamDeclarationColumn` · required

The stream's columns, in order.

- rule: {"repeated":{"minItems":"1"}}

### spec.streamDeclarations[].columns[].name

`string` · required

The column name (e.g. "TimeGenerated", "RawData").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.streamDeclarations[].columns[].type

`string` · required

The column type, from Azure's fixed vocabulary.

- rule: {"required":true,"string":{"in":["boolean","datetime","dynamic","int","long","real","string"]}}

### spec.tags

`map<string, string>`

Tags to apply to the data collection rule, merged over the
Planton-derived metadata tags (user values win on key conflicts).

## Validation Rules

- `dcr_stream_declaration_names_unique`: stream_declarations must carry unique stream_name values

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMonitorDataCollectionRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.data_collection_rule_id` | `string` | The rule's ARM resource ID (.../providers/Microsoft.Insights/dataCollectionRules/{name}) -- the value associations and Log Analytics workspaces reference. |
| `status.outputs.data_collection_rule_name` | `string` | The rule's resource name within its resource group. |
| `status.outputs.immutable_id` | `string` | The rule's immutable ID -- the identifier agents and the ingestion API address the rule by (stable across moves, unlike the ARM ID's casing). |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the rule's system-assigned identity, when one is configured -- grant this principal access on destinations (empty otherwise). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.identity.identityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.destinations.logAnalytics[].workspaceResourceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |
| `spec.destinations.eventHub.eventHubId` | AzureEventHub | `status.outputs.event_hub_id` |
| `spec.destinations.eventHubDirect.eventHubId` | AzureEventHub | `status.outputs.event_hub_id` |
| `spec.destinations.storageBlobs[].storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.destinations.storageBlobDirects[].storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.destinations.storageTableDirects[].storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureLogAnalyticsWorkspace | `spec.dataCollectionRuleId` | `status.outputs.data_collection_rule_id` |
| AzureMonitorDataCollectionRuleAssociation | `spec.dataCollectionRuleId` | `status.outputs.data_collection_rule_id` |

## See Also

- [Overview](../README.md)
