# Weekend Freeze

## Use Case

Stop new production rollouts, approvals, and phase advances all weekend, while leaving rollbacks open so a bad Friday release can still be undone.

## When to Use

- A team that does not staff weekends
- A first deploy policy for production targets

## What This Creates

- The Cloud Deploy API on the project
- A deploy policy that blocks CREATE, APPROVE, and ADVANCE on targets labeled `env: prod` every Saturday and Sunday, New York time

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `location` | `us-central1` | The region of your pipelines and targets. |
| `rules[].rolloutRestriction.timeWindows.timeZone` | `America/New_York` | The time zone your team works in. |
| `rules[].rolloutRestriction.actions` | CREATE, APPROVE, ADVANCE | Empty blocks every action, including rollbacks. |
| `selectors[].target.labels` | `env: prod` | The labels your production targets carry, or a target reference by ID. |
