# GCP Billing Budget

Sets a spending guardrail on a Cloud Billing account: a budgeted amount for a period, the projects, folders, services, or labels the spend is measured over, the thresholds that alert, and where the alerts go. A budget notifies -- by email, Cloud Monitoring channel, or Pub/Sub -- and a Pub/Sub notification into a Cloud Function is how teams cap spend automatically. Declare one budget per project or environment for the per-team guardrail, or one unfiltered budget for the whole bill.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Billing Budget** -- a `billing.Budget` on the billing account with its amount (fixed or last period's spend), filter (projects, folders, services, subaccounts, labels, credit treatment, period), threshold rules, notification rule, and ownership scope

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose principal holds `roles/billing.costsManager` on the billing account (a project role does not reach the account). Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Billing

- **The billing account ID** from the Cloud Billing console.
- **Notification targets** -- a `GcpPubSubTopic` or `GcpMonitoringNotificationChannel` resources when alerts should reach more than the billing administrators.

## Deploy

### Console

Open the deployment store, find **GCP Billing Budget**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Per-Project Monthly Guardrail** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBillingBudget
metadata:
  name: platform-prod-monthly
  org: acme-corp
  env: prod
spec:
  billingAccount: 012345-6789AB-CDEF01
  amount:
    specifiedAmount:
      currencyCode: USD
      units: 5000
  budgetFilter:
    projects:
      - valueFrom:
          kind: GcpProject
          name: platform-prod
          fieldPath: status.outputs.project_number
    calendarPeriod: MONTH
  thresholdRules:
    - thresholdPercent: 0.9
    - thresholdPercent: 1.0
      spendBasis: FORECASTED_SPEND
  notifications:
    pubsubTopic:
      valueFrom:
        kind: GcpPubSubTopic
        name: budget-alerts
        fieldPath: status.outputs.topic_id
    disableDefaultIamRecipients: true
```

```shell
planton apply -f billing-budget.yaml
```

This budgets USD 5,000 a month for one project, alerts at 90% of actual spend and at a forecasted 100%, and publishes every update to a Pub/Sub topic. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the budget to the project, topic, and notification channels deployed in the same InfraPipeline.

## Key Configuration

These are the most important decisions when configuring a budget. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Amount** -- `specifiedAmount` (a fixed amount in the account's currency) or `lastPeriodAmount: true` (last period's spend, a "no worse than last month" guardrail; calendar periods only).

**Scope** -- `budgetFilter.projects` (`GcpProject` references), `resourceAncestors` (folders and organizations), `services`, `labels`, and the period (`calendarPeriod` MONTH / QUARTER / YEAR or a `customPeriod`). `creditTypesTreatment: EXCLUDE_ALL_CREDITS` measures gross usage, the number that survives a promotion ending.

**Thresholds** -- `thresholdPercent` as a fraction (1.0 = 100%); `spendBasis: FORECASTED_SPEND` fires when Google projects an overshoot, before the money is spent.

**Recipients** -- `notifications.pubsubTopic` for automation, `monitoringNotificationChannels` for people, `disableDefaultIamRecipients` to keep the billing team out of one team's alerts.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `budgetFilter.projects[]` | `status.outputs.project_number` |
| **GcpFolder** | `budgetFilter.resourceAncestors[]` | `status.outputs.name` |
| **GcpPubSubTopic** | `notifications.pubsubTopic` | `status.outputs.topic_id` |
| **GcpMonitoringNotificationChannel** | `notifications.monitoringNotificationChannels[]` | `status.outputs.channel_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | `billingAccounts/{account}/budgets/{id}` | Audit, the Cloud Billing API |
| `budget_id` | The server-assigned id | Console links |
| `billing_account` | `billingAccounts/{id}` | Grouping budgets by account |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Per-project monthly guardrail** -- A fixed monthly amount on one project with forecasted and actual thresholds and a team channel. Start from the **Per-Project Monthly Guardrail** preset.

**No worse than last month** -- Last period's spend as the budget on a folder, gross of credits, into a Pub/Sub topic for automated response. Start from the **No Worse Than Last Month** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project), [**GCP Folder**](/cloud-catalog/gcp-folder) -- the scope a budget filters on
- [**GCP Pub/Sub Topic**](/cloud-catalog/gcp-pub-sub-topic) -- where automation receives budget notifications
- [**GCP Monitoring Notification Channel**](/cloud-catalog/gcp-monitoring-notification-channel) -- where people receive threshold alerts
