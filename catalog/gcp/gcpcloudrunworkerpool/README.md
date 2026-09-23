# GCP Cloud Run Worker Pool

A Cloud Run worker pool -- a pool of always-running container instances with no HTTP front door. Queue consumers, schedulers, event pullers, and background workers that do their own work instead of answering requests. It keeps the Cloud Run service's revision template (containers, volumes, direct VPC egress, GPUs) and revision model, but has no port, no URL, no ingress, no invoker IAM, and no request-driven autoscaling: instances scale MANUALLY to a count you set, or AUTOMATICALLY between bounds by a signal you drive. CPU is always allocated, so billing is per instance-hour.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Worker pool** -- a `cloud_run_v2_worker_pool` in your region with the container template, scaling posture, instance split, binary authorization, and encryption settings
- **API enablement** -- `run.googleapis.com` on the project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/run.admin` and `roles/iam.serviceAccountUser` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Resources

- **A container image** the Cloud Run service agent can pull -- Artifact Registry, or a public registry.
- **A VPC and subnet** when the worker reaches private resources (Memorystore, Cloud SQL private IP) through direct VPC egress.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunWorkerPool
metadata:
  name: orders-worker
spec:
  region: us-central1
  containers:
    - image: us-docker.pkg.dev/my-gcp-project/apps/orders-worker:1.4.0
      env:
        - name: SUBSCRIPTION
          value: projects/my-gcp-project/subscriptions/orders-events
      resources:
        cpu: "1"
        memory: 512Mi
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: orders-worker-sa
      fieldPath: status.outputs.email
  scaling:
    scalingMode: MANUAL
    manualInstanceCount: 2
```

```shell
planton apply -f worker-pool.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | `string` | Region of the pool. Immutable. |
| `containers` | `[]object` | The worker and any sidecars; at least one. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `workerPoolName` | `string` | `metadata.name` | Name in GCP. Immutable. |
| `description`, `labels`, `annotations` | — | — | Pool-object metadata. |
| `volumes` | `[]object` | — | Cloud SQL, Secret Manager, empty dir, GCS FUSE, NFS. |
| `serviceAccount` | `StringValueOrRef` | Compute default SA | Runtime identity (`GcpServiceAccount` reference). |
| `scaling` | `object` | Google's default (MANUAL) | `scalingMode` `MANUAL` with `manualInstanceCount`, or `AUTOMATIC` with `minInstanceCount` / `maxInstanceCount`. |
| `instanceSplits` | `[]object` | all on latest | `type` `INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST` / `_REVISION`, `revision`, `percent` (sum 100). |
| `encryptionKey` | `StringValueOrRef` | Google-managed | CMEK for the deployed images (`GcpKmsKey`). |
| `encryptionKeyRevocationAction` / `encryptionKeyShutdownDuration` | `string` | — | `PREVENT_NEW` or `SHUTDOWN` (with a seconds duration) if the key is revoked. |
| `revision`, `revisionLabels`, `revisionAnnotations` | — | — | Revision-level naming and metadata. |
| `vpcAccess` | `object` | — | Direct VPC egress (`networkInterfaces`) or a `connector`, with `egress`. |
| `nodeSelector` | `object` | — | `accelerator` (e.g. `nvidia-l4`) for GPU workers. |
| `gpuZonalRedundancyDisabled` | `bool` | `false` | Cheaper single-zone GPU capacity. |
| `launchStage` | `string` | `GA` | `ALPHA`, `BETA`, or `GA`. |
| `binaryAuthorization` | `object` | — | `useDefault` or `policy`, with `breakglassJustification`. |
| `deletionProtection` | `bool` | `true` | Destroy fails until set to `false`. Always sent explicitly. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

Per container: `name`, `image`, `command`, `args`, `env` (literal or `valueFromSecret`), `resources` (`cpu`, `memory`), `volumeMounts`, `workingDir`, `startupProbe` (HTTP / TCP / gRPC), `livenessProbe` (HTTP / gRPC), `dependsOn`.

### Validation Rules

- **`manualInstanceCount`** only under `MANUAL` (or unset); **`minInstanceCount` / `maxInstanceCount`** only under `AUTOMATIC`, and min ≤ max.
- **`REVISION`** splits name a `revision`; **`LATEST`** splits do not.
- **`encryptionKeyRevocationAction`** requires `encryptionKey`; **`encryptionKeyShutdownDuration`** requires `SHUTDOWN`.
- An env var takes a `value` or a `valueFromSecret`, not both; a volume has exactly one source; `vpcAccess` uses a connector or network interfaces, not both.
- Probe `timeoutSeconds` ≤ `periodSeconds`; a startup window (`failureThreshold` × `periodSeconds`) ≤ 240 s; **at most one probe `httpHeaders` entry** (see Important Notes).

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Full resource name |
| `worker_pool_name` | `string` | Bare name in GCP |
| `uid` | `string` | Server-generated identifier |
| `location`, `project_id` | `string` | Where it lives |
| `latest_created_revision`, `latest_ready_revision` | `string` | The rollout's two pointers |
| `observed_generation`, `etag` | `string` | Reconciliation state |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **No requests reach a worker pool.** There is no port, no URL, and no ingress; the worker pulls its own work (Pub/Sub, a queue, a schedule). Health probes must name the port a health listener binds.
- **Scaling is yours to drive.** `MANUAL` pins a count (`0` parks the pool without deleting it); `AUTOMATIC` moves between bounds on a signal you supply, never on traffic.
- **CPU is always allocated**, so an idle instance bills like a busy one.
- **Immutable:** `region`, `workerPoolName`, and a container's `dependsOn`. Everything else rolls a new revision in place.
- **Two levers wait on the Pulumi SDK.** A container's `sandboxLauncher` flag and more than one probe `httpHeaders` entry are not yet in the pinned pulumi-gcp SDK's worker-pool types; both are held out of the spec so both engines accept the same manifests, and return when the SDK carries them (re-evaluated at pulumi-gcp v10 GA).

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpCloudRun** -- the request-serving sibling; **GcpCloudRunJob** -- run-to-completion work
- **GcpServiceAccount** -- the runtime identity
- **GcpVpcNetwork** / **GcpSubnetwork** -- direct VPC egress; **GcpServerlessVpcConnector** -- the connector alternative
- **GcpRedisCluster**, **GcpCloudSql**, **GcpPubSubSubscription** -- what a worker typically talks to
- **GcpSecretManagerSecret** -- secret env vars and volumes; **GcpKmsKey** -- CMEK

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
