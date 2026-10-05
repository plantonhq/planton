# Kubernetes Velero

## When NOT to Use This

**One installation per cluster.** Velero's CRDs and its node-agent are
cluster-scoped, and one server owns the backup records in the store. The
Helm release name is therefore fixed to `velero` and never derives from
`metadata.name`.

Also not the right kind when:

- **You want application-level backups** — Velero backs up Kubernetes
  resources and volume data; database-consistent dumps (a point-in-time
  Postgres snapshot, a Mongo dump) are the application's own tooling,
  optionally triggered around Velero via backup hooks.
- **You only need the schedules, not the engine** — ad-hoc Backup and
  Restore operations are day-2 actions against the installed server (the
  `velero` CLI or the CRDs directly); this component installs the engine
  and, optionally, the recurring schedules.

## Overview

**KubernetesVelero** installs Velero — cluster backup and disaster
recovery — from the official Helm chart (`velero` at
`https://vmware-tanzu.github.io/helm-charts`). Velero backs up
Kubernetes resources (and optionally volume data) to an object store,
restores them on demand, and runs scheduled backups — the recovery path
for cluster loss, migration, and fat-finger deletions.

WHERE backups land is the central decision: the typed
`backup_storage.backend` oneof configures the default
BackupStorageLocation and its matching provider PLUGIN (an init
container Velero loads at start): S3/S3-compatible (including in-cluster
MinIO — the self-contained arm), Google Cloud Storage, or Azure Blob.
Volume DATA travels one of two ways: CSI snapshots (`enable_csi` + a
snapshot-capable CSI driver) or file-system backup
(`fs_backup.deploy_node_agent` — kopia-based, works on ANY volume type).

The typed spec covers the chart's meaningful configuration surface, with
a `helm_values` escape hatch (merged last, Helm `-f` semantics,
identical on both engines) for anything beyond it.

**Key design points:**

- **Backup records survive uninstall by default** — the chart ships its
  CRDs in the crds/ directory, which Helm installs once and, by its own
  contract, NEVER deletes on uninstall. `crds.cleanup_on_uninstall`
  (default false) is the DESTRUCTIVE opt-in upstream warns is meant for
  CI systems, not production; the stored objects in the bucket are never
  touched either way.
- **Credentials are secret-by-default; keyless where the cloud allows**
  — on EKS/GKE/AKS the IRSA / Workload Identity arms need no stored key
  at all; declared-credential arms materialize into Velero's credentials
  Secret in the plugin's expected format.
- **The plugin version pairs with the Velero release** — each backend
  arm loads the official plugin at the version paired with the chart's
  Velero release (v1.14.x plugins for Velero 1.18); `plugin_image`
  overrides for private mirrors or deliberate pins.
- **Restore-only standby is a first-class posture** —
  `server.restore_only_mode` blocks backups, schedules, and garbage
  collection: the stance for a DR-standby cluster reading another
  cluster's store.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`velero` is the upstream
  convention) — literal or a KubernetesNamespace reference
- **`spec.backup_storage`**: the backend oneof — Velero without a
  storage location backs up nothing

### Common

- **`spec.chart_version`**: pinned chart version (default `12.1.0`,
  which ships Velero 1.18 — chart and app versions move separately; the
  chart pin governs)
- **`spec.backup_storage.s3`**: bucket + region, `s3_url` +
  `force_path_style` for S3-compatible stores (MinIO, Ceph RGW, ...),
  `kms_key_id` (real S3 only), `ca_cert` for self-signed endpoints, and
  IRSA XOR access keys
- **`spec.backup_storage.gcs`**: bucket, and Workload Identity
  service-account email XOR a JSON key
- **`spec.backup_storage.azure_blob`**: storage account + container +
  resource group + subscription, and workload identity (with client ID)
  XOR a service principal
- **`spec.backup_storage.prefix`**: directory prefix within the bucket —
  the multi-cluster pattern: one bucket, one prefix per cluster
- **`spec.volume_snapshots`**: `enabled` (chart default true — disable
  on clusters without snapshot support), `enable_csi` (the modern
  snapshot path on EKS/GKE/AKS), `default_snapshot_move_data` (Velero's
  data mover — makes backups portable across clusters/regions; requires
  CSI and the node-agent)
- **`spec.fs_backup`**: `deploy_node_agent` (the DaemonSet required for
  file-system backup AND CSI data movement; it mounts the kubelet pod
  directory to read volume data) and `default_volumes_to_fs_backup`
  (all pod volumes without per-pod annotations)
- **`spec.schedules`**: recurring backups rendered as Schedule
  resources — cron expression, TTL (Velero default 720h),
  namespace/resource filters, label selector, per-schedule volume
  posture, `paused`
- **`spec.server`**: TTLs, item-operation timeout, GC frequency,
  `restore_only_mode`, log level/format
- **`spec.crds`**: `upgrade_on_install` (chart default true — the
  CRD-upgrade job that keeps crds/-directory CRDs current) and the
  destructive `cleanup_on_uninstall`
- **`spec.deployment` / `spec.prometheus`**: server sizing and
  scheduling; metrics (chart default on) and the opt-in ServiceMonitor
  (requires the Prometheus operator CRDs — the release fails without
  them)
- **`spec.helm_values`**: escape hatch for chart values beyond the typed
  fields — never the primary interface

## Environment Injection

How Velero reaches the object store, per backend arm:

| Environment | Arm | Keyless posture | Declared-credential posture |
|---|---|---|---|
| EKS / AWS S3 | `s3` | `irsa_role_arn` — the `eks.amazonaws.com/role-arn` annotation on the `velero-server` service account; no Secret exists at all | `access_keys` — materialized into the credentials Secret in the AWS plugin's `cloud` file format |
| S3-compatible (MinIO, Ceph RGW, ...) | `s3` with `s3_url` (+ `force_path_style`) | none — IRSA only mints AWS credentials (spec-enforced) | `access_keys` — the store's key pair |
| GKE / GCS | `gcs` | `workload_identity_service_account_email` — the `iam.gke.io/gcp-service-account` annotation plus the storage location's serviceAccount config | `service_account_key_json` — the JSON key as the `cloud` file |
| AKS / Azure Blob | `azure_blob` | `workload_identity_client_id` — the `azure.workload.identity/client-id` annotation plus the `azure.workload.identity/use` pod label; a `cloud` file still carries the non-secret parameters | `service_principal` — AZURE_* lines in the credentials Secret |

Leaving both postures unset means ambient node credentials. The
cloud-side half of each keyless contract is written against the chart's
fixed `velero-server` service account — which is why it is a
output.

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace |
| `release_name` | Helm release name (always `velero`) |
| `service_account_name` | Always `velero-server` — the subject cloud-side keyless bindings are written against |
| `backup_storage_location_name` | Always `default` — what Backup and Schedule resources reference through storageLocation |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **Cloud-side keyless identity** closes over the
  `service_account_name` output: the IRSA trust policy, GCP WI binding,
  or Entra federated credential names the `velero-server` service
  account in the installation namespace.
- **The bucket is cloud-side composition**: the S3 bucket / GCS bucket /
  blob container (and for DR, its replication) is declared by the cloud
  infrastructure; this component points at it. A standby cluster
  composes the same bucket with `server.restore_only_mode`.

## Examples

### EKS with IRSA and CSI snapshots

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesVelero
metadata:
  name: velero
spec:
  namespace:
    value: velero
  createNamespace: true
  backupStorage:
    s3:
      bucket: my-cluster-backups
      region: us-west-2
      irsaRoleArn:
        value: arn:aws:iam::111111111111:role/velero
    prefix: prod-cluster
  volumeSnapshots:
    enableCsi: true
  schedules:
    - name: daily-cluster
      schedule: "0 2 * * *"
      ttl: 720h
```

### Self-contained (in-cluster MinIO, file-system backup)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesVelero
metadata:
  name: velero
spec:
  namespace:
    value: velero
  createNamespace: true
  backupStorage:
    s3:
      bucket: velero-backups
      region: minio
      s3Url: http://minio.minio.svc:9000
      forcePathStyle: true
      accessKeys:
        accessKeyId: minio
        secretAccessKey: minio123
  volumeSnapshots:
    enabled: false # MinIO has no snapshot provider
  fsBackup:
    deployNodeAgent: true
    defaultVolumesToFsBackup: true
```

### DR-standby cluster (restore-only, reading the primary's store)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesVelero
metadata:
  name: velero
spec:
  namespace:
    value: velero
  createNamespace: true
  backupStorage:
    s3:
      bucket: my-cluster-backups
      region: us-west-2
      irsaRoleArn:
        value: arn:aws:iam::111111111111:role/velero-standby
    prefix: prod-cluster
  server:
    restoreOnlyMode: true # backups, schedules and GC are blocked
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
