# SeaweedFS

Deploys SeaweedFS — the Apache-2.0, S3-compatible distributed object store — from the official SeaweedFS Helm chart. One resource is one whole store: **master** servers coordinate the cluster, **volume** servers hold the object bytes (the only tier that grows with your data), and the **filer** serves the file/bucket namespace and hosts the S3 API. Each tier is sized and scaled independently, and storage deliberately rides PersistentVolumeClaims instead of the chart's hostPath default.

Applications need S3 for artifacts, backups, datasets, and media even where no cloud bucket service exists — on-prem fleets, air-gapped clusters, dev environments. SeaweedFS serves the S3 API from inside the cluster with a small footprint: an empty spec gives a working single-node store. Everything stays ClusterIP; external reachability composes from ingress and gateway kinds.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm release** (official `seaweedfs` chart, pinned `4.40.0`, named `metadata.name`) — the master, volume, and filer tiers on PersistentVolumeClaims (5Gi/30Gi/10Gi chart defaults), the S3 gateway (embedded on the filer, or its own Deployment when `s3.dedicated` is declared), the `<name>-s3` Service on port 8333, and the install hook that creates every bucket declared in `s3.buckets`
- **S3 credentials Secret** (`<name>-s3-secret`, when auth is on — the default) — admin and read-only access-key pairs, generated once, stable across upgrades, kept on uninstall
- **Admin auth Secret** (`<name>-admin-auth`, when the console is enabled without `existingAuthSecret`) — keys `user`/`password`; the console is never installed open
- **Kubernetes Namespace** — created only when `createNamespace` is true; otherwise the namespace must already exist

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **A StorageClass** for the data volumes — most managed clusters provide a default; reference a **Kubernetes StorageClass** per tier for explicit (SSD) placement. The module keeps data on PVCs, so pod restarts are never data loss.
- **The Prometheus Operator CRDs** — only if you set `serviceMonitorEnabled`; without the operator, the release fails to install.

## Deploy

### Console

Open the deployment store, find **SeaweedFS**, and click **Deploy**. The creation wizard walks you through namespace placement, the chart pin, the three tiers in dependency order, the S3 gateway posture, day-one buckets, the replication code, the admin console, observability, and the Helm-values escape hatch. Start from the **Dev single node preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSeaweedFs
metadata:
  name: dev-seaweedfs
  org: acme-corp
  env: dev
spec:
  namespace:
    value: dev-object-store
  createNamespace: true
  s3:
    buckets:
      - name: app-data
```

```shell
planton apply -f seaweedfs.yaml
```

This creates the smallest declarable store that actually serves: one master, one volume server, one filer, the S3 gateway embedded on the filer with authentication on (credentials in the `dev-seaweedfs-s3-secret` Secret), and one bucket created at install. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose the store behind its namespace with a reference, and the InfraPipeline orders the deploys:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: object-store-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then installs the store into it.

## Key Configuration

These are the most important decisions when configuring a SeaweedFS store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The S3 gateway is on by default** — this kind exists to serve S3. Untouched, the chart serves the API with authentication required and materializes admin + read-only credential pairs in the `<name>-s3-secret` Secret. Explicit `enabled: false` is the pure filer/POSIX posture; explicit `enable_auth: false` means every pod that can reach the Service can read AND write every bucket — a dev-only posture.

**The volume tier is the capacity decision** — `volume.dataVolume.size` is per volume-server pod and holds the actual object bytes (the chart default is 30Gi). Size it for the data you expect, and multiply by the replication code's copy count. Master and filer state stay small.

**Replication is placement, not backup** — the three-digit XYZ code (`"001"` = one extra copy on another server) protects against server loss, never against deletion or corruption written through the API. The copies must fit the declared topology: `"001"` needs at least 2 volume servers or writes fail.

**Buckets are declarative, deletion is not** — buckets in `s3.buckets` are created at install with per-bucket TTL (objects expire themselves — no cleanup job), S3 versioning, anonymous read, and Object Lock. Removing a bucket from the spec never deletes it or its data — cleanup is a deliberate manual act. Object Lock is a one-way door: once enabled it cannot be turned off, and locked objects resist deletion until retention expires.

**One filer until you wire an external store** — the filer's embedded LevelDB metadata store is per-pod. Raising `filer.replicas` without first wiring a shared store (Postgres/MySQL through `WEED_*` env vars and `helmValues`) gives each filer its own divergent namespace.

**The console is never installed open** — `admin.enabled` without `existingAuthSecret` generates the `<name>-admin-auth` Secret (keys `user`/`password`, surfaced in the outputs). Give it a small `dataVolume` to persist console configuration across restarts.

**`helmValues` merges last** — the escape hatch for chart surface the typed fields do not cover (scheduling, probes, SFTP, mTLS, external filer stores). Anything here silently overrides the typed fields on every deploy; never put secrets in it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `master.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesStorageClass** | `volume.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesStorageClass** | `filer.dataVolume.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesStorageClass** | `admin.dataVolume.storageClass` | `status.outputs.storage_class_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the store runs in | Application deployment manifests |
| `release_name` | Helm release name (= metadata.name) | Operational tooling |
| `s3_endpoint` | In-cluster S3 endpoint (port 8333, path-style; the filer service when embedded, the gateway service when dedicated). Empty when the gateway is disabled | Application S3 SDK configuration |
| `s3_credentials_secret_name` | The credentials Secret — admin and read-only access-key pairs. Empty when auth is disabled | Client authentication |
| `filer_service_name` | The filer Service (file-namespace HTTP API, port 8888) | POSIX/HTTP file access, diagnostics |
| `master_service_name` | The master Service (cluster coordination, port 9333) | Operational tooling |
| `admin_endpoint` | In-cluster admin console endpoint. Empty when the console is disabled | Operator access |
| `admin_auth_secret_name` | The console credentials Secret (keys `user`/`password`). Empty when the console is disabled | Operator authentication |
| `port_forward_command` | Copy-paste `kubectl port-forward` for the S3 endpoint | Local development access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev single node** — one master, one volume server, one filer on small PVCs, the S3 gateway embedded with auth on, and one bucket created at install: a real S3 endpoint for developers and CI without production ceremony. Start from the **Dev single node preset**.

**Production HA** — a 3-master Raft quorum, 3 volume servers on explicit fast storage under the `"001"` replication code, a dedicated 2-replica S3 gateway Deployment, resources on every tier, and the admin console with persisted state. Start from the **Production HA preset**.

**Artifact store** — three typed buckets from day one: CI artifacts that expire themselves after 30 days (TTL), release artifacts with versioning so an overwrite never destroys a prior release, and a public-assets bucket with anonymous reads. Start from the **Artifact store preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — referenced placement; the InfraPipeline orders namespace-first.
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — SSD-backed classes for the tier data volumes.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — bring-your-own S3 identities and console credentials.
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) and [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) — HTTP exposure over the exported S3 endpoint or admin console.
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) and [**Kubernetes CronJob**](/infra-catalog/kubernetes-cron-job) — the workloads that consume the exported `s3_endpoint` and credentials Secret for artifacts, backups, and media.
