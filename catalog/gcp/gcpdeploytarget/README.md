# GCP Deploy Target

Declares a Cloud Deploy target: one place a delivery pipeline's stage deploys a release to -- a GKE cluster, a fleet-registered cluster, a Cloud Run region, a group of other targets, or a custom target type -- together with how Cloud Deploy runs the render, deploy, verify, and hook jobs for it (service account, artifact bucket, private worker pool, timeout).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `clouddeploy.googleapis.com` on the target's project (never disabled on destroy)
- **Target** -- one `clouddeploy_target`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/clouddeploy.admin` (or the target permissions in `iac/permissions.yaml`) on the target's project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

None. A Cloud Run target needs only a project and a region.

### Optional Dependencies

- **`GcpGkeCluster`** -- the cluster of a `gke` target or an associated entity (`cluster_id`), or a fleet-joined cluster of an `anthosCluster` target (`fleet_membership`).
- **`GcpGkeFleetMembership`** -- an explicitly registered cluster (`name`).
- **`GcpDeployTarget`** -- the children of a `multiTarget` (`target_id`).
- **`GcpDeployCustomTargetType`** -- the type of a `customTarget` (`name`).
- **`GcpServiceAccount`**, **`GcpGcsBucket`**, **`GcpCloudBuildWorkerPool`** -- the execution service account, artifact bucket, and private pool.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployTarget
metadata:
  name: prod
spec:
  projectId:
    value: acme-delivery
  location: us-central1
  description: Production on Cloud Run
  requireApproval: true
  run:
    location: projects/acme-prod/locations/us-central1
  executionConfigs:
    - usages:
        - RENDER
        - DEPLOY
      serviceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: prod-deployer
          fieldPath: status.outputs.email
```

```shell
planton apply -f deploy-target.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The target's region, the region of the pipelines that deploy to it. Immutable. |
| one target type | block | Exactly one of `gke`, `anthosCluster`, `run`, `multiTarget`, `customTarget`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The target's project (`GcpProject` ref). Immutable. |
| `targetId` | `string` | `metadata.name` | The ID pipeline stages name. Immutable. |
| `description` | `string` | -- | Up to 255 characters. |
| `labels` | `map` | -- | Labels; deploy policies and automations can select targets by them. |
| `annotations` | `map` | -- | AIP-128 annotations. |
| `requireApproval` | `bool` | `false` | Every rollout waits for an approval. |
| `deployParameters` | `map` | -- | Values substituted at `# from-param:` markers in rendered manifests. |
| `gke` | block | -- | `cluster` (full name or `GcpGkeCluster` ref), `internalIp`, `dnsEndpoint`, `proxyUrl`. |
| `anthosCluster` | block | -- | `membership` (full name, `GcpGkeFleetMembership` or `GcpGkeCluster` ref). |
| `run` | block | -- | `location` as `projects/{project}/locations/{region}`. |
| `multiTarget` | block | -- | `targetIds` (bare IDs or `GcpDeployTarget` refs; at least one). |
| `customTarget` | block | -- | `customTargetType` (full name or `GcpDeployCustomTargetType` ref). |
| `associatedEntities[]` | list | -- | `entityId`; `gkeClusters[]` (`cluster`, `internalIp`, `proxyUrl`); `anthosClusters[]` (`membership`). |
| `executionConfigs[]` | list | Cloud Build default pool | `usages`; `workerPool`, `serviceAccount`, `artifactStorage`, `executionTimeout`, `verbose`; or the `defaultPool` / `privatePool` block forms. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one target type.
- `gke.dnsEndpoint` and `gke.internalIp` are not both true.
- Literal cluster, membership, custom target type, and worker pool values are full resource names; `run.location` is `projects/{project}/locations/{region}`; child target IDs are bare IDs; artifact storage is a `gs://` location.
- Entity IDs are unique and follow Google's ID rule.
- Each usage (`RENDER`, `DEPLOY`, `VERIFY`, `PREDEPLOY`, `POSTDEPLOY`) appears in at most one execution configuration, and declared configurations cover `RENDER` and `DEPLOY`; at most one of `defaultPool` or `privatePool` per configuration.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/targets/{target_id}` |
| `target_id` | `string` | The target's ID, what pipeline stages and multi-targets reference |
| `uid` | `string` | Google's unique identifier for the target |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Same project and region as the pipeline.** A delivery pipeline deploys only to targets in its own project and region; the workload itself can live elsewhere (`run.location`, a cluster in another project).
- **The execution service account does the deploying.** It needs `roles/clouddeploy.jobRunner`, plus access to what it deploys to; the default is the project's compute service account.
- **Private clusters need a private pool.** Reaching a GKE control plane on its internal IP means running the jobs on a `GcpCloudBuildWorkerPool` with a route to that network.
- **Destroy** deletes the target; a pipeline stage that still names it can no longer deploy.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpDeliveryPipeline** -- the pipeline whose stages deploy to the target
- **GcpDeployCustomTargetType** -- the type behind a custom target
- **GcpDeployPolicy** -- rollout restrictions that select targets
- **GcpCloudBuildWorkerPool** -- private machines for the target's jobs
- **GcpGkeCluster** / **GcpGkeFleetMembership** -- clusters to deploy to

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
