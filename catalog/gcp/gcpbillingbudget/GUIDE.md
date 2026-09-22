# GcpBillingBudget Guide

The judgment this guide protects: a budget is the only guardrail between a
runaway workload and a bill nobody approved, and it works only if the alert
reaches someone who can act. Decide the amount, the scope, and the
recipient together; each alone is decoration.

## A budget notifies, it does not stop anything

Crossing a threshold sends a notification -- to the billing
administrators by default, to `monitoringNotificationChannels` for people,
to a `pubsubTopic` for machines. Nothing pauses. The pattern that caps
spend is the Pub/Sub notification driving a Cloud Function that disables
billing on the project (Google documents it); declare the topic here and
the function separately. Without that, a budget is a smoke alarm with no
sprinkler -- still worth having, but know what it is.

## Forecast before actual

`thresholdPercent` compares against `CURRENT_SPEND` by default: you learn
you passed 90% after you passed it. `FORECASTED_SPEND` fires when Google's
projection says the period will overshoot -- days earlier, while a fix
still saves money. Declare both: a forecasted 100% to act early, an
actual 100% as the record.

## Scope one budget per team, not one per account

An unfiltered budget on the billing account is the finance view. The
guardrail that changes behavior is per project or environment: a
`budgetFilter` on the project (`GcpProject` reference, resolved to the
project number Google wants), a folder, or a label, with the team's own
channel as the recipient and `disableDefaultIamRecipients` so the whole
billing team is not paged for one team's overrun. `lastPeriodAmount`
gives a "no worse than last month" budget that follows the team's own
baseline instead of a number someone has to keep updating.

## Credits change the meaning of the number

`INCLUDE_ALL_CREDITS` (the default) measures what you pay after discounts
and promotions; `EXCLUDE_ALL_CREDITS` measures gross usage, the number that
survives when a promotion ends. A budget meant to catch a workload growing
should exclude credits; one meant to track the invoice should include
them.

## The budget belongs to the account

The deploying principal needs `roles/billing.costsManager` (or
`billing.admin`) on the billing account -- no project role reaches it. The
budget outlives the projects it filters on, so an environment teardown
leaves it alerting on nothing until it is removed; `deletionPolicy:
PREVENT` protects a production guardrail from a careless destroy.
