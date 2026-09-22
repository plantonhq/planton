# No Worse Than Last Month

## Use Case

A guardrail that follows the team's own baseline: last calendar month's spend is this month's budget, so growth -- not a stale number -- is what alerts. Scoped to a folder so every project under it counts, measured gross of credits so a promotion ending does not masquerade as growth, and published to a Pub/Sub topic so a function can respond.

## When to Use

- A folder of projects whose spend is expected to stay flat
- Detecting growth without maintaining a fixed amount per team
- Feeding a cost-anomaly pipeline (Cloud Function, Slack bridge, ticketing) rather than emailing people

## What This Creates

- A budget equal to last month's spend across the `data-platform` folder (by reference), gross of credits, reset monthly
- Thresholds at a forecasted 100% and an actual 120%
- Every budget update published to the `budget-events` Pub/Sub topic; default administrator emails off
- `ownershipScope: BILLING_ACCOUNT`, `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `budgetFilter.resourceAncestors` | the `data-platform` folder | Your folder (a `GcpFolder` reference) or `organizations/{id}`; or switch to `projects` for one project. |
| `thresholdRules` | forecast 100% / actual 120% | Tighten the actual threshold once the baseline is trusted. |
| `notifications.pubsubTopic` | `budget-events` | Your topic; the Billing service agent must be a publisher on it. Add `monitoringNotificationChannels` for people. |
| `ownershipScope` | `BILLING_ACCOUNT` | `ALL_USERS` to let every principal with budget permissions read it. |

`lastPeriodAmount` requires a calendar period (`MONTH`, `QUARTER`, `YEAR`); a `customPeriod` budget needs a fixed `specifiedAmount`.
