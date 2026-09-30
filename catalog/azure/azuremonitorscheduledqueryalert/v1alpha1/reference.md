# AzureMonitorScheduledQueryAlert

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureMonitorScheduledQueryAlertSpec** defines the configuration for
creating an Azure Monitor scheduled query alert rule
(Microsoft.Insights/scheduledQueryRules) -- the log-search alert.

A scheduled query alert runs a KQL query against a Log Analytics Workspace
(or an Application Insights resource) on a schedule, compares the result to
a threshold, and fires action groups when the condition holds. It is the
alerting half of the logging pipeline: diagnostic settings route logs into
the workspace, and query alerts watch what arrives -- error spikes,
security events, missing heartbeats, business anomalies, anything KQL can
express.

Two evaluation styles per criterion:
  - **row count** -- time_aggregation_method COUNT compares the number of
    rows the query returns ("more than 5 exceptions in 10 minutes")
  - **metric measurement** -- other aggregations compute over a numeric
    column named in metric_measure_column ("average latency column > 500")

The rule is regional: it must be created in the same region as the
workspace it queries.

## Example

```yaml
# Offline-plan test manifest. Exercises the deep seams: a
# metric-measurement criterion (Average over a projected column), a
# dimension split, failing periods, the mute duration, a query lookback
# override, a system-assigned identity, per-resource targeting, and the
# action block's custom properties + email subject.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorScheduledQueryAlert
metadata:
  name: test-query-alert
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: test-rg
  alertName: api-latency-p95
  scope:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.OperationalInsights/workspaces/test-law
  displayName: API latency
  description: Average request duration over 500ms
  enabled: true
  severity: 1
  evaluationFrequency: PT15M
  windowDuration: PT30M
  criteria:
    - query: |
        AppRequests
        | summarize avg(DurationMs) by bin(TimeGenerated, 5m), AppRoleName
      timeAggregationMethod: AVERAGE
      operator: GREATER_THAN
      threshold: 500
      metricMeasureColumn: avg_DurationMs
      resourceIdColumn: _ResourceId
      dimensions:
        - name: AppRoleName
          operator: INCLUDE
          values:
            - "*"
      failingPeriods:
        minimumFailingPeriodsToTriggerAlert: 3
        numberOfEvaluationPeriods: 5
  queryTimeRangeOverride: P1D
  muteActionsAfterAlertDuration: PT30M
  workspaceAlertsStorageEnabled: false
  skipQueryValidation: true
  identity:
    type: SYSTEM_ASSIGNED
  targetResourceTypes:
    - Microsoft.OperationalInsights/workspaces
  action:
    actionGroupIds:
      - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Insights/actionGroups/test-ag
    customProperties:
      source: planton
    emailSubject: "[ALERT] API latency"
  tags:
    team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.alertName` | `string` | yes |  |  |
| `spec.scope` | `string \| valueFrom` | yes |  | AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`), AzureApplicationInsights (`status.outputs.application_insights_id`) |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.severity` | `int32` |  | `3` |  |
| `spec.evaluationFrequency` | `string` |  | `PT5M` |  |
| `spec.windowDuration` | `string` |  | `PT5M` |  |
| `spec.criteria` | `[]AzureMonitorScheduledQueryAlertCriteria` | yes |  |  |
| `spec.criteria[].query` | `string` | yes |  |  |
| `spec.criteria[].timeAggregationMethod` | `enum` |  |  |  |
| `spec.criteria[].operator` | `enum` |  |  |  |
| `spec.criteria[].threshold` | `double` |  |  |  |
| `spec.criteria[].metricMeasureColumn` | `string` |  |  |  |
| `spec.criteria[].resourceIdColumn` | `string` |  |  |  |
| `spec.criteria[].dimensions` | `[]AzureMonitorScheduledQueryAlertDimension` |  |  |  |
| `spec.criteria[].dimensions[].name` | `string` | yes |  |  |
| `spec.criteria[].dimensions[].operator` | `enum` |  |  |  |
| `spec.criteria[].dimensions[].values` | `[]string` | yes |  |  |
| `spec.criteria[].failingPeriods` | `AzureMonitorScheduledQueryAlertFailingPeriods` |  |  |  |
| `spec.criteria[].failingPeriods.minimumFailingPeriodsToTriggerAlert` | `int32` |  |  |  |
| `spec.criteria[].failingPeriods.numberOfEvaluationPeriods` | `int32` |  |  |  |
| `spec.queryTimeRangeOverride` | `string` |  |  |  |
| `spec.autoMitigationEnabled` | `bool` |  |  |  |
| `spec.muteActionsAfterAlertDuration` | `string` |  |  |  |
| `spec.workspaceAlertsStorageEnabled` | `bool` |  |  |  |
| `spec.skipQueryValidation` | `bool` |  |  |  |
| `spec.identity` | `AzureMonitorScheduledQueryAlertIdentity` |  |  |  |
| `spec.identity.type` | `enum` |  |  |  |
| `spec.identity.userAssignedIdentityIds` | `[]string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.identity_id`) |
| `spec.targetResourceTypes` | `[]string` |  |  |  |
| `spec.action` | `AzureMonitorScheduledQueryAlertAction` |  |  |  |
| `spec.action.actionGroupIds` | `[]string \| valueFrom` |  |  | AzureMonitorActionGroup (`status.outputs.action_group_id`) |
| `spec.action.customProperties` | `map<string, string>` |  |  |  |
| `spec.action.emailSubject` | `string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region where the alert rule will be created -- must match the
region of the workspace (or Application Insights resource) it queries.
Examples: "eastus", "westus2", "westeurope".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure Resource Group where the alert rule will be created.
Can be a literal string or a reference to an AzureResourceGroup output.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.alertName

`string` · required

The name of the alert rule, unique within the resource group.

**ForceNew**: Changing this destroys and recreates the rule.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"260"}}

### spec.scope

`string | valueFrom` · required

The resource the query runs against -- a Log Analytics Workspace (the
common case, and the default reference) or an Application Insights
resource (a valueFrom naming that kind; both are declared candidates,
so each one's resource ID output fills in).
Azure allows exactly one scope per rule.

**ForceNew**: Changing this destroys and recreates the rule.

- references: AzureLogAnalyticsWorkspace (`status.outputs.workspace_id`), AzureApplicationInsights (`status.outputs.application_insights_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureLogAnalyticsWorkspace, name: <that resource's name>, fieldPath: status.outputs.workspace_id}} -- a bare string does not parse

### spec.displayName

`string`

The display name shown in the portal and notifications. When empty,
Azure shows the resource name.

### spec.description

`string`

A human-readable description delivered with every notification --
the place for runbook links and on-call context.

### spec.enabled

`bool` · optional (explicit presence)

Whether the rule evaluates at all. Disable during maintenance windows
instead of deleting the rule.
Default: true

- default: `true`

### spec.severity

`int32` · optional (explicit presence)

The alert severity: 0 (critical) through 4 (verbose). Severity drives
notification routing and on-call urgency conventions.
Default: 3 (informational -- raise deliberately for paging conditions)

- default: `3`
- rule: {"int32":{"lte":4,"gte":0}}

### spec.evaluationFrequency

`string` · optional (explicit presence)

How often the query runs, as an ISO 8601 duration. One of PT1M, PT5M,
PT10M, PT15M, PT30M, PT45M, PT1H, PT2H, PT3H, PT4H, PT5H, PT6H, P1D.
Frequent evaluation costs more; PT5M suits most operational alerts.
Default: PT5M

- default: `PT5M`
- rule: {"string":{"in":["PT1M","PT5M","PT10M","PT15M","PT30M","PT45M","PT1H","PT2H","PT3H","PT4H","PT5H","PT6H","P1D"]}}

### spec.windowDuration

`string` · optional (explicit presence)

The time range each evaluation queries over, as an ISO 8601 duration.
One of PT1M, PT5M, PT10M, PT15M, PT30M, PT45M, PT1H, PT2H, PT3H, PT4H,
PT5H, PT6H, P1D, P2D. Must be at least the evaluation frequency.
Default: PT5M

- default: `PT5M`
- rule: {"string":{"in":["PT1M","PT5M","PT10M","PT15M","PT30M","PT45M","PT1H","PT2H","PT3H","PT4H","PT5H","PT6H","P1D","P2D"]}}

### spec.criteria

`[]AzureMonitorScheduledQueryAlertCriteria` · required

The conditions the query results are evaluated against. Azure evaluates
each criterion independently; the rule fires when any holds.

- rule: {"repeated":{"minItems":"1"}}

### spec.criteria[].query

`string` · required

The KQL query to run. Its result rows (or a numeric column of them)
are what the condition evaluates.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.criteria[].timeAggregationMethod

`enum`

How the query results aggregate before comparison: COUNT compares the
number of result rows; AVERAGE/MINIMUM/MAXIMUM/TOTAL compute over the
numeric column named in metric_measure_column (Azure requires that
column for the non-COUNT aggregations).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_scheduled_query_alert_time_aggregation_unspecified` -- Not specified -- invalid; choose an explicit aggregation.
- `COUNT` -- The number of result rows -- the row-count evaluation style.
- `AVERAGE` -- The mean of metric_measure_column across result rows.
- `MINIMUM` -- The smallest metric_measure_column value.
- `MAXIMUM` -- The largest metric_measure_column value.
- `TOTAL` -- The sum of metric_measure_column across result rows.

### spec.criteria[].operator

`enum`

The comparison between the aggregated value and the threshold.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_scheduled_query_alert_operator_unspecified` -- Not specified -- invalid; choose an explicit operator.
- `EQUAL` -- The aggregated value equals the threshold.
- `GREATER_THAN` -- The aggregated value exceeds the threshold.
- `GREATER_THAN_OR_EQUAL` -- The aggregated value is at least the threshold.
- `LESS_THAN` -- The aggregated value is below the threshold.
- `LESS_THAN_OR_EQUAL` -- The aggregated value is at most the threshold.

### spec.criteria[].threshold

`double`

The threshold the aggregated value is compared against. Zero is a
meaningful threshold ("any result row" with COUNT + GREATER_THAN).

### spec.criteria[].metricMeasureColumn

`string`

The numeric column the non-COUNT aggregations compute over. The query
must project this column.

### spec.criteria[].resourceIdColumn

`string`

The column carrying an ARM resource ID when the rule alerts per
resource (pairs with the spec's target_resource_types).

### spec.criteria[].dimensions

`[]AzureMonitorScheduledQueryAlertDimension`

Dimension splits -- each combination of the named columns' values is
evaluated (and alerts) independently.

### spec.criteria[].dimensions[].name

`string` · required

The result column to split on.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.criteria[].dimensions[].operator

`enum`

How the values list applies: INCLUDE evaluates only matching values
(each independently), EXCLUDE evaluates everything else.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_scheduled_query_alert_dimension_operator_unspecified` -- Not specified -- invalid; choose an explicit operator.
- `INCLUDE` -- Evaluate only the listed values, each independently.
- `EXCLUDE` -- Evaluate every value except the listed ones.

### spec.criteria[].dimensions[].values

`[]string` · required

The column values the operator applies to. "*" with INCLUDE splits
across every value.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.criteria[].failingPeriods

`AzureMonitorScheduledQueryAlertFailingPeriods`

Requires the condition to hold across multiple evaluation periods
before firing -- the flap damper. When absent, one breaching
evaluation fires the alert.

- rule: minimum_failing_periods_to_trigger_alert cannot exceed number_of_evaluation_periods -- the alert can never fire otherwise

### spec.criteria[].failingPeriods.minimumFailingPeriodsToTriggerAlert

`int32`

How many of the examined periods must breach for the alert to fire
(1-6). Must not exceed number_of_evaluation_periods.

- rule: {"int32":{"lte":6,"gte":1}}

### spec.criteria[].failingPeriods.numberOfEvaluationPeriods

`int32`

How many recent evaluation periods are examined (1-6).

- rule: {"int32":{"lte":6,"gte":1}}

### spec.queryTimeRangeOverride

`string`

Overrides the query's default lookback when the query needs more
history than the evaluation window (for example a query comparing
against last week). One of PT5M, PT10M, PT15M, PT20M, PT30M, PT45M,
PT1H, PT2H, PT3H, PT4H, PT5H, PT6H, P1D, P2D.

- rule: query_time_range_override must be one of PT5M, PT10M, PT15M, PT20M, PT30M, PT45M, PT1H, PT2H, PT3H, PT4H, PT5H, PT6H, P1D, P2D

### spec.autoMitigationEnabled

`bool`

Whether the alert auto-resolves when the condition stops holding.
Mutually exclusive with mute_actions_after_alert_duration.
Default: false (Azure's default -- each firing is its own alert)

### spec.muteActionsAfterAlertDuration

`string`

Suppresses repeat action-group firings for the given duration after an
alert fires -- the noise dial for flapping conditions. One of PT5M,
PT10M, PT15M, PT30M, PT45M, PT1H, PT2H, PT3H, PT4H, PT5H, PT6H, P1D,
P2D. Mutually exclusive with auto_mitigation_enabled.

- rule: mute_actions_after_alert_duration must be one of PT5M, PT10M, PT15M, PT30M, PT45M, PT1H, PT2H, PT3H, PT4H, PT5H, PT6H, P1D, P2D

### spec.workspaceAlertsStorageEnabled

`bool`

Whether Azure verifies the alert has access to dedicated
workspace-alerts storage (for estates that route alert state to
customer-managed storage).
Default: false

### spec.skipQueryValidation

`bool`

Whether to skip validating the query at rule creation -- needed when
the queried table appears only after data starts flowing (custom logs).

### spec.identity

`AzureMonitorScheduledQueryAlertIdentity`

The rule's managed identity, used to run the query with Entra
permissions instead of the alert service's built-in access -- required
when the workspace enforces Entra-only query access, or when the query
crosses resources the service identity cannot read.

- rule: user_assigned_identity_ids is required with USER_ASSIGNED and must be empty with SYSTEM_ASSIGNED

### spec.identity.type

`enum`

The identity model: SYSTEM_ASSIGNED (Azure creates and rotates a
service principal bound to the rule's lifecycle) or USER_ASSIGNED
(bring an identity from user_assigned_identity_ids, grantable before
the rule exists).

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `azure_monitor_scheduled_query_alert_identity_type_unspecified` -- Not specified -- invalid; choose an explicit identity model.
- `SYSTEM_ASSIGNED` -- Azure creates a service principal bound to the rule's lifecycle.
- `USER_ASSIGNED` -- Bring your own AzureUserAssignedIdentity -- grantable before the rule exists.

### spec.identity.userAssignedIdentityIds

`[]string | valueFrom`

The user-assigned identities to attach -- required when (and only
meaningful when) type is USER_ASSIGNED. Each entry references an
AzureUserAssignedIdentity's ARM id.

- references: AzureUserAssignedIdentity (`status.outputs.identity_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.identity_id}} -- a bare string does not parse

### spec.targetResourceTypes

`[]string`

The resource types the rule targets when the query projects a
resource-id column (for example "Microsoft.Compute/virtualMachines") --
lets one rule alert per resource.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.action

`AzureMonitorScheduledQueryAlertAction`

The action to take when the alert fires.

### spec.action.actionGroupIds

`[]string | valueFrom`

The action groups to notify. Each entry can be a literal ARM ID or a
reference to an AzureMonitorActionGroup output.

- references: AzureMonitorActionGroup (`status.outputs.action_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMonitorActionGroup, name: <that resource's name>, fieldPath: status.outputs.action_group_id}} -- a bare string does not parse

### spec.action.customProperties

`map<string, string>`

Custom properties merged into the alert payload delivered to the
action groups.

### spec.action.emailSubject

`string`

Overrides the subject line of email notifications.

### spec.tags

`map<string, string>`

Tags to apply to the alert rule, merged over the Planton-derived
metadata tags (user values win on key conflicts).

## Validation Rules

- `scheduled_query_alert_mute_xor_auto_mitigation`: auto_mitigation_enabled and mute_actions_after_alert_duration are mutually exclusive -- an auto-resolving alert cannot also mute repeat firings

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMonitorScheduledQueryAlert, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.scheduled_query_alert_id` | `string` | The Azure Resource Manager ID of the scheduled query alert rule. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Insights/scheduledQueryRules/{name} |
| `status.outputs.scheduled_query_alert_name` | `string` | The name of the scheduled query alert rule. |
| `status.outputs.identity_principal_id` | `string` | The principal ID of the rule's system-assigned managed identity. Empty unless the identity block enables SYSTEM_ASSIGNED. Grant this principal read access on the queried workspace when the workspace enforces Entra-only query access. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.scope` | AzureLogAnalyticsWorkspace | `status.outputs.workspace_id` |
| `spec.scope` | AzureApplicationInsights | `status.outputs.application_insights_id` |
| `spec.identity.userAssignedIdentityIds` | AzureUserAssignedIdentity | `status.outputs.identity_id` |
| `spec.action.actionGroupIds` | AzureMonitorActionGroup | `status.outputs.action_group_id` |

## See Also

- [Overview](../README.md)
