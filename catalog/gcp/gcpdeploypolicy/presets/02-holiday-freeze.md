# Holiday Freeze

## Use Case

Freeze every pipeline in the location for people over the year-end holidays, and stop new rollouts every Friday afternoon, while the pipelines' own automations keep running.

## When to Use

- A company-wide change freeze with dates
- Quiet hours before the weekend

## What This Creates

- The Cloud Deploy API on the project
- A deploy policy with a dated year-end window that blocks every action invoked by a person, and a weekly Friday 15:00-24:00 window that blocks new rollouts, both Berlin time, on every pipeline in the location

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `rules[0].rolloutRestriction.timeWindows.oneTimeWindows` | December 19 17:00 to January 4 09:00 | This year's freeze dates. |
| `rules[0].rolloutRestriction.invokers` | `USER` | Add `DEPLOY_AUTOMATION` to stop automated promotions too. |
| `rules[1].rolloutRestriction.timeWindows.weeklyWindows` | Friday 15:00-24:00 | Your quiet hours. |
| `selectors[].deliveryPipeline.id` | `*` | A `GcpDeliveryPipeline` reference to freeze one pipeline. |
