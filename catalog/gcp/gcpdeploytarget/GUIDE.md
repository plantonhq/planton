# GcpDeployTarget Guide

The judgment this guide protects: a target is where a release lands and who lands it. Pick one target type, keep the target in its pipeline's project and region, and give the deploy jobs an identity that can reach the workload.

## Choosing the target type

Exactly one type is set. `run` names a Cloud Run region as `projects/{project}/locations/{region}`; the project can differ from the target's own, which is how one delivery project deploys into separate staging and production projects. `gke` names a cluster by full name (a `GcpGkeCluster` reference gives its `cluster_id`). `anthosCluster` names a fleet membership, either a `GcpGkeFleetMembership`'s `name` or the `fleet_membership` a `GcpGkeCluster` exports when it joined a fleet through `fleetProject`. `multiTarget` lists other targets by bare ID (`target_id` references); a rollout deploys to all of them in parallel, which is how a multi-region production stage works. `customTarget` points at a `GcpDeployCustomTargetType` for anything Cloud Deploy does not deploy natively.

## Where it lives

The target's `location` and project are Cloud Deploy's, not the workload's: the delivery pipelines that deploy to it must be in the same project and region, and they name it by `target_id` in their stages. Both are fixed at creation; changing either replaces the target.

## Who deploys

Every render, deploy, verify, and hook job runs as a Cloud Build build. With no `executionConfigs`, they run on Cloud Build's default pool as the project's default compute service account, writing to a bucket Cloud Deploy creates in the target's region. Declare configurations to change that, grouping jobs by usage: each of `RENDER`, `DEPLOY`, `VERIFY`, `PREDEPLOY`, and `POSTDEPLOY` may appear in only one configuration, and once you declare any, `RENDER` and `DEPLOY` must both be covered. The service account needs `roles/clouddeploy.jobRunner`, plus what the deployment touches: for Cloud Run, `roles/run.developer` and `roles/iam.serviceAccountUser` on the service's runtime identity; for GKE, `roles/container.developer` on the cluster. Whoever creates releases needs `roles/iam.serviceAccountUser` on it. The artifact bucket (`gs://bucket` or `gs://bucket/path`) must be writable by it.

The pool can be chosen two ways that mean the same thing: the top-level `workerPool`, `serviceAccount`, and `artifactStorage`, or the `defaultPool` / `privatePool` blocks that carry the same fields. Use one form per configuration; the spec refuses both blocks at once.

## Private clusters

A GKE target whose control plane is reachable only privately needs `gke.internalIp` and jobs that run on a `GcpCloudBuildWorkerPool` peered into that network (`executionConfigs[].workerPool`). The DNS endpoint (`gke.dnsEndpoint`) avoids the private pool when the cluster has one enabled; the two cannot both be true.

## Lifecycle

Everything except `location` and `targetId` updates in place, including approval, parameters, and execution configurations. Destroy deletes the target: rollout history stays, but a pipeline stage that still names the target can no longer deploy, so in a chart the pipeline is destroyed first.
