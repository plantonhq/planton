# Velero

Installs Velero — cluster backup and disaster recovery — from the official Helm chart. Velero backs up Kubernetes resources (and optionally volume data) to an object store, restores them on demand, and runs scheduled backups — the recovery path for cluster loss, migration, and fat-finger deletions. The backend picks where backups land (S3/S3-compatible including in-cluster MinIO, Google Cloud Storage, or Azure Blob) and loads the matching provider plugin. One installation per cluster: the CRDs are cluster-scoped and one server owns the store's backup records.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm Release** (`velero`) -- the Velero server Deployment with the backend's provider plugin as an init container, the default BackupStorageLocation (named `default`), the default VolumeSnapshotLocation (unless snapshots are disabled), and any declared Schedule resources
- **CRDs** -- Backup, Restore, Schedule, and companions — surviving uninstall by Helm's own contract for `crds/`-directory CRDs, so backup records outlive the release
- **Node-agent DaemonSet** (optional) -- when file-system backup or CSI snapshot data movement is enabled; runs privileged with host-path mounts of the kubelet pod directory — that is how it reads volume data
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true (`velero` is the upstream convention)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- An object store to land backups in: an S3 bucket, an S3-compatible endpoint (MinIO, Ceph RGW, ...), a GCS bucket, or an Azure Blob container.
- Cloud-side identity for the keyless postures: an IRSA role, GCP Workload Identity binding, or Azure federated credential written against the `velero-server` service account; S3-compatible endpoints authenticate with access keys.
- For CSI snapshots: a snapshot-capable CSI driver AND a VolumeSnapshotClass labeled for Velero — without the labeled class, backups report Completed while volumes were never snapshotted.
- With the Prometheus ServiceMonitor: the Prometheus operator CRDs — the release fails to install without them.

## Deploy

### Console

Open the deployment store, find **Velero**, and click **Deploy**. The creation wizard walks you through placement, the chart pin and CRD lifecycle, the backup storage backend with its credential posture, volume snapshots, file-system backup, backup schedules, server behavior, observability, and scheduling. Start from the **EKS S3 IRSA** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesVelero
metadata:
  name: velero
  org: acme-corp
  env: prod
spec:
  namespace:
    value: velero
  createNamespace: true
  backupStorage:
    s3:
      bucket: acme-cluster-backups
      region: us-west-2
      irsaRoleArn:
        value: arn:aws:iam::111111111111:role/velero
    prefix: prod-cluster
  schedules:
    - name: daily-cluster
      schedule: "0 2 * * *"
      ttl: 720h
```

```shell
planton apply -f velero.yaml
```

This installs the Velero server with the AWS plugin, a default BackupStorageLocation in the `acme-cluster-backups` bucket under the `prod-cluster` prefix (keyless via IRSA), and a nightly schedule whose backups are garbage-collected after 30 days. An Infra Job tracks the provisioning in real time.

### InfraChart

When the IRSA role is managed in the same InfraChart, wire the credential reference with `valueFrom` so the role exists before Velero starts:

```yaml
spec:
  namespace:
    value: velero
  createNamespace: true
  backupStorage:
    s3:
      bucket: acme-cluster-backups
      region: us-west-2
      irsaRoleArn:
        valueFrom:
          kind: AwsIamRole
          name: velero-backup-role
          fieldPath: status.outputs.role_arn
    prefix: prod-cluster
```

The InfraPipeline deploys the referenced IAM role first, then installs Velero against it. The GCS and Azure Blob backends compose the same way through `workloadIdentityServiceAccountEmail` and `workloadIdentityClientId`.

## Key Configuration

These are the most important decisions when configuring Velero. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The store is the source of truth** -- deleting a Backup resource from the cluster does not remove its records from the bucket, and Velero's location sync re-imports whatever the store holds. A backup name can never be reused against a prefix that still holds a prior backup's records, and two clusters must never share a prefix unless one is deliberately restoring from the other.

**Prefix per cluster** -- when several clusters share one bucket, `prefix` is what keeps their records apart. It is the single most consequential string in the spec.

**Credential posture per backend** -- each backend offers keyless (IRSA / GCP Workload Identity / Azure Workload Identity), declared keys (stored as managed secrets), or NEITHER — ambient node credentials. S3-compatible endpoints (MinIO, Ceph RGW) cannot use IRSA; they authenticate with access keys.

**The volume-data path** -- CSI snapshots on clusters with snapshot-capable drivers (add snapshot data movement to make backups portable across clusters), or file-system backup via the node agent for everything else. Verify volume snapshots appear in a backup's own resource list before trusting it — a missing VolumeSnapshotClass label fails silently.

**CRDs and records survive removal** -- Helm never deletes `crds/`-directory CRDs on uninstall, and bucket objects are never touched either way. The cleanup-on-uninstall dial is DESTRUCTIVE and reserved by upstream for CI — never production.

**Restore-only standby** -- a DR-standby cluster reads the same store with restore-only mode when recovery time matters.

**Observability you can trust** -- DR you cannot observe is DR you cannot trust; the ServiceMonitor exposes backup success/failure for alerting.

**The escape hatch** -- `helmValues` carries additional chart values as a YAML document, merged LAST — never the substitute for typed fields, and never a place for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Kubernetes Namespace | `spec.namespace` | `spec.name` |
| AWS IAM Role | `spec.backupStorage.s3.irsaRoleArn` | `status.outputs.role_arn` |
| GCP Service Account | `spec.backupStorage.gcs.workloadIdentityServiceAccountEmail` | `status.outputs.email` |
| Azure User Assigned Identity | `spec.backupStorage.azureBlob.workloadIdentityClientId` | `status.outputs.client_id` |

Each credential reference is also the deploy-ordering edge: the identity (and the grants riding it) exists before Velero starts. Literal values cover identities created outside Planton.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Installation namespace | Debugging and composition |
| `release_name` | Helm release name (always `velero`) | Debugging the release (`helm status`) |
| `service_account_name` | Always `velero-server` | The subject cloud-side keyless bindings are written against (IRSA trust policy, GCP WI binding, Azure federated credential) |
| `backup_storage_location_name` | Always `default` | What Backup and Schedule resources reference through `storageLocation` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EKS with S3 and IRSA** -- keyless backups to S3, CSI snapshots via the EBS driver, a nightly 30-day schedule. Start from the **EKS S3 IRSA** preset.

**GKE with GCS Workload Identity** -- keyless backups to a GCS bucket. Start from the **GKE GCS Workload Identity** preset.

**Self-contained MinIO** -- an in-cluster S3-compatible store with access keys and file-system backup — the air-gapped/dev posture. Start from the **MinIO Self-Contained** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the placement target; also the unit most restores operate on.
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the IRSA role behind keyless S3 backups (`irsaRoleArn`).
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the Workload Identity subject behind keyless GCS backups.
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the federated identity behind keyless Azure Blob backups.
- [**Kubernetes StatefulSet**](/infra-catalog/kubernetes-stateful-set) -- the stateful workloads whose volume data CSI snapshots or file-system backup capture.
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- provides the Prometheus operator CRDs the ServiceMonitor needs and turns backup health into alerts.
