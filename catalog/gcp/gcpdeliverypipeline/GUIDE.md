# GcpDeliveryPipeline Guide

The judgment this guide protects: a pipeline is the promotion flow, not the release. It names targets by bare ID in its own project and region, it says how each rollout runs, and its automations act with a service account's authority. Destroying it destroys the release history under it.

## Stages and targets

Each stage names a target by its last name segment (`web-prod`, never `projects/.../targets/web-prod`). Google looks it up in the pipeline's project and region, so a `GcpDeployTarget` reference only works when that target lives there too. A pipeline can be created before its targets exist -- Google reports the missing ones in the pipeline's condition -- but a release cannot roll out to a missing stage. Stage `profiles` pick Skaffold profiles for rendering; `deployParameters` substitute values into the rendered manifests, optionally only on child targets whose labels match.

## Choosing a strategy

Set one of `standard` or `canary` per stage; with neither, the stage gets a standard deploy without verification. Standard deploys everything at once. `verify` runs `skaffold verify` (or `verifyConfig`'s container tasks) afterwards, `predeploy` and `postdeploy` run Skaffold custom actions or container tasks (one or the other, not both), and `analysis` watches the rollout for a duration, failing it when a Cloud Monitoring alert policy fires or a custom check's container fails.

Canary is for Cloud Run and GKE targets. `canaryDeployment` lists ascending percentages, the same jobs for every phase; `customCanaryDeployment` lets each phase choose its own percentage, profiles, and jobs. `runtimeConfig` says how traffic moves: on Cloud Run, Cloud Deploy rewrites revision traffic, and Google requires `automaticTrafficControl: true` for a `canaryDeployment`; on GKE, choose `gatewayServiceMesh` (a Gateway API HTTPRoute, optionally deployed to more clusters through `routeDestinations`) or `serviceNetworking` (Pod counts behind a Service). The canary paths' predeploy and postdeploy take `actions` only: the pinned 8.3 provider schema declares no `tasks` argument on those paths (only on the standard strategy's), so neither engine can send them.

## Automations

An automation applies rules to the targets its selector picks (bare IDs or `*`). `promoteReleaseRule` promotes a release after it succeeds, to `@next` or a named stage, optionally after a `wait`. `advanceRolloutRule` moves a healthy canary to its next phase. `repairRolloutRule` lists repair phases in order -- each a `retry` (attempts, wait, backoff) or a `rollback` -- and Google requires at least one. `timedPromoteReleaseRule` promotes on a cron schedule in a time zone. Rule IDs are unique within an automation; Google caps a pipeline at 250 rules across its automations.

Every automation runs as the service account in `serviceAccount`. Two grants make it work: the identity deploying this block needs `iam.serviceAccounts.actAs` on that account (`roles/iam.serviceAccountUser`), and the account itself needs permission to create releases and rollouts (`roles/clouddeploy.operator` or narrower) plus `actAs` on the targets' execution service accounts. Grant them through `GcpServiceAccount.iamMembers` or project IAM before the automation first fires.

## Lifecycle

`location` and `deliveryPipelineId` are fixed at creation; stages, strategies, labels, and automations update in place, and an automation's ID is its identity. The provider always deletes the pipeline with `force=true`, which removes every release, rollout, and automation under it -- there is no softer option -- so set `deletionPolicy: PREVENT` on a pipeline whose history is worth keeping. Cloud Deploy bills per active delivery pipeline (its pricing page defines active, and the first one per billing account is free); a pipeline nobody releases to is not active. Render, deploy, and verify jobs bill separately as Cloud Build minutes.
