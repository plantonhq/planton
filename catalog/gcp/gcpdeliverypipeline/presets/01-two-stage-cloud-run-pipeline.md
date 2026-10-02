# Two-Stage Cloud Run Pipeline

## Use Case

Ship a Cloud Run service through dev and then prod, with each release that succeeds on dev promoted to prod on its own.

## When to Use

- A first delivery pipeline for a Cloud Run service
- Teams that want hands-off promotion, with any approval gate set on the prod target itself

## What This Creates

- The Cloud Deploy API on the project
- A delivery pipeline in us-central1 with two stages, `dev` then `prod`, both using the standard strategy
- An automation that promotes a release to the next stage ten minutes after its rollout succeeds on `dev`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `serialPipeline.stages[].targetId` | `dev`, `prod` | Point at your own GcpDeployTarget resources (reference them with `valueFrom`). |
| `automations[].serviceAccount` | `deploy-automation@...` | The service account the automation runs as; it needs `roles/clouddeploy.operator` or narrower on the pipeline. |
| `automations[].rules[].promoteReleaseRule.wait` | `600s` | How long a release soaks on dev before it moves on. |
| `serialPipeline.stages[].strategy.standard.verify` | `false` | Set `true` to run your Skaffold verify job after each deploy. |
