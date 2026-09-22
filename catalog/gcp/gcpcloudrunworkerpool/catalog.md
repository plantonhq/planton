# GCP Cloud Run Worker Pool

Cloud Run's shape for work that is not a request: a pool of always-running container instances with no HTTP front door. Queue consumers, schedulers, event pullers, ML batch workers -- anything that should be up all the time and pull its own work. It keeps everything a Cloud Run service has for building the instance (containers and sidecars, volumes, direct VPC egress, GPUs, CMEK) and drops the request path: no port, no URL, no ingress, and instances scaled to a count you set or between bounds by a signal you drive, never by traffic.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Worker pool** -- a `cloudrunv2.WorkerPool` in the chosen region with the container template, scaling posture, instance split across revisions, binary authorization, and encryption settings
- **API enablement** -- the Cloud Run Admin API on the project, never disabled on destroy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/run.admin` and `roles/iam.serviceAccountUser` on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Resources

- **A container image** in Artifact Registry (or a public registry) the Cloud Run service agent can pull.
- **A `GcpServiceAccount`** for the runtime identity -- give real workers a least-privilege identity instead of the Compute default.
- **A VPC and subnet** when the worker reaches private resources through direct VPC egress.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Run Worker Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Queue Consumer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunWorkerPool
metadata:
  name: orders-worker
  org: acme-corp
  env: prod
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
  vpcAccess:
    networkInterfaces:
      - subnetwork:
          valueFrom:
            kind: GcpSubnetwork
            name: workers-subnet
            fieldPath: status.outputs.subnetwork_name
    egress: PRIVATE_RANGES_ONLY
  scaling:
    scalingMode: MANUAL
    manualInstanceCount: 2
```

```shell
planton apply -f worker-pool.yaml
```

This runs two instances of the orders worker, pulling from a Pub/Sub subscription and reaching private resources through direct VPC egress. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire `serviceAccount`, the subnet, secrets, and the cache or database the worker talks to through ValueFromRef in one InfraPipeline; deployment pipelines inject the built image into a container whose `image` is left blank.

## Key Configuration

These are the most important decisions when configuring a worker pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**How many instances** -- `scaling.scalingMode: MANUAL` (Google's default) runs exactly `manualInstanceCount` instances (`0` parks the pool); `AUTOMATIC` lets Cloud Run move between `minInstanceCount` and `maxInstanceCount` on a signal you drive, such as a queue-depth metric.

**The container** -- `containers[].image`, `env` (literal or Secret Manager), `resources.cpu` / `memory`, probes on the port a health listener binds, sidecars ordered by `dependsOn`.

**Private networking** -- `vpcAccess.networkInterfaces` for direct VPC egress (recommended) or a `connector`; `egress: ALL_TRAFFIC` routes public egress through the VPC too.

**Rollouts** -- every template change is a new revision; `instanceSplits` moves instances between revisions gradually.

**Destroy semantics** -- `deletionProtection` (default `true`) blocks destroys until flipped; `deletionPolicy: PREVENT` is a second guard.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** | `serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** | `encryptionKey` | `status.outputs.key_id` |
| **GcpServerlessVpcConnector** | `vpcAccess.connector` | `status.outputs.self_link` |
| **GcpVpcNetwork** / **GcpSubnetwork** | `vpcAccess.networkInterfaces[].network` / `.subnetwork` | `status.outputs.network_name` / `status.outputs.subnetwork_name` |
| **GcpCloudSql** | `volumes[].cloudSqlInstance.instances[]` | `status.outputs.connection_name` |
| **GcpGcsBucket** | `volumes[].gcs.bucket` | `status.outputs.bucket_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Full resource name | IAM bindings, monitoring |
| `worker_pool_name` | Bare name in GCP | Dashboards |
| `latest_ready_revision` | The revision serving instances | Rollout checks |
| `uid`, `location`, `project_id`, `observed_generation`, `etag` | Identity and reconciliation state | Audit |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Queue consumer** -- A fixed number of instances pulling from Pub/Sub over direct VPC egress. Start from the **Queue Consumer** preset.

**Autoscaled worker** -- Bounds instead of a count, a secret-backed env var, a liveness probe. Start from the **Autoscaled Worker** preset.

**GPU inference worker** -- An L4 GPU per instance, single-zone for cost, a startup probe with a long window. Start from the **GPU Inference Worker** preset.

## Works With

- [**GCP Cloud Run**](/cloud-catalog/gcp-cloud-run) -- the request-serving sibling
- [**GCP Cloud Run Job**](/cloud-catalog/gcp-cloud-run-job) -- run-to-completion work
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the runtime identity
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- direct VPC egress
- [**GCP Redis Cluster**](/cloud-catalog/gcp-redis-cluster) -- a cache a worker reaches privately
- [**GCP Pub/Sub Subscription**](/cloud-catalog/gcp-pub-sub-subscription) -- the queue a worker pulls from
