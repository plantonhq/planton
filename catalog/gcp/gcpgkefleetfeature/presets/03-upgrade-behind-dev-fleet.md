# Upgrade Behind the Dev Fleet

## Use Case

Let the production fleet take a GKE upgrade only after it has soaked for seven days in the dev fleet.

## When to Use

- Separate dev and prod fleets on the same release channel
- Catching upgrade regressions in dev before they reach prod

## What This Creates

- The Fleet API on the prod fleet's host project
- The clusterupgrade feature on the prod fleet, consuming upgrades the dev fleet completed

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `clusterupgrade.upstreamFleets` | the `dev-fleet` reference | The fleet upgrades soak in first (Google accepts one). |
| `clusterupgrade.postConditions.soaking` | `604800s` (7 days) | Longer or shorter soak, at most 30 days. |
