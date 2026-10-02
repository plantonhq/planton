# Per-Project Monthly Guardrail

## Use Case

The budget that changes behavior: a fixed monthly amount for one project, alerting the team that owns it -- not the whole billing organization -- at half, at 90% of actual spend, and early when Google forecasts the month will overshoot. Declare one per production project or environment.

## When to Use

- Every production project, as the standing guardrail
- A new environment whose cost is not yet understood
- Any project a single team owns and should be paged about

## What This Creates

- A budget of USD 8,000 per calendar month on the `payments-prod` project (by reference, resolved to its project number), net of credits
- Thresholds at 50% and 90% of actual spend, and 100% of forecasted spend
- Alerts to the `payments-oncall` notification channel only (default administrator emails off)
- `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `billingAccount` | `012345-6789AB-CDEF01` | Your billing account ID. |
| `amount.specifiedAmount` | USD 8,000 | The monthly amount and the account's currency. |
| `budgetFilter.projects` | `payments-prod` | Your project (a `GcpProject` reference or `projects/{number}`); add a `labels` pair to narrow further. |
| `thresholdRules` | 50% / 90% / forecast 100% | Your alerting ladder; forecasted thresholds fire before the money is spent. |
| `notifications` | one channel | Add `pubsubTopic` (a `GcpPubSubTopic`) to drive automated response. |
| `creditTypesTreatment` | `INCLUDE_ALL_CREDITS` | `EXCLUDE_ALL_CREDITS` to watch gross usage, the number that survives a promotion ending. |

A budget notifies; to cap spend, route the Pub/Sub notification into a Cloud Function that disables billing on the project.
