# AzureMonitorDiagnosticSetting

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureMonitorDiagnosticSettingSpec** defines the configuration for creating
an Azure Monitor diagnostic setting.

A diagnostic setting is how a resource's platform telemetry leaves the
resource: it selects which log categories and metrics the target resource
emits and routes them to one or more destinations -- a Log Analytics
Workspace (queryable, alertable), a Storage Account (cheap archival), an
Event Hub (streaming to SIEMs and external systems), or a partner
monitoring solution. Without a diagnostic setting, most Azure resources
emit nothing beyond basic platform metrics.

The setting is an extension resource: it lives ON the target resource
(any ARM resource -- a Key Vault, an AKS cluster, an Application Gateway, a
subscription itself), and a target can carry up to five settings, each
routing a different selection to different destinations.

Which log categories exist is defined per resource type by Azure (for
example Key Vault exposes AuditEvent; Application Gateway exposes access
and firewall logs). Category names are discoverable in the portal's
"Diagnostic settings" blade or via
`az monitor diagnostic-settings categories list --resource <id>`.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: a category AND a
# category-group log entry, a metric entry, and three destinations at
# once (workspace with the Dedicated table layout, storage archival, and
# an Event Hub stream with a named hub).
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorDiagnosticSetting
metadata:
  name: test-diag
  org: test-org
  env: dev
spec:
  settingName: route-to-observability
  targetResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.KeyVault/vaults/test-vault
  enabledLogs:
    - category: AuditEvent
    - categoryGroup: allLogs
  enabledMetrics:
    - category: AllMetrics
  logAnalyticsWorkspaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.OperationalInsights/workspaces/test-law
  logAnalyticsDestinationType: DEDICATED
  storageAccountId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/testarchive
  eventhubAuthorizationRuleId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.EventHub/namespaces/test-ns/authorizationRules/RootManageSharedAccessKey
  eventhubName:
    value: diagnostics
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.settingName` | `string` | yes |  |  |
| `spec.targetResourceId` | `string \| valueFrom` | yes |  |  |
| `spec.enabledLogs` | `[]AzureMonitorDiagnosticSettingLog` |  |  |  |
| `spec.enabledLogs[].category` | `string` |  |  |  |
| `spec.enabledLogs[].categoryGroup` | `string` |  |  |  |
| `spec.enabledMetrics` | `[]AzureMonitorDiagnosticSettingMetric` |  |  |  |
| `spec.enabledMetrics[].category` | `string` | yes |  |  |
| `spec.logAnalyticsWorkspaceId` | `string \| valueFrom` |  |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`) |
| `spec.logAnalyticsDestinationType` | `enum` |  |  |  |
| `spec.storageAccountId` | `string \| valueFrom` |  |  | AzureStorageAccount (`status.outputs.storage_account_id`) |
| `spec.eventhubAuthorizationRuleId` | `string \| valueFrom` |  |  | AzureEventHubAuthorizationRule (`status.outputs.authorization_rule_id`) |
| `spec.eventhubName` | `string \| valueFrom` |  |  | AzureEventHub (`status.outputs.event_hub_name`) |
| `spec.partnerSolutionId` | `string` |  |  |  |

## Field Details

### spec.settingName

`string` · required

The name of the diagnostic setting, unique among the target resource's
settings. Length: 1 to 260 characters; may not contain
`<`, `>`, `*`, `%`, `&`, `:`, `\`, `?`, `+`, or `/`.

**ForceNew**: Changing this destroys and recreates the setting.

- rule: diagnostic setting name may not contain any of < > * % & : \ ? + /
- rule: {"required":true,"string":{"minLen":"1","maxLen":"260"}}

### spec.targetResourceId

`string | valueFrom` · required

The ARM ID of the resource whose telemetry this setting routes -- any
Azure resource that emits diagnostics (a vault, a cluster, a gateway, a
database, a subscription). There is no default kind because no single
kind dominates as a target: reference the resource's `*_id` output
explicitly with valueFrom (kind + fieldPath), or pass a literal ARM ID.

Containment-exempt: a diagnostic setting WATCHES its target and
routes what it emits; it is an extension resource, not a tenant.
Vaults, clusters, and gateways are all containers, and a setting
wired to one by reference would otherwise be drawn inside the very
thing it monitors.

**ForceNew**: Changing this destroys and recreates the setting.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.enabledLogs

`[]AzureMonitorDiagnosticSettingLog`

The log categories to enable on the target. Each entry names either a
single category or a category group. At least one log or metric entry
must be enabled -- Azure has no concept of an empty diagnostic setting
(the API accepts one, then the setting is unreadable).

- rule: each log entry names exactly one of category (a single log category) or category_group (an Azure-curated bundle like allLogs)

### spec.enabledLogs[].category

`string`

A single log category to enable (for example "AuditEvent" on a Key
Vault, "kube-audit" on an AKS cluster). Category names are defined per
resource type by Azure. Exactly one of category or category_group.

### spec.enabledLogs[].categoryGroup

`string`

A category group to enable -- Azure's curated bundles, "allLogs" or
"audit", which track new categories automatically as Azure adds them.
Exactly one of category or category_group.

### spec.enabledMetrics

`[]AzureMonitorDiagnosticSettingMetric`

The metric categories to enable on the target. Most resource types
expose a single "AllMetrics" category; some expose finer categories.

### spec.enabledMetrics[].category

`string` · required

The metric category to enable. Most resource types expose "AllMetrics";
some expose finer categories -- discover them per resource type the same
way as log categories.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.logAnalyticsWorkspaceId

`string | valueFrom`

The Log Analytics Workspace to send the selected telemetry to -- the
destination that makes logs queryable with KQL and alertable with
scheduled query rules. Can be a literal workspace ARM ID or a reference
to an AzureLogAnalyticsWorkspace output.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.logAnalyticsDestinationType

`enum`

How logs land in the Log Analytics Workspace: DEDICATED writes to
resource-specific tables (the modern layout -- typed columns, better
query performance); AZURE_DIAGNOSTICS writes everything to the legacy
shared AzureDiagnostics table. Only meaningful when
log_analytics_workspace_id is set; some resource types support only one
layout and Azure decides for them when unspecified.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_monitor_diagnostic_setting_log_analytics_destination_type_unspecified` -- Not specified -- Azure picks the layout the target resource type supports (resource-specific tables where available).
- `DEDICATED` -- Resource-specific tables -- the modern layout with typed columns and better query performance. Prefer this where the resource type supports it.
- `AZURE_DIAGNOSTICS` -- The legacy shared AzureDiagnostics table.

### spec.storageAccountId

`string | valueFrom`

The Storage Account to archive the selected telemetry to -- the cheap
long-term retention destination. Can be a literal storage account ARM ID
or a reference to an AzureStorageAccount output.

- references: AzureStorageAccount (`status.outputs.storage_account_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.storage_account_id}} -- a bare string does not parse

### spec.eventhubAuthorizationRuleId

`string | valueFrom`

The Event Hub namespace authorization rule that authorizes streaming the
selected telemetry to Event Hubs -- the destination for SIEMs and
external analytics systems. A NAMESPACE-scoped rule with send rights.
Can be a literal ARM ID or a reference to an
AzureEventHubAuthorizationRule output; pair with eventhub_name to pick
a specific hub.

- references: AzureEventHubAuthorizationRule (`status.outputs.authorization_rule_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHubAuthorizationRule, name: <that resource's name>, fieldPath: status.outputs.authorization_rule_id}} -- a bare string does not parse

### spec.eventhubName

`string | valueFrom`

The Event Hub (within the authorized namespace) to stream to. When
empty, Azure routes to a default hub per category. Only meaningful with
eventhub_authorization_rule_id. Can be a literal name or a reference to
an AzureEventHub output.

- references: AzureEventHub (`status.outputs.event_hub_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventHub, name: <that resource's name>, fieldPath: status.outputs.event_hub_name}} -- a bare string does not parse

### spec.partnerSolutionId

`string`

The ARM ID of an Azure Native ISV partner solution (for example Elastic
or Datadog resources created through Azure Marketplace) to send the
selected telemetry to. Pass the literal ARM ID of the partner resource.

## Validation Rules

- `diagnostic_setting_at_least_one_category`: enable at least one log or metric category -- a diagnostic setting that routes nothing is rejected by Azure
- `diagnostic_setting_at_least_one_destination`: route the telemetry somewhere: set at least one of log_analytics_workspace_id, storage_account_id, eventhub_authorization_rule_id, or partner_solution_id
- `diagnostic_setting_eventhub_name_requires_rule`: eventhub_name picks a hub within the namespace that eventhub_authorization_rule_id authorizes -- set the authorization rule too

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMonitorDiagnosticSetting, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.diagnostic_setting_id` | `string` | The Azure Resource Manager ID of the diagnostic setting -- an extension resource ID scoped under the target resource. Format: {target_resource_id}/providers/Microsoft.Insights/diagnosticSettings/{name} |
| `status.outputs.diagnostic_setting_name` | `string` | The name of the diagnostic setting. |
| `status.outputs.target_resource_id` | `string` | The ARM ID of the target resource the setting routes telemetry from -- resolved from the spec reference, exported for downstream visibility. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.logAnalyticsWorkspaceId` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |
| `spec.storageAccountId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.eventhubAuthorizationRuleId` | AzureEventHubAuthorizationRule | `status.outputs.authorization_rule_id` |
| `spec.eventhubName` | AzureEventHub | `status.outputs.event_hub_name` |

## See Also

- [Overview](../README.md)
