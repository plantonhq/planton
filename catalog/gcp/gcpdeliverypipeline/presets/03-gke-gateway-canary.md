# GKE Gateway Canary

## Use Case

Roll a GKE workload out to prod in phases you define, with Cloud Deploy shifting traffic through a Gateway API HTTPRoute.

## When to Use

- Services on GKE that already sit behind a Gateway API `HTTPRoute`
- Releases that need their own phase names, percentages, and verify steps

## What This Creates

- The Cloud Deploy API on the project
- A delivery pipeline with a `gke-staging` stage (standard, with verify) and a `gke-prod` stage using a custom canary: 10%, then 50%, then stable, verified at each canary step
- Traffic split by Cloud Deploy through the `api-route` HTTPRoute in front of the `api` Service and Deployment

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `...gatewayServiceMesh.httpRoute` | `api-route` | The HTTPRoute that fronts your Service. |
| `...gatewayServiceMesh.service` / `deployment` | `api` | Your workload's Service and Deployment names. |
| `...customCanaryDeployment.phaseConfigs` | `10`, `50`, `100` | Your own phases; the last one must be 100 (stable). |
| `...gatewayServiceMesh.stableCutbackDuration` | `300s` | How long the old version keeps traffic while the route cuts back to stable. |
