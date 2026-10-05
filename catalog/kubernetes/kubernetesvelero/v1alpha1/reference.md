# KubernetesVelero

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesVeleroSpec** installs Velero — cluster backup and disaster
recovery — from the official Helm chart (`velero` at
https://vmware-tanzu.github.io/helm-charts). Velero backs up Kubernetes
resources (and optionally volume data) to an object store, restores them
on demand, and runs scheduled backups — the recovery path for cluster
loss, migration, and fat-finger deletions.

WHERE backups land is the central decision: the typed
`backup_storage.backend` oneof configures the default
BackupStorageLocation and its matching provider PLUGIN (an init
container Velero loads at start): S3/S3-compatible (including in-cluster
MinIO — the self-contained arm), Google Cloud Storage, or Azure Blob.
Credentials are secret-by-default; on EKS/GKE/AKS the keyless posture
(IRSA / Workload Identity annotations) needs no stored key at all.

Volume DATA travels one of two ways: CSI snapshots (features:
EnableCSI + a snapshot-capable CSI driver) or file-system backup
(`fs_backup.deploy_node_agent` — kopia-based, works on ANY volume type
including clusters without snapshot support).

ONE INSTALLATION PER CLUSTER: Velero's CRDs and its node-agent are
cluster-scoped, and one server owns the backup records in the store.
The Helm release name is therefore fixed to "velero".

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface test manifest: exercises every typed arm of the spec so the
# offline plan proofs cover what the live lanes may not. Not a realistic
# production shape — see presets for those.
#
# Backend: the self-contained S3-compatible posture (in-cluster MinIO via
# s3_url + path-style + access keys); volume snapshots are OFF (MinIO has
# no snapshot provider) so volume data rides file-system backup.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesVelero
metadata:
  name: hack-velero
spec:
  namespace:
    value: hack-velero
  createNamespace: true
  chartVersion: "12.1.0"
  crds:
    upgradeOnInstall: true
    cleanupOnUninstall: false
  backupStorage:
    s3:
      bucket: velero-backups
      region: minio
      s3Url: http://minio.minio.svc:9000
      forcePathStyle: true
      caCert: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJoVENDQVN1Z0F3SUJBZ0lTCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
      accessKeys:
        accessKeyId: minio
        secretAccessKey: minio123
    prefix: hack-cluster
  volumeSnapshots:
    enabled: false
  fsBackup:
    deployNodeAgent: true
    defaultVolumesToFsBackup: true
    nodeAgentResources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: "1"
        memory: 1Gi
    nodeAgentTolerations:
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
  schedules:
    - name: daily-cluster
      schedule: "0 2 * * *"
      ttl: 720h
      includedNamespaces:
        - default
        - hack-apps
      excludedResources:
        - events
      labelSelector:
        backup: enabled
      includeClusterResources: true
      snapshotVolumes: false
      defaultVolumesToFsBackup: true
      storageLocation: default
  server:
    defaultBackupTtl: 720h
    defaultItemOperationTimeout: 4h
    garbageCollectionFrequency: 1h
    logLevel: info
    logFormat: text
  deployment:
    resources:
      requests:
        cpu: 500m
        memory: 128Mi
      limits:
        cpu: "1"
        memory: 512Mi
    priorityClassName: system-cluster-critical
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: node-role.kubernetes.io/control-plane
        operator: Exists
        effect: NoSchedule
  prometheus:
    enabled: true
    serviceMonitor: false
  helmValues: |
    configuration:
      backupSyncPeriod: 1m
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `12.1.0` |  |
| `spec.crds` | `KubernetesVeleroCrds` |  |  |  |
| `spec.crds.upgradeOnInstall` | `bool` |  | `true` |  |
| `spec.crds.cleanupOnUninstall` | `bool` |  |  |  |
| `spec.backupStorage` | `KubernetesVeleroBackupStorage` | yes |  |  |
| `spec.backupStorage.s3` | `KubernetesVeleroS3Backend` |  |  |  |
| `spec.backupStorage.s3.bucket` | `string` | yes |  |  |
| `spec.backupStorage.s3.region` | `string` | yes |  |  |
| `spec.backupStorage.s3.s3Url` | `string` |  |  |  |
| `spec.backupStorage.s3.forcePathStyle` | `bool` |  |  |  |
| `spec.backupStorage.s3.kmsKeyId` | `string` |  |  |  |
| `spec.backupStorage.s3.caCert` | `string` |  |  |  |
| `spec.backupStorage.s3.irsaRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.backupStorage.s3.accessKeys` | `KubernetesVeleroS3AccessKeys` |  |  |  |
| `spec.backupStorage.s3.accessKeys.accessKeyId` | `string` | yes |  |  |
| `spec.backupStorage.s3.accessKeys.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.backupStorage.gcs` | `KubernetesVeleroGcsBackend` |  |  |  |
| `spec.backupStorage.gcs.bucket` | `string` | yes |  |  |
| `spec.backupStorage.gcs.workloadIdentityServiceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.backupStorage.gcs.serviceAccountKeyJson` | `string` (sensitive) |  |  |  |
| `spec.backupStorage.azureBlob` | `KubernetesVeleroAzureBlobBackend` |  |  |  |
| `spec.backupStorage.azureBlob.storageAccount` | `string` | yes |  |  |
| `spec.backupStorage.azureBlob.container` | `string` | yes |  |  |
| `spec.backupStorage.azureBlob.resourceGroup` | `string` | yes |  |  |
| `spec.backupStorage.azureBlob.subscriptionId` | `string` | yes |  |  |
| `spec.backupStorage.azureBlob.workloadIdentityClientId` | `string \| valueFrom` |  |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.backupStorage.azureBlob.servicePrincipal` | `KubernetesVeleroAzureServicePrincipal` |  |  |  |
| `spec.backupStorage.azureBlob.servicePrincipal.tenantId` | `string` | yes |  |  |
| `spec.backupStorage.azureBlob.servicePrincipal.clientId` | `string` | yes |  |  |
| `spec.backupStorage.azureBlob.servicePrincipal.clientSecret` | `string` (sensitive) | yes |  |  |
| `spec.backupStorage.prefix` | `string` |  |  |  |
| `spec.backupStorage.pluginImage` | `string` |  |  |  |
| `spec.volumeSnapshots` | `KubernetesVeleroVolumeSnapshots` |  |  |  |
| `spec.volumeSnapshots.enabled` | `bool` |  | `true` |  |
| `spec.volumeSnapshots.enableCsi` | `bool` |  |  |  |
| `spec.volumeSnapshots.defaultSnapshotMoveData` | `bool` |  |  |  |
| `spec.fsBackup` | `KubernetesVeleroFsBackup` |  |  |  |
| `spec.fsBackup.deployNodeAgent` | `bool` |  |  |  |
| `spec.fsBackup.defaultVolumesToFsBackup` | `bool` |  |  |  |
| `spec.fsBackup.nodeAgentResources` | `ContainerResources` |  |  |  |
| `spec.fsBackup.nodeAgentResources.limits` | `CpuMemory` |  |  |  |
| `spec.fsBackup.nodeAgentResources.limits.cpu` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentResources.limits.memory` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentResources.requests` | `CpuMemory` |  |  |  |
| `spec.fsBackup.nodeAgentResources.requests.cpu` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentResources.requests.memory` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentTolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.fsBackup.nodeAgentTolerations[].key` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentTolerations[].operator` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentTolerations[].value` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentTolerations[].effect` | `string` |  |  |  |
| `spec.fsBackup.nodeAgentTolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.schedules` | `[]KubernetesVeleroSchedule` |  |  |  |
| `spec.schedules[].name` | `string` | yes |  |  |
| `spec.schedules[].schedule` | `string` | yes |  |  |
| `spec.schedules[].paused` | `bool` |  |  |  |
| `spec.schedules[].ttl` | `string` |  |  |  |
| `spec.schedules[].includedNamespaces` | `[]string` |  |  |  |
| `spec.schedules[].excludedNamespaces` | `[]string` |  |  |  |
| `spec.schedules[].includedResources` | `[]string` |  |  |  |
| `spec.schedules[].excludedResources` | `[]string` |  |  |  |
| `spec.schedules[].labelSelector` | `map<string, string>` |  |  |  |
| `spec.schedules[].includeClusterResources` | `bool` |  |  |  |
| `spec.schedules[].snapshotVolumes` | `bool` |  |  |  |
| `spec.schedules[].defaultVolumesToFsBackup` | `bool` |  |  |  |
| `spec.schedules[].storageLocation` | `string` |  |  |  |
| `spec.server` | `KubernetesVeleroServer` |  |  |  |
| `spec.server.defaultBackupTtl` | `string` |  |  |  |
| `spec.server.defaultItemOperationTimeout` | `string` |  |  |  |
| `spec.server.garbageCollectionFrequency` | `string` |  |  |  |
| `spec.server.restoreOnlyMode` | `bool` |  |  |  |
| `spec.server.logLevel` | `string` |  | `info` |  |
| `spec.server.logFormat` | `string` |  | `text` |  |
| `spec.deployment` | `KubernetesVeleroDeployment` |  |  |  |
| `spec.deployment.resources` | `ContainerResources` |  |  |  |
| `spec.deployment.resources.limits` | `CpuMemory` |  |  |  |
| `spec.deployment.resources.limits.cpu` | `string` |  |  |  |
| `spec.deployment.resources.limits.memory` | `string` |  |  |  |
| `spec.deployment.resources.requests` | `CpuMemory` |  |  |  |
| `spec.deployment.resources.requests.cpu` | `string` |  |  |  |
| `spec.deployment.resources.requests.memory` | `string` |  |  |  |
| `spec.deployment.priorityClassName` | `string` |  |  |  |
| `spec.deployment.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.deployment.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.deployment.tolerations[].key` | `string` |  |  |  |
| `spec.deployment.tolerations[].operator` | `string` |  |  |  |
| `spec.deployment.tolerations[].value` | `string` |  |  |  |
| `spec.deployment.tolerations[].effect` | `string` |  |  |  |
| `spec.deployment.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.prometheus` | `KubernetesVeleroPrometheus` |  |  |  |
| `spec.prometheus.enabled` | `bool` |  | `true` |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install Velero into ("velero" is the upstream
convention). Accepts a literal namespace name or a reference to a
KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "12.1.0", which ships Velero
1.18 — chart and app versions move separately; the chart pin
governs). Pin deliberately; upgrades re-run the release with the new
chart. Pick versions from the chart repository's index
(`helm search repo`): the served chart is the contract — the upstream
source tree's Chart.yaml can claim a version at a tag that was never
served.

- default: `12.1.0`

### spec.crds

`KubernetesVeleroCrds`

Velero custom resource definitions (Backup, Restore, Schedule,
BackupStorageLocation, ...) lifecycle. Helm installs the CRDs from
the chart's crds/ directory and — by Helm's own contract — NEVER
deletes them on uninstall, so backup records always survive removing
the component (the DR-safety posture).

### spec.crds.upgradeOnInstall

`bool` · optional (explicit presence)

Run the chart's CRD-upgrade job on install/upgrade (an init job that
re-applies the pinned CRDs). Chart default: true — Helm never
upgrades crds/-directory CRDs on its own, so this is what keeps them
current across chart upgrades.

- default: `true`

### spec.crds.cleanupOnUninstall

`bool`

DESTRUCTIVE: delete the Velero CRDs (and therefore every Backup /
Restore / Schedule / BackupStorageLocation record in the cluster) on
uninstall. Chart default: false — backup records survive removal, and
upstream warns this switch is meant for CI systems, not production.
The stored objects in the bucket are never touched either way.

### spec.backupStorage

`KubernetesVeleroBackupStorage` · required

Where backups are stored: the default BackupStorageLocation and the
provider plugin that serves it. Required — Velero without a storage
location backs up nothing.

- rule: {"required":true}

### spec.backupStorage.s3

`KubernetesVeleroS3Backend`

AWS S3 — or ANY S3-compatible store (MinIO, Ceph RGW, DigitalOcean
Spaces, Backblaze B2, ...) via the s3_url override.

- rule: irsa_role_arn and access_keys are alternative credential postures — set at most one (neither = ambient node credentials)
- rule: an S3-compatible endpoint (s3_url) authenticates with access_keys — IRSA only mints AWS credentials

### spec.backupStorage.s3.bucket

`string` · required

Bucket to store backups in.

- rule: {"required":true}

### spec.backupStorage.s3.region

`string` · required

AWS region of the bucket. Required for real S3; for S3-compatible
stores use the store's expected region value (MinIO accepts any,
"minio" by convention).

- rule: {"required":true}

### spec.backupStorage.s3.s3Url

`string`

S3-COMPATIBLE ARM: endpoint URL of the store (e.g.
http://minio.minio.svc:9000 for in-cluster MinIO). Empty = real AWS
S3.

- rule: s3_url must be an http(s) endpoint URL (e.g. http://minio.minio.svc:9000)

### spec.backupStorage.s3.forcePathStyle

`bool`

Use path-style addressing (bucket in the path instead of the
subdomain) — required by MinIO and most S3-compatible stores.

### spec.backupStorage.s3.kmsKeyId

`string`

KMS key (id, alias, or ARN) for server-side encryption of backups
(real S3 only).

### spec.backupStorage.s3.caCert

`string`

Base64-encoded CA bundle for verifying a self-signed s3_url
endpoint's TLS certificate.

### spec.backupStorage.s3.irsaRoleArn

`string | valueFrom`

IAM role ARN for IRSA: annotates Velero's service account so it
reaches S3 without stored keys — the keyless posture on EKS.
Reference an AwsIamRole's role_arn output -- the reference is also the
deploy-ordering edge, so the role exists before Velero starts -- or
pass a literal ARN (arn:aws:iam::<account>:role/<name>). Mutually
exclusive with access_keys.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.backupStorage.s3.accessKeys

`KubernetesVeleroS3AccessKeys`

Static access keys, materialized into Velero's credentials Secret
(the `cloud` credentials-file format the AWS plugin reads). The
declared-credential arm — for S3-compatible stores and clusters
without IRSA.

### spec.backupStorage.s3.accessKeys.accessKeyId

`string` · required

Access key ID — the public identifier of the key pair, not a secret
(the guard's name heuristic treats *_id names as references, so no
exemption annotation is needed); only the paired secret access key is
a credential. For MinIO this is the access key / username.

- rule: {"required":true}

### spec.backupStorage.s3.accessKeys.secretAccessKey

`string` · required · sensitive

Secret access key (for MinIO: the secret key / password).

- rule: {"required":true}

### spec.backupStorage.gcs

`KubernetesVeleroGcsBackend`

Google Cloud Storage.

- rule: workload_identity_service_account_email and service_account_key_json are alternative credential postures — set at most one (neither = ambient node credentials)

### spec.backupStorage.gcs.bucket

`string` · required

GCS bucket to store backups in.

- rule: {"required":true}

### spec.backupStorage.gcs.workloadIdentityServiceAccountEmail

`string | valueFrom`

GCP service-account email for Workload Identity: annotates Velero's
Kubernetes service account (and is set on the storage location) so
Velero reaches GCS without a stored key — the keyless posture on
GKE. Reference a GcpServiceAccount's email output -- the reference is
also the deploy-ordering edge, so the identity (and the grants riding
it) exists before Velero starts -- or pass a literal email
(<id>@<project>.iam.gserviceaccount.com). Mutually exclusive with
service_account_key_json.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.backupStorage.gcs.serviceAccountKeyJson

`string` · sensitive

GCP service-account key (the JSON key file's content), materialized
into Velero's credentials Secret. The declared-credential arm for
non-GKE clusters backing up to GCS.

### spec.backupStorage.azureBlob

`KubernetesVeleroAzureBlobBackend`

Azure Blob Storage.

- rule: workload_identity_client_id and service_principal are alternative credential postures — set at most one (neither = ambient node credentials)

### spec.backupStorage.azureBlob.storageAccount

`string` · required

Storage-account name holding the container.

- rule: {"required":true}

### spec.backupStorage.azureBlob.container

`string` · required

Blob container to store backups in.

- rule: {"required":true}

### spec.backupStorage.azureBlob.resourceGroup

`string` · required

Resource group of the storage account.

- rule: {"required":true}

### spec.backupStorage.azureBlob.subscriptionId

`string` · required

Subscription the storage account lives in.

- rule: {"required":true}

### spec.backupStorage.azureBlob.workloadIdentityClientId

`string | valueFrom`

Entra client ID for Azure AD Workload Identity: annotates Velero's
Kubernetes service account (azure.workload.identity/client-id) and
labels its pods (azure.workload.identity/use) so the AKS webhook
injects a federated token — the keyless posture on AKS. Reference an
AzureUserAssignedIdentity's client_id output -- the reference is also
the deploy-ordering edge, so the identity (and the federated
credential and grants riding it) exists before Velero starts -- or
pass a literal client ID (GUID). Setting it selects the
workload-identity posture; mutually exclusive with service_principal.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.backupStorage.azureBlob.servicePrincipal

`KubernetesVeleroAzureServicePrincipal`

Service-principal credentials, materialized into Velero's
credentials Secret. The declared-credential arm.

### spec.backupStorage.azureBlob.servicePrincipal.tenantId

`string` · required

Entra tenant ID.

- rule: {"required":true}

### spec.backupStorage.azureBlob.servicePrincipal.clientId

`string` · required

Application (client) ID of the service principal (a public
identifier — not a secret).

- rule: {"required":true}

### spec.backupStorage.azureBlob.servicePrincipal.clientSecret

`string` · required · sensitive

Client secret of the service principal.

- rule: {"required":true}

### spec.backupStorage.prefix

`string`

Directory prefix under which Velero data is stored within the
bucket/container — the multi-cluster pattern: one bucket, one prefix
per cluster.

Backup NAMES are keys under this prefix, and the store is the source of
truth: deleting a Backup resource from the cluster does NOT remove its
records from the bucket, and Velero's location sync re-imports whatever
the store holds (verified live) — so a backup name can never be reused
against a prefix that still holds a prior backup's records ("already
exists in object storage"). Give ad-hoc backups unique names, let each
backup's TTL expire its store records, and never point two clusters at
the same prefix unless the second is deliberately restoring the
first's backups.

### spec.backupStorage.pluginImage

`string`

Override the provider plugin image (default: the arm's official
plugin at the version paired with the chart's Velero release, e.g.
velero/velero-plugin-for-aws:v1.14.2 for Velero 1.18). Set for
private registry mirrors or deliberate version pins.

### spec.volumeSnapshots

`KubernetesVeleroVolumeSnapshots`

Volume snapshotting via the cloud provider / CSI. Controls the
default VolumeSnapshotLocation and the CSI feature flag.

- rule: default_snapshot_move_data rides CSI snapshots — set enable_csi too

### spec.volumeSnapshots.enabled

`bool` · optional (explicit presence)

Create the default VolumeSnapshotLocation for the backend arm's
provider. Chart default: true. Disable on clusters without snapshot
support (kind, bare metal without CSI snapshots) — file-system backup
then carries volume data.

- default: `true`

### spec.volumeSnapshots.enableCsi

`bool`

Enable CSI snapshot support (features: EnableCSI): Velero snapshots
PVCs through the CSI snapshot API instead of provider-native
instance snapshots — the modern path on EKS/GKE/AKS with CSI
drivers.

Requires the cluster's snapshot controller AND a VolumeSnapshotClass
for the volume's CSI driver labeled
`velero.io/csi-volumesnapshot-class: "true"` — without that class the
backup still reports Completed while the volumes were never
snapshotted (they ride fs-backup if enabled, or carry no data at
all). Verify volume snapshots appear in the backup's own resource
list before trusting it.

### spec.volumeSnapshots.defaultSnapshotMoveData

`bool`

Move CSI snapshot data INTO the backup store by default (Velero's
data mover) instead of leaving snapshots in the cloud provider —
makes backups portable across clusters/regions. Requires enable_csi
and the node-agent.

### spec.fsBackup

`KubernetesVeleroFsBackup`

File-system backup (kopia): the node-agent DaemonSet reads volume
data directly from pods — the volume-data path that works on ANY
volume type, including clusters without CSI snapshot support.

- rule: default_volumes_to_fs_backup requires deploy_node_agent — the node-agent is what reads volume data

### spec.fsBackup.deployNodeAgent

`bool`

Deploy the node-agent DaemonSet (required for file-system backup AND
for CSI snapshot data movement). Runs privileged host-path mounts of
the kubelet pod directory — that is how it reads volume data.

### spec.fsBackup.defaultVolumesToFsBackup

`bool`

Back up ALL pod volumes via file-system backup by default, without
per-pod annotations (Velero's --default-volumes-to-fs-backup).
Without this, only pods annotated backup.velero.io/backup-volumes
get fs-backup.

### spec.fsBackup.nodeAgentResources

`ContainerResources`

Node-agent container resources.

### spec.fsBackup.nodeAgentResources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.fsBackup.nodeAgentResources.limits.cpu

`string`

### spec.fsBackup.nodeAgentResources.limits.memory

`string`

### spec.fsBackup.nodeAgentResources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.fsBackup.nodeAgentResources.requests.cpu

`string`

### spec.fsBackup.nodeAgentResources.requests.memory

`string`

### spec.fsBackup.nodeAgentTolerations

`[]WorkloadToleration`

Node-agent tolerations (the agent must run on EVERY node whose
volumes need backup — tolerate dedicated-pool taints here).

### spec.fsBackup.nodeAgentTolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.fsBackup.nodeAgentTolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.fsBackup.nodeAgentTolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.fsBackup.nodeAgentTolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.fsBackup.nodeAgentTolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.schedules

`[]KubernetesVeleroSchedule`

Scheduled backups to create with the installation (rendered as
Schedule resources by the chart). Each entry is a cron schedule plus
a backup template.

- rule: each schedule needs a distinct name

### spec.schedules[].name

`string` · required

Schedule name (the Schedule object's name; backups are named
<name>-<timestamp>).

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.schedules[].schedule

`string` · required

Cron expression (five fields, or @daily/@hourly/...) — when backups
run.

- rule: {"required":true,"string":{"pattern":"^(@(annually|yearly|monthly|weekly|daily|midnight|hourly))|((.+)\\s(.+)\\s(.+)\\s(.+)\\s(.+))$"}}

### spec.schedules[].paused

`bool`

Create the schedule paused (backups only run after unpausing).

### spec.schedules[].ttl

`string`

How long backups from this schedule are kept before garbage
collection (Go duration; Velero default 720h = 30 days).

- rule: ttl must be a Go duration such as "720h"

### spec.schedules[].includedNamespaces

`[]string`

Namespaces to include (empty = all namespaces).

### spec.schedules[].excludedNamespaces

`[]string`

Namespaces to exclude.

### spec.schedules[].includedResources

`[]string`

Resource types to include (e.g. "deployments", "configmaps"; empty =
all).

### spec.schedules[].excludedResources

`[]string`

Resource types to exclude.

### spec.schedules[].labelSelector

`map<string, string>`

Only back up resources matching this label selector.

### spec.schedules[].includeClusterResources

`bool` · optional (explicit presence)

Include cluster-scoped resources: unset = auto (Velero includes
those associated with included namespaces), true = all, false =
none.

### spec.schedules[].snapshotVolumes

`bool` · optional (explicit presence)

Snapshot volumes for this schedule: unset = Velero's default
behavior; false = resources only.

### spec.schedules[].defaultVolumesToFsBackup

`bool` · optional (explicit presence)

Use file-system backup for all pod volumes in this schedule
(overrides the install-wide fs_backup default per schedule).

### spec.schedules[].storageLocation

`string`

Storage location for this schedule's backups (empty = the default
BackupStorageLocation).

### spec.server

`KubernetesVeleroServer`

Velero server tuning (TTLs, timeouts, logging).

### spec.server.defaultBackupTtl

`string`

Default TTL for backups without an explicit one (Velero default
720h).

- rule: default_backup_ttl must be a Go duration such as "720h"

### spec.server.defaultItemOperationTimeout

`string`

Ceiling for a single item operation (large-volume uploads; Velero
default 4h).

- rule: default_item_operation_timeout must be a Go duration such as "4h"

### spec.server.garbageCollectionFrequency

`string`

How often expired backups are garbage-collected (Velero default 1h).

- rule: garbage_collection_frequency must be a Go duration such as "1h"

### spec.server.restoreOnlyMode

`bool`

Run the server in restore-only mode (backups, schedules and GC are
blocked) — the posture for a DR-standby cluster reading another
cluster's store.

### spec.server.logLevel

`string` · optional (explicit presence)

Server log level: debug, info (default), warning, error, fatal,
panic.

- default: `info`
- rule: log_level must be one of 'debug', 'info', 'warning', 'error', 'fatal' or 'panic'

### spec.server.logFormat

`string` · optional (explicit presence)

Server log format: "text" (default) or "json".

- default: `text`
- rule: log_format must be 'text' or 'json'

### spec.deployment

`KubernetesVeleroDeployment`

Deployment sizing and platform scheduling for the Velero server pod.

### spec.deployment.resources

`ContainerResources`

Velero server container resources. Empty = chart defaults (no
requests/limits; upstream's documented starting point is 500m/128Mi
requests).

### spec.deployment.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.deployment.resources.limits.cpu

`string`

### spec.deployment.resources.limits.memory

`string`

### spec.deployment.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.deployment.resources.requests.cpu

`string`

### spec.deployment.resources.requests.memory

`string`

### spec.deployment.priorityClassName

`string`

PriorityClass for the server pod.

### spec.deployment.nodeSelector

`map<string, string>`

Node selector for the server pod.

### spec.deployment.tolerations

`[]WorkloadToleration`

Tolerations for the server pod.

### spec.deployment.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.deployment.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.deployment.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.deployment.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.deployment.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.prometheus

`KubernetesVeleroPrometheus`

Velero's own Prometheus telemetry.

- rule: service_monitor requires prometheus metrics to be enabled — the ServiceMonitor would have no metrics endpoint to scrape

### spec.prometheus.enabled

`bool` · optional (explicit presence)

Expose Velero's /metrics endpoint (backup successes/failures,
durations, validation results). Chart default: true — DR you cannot
observe is DR you cannot trust.

- default: `true`

### spec.prometheus.serviceMonitor

`bool`

Create a ServiceMonitor for scrape discovery. Requires the
Prometheus operator CRDs on the cluster — the release FAILS to
install without them.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (additional BackupStorageLocations/VolumeSnapshotLocations,
image overrides, RBAC tuning, extra volumes, plugin-config
ConfigMaps, ...) — never the substitute for them. Do not put secrets
here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesVelero, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace Velero was installed into (the resolved spec.namespace). |
| `status.outputs.release_name` | `string` | Helm release name — fixed "velero" (one installation per cluster; the server owns the cluster's backup records). |
| `status.outputs.service_account_name` | `string` | Name of the Velero server's Kubernetes service account — the subject cloud-side keyless bindings (IRSA trust policies, GCP WI bindings, Azure federated credentials) are written against. |
| `status.outputs.backup_storage_location_name` | `string` | Name of the default BackupStorageLocation ("default") — what Backup and Schedule resources reference through storageLocation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.backupStorage.s3.irsaRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.backupStorage.gcs.workloadIdentityServiceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.backupStorage.azureBlob.workloadIdentityClientId` | AzureUserAssignedIdentity | `status.outputs.client_id` |

## See Also

- [Overview](../README.md)
