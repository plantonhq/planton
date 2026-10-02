# Cloud Run Canary with Repair

## Use Case

Roll a Cloud Run service out to prod in steps -- 25% of traffic, then 50%, then all of it -- and let Cloud Deploy move through the steps and undo a bad release on its own.

## When to Use

- Production services where a bad release must reach only part of the traffic first
- Teams that want canary phases to advance after a soak, and failed deploys retried and then rolled back, without someone watching

## What This Creates

- The Cloud Deploy API on the project
- A delivery pipeline with a `staging` stage (standard, with verify) and a `prod` stage using a Cloud Run canary at 25% and 50%, with verify on each phase
- An automation on `prod` that advances each canary phase after 15 minutes, retries a failed deploy or verify job twice with exponential backoff, and then rolls back

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `...canaryDeployment.percentages` | `25`, `50` | The traffic steps before stable; each must be below 100. |
| `...advanceRolloutRule.wait` | `900s` | How long each canary phase soaks before it advances. |
| `...repairPhases[].retry.attempts` | `"2"` | How many times a failed job is retried before the rollback. |
| `...runtimeConfig.cloudRun.automaticTrafficControl` | `true` | Must stay `true` for this canary shape (Google's rule); switch to a custom canary to manage traffic yourself. |
