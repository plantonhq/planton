# GCP Billing Budget

Sets a spending guardrail on a Cloud Billing account: a budgeted amount for a period, the projects, folders, services, or labels the spend is measured over, the thresholds that alert, and where the alerts go. A budget never stops spending by itself — it notifies; a Pub/Sub notification into a Cloud Function is how teams cap spend automatically. Declare one budget per project or environment for the per-team guardrail pattern, or one unfiltered budget per billing account for the whole bill.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Billing budget** -- the `billing_budget` on the billing account with its amount, filter, threshold rules, and notification rule

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose principal holds `roles/billing.costsManager` (or `roles/billing.admin`) on the billing account; a project-level role is not enough, because the budget lives on the account.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Billing

- **The billing account ID** -- from the Cloud Billing console (`012345-6789AB-CDEF01`).
- **Notification targets** -- a `GcpPubSubTopic` the Billing service agent may publish to, or `GcpMonitoringNotificationChannel` resources, when alerts should go beyond the default administrator emails.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBillingBudget
metadata:
  name: platform-prod-monthly
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
  thresholdRules:
    - thresholdPercent: 0.9
    - thresholdPercent: 1.0
      spendBasis: FORECASTED_SPEND
```

```shell
planton apply -f billing-budget.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `billingAccount` | `string` | The billing account ID (`012345-6789AB-CDEF01`) or `billingAccounts/{id}`. Immutable. |
| `amount` | `object` | Exactly one of `specifiedAmount { currencyCode, units, nanos }` or `lastPeriodAmount: true`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Console name, up to 60 characters. |
| `budgetFilter` | `object` | everything, monthly | `projects` (`GcpProject` references or `projects/{number}`), `resourceAncestors` (`GcpFolder` references or `organizations/{id}`), `services`, `subaccounts`, `labels` (one pair), `creditTypesTreatment` (`INCLUDE_ALL_CREDITS` default, `EXCLUDE_ALL_CREDITS`, `INCLUDE_SPECIFIED_CREDITS` + `creditTypes`), `calendarPeriod` (`MONTH`, `QUARTER`, `YEAR`) or `customPeriod { startDate, endDate }`. |
| `thresholdRules` | `[]object` | — | `thresholdPercent` (> 0; 1.0 = 100%) and `spendBasis` (`CURRENT_SPEND` default, `FORECASTED_SPEND`). |
| `notifications` | `object` | administrator emails | `pubsubTopic` (`GcpPubSubTopic` reference), `monitoringNotificationChannels` (up to 5 `GcpMonitoringNotificationChannel` references), `disableDefaultIamRecipients`, `enableProjectLevelRecipients`, `schemaVersion` (`1.0`). At least one target. |
| `ownershipScope` | `string` | Google's default | `ALL_USERS` or `BILLING_ACCOUNT`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- **`amount`** carries exactly one arm; **`lastPeriodAmount`** is rejected with a `customPeriod`.
- **`calendarPeriod`** and **`customPeriod`** are alternatives; **`creditTypes`** needs `INCLUDE_SPECIFIED_CREDITS`.
- **`notifications`** must name a topic, a channel, or both.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `billingAccounts/{account}/budgets/{id}` |
| `budget_id` | `string` | The server-assigned id |
| `billing_account` | `string` | `billingAccounts/{id}` |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **A budget notifies; it does not cap.** To stop spending, route the Pub/Sub notification into a Cloud Function that disables billing on the project.
- **Forecasted thresholds warn early.** `spendBasis: FORECASTED_SPEND` fires when Google projects the period will overshoot, before the money is spent.
- **The budget is account-scoped.** The deploying principal needs a billing-account role, and the budget survives the projects it filters on.
- **Cost**: budgets are free.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpProject](/docs/catalog/gcp/gcpproject) — the projects a budget filters on
- [GcpFolder](/docs/catalog/gcp/gcpfolder) — the folders a budget filters on
- [GcpPubSubTopic](/docs/catalog/gcp/gcppubsubtopic) — where automation receives budget notifications
- [GcpMonitoringNotificationChannel](/docs/catalog/gcp/gcpmonitoringnotificationchannel) — where people receive threshold alerts

## Additional Resources

- [Create, edit, or delete budgets and budget alerts](https://cloud.google.com/billing/docs/how-to/budgets)
- [Manage programmatic budget alert notifications](https://cloud.google.com/billing/docs/how-to/budgets-programmatic-notifications)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
