# GcpBillingBudget

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBillingBudgetSpec sets a spending guardrail on a Cloud Billing
account: a budgeted amount for a period, the projects, folders, services,
or labels the spend is measured over, the thresholds that alert, and
where the alerts go. A budget never stops spending by itself -- it
notifies; a Pub/Sub notification into a Cloud Function is how teams cap
spend automatically. Declare one budget per project or environment for
the per-team guardrail pattern, or one unfiltered budget per billing
account for the whole bill.

The budget lives on the billing account, not in a project, so the
principal needs roles/billing.costsManager (or billing.admin) on the
account. Everything but the billing account changes in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBillingBudget
metadata:
  name: platform-prod-monthly
spec:
  # The billing account the budget belongs to; the module accepts the bare
  # ID or billingAccounts/{id}.
  billingAccount: 012345-6789AB-CDEF01
  displayName: Platform production -- monthly
  # A fixed monthly budget of USD 5,000.
  amount:
    specifiedAmount:
      currencyCode: USD
      units: 5000
  # Only this project's spend counts, reset every month, net of all
  # credits.
  budgetFilter:
    projects:
      - valueFrom:
          kind: GcpProject
          name: platform-prod
          fieldPath: status.outputs.project_number
    calendarPeriod: MONTH
    creditTypesTreatment: INCLUDE_ALL_CREDITS
  # Alert at half, at 90% of actual spend, and when the forecast says the
  # month will overshoot.
  thresholdRules:
    - thresholdPercent: 0.5
    - thresholdPercent: 0.9
    - thresholdPercent: 1.0
      spendBasis: FORECASTED_SPEND
  # Alerts go to a Pub/Sub topic (for automation) and an on-call channel
  # (for people), and not to every billing administrator.
  notifications:
    pubsubTopic:
      valueFrom:
        kind: GcpPubSubTopic
        name: budget-alerts
        fieldPath: status.outputs.topic_id
    monitoringNotificationChannels:
      - valueFrom:
          kind: GcpMonitoringNotificationChannel
          name: platform-oncall-email
          fieldPath: status.outputs.channel_name
    disableDefaultIamRecipients: true
  ownershipScope: ALL_USERS
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.billingAccount` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.amount` | `GcpBillingBudgetAmount` | yes |  |  |
| `spec.amount.specifiedAmount` | `GcpBillingBudgetSpecifiedAmount` |  |  |  |
| `spec.amount.specifiedAmount.currencyCode` | `string` |  |  |  |
| `spec.amount.specifiedAmount.units` | `int64` |  |  |  |
| `spec.amount.specifiedAmount.nanos` | `int32` |  |  |  |
| `spec.amount.lastPeriodAmount` | `bool` |  |  |  |
| `spec.budgetFilter` | `GcpBillingBudgetFilter` |  |  |  |
| `spec.budgetFilter.projects` | `[]string \| valueFrom` |  |  | GcpProject (`status.outputs.project_number`) |
| `spec.budgetFilter.resourceAncestors` | `[]string \| valueFrom` |  |  | GcpFolder (`status.outputs.name`) |
| `spec.budgetFilter.services` | `[]string` |  |  |  |
| `spec.budgetFilter.subaccounts` | `[]string` |  |  |  |
| `spec.budgetFilter.labels` | `map<string, string>` |  |  |  |
| `spec.budgetFilter.creditTypesTreatment` | `string` |  | `INCLUDE_ALL_CREDITS` |  |
| `spec.budgetFilter.creditTypes` | `[]string` |  |  |  |
| `spec.budgetFilter.calendarPeriod` | `string` |  |  |  |
| `spec.budgetFilter.customPeriod` | `GcpBillingBudgetCustomPeriod` |  |  |  |
| `spec.budgetFilter.customPeriod.startDate` | `GcpBillingBudgetDate` | yes |  |  |
| `spec.budgetFilter.customPeriod.startDate.year` | `int32` | yes |  |  |
| `spec.budgetFilter.customPeriod.startDate.month` | `int32` | yes |  |  |
| `spec.budgetFilter.customPeriod.startDate.day` | `int32` | yes |  |  |
| `spec.budgetFilter.customPeriod.endDate` | `GcpBillingBudgetDate` |  |  |  |
| `spec.budgetFilter.customPeriod.endDate.year` | `int32` | yes |  |  |
| `spec.budgetFilter.customPeriod.endDate.month` | `int32` | yes |  |  |
| `spec.budgetFilter.customPeriod.endDate.day` | `int32` | yes |  |  |
| `spec.thresholdRules` | `[]GcpBillingBudgetThresholdRule` |  |  |  |
| `spec.thresholdRules[].thresholdPercent` | `double` | yes |  |  |
| `spec.thresholdRules[].spendBasis` | `string` |  | `CURRENT_SPEND` |  |
| `spec.notifications` | `GcpBillingBudgetNotifications` |  |  |  |
| `spec.notifications.pubsubTopic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.notifications.monitoringNotificationChannels` | `[]string \| valueFrom` |  |  | GcpMonitoringNotificationChannel (`status.outputs.channel_name`) |
| `spec.notifications.disableDefaultIamRecipients` | `bool` |  |  |  |
| `spec.notifications.enableProjectLevelRecipients` | `bool` |  |  |  |
| `spec.notifications.schemaVersion` | `string` |  | `1.0` |  |
| `spec.ownershipScope` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.billingAccount

`string` · required

The Cloud Billing account the budget belongs to, as its ID
(`012345-6789AB-CDEF01`) or `billingAccounts/{id}`. Required; the
module normalizes to the resource name. Immutable.

- rule: billing_account must be a billing account ID such as 012345-6789AB-CDEF01, optionally prefixed with billingAccounts/
- rule: {"required":true}

### spec.displayName

`string`

Name shown in the Cloud Billing console, up to 60 characters. Defaults
to metadata.name.

- rule: {"string":{"maxLen":"60"}}

### spec.amount

`GcpBillingBudgetAmount` · required

The budgeted amount: a fixed amount or last period's spend. Required.

- rule: {"required":true}
- rule: exactly one of specified_amount or last_period_amount sets the budget

### spec.amount.specifiedAmount

`GcpBillingBudgetSpecifiedAmount`

A fixed amount in the billing account's currency.

### spec.amount.specifiedAmount.currencyCode

`string`

The 3-letter ISO 4217 currency code. Must match the billing account's
currency; unset takes the account's currency.

- rule: currency_code must be a 3-letter ISO 4217 code such as USD

### spec.amount.specifiedAmount.units

`int64`

The whole units of the amount in currency_code, e.g. 1000 for a budget
of one thousand US dollars when the currency is USD. 0 with a non-zero
nanos is a sub-unit budget.

- rule: {"int64":{"gte":"0"}}

### spec.amount.specifiedAmount.nanos

`int32`

Fractional part of the amount in nano units (10^-9), 0 to 999,999,999.
750,000,000 with units 1 is 1.75 in the currency.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.amount.lastPeriodAmount

`bool`

Budget the previous calendar period's spend: the budget for this period
is what was spent last period. Only with a calendar_period filter,
never with a custom_period.

### spec.budgetFilter

`GcpBillingBudgetFilter`

Which spend the budget counts. Omit to count everything the billing
account pays for, reset monthly.

- rule: calendar_period and custom_period are alternatives; set at most one
- rule: credit_types applies only with credit_types_treatment INCLUDE_SPECIFIED_CREDITS

### spec.budgetFilter.projects

`[]string | valueFrom`

Only usage from these projects counts: GcpProject references (resolved
to project numbers) or `projects/{number}` literals. Omitted means every
project the billing account pays for.

- references: GcpProject (`status.outputs.project_number`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_number}} -- a bare string does not parse

### spec.budgetFilter.resourceAncestors

`[]string | valueFrom`

Only usage under these folders or organizations counts: GcpFolder
references (resolved to `folders/{id}`) or `organizations/{id}`
literals.

- references: GcpFolder (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.budgetFilter.services

`[]string`

Only usage of these services counts, as `services/{service_id}` (the
id from the Cloud Billing catalog, e.g. services/6F81-5844-456A for
Compute Engine). Unset sends nothing.

### spec.budgetFilter.subaccounts

`[]string`

Only usage from these subaccounts counts, as
`billingAccounts/{account_id}`; naming the parent account includes
usage from the parent and every subaccount.

### spec.budgetFilter.labels

`map<string, string>`

Only usage carrying this label counts. One key with one value (the API
accepts a single pair). Unset sends nothing.

- rule: {"map":{"maxPairs":"1"}}

### spec.budgetFilter.creditTypesTreatment

`string` · optional (explicit presence)

How credits count against the budget: INCLUDE_ALL_CREDITS (the
default -- spend net of every credit), EXCLUDE_ALL_CREDITS (gross
spend), INCLUDE_SPECIFIED_CREDITS (net of the credit_types listed).

- default: `INCLUDE_ALL_CREDITS`
- rule: credit_types_treatment must be INCLUDE_ALL_CREDITS, EXCLUDE_ALL_CREDITS, or INCLUDE_SPECIFIED_CREDITS

### spec.budgetFilter.creditTypes

`[]string`

The credit types subtracted under INCLUDE_SPECIFIED_CREDITS, e.g.
COMMITTED_USAGE_DISCOUNT, SUSTAINED_USAGE_DISCOUNT, PROMOTION,
FREE_TIER.

### spec.budgetFilter.calendarPeriod

`string`

The recurring period the budget resets on: MONTH (the default when
neither period is set), QUARTER, or YEAR. Alternative to custom_period.

- rule: calendar_period must be MONTH, QUARTER, or YEAR

### spec.budgetFilter.customPeriod

`GcpBillingBudgetCustomPeriod`

A fixed date range instead of a recurring period. Alternative to
calendar_period; incompatible with last_period_amount.

### spec.budgetFilter.customPeriod.startDate

`GcpBillingBudgetDate` · required

- rule: {"required":true}

### spec.budgetFilter.customPeriod.startDate.year

`int32` · required

- rule: {"required":true,"int32":{"lte":9999,"gte":1}}

### spec.budgetFilter.customPeriod.startDate.month

`int32` · required

- rule: {"required":true,"int32":{"lte":12,"gte":1}}

### spec.budgetFilter.customPeriod.startDate.day

`int32` · required

- rule: {"required":true,"int32":{"lte":31,"gte":1}}

### spec.budgetFilter.customPeriod.endDate

`GcpBillingBudgetDate`

### spec.budgetFilter.customPeriod.endDate.year

`int32` · required

- rule: {"required":true,"int32":{"lte":9999,"gte":1}}

### spec.budgetFilter.customPeriod.endDate.month

`int32` · required

- rule: {"required":true,"int32":{"lte":12,"gte":1}}

### spec.budgetFilter.customPeriod.endDate.day

`int32` · required

- rule: {"required":true,"int32":{"lte":31,"gte":1}}

### spec.thresholdRules

`[]GcpBillingBudgetThresholdRule`

The thresholds that alert, each a share of the budget against current
or forecasted spend. Omit for a budget that only reports.

### spec.thresholdRules[].thresholdPercent

`double` · required

The share of the budget that triggers the alert, as a 1.0-based
fraction: 0.5 is 50%, 1.0 is 100%, 1.2 is 120% (over budget). Must be
greater than 0.

- rule: {"required":true,"double":{"gt":0}}

### spec.thresholdRules[].spendBasis

`string` · optional (explicit presence)

What the threshold is compared against: CURRENT_SPEND (actual spend so
far, the default) or FORECASTED_SPEND (Google's projection of the
period's total, which alerts before the money is spent).

- default: `CURRENT_SPEND`
- rule: spend_basis must be CURRENT_SPEND or FORECASTED_SPEND

### spec.notifications

`GcpBillingBudgetNotifications`

Where alerts go beyond the default administrator emails.

- rule: set pubsub_topic, monitoring_notification_channels, or both

### spec.notifications.pubsubTopic

`string | valueFrom`

The Pub/Sub topic budget notifications are published to (the full
budget state on every update, as JSON): a GcpPubSubTopic reference or
`projects/{project}/topics/{topic}`. The Billing service agent must be
a publisher on it.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.notifications.monitoringNotificationChannels

`[]string | valueFrom`

Cloud Monitoring notification channels (email, SMS, Slack, PagerDuty)
that receive threshold alerts: GcpMonitoringNotificationChannel
references or `projects/{project}/notificationChannels/{id}`. Up to 5.

- references: GcpMonitoringNotificationChannel (`status.outputs.channel_name`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMonitoringNotificationChannel, name: <that resource's name>, fieldPath: status.outputs.channel_name}} -- a bare string does not parse

### spec.notifications.disableDefaultIamRecipients

`bool`

Stop the default emails to the billing account's administrators and
users; only the channels above are notified.

### spec.notifications.enableProjectLevelRecipients

`bool`

Also email the owners of the projects the budget filters on (only
with a single-project filter).

### spec.notifications.schemaVersion

`string` · optional (explicit presence)

The JSON schema version of the Pub/Sub notification. Only "1.0"
exists; sent as such.

- default: `1.0`
- rule: schema_version must be 1.0

### spec.ownershipScope

`string`

Who may read the budget's data: ALL_USERS (anyone with billing.budgets
permissions on the account) or BILLING_ACCOUNT (only billing account
administrators). Unset lets Google apply its default.

- rule: ownership_scope must be ALL_USERS or BILLING_ACCOUNT

### spec.deletionPolicy

`string`

What destroy does to the budget:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the budget and its alerts are deleted
  "PREVENT" -- destroy FAILS; keeps a guardrail production spend
               depends on
  "ABANDON" -- the budget leaves management but keeps alerting

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `last_period_amount_needs_calendar_period`: last_period_amount budgets the previous calendar period, so the filter must not set custom_period

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBillingBudget, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The budget's resource name, `billingAccounts/{account}/budgets/{id}` -- the handle the Cloud Billing API and console address it by. |
| `status.outputs.budget_id` | `string` | The server-assigned budget id (the last segment of name). |
| `status.outputs.billing_account` | `string` | The billing account the budget belongs to, as `billingAccounts/{id}`. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.budgetFilter.projects` | GcpProject | `status.outputs.project_number` |
| `spec.budgetFilter.resourceAncestors` | GcpFolder | `status.outputs.name` |
| `spec.notifications.pubsubTopic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.notifications.monitoringNotificationChannels` | GcpMonitoringNotificationChannel | `status.outputs.channel_name` |

## See Also

- [Overview](../README.md)
