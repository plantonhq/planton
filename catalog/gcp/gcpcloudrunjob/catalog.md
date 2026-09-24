# GCP Cloud Run Job

Deploys a run-to-completion batch workload on Google Cloud Run v2: a task template (containers, volumes, networking, hardware) plus an execution model (task count, parallelism, per-attempt timeout, retries). Each run — an "execution" — stamps out the tasks and exits; there is no endpoint and no traffic, and the only health check is an optional startup probe gating task start. Every composition seam — the GCP project, the runtime service account, the KMS key, Cloud SQL instances, GCS buckets, VPC networks and subnetworks — accepts ValueFromRef wiring.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Cloud Run v2 Job** -- a job definition in the specified GCP project and region with the provided task template and execution model
- **Task Template** -- the containers every task runs (images, env vars, resources), with sidecar startup ordering and volume mounts
- **Volumes** -- Cloud SQL Unix sockets, Secret Manager file projections, ephemeral scratch space, GCS FUSE mounts, or NFS shares
- **VPC Access** -- created only when `template.vpcAccess` is configured; Direct VPC Egress network interfaces or a Serverless VPC Access connector
- **GCP Labels and Annotations** -- billing-visible labels and tool-preserved annotations

The resource owns the job DEFINITION, not individual runs — trigger executions with `gcloud run jobs execute`, Cloud Scheduler (cron), or Eventarc.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the job will be created. Provide the project ID directly or reference a GcpProject Cloud Resource via ValueFromRef.
- **Artifact Registry or container registry** with the task image pushed and accessible to the Cloud Run service agent.
- **Cloud Run Admin API** enabled in the target project.
- **VPC network and subnetwork** (if using Direct VPC Egress) -- the subnetwork must be in the job's region with address headroom for `parallelism` concurrent tasks.
- **Secret Manager secrets** (if referenced through `valueFromSecret`) -- the runtime service account needs `roles/secretmanager.secretAccessor` on each. A variable's `secretValue` needs neither: the component creates its secret and grants that access itself.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Run Job**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Batch ETL — basic** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunJob
metadata:
  name: nightly-etl
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  region: us-central1
  template:
    containers:
      - image: us-docker.pkg.dev/acme-prod-12345/registry/nightly-etl:1.0.0
        resources:
          cpu: "2"
          memory: 4Gi
    timeoutSeconds: 3600
  taskCount: 20
  parallelism: 5
```

```shell
planton apply -f cloud-run-job.yaml
```

This creates a job whose executions run 20 tasks (5 at a time), each with 2 vCPU / 4Gi and a one-hour per-attempt budget. A Stack Job tracks the provisioning in real time. Trigger a run:

```shell
gcloud run jobs execute nightly-etl --region us-central1
```

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the job to resources deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  template:
    serviceAccount:
      valueFrom:
        kind: GcpServiceAccount
        name: etl-runtime
        fieldPath: status.outputs.email
    volumes:
      - name: warehouse
        cloudSqlInstance:
          instances:
            - valueFrom:
                kind: GcpCloudSql
                name: analytics-db
                fieldPath: status.outputs.connection_name
```

The InfraPipeline resolves the dependency graph, deploys the project, service account, and database first, then provisions the job.

## Key Configuration

These are the most important decisions when configuring a Cloud Run job. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Task template** -- `template.containers` holds the task container plus any sidecars (a Cloud SQL Auth Proxy, a collector), ordered with `dependsOn`. Success is exit code 0; an optional `startupProbe` (HTTP, TCP, or gRPC against the container's declared `ports`) gates task start within a 240-second window — a container that never passes is shut down and retried per `maxRetries`. Tasks learn their shard from `$CLOUD_RUN_TASK_INDEX` and `$CLOUD_RUN_TASK_COUNT`.

**Run on deploy** -- The execution tokens make a deploy itself trigger a run: `startExecutionToken` counts the job READY as soon as the triggered execution starts, while `runExecutionToken` waits for it to complete — deploy-and-verify semantics for migrations and one-shot setup work. Change the token on a later update to trigger another run; set at most one of the two.

**Execution model** -- `taskCount` is the work units per run (GCP default 1); `parallelism` caps concurrency (unset = maximum possible, and it can never exceed the task count). `template.timeoutSeconds` bounds each ATTEMPT (up to 24h; every retry gets a fresh budget) and `template.maxRetries` is the per-task retry budget — set 0 for non-idempotent work.

**Volumes** -- the same five sources as the Cloud Run service: Cloud SQL sockets, Secret Manager files, scratch space (`emptyDir` — memory-backed scratch counts against the task's memory limit), GCS FUSE, and NFS. GCS/NFS require `EXECUTION_ENVIRONMENT_GEN2`.

**VPC connectivity** -- `template.vpcAccess.networkInterfaces` for Direct VPC Egress (recommended) or `template.vpcAccess.connector`, never both. At parallelism N, N concurrent tasks draw N subnetwork addresses.

**Identity** -- `template.serviceAccount` is the identity every task exercises. Batch jobs usually touch the most sensitive surfaces (whole tables, whole buckets) — give production jobs a dedicated least-privilege account.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** (optional) | `template.serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** (optional) | `template.encryptionKey` | `status.outputs.key_id` |
| **GcpCloudSql** (optional) | `template.volumes[].cloudSqlInstance.instances` | `status.outputs.connection_name` |
| **GcpGcsBucket** (optional) | `template.volumes[].gcs.bucket` | `status.outputs.bucket_id` |
| **GcpServerlessVpcConnector** (optional) | `template.vpcAccess.connector` | `status.outputs.self_link` |
| **GcpVpcNetwork** (optional) | `template.vpcAccess.networkInterfaces[].network` | `status.outputs.network_name` |
| **GcpSubnetwork** (optional) | `template.vpcAccess.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `job_name` | Name of the Cloud Run job in GCP | Scheduler/Eventarc triggers, gcloud automation |
| `location` | Region the job runs in | Regional composition |
| `uid` | Server-assigned unique identifier | Audit and correlation |
| `latest_created_execution` | Most recently created execution (empty until the first run) | Run tracking, pipeline observability |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Batch ETL** -- a single task (or a small sharded set) loading data on a schedule, triggered by Cloud Scheduler. Start from the **Batch ETL — basic** preset.

**Parallel fan-out** -- many tasks at capped parallelism, each selecting its slice from the task index — dataset processing without a queue. Start from the **Parallel VPC cleanup** preset.

**GPU batch inference** -- `template.nodeSelector.accelerator` gives every task a GPU; the zonal-redundancy opt-out (`gpuZonalRedundancyDisabled`) trades single-zone risk for cheaper capacity — often right for restartable batch work. Start from the **GPU batch inference** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the job is created
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- provides the least-privilege runtime identity
- [**GCP Cloud SQL**](/cloud-catalog/gcp-cloud-sql) -- exposes databases as managed Unix sockets via the Cloud SQL volume
- [**GCP GCS Bucket**](/cloud-catalog/gcp-gcs-bucket) -- mounts input/output object storage via Cloud Storage FUSE
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- provides the VPC for Direct VPC Egress connectivity
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- provides the subnetwork tasks draw IPs from
- [**GCP Cloud Run**](/cloud-catalog/gcp-cloud-run) -- the request-serving sibling for HTTP APIs and websites
