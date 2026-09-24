# GCP Cloud Run Job

Deploys a Cloud Run job — a run-to-completion container workload that executes a fixed number of parallel tasks and exits. Unlike [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) (request-serving, traffic splitting, liveness/readiness probes), a job is batch-shaped: define the task template once, then each execution stamps out `taskCount` tasks with up to `parallelism` running concurrently. A startup probe (HTTP/TCP/gRPC) is supported — it gates task start, and a container that never passes is shut down and retried per `maxRetries`.

## What Gets Created

When you deploy a GcpCloudRunJob resource, Planton provisions:

- **Cloud Run job** — a `google_cloud_run_v2_job` in the chosen region; the Cloud Run Admin API is enabled automatically
- **Task template** — containers, volumes, VPC access, identity, and per-task limits as configured
- **Platform attribution** — organization, environment, and resource labels on the job object

Individual executions (`gcloud run jobs execute`, Cloud Scheduler, Eventarc) are separate API objects — this resource owns the job definition, not each run.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A container image** built for batch/CLI work ([GcpArtifactRegistryRepo](/docs/catalog/gcp/gcpartifactregistryrepo))
- **A service account** ([GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount)) when tasks call other GCP APIs
- **VPC and subnetwork** ([GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork), [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork)) for direct VPC egress
- **Secret Manager secrets** if referencing them in env vars (`value_from_secret`) or volumes; a variable's `secret_value` needs none -- the module creates its secret, replicated in the job's region and readable only by the task identity

## Quick Start

Create a file `job.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunJob
metadata:
  name: my-etl
spec:
  region: us-central1
  template:
    containers:
      - image: us-docker.pkg.dev/my-project/my-repo/etl:1.0.0
        resources:
          cpu: "1"
          memory: 512Mi
  taskCount: 1
  deletionProtection: false
```

Deploy:

```shell
planton apply -f job.yaml
```

Run once:

```shell
gcloud run jobs execute my-etl --region us-central1
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Region the job is deployed in. Immutable. | Required |
| `template.containers` | `list` | Containers that make up one task. | At least 1; each needs `image` |

### Job Identity & Metadata

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project. Can reference GcpProject. |
| `jobName` | `string` | `metadata.name` | Job name in GCP (1–63 chars). Immutable. |
| `labels` | `map` | — | User labels on the job object; merged beneath platform attribution. |
| `annotations` | `map` | — | Unstructured metadata for external tools, on the job object. |
| `executionLabels` | `map` | — | Labels stamped on every execution the job creates. |
| `executionAnnotations` | `map` | — | Annotations stamped on every execution. |
| `deletionProtection` | `bool` | `true` | Destroy fails until explicitly set false. |
| `deletionPolicy` | `string` | `DELETE` | Destroy stance: `PREVENT` fails destroys; `ABANDON` unmanages without deleting. |

### Execution Controls

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `taskCount` | `int` | GCP default (1) | Tasks each execution runs. |
| `parallelism` | `int` | max possible | Max tasks running concurrently (≤ taskCount). |
| `launchStage` | `string` | GA | ALPHA/BETA for preview features. |
| `gpuZonalRedundancyDisabled` | `bool` | false | Single-zone GPU (requires `template.nodeSelector`). |
| `binaryAuthorization` | object | — | Image attestation deploy gate. |
| `startExecutionToken` | `string` | — | Run-on-deploy: a unique suffix triggers an execution on create/update; job READY when it starts. Mutually exclusive with `runExecutionToken`. |
| `runExecutionToken` | `string` | — | Run-on-deploy with deploy-and-verify semantics: job READY only when the triggered execution completes. |

### Task Template (`template`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `containers[]` | list | — | Worker and sidecar containers per task; each may declare `ports` and a `startupProbe` (HTTP/TCP/gRPC, 240-second startup window), plus `volumeMounts` with `subPath`. |
| `volumes[]` | list | — | Cloud SQL, Secret, emptyDir, GCS (with gcsfuse `mountOptions`), NFS. |
| `serviceAccount` | `StringValueOrRef` | Compute default SA | Runtime identity per task. |
| `executionEnvironment` | enum | GEN2 | GEN1 (fast) vs GEN2 (full Linux, GCS/NFS). |
| `encryptionKey` | `StringValueOrRef` | — | CMEK for container images (GcpKmsKey ref). |
| `timeoutSeconds` | `int` | 600 | Per-attempt task timeout (1–86400). |
| `maxRetries` | `int` | 3 | Retries per task before failure (≥ 0). |
| `vpcAccess` | object | — | Connector XOR direct VPC egress. |
| `nodeSelector.accelerator` | `string` | — | GPU type (e.g. `nvidia-l4`). |

### Stack Outputs

| Output | Description |
|--------|-------------|
| `jobName` | Name of the job in GCP |
| `location` | Deployment region |
| `uid` | Server-assigned unique ID |
| `latestCreatedExecution` | Most recent execution name (empty until first run) |

## Deliberately not modeled (recorded reasons)

Everything the pinned GA provider can configure on `google_cloud_run_v2_job` is representable through this component, except the entries below — each a recorded decision (the machine-checked record lives in `iac/provider-parity.yaml`):

| Excluded Feature | Why |
|---|---|
| `tags` (resource-manager tags) | GA at the pin but not yet bridged by the pinned Pulumi SDK; modeling it only in Terraform would break cross-engine parity. Re-evaluate at the next pulumi-gcp bump. |
| `client` / `client_version` | API-client telemetry strings with no user-facing behavior. |
| Job IAM (`google_cloud_run_v2_job_iam_*`) | Jobs have no public-serving toggle (nothing to invoke over HTTP); fine-grained execution grants to specific identities are IAM-family territory. |

## Related Components

- [GcpCloudRun](/docs/catalog/gcp/gcpcloudrun) — request-serving sibling (services, not jobs)
- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — Cloud SQL volume source
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — task runtime identity
- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — customer-managed encryption

## Presets

- [01-batch-etl-basic](presets/01-batch-etl-basic.yaml) — single-task nightly ETL
- [02-parallel-vpc-cleanup](presets/02-parallel-vpc-cleanup.yaml) — parallel workers with VPC + secrets
- [03-gpu-batch-inference](presets/03-gpu-batch-inference.yaml) — GPU map-reduce inference

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
