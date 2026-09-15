# KubernetesOpenBao

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesOpenBaoSpec** installs OpenBao — the open-source,
Linux Foundation-governed secrets manager (MPL-2.0 fork of Vault) —
from the official `openbao` chart
(https://openbao.github.io/openbao-helm, chart 0.28.x = OpenBao 2.6.x).

THE SEAL LIFECYCLE — the fact everything else follows from: a fresh
OpenBao server starts UNINITIALIZED and SEALED. Initialization
(`bao operator init` — generates the unseal key shares and the
initial root token) and unsealing are RUNTIME operations performed
against the API after deploy; no Kubernetes deployment tool can do
them declaratively, and this component deliberately does not try.
Until a server is initialized and unsealed, its pod reports
NotReady BY DESIGN (the readiness probe is `bao status`, which
exits non-zero for sealed servers) — the chart keeps sealed pods
addressable through its Services (publishNotReadyAddresses), so
`kubectl port-forward` and the DNS names work for the init/unseal
calls. A deployment that never becomes "ready" until you initialize
it is the designed behavior, not a failure. Auto-unseal (below)
removes the UNSEAL step from restarts, but the one-time
initialization is always yours.

ONE SERVER MODE at a time: dev XOR standalone XOR ha (Raft). When
no mode is declared, standalone is used — the chart's own default.
The server always runs as a StatefulSet with an OnDelete update
strategy (config changes never roll pods automatically; delete pods
to pick up config).

NAME BUDGET: `metadata.name` is the Helm release name, and the chart
derives every Service name by suffixing it (`-internal` always,
`-agent-injector-svc` with the injector), while the module names the
backup CronJob `<name>-backup`. Kubernetes caps Service names at 63
characters and CronJob names at 52, so the longest name that fits is
54 characters, 44 with `injector.enabled`, and 45 with `backup`
declared (the tightest wins when both apply). A longer name fails the
deploy before anything is created, with a message naming the budget.

## Example

```yaml
# Full-surface development manifest — exercises every module-rendered arm
# so the offline plan/preview proofs cover what the kind-cluster lanes
# exclude (HA Raft with synthesized retry_join, TLS listener wiring, a
# declared-credential auto-unseal seal, the injector, metrics +
# ServiceMonitor, audit storage, the KEYLESS S3 backup arm through EKS
# IRSA — the one store posture no lane can prove: the kind lanes back up
# with declared keys to an in-cluster store and the GKE lanes prove GCS
# and R2, so the IRSA arm's rendering lives here).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenBao
metadata:
  name: bao-dev
spec:
  namespace:
    value: openbao-dev
  createNamespace: true
  server:
    ha:
      replicas: 3
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 512Mi
    dataStorage:
      size: 10Gi
    auditStorage:
      size: 5Gi
    logLevel: info
    logFormat: json
    scheduling:
      nodeSelector:
        kubernetes.io/os: linux
      tolerations:
        - key: dedicated
          operator: Equal
          value: secrets
          effect: NoSchedule
  tls:
    enabled: true
    certSecretName:
      value: bao-dev-tls
  autoUnseal:
    awsKms:
      region: us-west-2
      kmsKeyId: alias/openbao-unseal
      accessKeyId: AKIAEXAMPLEDEVONLY
      secretAccessKey: dev-only-placeholder-secret-key
  injector:
    enabled: true
    replicas: 1
    failurePolicy: Ignore
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        cpu: 250m
        memory: 128Mi
  uiEnabled: true
  networkPolicyEnabled: true
  metrics:
    enabled: true
    serviceMonitorEnabled: true
  backup:
    schedule: "0 */6 * * *"
    retentionDays: 7
    objectStore:
      prefix: openbao/bao-dev
      s3:
        bucket: bao-dev-snapshots
        region: us-west-2
        keyless: true
    workloadIdentity:
      eks:
        roleArn:
          value: arn:aws:iam::111122223333:role/bao-dev-backup
    auth:
      mountPath: kubernetes
  serviceAccount:
    annotations:
      eks.amazonaws.com/role-arn: arn:aws:iam::111122223333:role/bao-dev-unseal
    authDelegatorEnabled: true
  helmValues: |
    server:
      annotations:
        example.planton.ai/full-surface: "true"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.28.6` |  |
| `spec.server` | `KubernetesOpenBaoServer` |  |  |  |
| `spec.server.dev` | `KubernetesOpenBaoDevMode` |  |  |  |
| `spec.server.standalone` | `KubernetesOpenBaoStandaloneMode` |  |  |  |
| `spec.server.ha` | `KubernetesOpenBaoHaMode` |  |  |  |
| `spec.server.ha.replicas` | `int32` |  | `3` |  |
| `spec.server.resources` | `ContainerResources` |  |  |  |
| `spec.server.resources.limits` | `CpuMemory` |  |  |  |
| `spec.server.resources.limits.cpu` | `string` |  |  |  |
| `spec.server.resources.limits.memory` | `string` |  |  |  |
| `spec.server.resources.requests` | `CpuMemory` |  |  |  |
| `spec.server.resources.requests.cpu` | `string` |  |  |  |
| `spec.server.resources.requests.memory` | `string` |  |  |  |
| `spec.server.dataStorage` | `KubernetesOpenBaoStorage` |  |  |  |
| `spec.server.dataStorage.size` | `string` |  | `10Gi` |  |
| `spec.server.dataStorage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.server.auditStorage` | `KubernetesOpenBaoStorage` |  |  |  |
| `spec.server.auditStorage.size` | `string` |  | `10Gi` |  |
| `spec.server.auditStorage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.server.logLevel` | `string` |  | `info` |  |
| `spec.server.logFormat` | `string` |  | `standard` |  |
| `spec.server.scheduling` | `KubernetesOpenBaoScheduling` |  |  |  |
| `spec.server.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.server.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.server.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.server.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.server.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.server.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.server.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.tls` | `KubernetesOpenBaoTls` |  |  |  |
| `spec.tls.enabled` | `bool` |  |  |  |
| `spec.tls.certSecretName` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.autoUnseal` | `KubernetesOpenBaoAutoUnseal` |  |  |  |
| `spec.autoUnseal.awsKms` | `KubernetesOpenBaoAwsKmsSeal` |  |  |  |
| `spec.autoUnseal.awsKms.region` | `string` | yes |  |  |
| `spec.autoUnseal.awsKms.kmsKeyId` | `string` | yes |  |  |
| `spec.autoUnseal.awsKms.accessKeyId` | `string` |  |  |  |
| `spec.autoUnseal.awsKms.secretAccessKey` | `string` (sensitive) |  |  |  |
| `spec.autoUnseal.gcpKms` | `KubernetesOpenBaoGcpKmsSeal` |  |  |  |
| `spec.autoUnseal.gcpKms.project` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.autoUnseal.gcpKms.region` | `string` | yes |  |  |
| `spec.autoUnseal.gcpKms.keyRing` | `string \| valueFrom` | yes |  | GcpKmsKeyRing (`status.outputs.key_ring_name`) |
| `spec.autoUnseal.gcpKms.cryptoKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_name`) |
| `spec.autoUnseal.gcpKms.workloadIdentityServiceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.autoUnseal.azureKeyVault` | `KubernetesOpenBaoAzureKeyVaultSeal` |  |  |  |
| `spec.autoUnseal.azureKeyVault.vaultName` | `string` | yes |  |  |
| `spec.autoUnseal.azureKeyVault.keyName` | `string` | yes |  |  |
| `spec.autoUnseal.azureKeyVault.tenantId` | `string` | yes |  |  |
| `spec.autoUnseal.azureKeyVault.clientId` | `string` |  |  |  |
| `spec.autoUnseal.azureKeyVault.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.autoUnseal.transit` | `KubernetesOpenBaoTransitSeal` |  |  |  |
| `spec.autoUnseal.transit.address` | `string` | yes |  |  |
| `spec.autoUnseal.transit.keyName` | `string` | yes |  |  |
| `spec.autoUnseal.transit.mountPath` | `string` |  | `transit/` |  |
| `spec.autoUnseal.transit.token` | `string` (sensitive) |  |  |  |
| `spec.injector` | `KubernetesOpenBaoInjector` |  |  |  |
| `spec.injector.enabled` | `bool` |  |  |  |
| `spec.injector.replicas` | `int32` |  | `1` |  |
| `spec.injector.failurePolicy` | `string` |  | `Ignore` |  |
| `spec.injector.resources` | `ContainerResources` |  |  |  |
| `spec.injector.resources.limits` | `CpuMemory` |  |  |  |
| `spec.injector.resources.limits.cpu` | `string` |  |  |  |
| `spec.injector.resources.limits.memory` | `string` |  |  |  |
| `spec.injector.resources.requests` | `CpuMemory` |  |  |  |
| `spec.injector.resources.requests.cpu` | `string` |  |  |  |
| `spec.injector.resources.requests.memory` | `string` |  |  |  |
| `spec.uiEnabled` | `bool` |  | `true` |  |
| `spec.networkPolicyEnabled` | `bool` |  |  |  |
| `spec.metrics` | `KubernetesOpenBaoMetrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  |  |  |
| `spec.metrics.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.serviceAccount` | `KubernetesOpenBaoServiceAccount` |  |  |  |
| `spec.serviceAccount.annotations` | `map<string, string>` |  |  |  |
| `spec.serviceAccount.authDelegatorEnabled` | `bool` |  | `true` |  |
| `spec.helmValues` | `string` |  |  |  |
| `spec.backup` | `KubernetesOpenBaoBackup` |  |  |  |
| `spec.backup.schedule` | `string` |  | `0 * * * *` |  |
| `spec.backup.retentionDays` | `int32` |  | `14` |  |
| `spec.backup.objectStore` | `KubernetesOpenBaoBackupObjectStore` | yes |  |  |
| `spec.backup.objectStore.prefix` | `string` |  |  |  |
| `spec.backup.objectStore.s3` | `KubernetesOpenBaoS3ObjectStore` |  |  |  |
| `spec.backup.objectStore.s3.bucket` | `string` | yes |  |  |
| `spec.backup.objectStore.s3.region` | `string` |  |  |  |
| `spec.backup.objectStore.s3.endpointUrl` | `string \| valueFrom` |  |  | KubernetesSeaweedFs (`status.outputs.s3_endpoint`) |
| `spec.backup.objectStore.s3.forcePathStyle` | `bool` |  |  |  |
| `spec.backup.objectStore.s3.caPem` | `string` |  |  |  |
| `spec.backup.objectStore.s3.keyless` | `bool` |  |  |  |
| `spec.backup.objectStore.s3.accessKeys` | `KubernetesOpenBaoS3AccessKeys` |  |  |  |
| `spec.backup.objectStore.s3.accessKeys.accessKeyId` | `string` | yes |  |  |
| `spec.backup.objectStore.s3.accessKeys.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.backup.objectStore.gcs` | `KubernetesOpenBaoGcsObjectStore` |  |  |  |
| `spec.backup.objectStore.gcs.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.backup.objectStore.gcs.keyless` | `bool` |  |  |  |
| `spec.backup.objectStore.gcs.serviceAccountKey` | `string \| valueFrom` (sensitive) |  |  | GcpServiceAccount (`status.outputs.key_base64`) |
| `spec.backup.objectStore.azureBlob` | `KubernetesOpenBaoAzureBlobObjectStore` |  |  |  |
| `spec.backup.objectStore.azureBlob.storageAccount` | `string` |  |  |  |
| `spec.backup.objectStore.azureBlob.container` | `string` | yes |  |  |
| `spec.backup.objectStore.azureBlob.keyless` | `bool` |  |  |  |
| `spec.backup.objectStore.azureBlob.storageKey` | `string` (sensitive) |  |  |  |
| `spec.backup.objectStore.azureBlob.connectionString` | `string` (sensitive) |  |  |  |
| `spec.backup.objectStore.r2` | `KubernetesOpenBaoR2ObjectStore` |  |  |  |
| `spec.backup.objectStore.r2.bucket` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.bucket_name`) |
| `spec.backup.objectStore.r2.accountId` | `string \| valueFrom` | yes |  | CloudflareR2Bucket (`status.outputs.account_id`) |
| `spec.backup.objectStore.r2.jurisdiction` | `string \| valueFrom` |  |  | CloudflareR2Bucket (`status.outputs.jurisdiction`) |
| `spec.backup.objectStore.r2.credentials` | `KubernetesOpenBaoR2Credentials` | yes |  |  |
| `spec.backup.objectStore.r2.credentials.accessKeyId` | `string \| valueFrom` | yes |  | CloudflareAccountApiToken (`status.outputs.r2_access_key_id`) |
| `spec.backup.objectStore.r2.credentials.secretAccessKey` | `string \| valueFrom` (sensitive) | yes |  | CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`) |
| `spec.backup.workloadIdentity` | `KubernetesWorkloadIdentity` |  |  |  |
| `spec.backup.workloadIdentity.gke` | `KubernetesWorkloadIdentityGke` |  |  |  |
| `spec.backup.workloadIdentity.gke.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.backup.workloadIdentity.eks` | `KubernetesWorkloadIdentityEksIrsa` |  |  |  |
| `spec.backup.workloadIdentity.eks.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.backup.workloadIdentity.aks` | `KubernetesWorkloadIdentityAks` |  |  |  |
| `spec.backup.workloadIdentity.aks.clientId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.backup.workloadIdentity.aks.tenantId` | `string` |  |  |  |
| `spec.backup.auth` | `KubernetesOpenBaoBackupAuth` |  |  |  |
| `spec.backup.auth.mountPath` | `string` |  | `kubernetes` |  |
| `spec.backup.auth.role` | `string` |  |  |  |
| `spec.backup.images` | `KubernetesOpenBaoBackupImages` |  |  |  |
| `spec.backup.images.openbao` | `ContainerImage` |  |  |  |
| `spec.backup.images.openbao.repo` | `string` |  |  |  |
| `spec.backup.images.openbao.tag` | `string` |  |  |  |
| `spec.backup.images.openbao.pullSecretName` | `string` |  |  |  |
| `spec.backup.images.rclone` | `ContainerImage` |  |  |  |
| `spec.backup.images.rclone.repo` | `string` |  |  |  |
| `spec.backup.images.rclone.tag` | `string` |  |  |  |
| `spec.backup.images.rclone.pullSecretName` | `string` |  |  |  |
| `spec.backup.resources` | `ContainerResources` |  |  |  |
| `spec.backup.resources.limits` | `CpuMemory` |  |  |  |
| `spec.backup.resources.limits.cpu` | `string` |  |  |  |
| `spec.backup.resources.limits.memory` | `string` |  |  |  |
| `spec.backup.resources.requests` | `CpuMemory` |  |  |  |
| `spec.backup.resources.requests.cpu` | `string` |  |  |  |
| `spec.backup.resources.requests.memory` | `string` |  |  |  |
| `spec.restore` | `KubernetesOpenBaoRestore` |  |  |  |
| `spec.restore.snapshotKey` | `string` |  |  |  |
| `spec.restore.latest` | `bool` |  |  |  |
| `spec.restore.rootToken` | `KubernetesSecretKey` | yes |  |  |
| `spec.restore.rootToken.name` | `string` |  |  |  |
| `spec.restore.rootToken.key` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into (conventionally "openbao"). Accepts a
literal namespace name or a reference to a KubernetesNamespace
resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "0.28.6" = OpenBao v2.6.1 —
the chart's appVersion pins the server image). Versions must
exist in the SERVED index at https://openbao.github.io/openbao-helm.

- default: `0.28.6`

### spec.server

`KubernetesOpenBaoServer`

The OpenBao server: mode, sizing, storage, and logging.

### spec.server.dev

`KubernetesOpenBaoDevMode`

Dev mode: in-memory, auto-initialized, auto-unsealed, root
token literally "root". NEVER for real secrets — all data is
lost on every restart, and the root token is plaintext in the
pod spec (readable by anyone who can get pods). Exists so the
component can be evaluated and composed against without the
init/unseal ceremony. No PVC is created in dev mode, and
workload-identity ServiceAccount annotations are NOT applied
(a chart behavior — dev mode drops them).

### spec.server.standalone

`KubernetesOpenBaoStandaloneMode`

Standalone: one instance, `storage "file"` on the data PVC.
The production shape for single-instance installs.

### spec.server.ha

`KubernetesOpenBaoHaMode`

High availability with integrated Raft storage: every replica
persists to its own data PVC and the cluster elects a leader.
This module renders `retry_join` stanzas for every peer (the
chart alone ships NONE — without them a multi-replica Raft
install never forms a cluster and each pod sits uninitialized
and independent). Bootstrap: initialize pod-0 and unseal every
pod; joins then happen automatically through retry_join.

### spec.server.ha.replicas

`int32` · optional (explicit presence)

Number of server replicas (Raft peers). Odd counts (3, 5)
tolerate minority loss; 3 is the standard production shape. A
single replica is a legal Raft cluster of one (useful in labs).
Remember the chart's default required anti-affinity: replicas
beyond the node count stay Pending (see scheduling).

- default: `3`
- rule: {"int32":{"lte":11,"gte":1}}

### spec.server.resources

`ContainerResources`

CPU and memory for the server container. The chart ships no
defaults; these are modest laboratory defaults — size real
installs to the workload.

### spec.server.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.server.resources.limits.cpu

`string`

### spec.server.resources.limits.memory

`string`

### spec.server.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.server.resources.requests.cpu

`string`

### spec.server.resources.requests.memory

`string`

### spec.server.dataStorage

`KubernetesOpenBaoStorage`

The data volume (file storage in standalone, Raft storage in HA).
Ignored in dev mode (in-memory). One PVC per replica, mounted at
/openbao/data.

### spec.server.dataStorage.size

`string` · optional (explicit presence)

Volume size (e.g. "10Gi").

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.server.dataStorage.storageClass

`string | valueFrom`

StorageClass name. Empty uses the cluster's default class.
Accepts a literal name or a reference to a
KubernetesStorageClass resource.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.server.auditStorage

`KubernetesOpenBaoStorage`

Optional dedicated volume for file audit logs, mounted at
/openbao/audit. Creating the volume does NOT enable auditing —
after initialization run
`bao audit enable file file_path=/openbao/audit/audit.log`.

### spec.server.auditStorage.size

`string` · optional (explicit presence)

Volume size (e.g. "10Gi").

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.server.auditStorage.storageClass

`string | valueFrom`

StorageClass name. Empty uses the cluster's default class.
Accepts a literal name or a reference to a
KubernetesStorageClass resource.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.server.logLevel

`string` · optional (explicit presence)

Server log verbosity: trace, debug, info (default), warn, error.

- default: `info`
- rule: Log level must be one of: trace, debug, info, warn, error.

### spec.server.logFormat

`string` · optional (explicit presence)

Server log format: standard (default) or json.

- default: `standard`
- rule: Log format must be either "standard" or "json".

### spec.server.scheduling

`KubernetesOpenBaoScheduling`

Pod scheduling constraints for the server pods. NOTE the chart
ships a REQUIRED pod anti-affinity on hostname by default, so an
HA cluster needs as many schedulable nodes as replicas; relax it
through `helm_values` (`server.affinity: ""`) when running
multiple replicas on fewer nodes (labs only — co-located Raft
replicas share their node's fate).

### spec.server.scheduling.nodeSelector

`map<string, string>`

Schedule onto nodes carrying these labels.

### spec.server.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes.

### spec.server.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.server.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.server.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.server.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.server.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.tls

`KubernetesOpenBaoTls`

End-to-end TLS for the OpenBao listener. When unset, the server
listens on plain HTTP inside the cluster (the chart default).
Enabling TLS is a COMPOSITE change this module owns end to end:
the listener gets tls_cert_file/tls_key_file, the certificate
Secret is mounted, and every derived URL and probe switches to
https. (The chart's own `global.tlsDisable` value alone does NOT
configure the listener — flipping it without listener changes
produces a plaintext server addressed as https, an instant
outage. This module renders all the pieces together.)

- rule: TLS needs certificate material: set cert_secret_name to a kubernetes.io/tls Secret (or reference a KubernetesCertificate) when TLS is enabled.

### spec.tls.enabled

`bool`

Enable TLS on the listener (port 8200) and the cluster port
(8201 always uses OpenBao's own cluster TLS regardless).

### spec.tls.certSecretName

`string | valueFrom`

Name of a kubernetes.io/tls Secret in the install namespace
carrying tls.crt / tls.key (and optionally ca.crt) for the
server. Accepts a literal name or a reference to a
KubernetesCertificate resource (cert-manager) — the natural
issuer: point the certificate's dnsNames at
`<name>.<namespace>.svc` and this component's derived DNS names.
With `backup` declared the dnsNames must also include the
active-leader Service, `<name>-active.<namespace>.svc`: the backup
and restore jobs address the leader through it, and a certificate
without that name fails every run with an x509 error whose log line
names the Service to add. Required when enabled.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.autoUnseal

`KubernetesOpenBaoAutoUnseal`

Auto-unseal: delegate master-key protection to an external KMS so
servers unseal themselves on startup (initialization is still a
one-time manual step; it produces RECOVERY keys instead of unseal
keys). Exactly one seal backend may be declared.

VERSION HORIZON (verified at OpenBao v2.6.1): the cloud KMS seal
mechanisms (awskms, gcpckms, azurekeyvault) are built in but
DEPRECATED — upstream moves them to external KMS plugins in
v2.7.0. This module renders the seal stanza for the pinned 2.6.x
line; expect the rendering to gain a plugin declaration when the
chart pin crosses 2.7.

### spec.autoUnseal.awsKms

`KubernetesOpenBaoAwsKmsSeal`

AWS KMS. Natural on EKS (IRSA for keyless auth).

### spec.autoUnseal.awsKms.region

`string` · required

AWS region of the KMS key (e.g. "us-west-2").

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.awsKms.kmsKeyId

`string` · required

KMS key ID or full ARN of a SYMMETRIC encrypt/decrypt key.

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.awsKms.accessKeyId

`string`

Static credentials — only when keyless (IRSA / instance profile)
is unavailable. The module materializes them into a Secret
(`<name>-seal-credentials`) delivered as environment variables;
nothing credential-bearing lands in the config ConfigMap.

### spec.autoUnseal.awsKms.secretAccessKey

`string` · sensitive

The secret access key paired with access_key_id.

### spec.autoUnseal.gcpKms

`KubernetesOpenBaoGcpKmsSeal`

GCP Cloud KMS. Natural on GKE (Workload Identity).

### spec.autoUnseal.gcpKms.project

`string | valueFrom` · required

GCP project containing the KMS key ring.

containment_exempt: names where the unseal key lives — the server
runs in the cluster, not the project.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.autoUnseal.gcpKms.region

`string` · required

KMS key ring region (e.g. "global", "us-central1").

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.gcpKms.keyRing

`string | valueFrom` · required

Key ring name.

containment_exempt: an unseal-key source the server calls out to —
access, never placement.

- references: GcpKmsKeyRing (`status.outputs.key_ring_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKeyRing, name: <that resource's name>, fieldPath: status.outputs.key_ring_name}} -- a bare string does not parse

### spec.autoUnseal.gcpKms.cryptoKey

`string | valueFrom` · required

Crypto key (symmetric encrypt/decrypt) used to wrap the master
key. The identity running OpenBao needs TWO roles on it:
roles/cloudkms.cryptoKeyEncrypterDecrypter to wrap on init and
unwrap on every unseal, AND roles/cloudkms.viewer — the server reads
the key's metadata when it configures the seal at start (a
key-existence check), and the encrypter-decrypter role does not
carry cloudkms.cryptoKeys.get. With only the first role the pod
crash-loops with "Error configuring seal \"gcpckms\": ... Permission
'cloudkms.cryptoKeys.get' denied" and init never opens. Two
GcpKmsKeyIamMember resources, one per role, scoped to the key.

- references: GcpKmsKey (`status.outputs.key_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_name}} -- a bare string does not parse

### spec.autoUnseal.gcpKms.workloadIdentityServiceAccount

`string | valueFrom`

GKE Workload Identity: the GCP service account email to annotate
the server ServiceAccount with (iam.gke.io/gcp-service-account).
Leave empty to rely on node/ambient credentials. NOTE dev mode
drops ServiceAccount annotations (chart behavior) — auto-unseal
with workload identity requires standalone or ha mode.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.autoUnseal.azureKeyVault

`KubernetesOpenBaoAzureKeyVaultSeal`

Azure Key Vault. Natural on AKS (Workload Identity / MSI).

### spec.autoUnseal.azureKeyVault.vaultName

`string` · required

Key Vault name (the vault, not the key).

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.azureKeyVault.keyName

`string` · required

Name of the key inside the vault.

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.azureKeyVault.tenantId

`string` · required

Entra (Azure AD) tenant ID.

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.azureKeyVault.clientId

`string`

Service-principal client ID — only when keyless (AKS Workload
Identity / Managed Identity) is unavailable.

### spec.autoUnseal.azureKeyVault.clientSecret

`string` · sensitive

Service-principal client secret paired with client_id. Delivered
as environment variables from a module-owned Secret; never lands
in the config ConfigMap.

### spec.autoUnseal.transit

`KubernetesOpenBaoTransitSeal`

Transit engine of another OpenBao/Vault instance.

### spec.autoUnseal.transit.address

`string` · required

Address of the central instance (e.g. "https://bao.example.com:8200").
The satellite depends on the central instance being reachable
and unsealed at every startup.

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.transit.keyName

`string` · required

Transit key name used to wrap the master key. The key need not exist
beforehand: the transit engine creates a key on its first encrypt
(the server's own startup test-encrypt does it) — the ENGINE must
exist, the key may be born there.

- rule: {"string":{"minLen":"1"}}

### spec.autoUnseal.transit.mountPath

`string` · optional (explicit presence)

Transit engine mount path. The engine must be enabled on the central
instance before this satellite starts (`bao secrets enable
-path=transit transit` there); a satellite whose seal finds no engine
exits at startup with "Error configuring seal" and crash-loops until
the engine exists.

- default: `transit/`

### spec.autoUnseal.transit.token

`string` · sensitive

Token authorized for encrypt/decrypt on the transit key.
Delivered as environment variables from a module-owned Secret;
never lands in the config ConfigMap.

### spec.injector

`KubernetesOpenBaoInjector`

The OpenBao Agent Injector: a MutatingWebhookConfiguration that
intercepts pod creation CLUSTER-WIDE and injects secret-fetching
agent sidecars into annotated pods.

OFF by default here — a deliberate divergence from the chart
(whose default installs the webhook for every pod create/update
in the cluster). Enable it only when workloads will actually use
agent injection annotations. The webhook fails OPEN by default
(failure_policy Ignore), so injector downtime never blocks pod
creation — it silently skips injection instead.

### spec.injector.enabled

`bool`

Deploy the injector. See the spec-level comment for the
cluster-wide webhook blast radius.

### spec.injector.replicas

`int32` · optional (explicit presence)

Injector replicas. Above 1, leader election activates and the
chart creates a HARD-CODED Secret `openbao-injector-certs` —
only one multi-replica injector can exist per namespace.
The injector pods also carry a required anti-affinity, so
replicas need distinct nodes.

- default: `1`
- rule: {"int32":{"lte":5,"gte":1}}

### spec.injector.failurePolicy

`string` · optional (explicit presence)

Webhook failure policy: "Ignore" (default — injector downtime
skips injection, pods still schedule) or "Fail" (pod creation
BLOCKS while the injector is down; only for clusters that treat
missing secrets as worse than blocked deploys).

- default: `Ignore`
- rule: Webhook failure policy must be either "Ignore" (fail open) or "Fail" (fail closed).

### spec.injector.resources

`ContainerResources`

CPU and memory for the injector container.

### spec.injector.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.injector.resources.limits.cpu

`string`

### spec.injector.resources.limits.memory

`string`

### spec.injector.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.injector.resources.requests.cpu

`string`

### spec.injector.resources.requests.memory

`string`

### spec.uiEnabled

`bool` · optional (explicit presence)

Serve the built-in web UI and create the `<name>-ui` Service for
it. Defaults to true — the UI is part of the product experience.
(Exposure beyond the cluster composes from Gateway API kinds
referencing the exported service handles; this component never
creates ingress resources.)

- default: `true`

### spec.networkPolicyEnabled

`bool`

Render the chart's NetworkPolicy for the server pods (ingress on
8200/8201 from all namespaces by default). Off by default; most
clusters manage network policy through first-class
KubernetesNetworkPolicy resources instead.

### spec.metrics

`KubernetesOpenBaoMetrics`

Prometheus metrics. Enabling this renders the telemetry stanza
into the server config (prometheus_retention_time plus
unauthenticated_metrics_access on the listener — the metrics
endpoint is UNAUTHENTICATED when enabled, scoped to
/v1/sys/metrics) and optionally a ServiceMonitor.

- rule: A ServiceMonitor without the telemetry stanza scrapes an endpoint that rejects every request — enable metrics too.

### spec.metrics.enabled

`bool`

Render the telemetry stanza (prometheus_retention_time = "30s",
disable_hostname) and allow UNAUTHENTICATED access to
/v1/sys/metrics on the listener. Without this, the metrics
endpoint requires an OpenBao token and Prometheus cannot scrape.

### spec.metrics.serviceMonitorEnabled

`bool`

Also create a ServiceMonitor (requires the Prometheus Operator
CRDs — a KubernetesKubePrometheusStack — on the cluster; the
install FAILS without them). In HA mode the chart scrapes only
the active node.

### spec.serviceAccount

`KubernetesOpenBaoServiceAccount`

Server ServiceAccount identity: cloud workload-identity
annotations and the Kubernetes-auth delegation binding.

### spec.serviceAccount.annotations

`map<string, string>`

Annotations for the server ServiceAccount — the cloud
workload-identity seam (eks.amazonaws.com/role-arn,
iam.gke.io/gcp-service-account, azure.workload.identity/client-id).
NOTE: dev mode drops these (chart behavior).

### spec.serviceAccount.authDelegatorEnabled

`bool` · optional (explicit presence)

Bind the ServiceAccount to the cluster's system:auth-delegator
role (a ClusterRoleBinding). Required for OpenBao's Kubernetes
AUTH METHOD to validate workload tokens via TokenReview — leave
on unless the cluster forbids the binding and you will not use
Kubernetes auth.

- default: `true`

### spec.helmValues

`string`

Advanced escape hatch: raw Helm values merged LAST (Helm `-f`
semantics) over everything this spec renders — later keys win.
Use it for the chart surfaces deliberately not modeled (CSI
provider, injector webhook selectors, extra volumes, affinity
overrides). The module re-pins `fullnameOverride` after the
merge, so resource naming cannot be overridden. YAML document as
a string.

### spec.backup

`KubernetesOpenBaoBackup`

Scheduled Raft snapshots to an object store — the disaster-recovery
story for this vault. Omitted = no backups (a deliberate choice to
make, not a default to forget). Declaring it renders a CronJob that
takes a snapshot through OpenBao's own API and ships it with rclone
to S3, Google Cloud Storage, Azure Blob, or Cloudflare R2, in that
store's own vocabulary, wired by reference to the catalog's bucket,
identity, and token kinds; retention prunes older objects.

RAFT ONLY: snapshots exist only for integrated Raft storage
(`server.ha`; a single replica is a legal Raft cluster of one).
Standalone file storage and dev mode have no snapshot API, and the
server refuses the call ("raft storage is not in use").

ONE STEP THE MODULE CANNOT TAKE: the job logs in through OpenBao's
Kubernetes auth method, and the policy, auth mount, and role it
needs live INSIDE OpenBao, which is sealed and uninitialized at
deploy time. After initialization, run the four-command recipe on
the `auth` field once; until then every backup run fails its login
and its log prints the same four commands with the real names.

- rule: A keyless store authenticates with the backup job's cloud identity — declare backup.workload_identity (the GCP service account email, AWS IAM role ARN, or Azure client id the job's ServiceAccount federates with), or declare the store's keys instead.

### spec.backup.schedule

`string` · optional (explicit presence)

When to snapshot, as a standard 5-field cron expression (minute
hour day-of-month month day-of-week). Default hourly. A Raft
snapshot is a full copy of the state, so the schedule is a trade
between recovery-point objective and object-store churn; hourly
suits most vaults, every 15 minutes a busy one.

- default: `0 * * * *`
- rule: Schedule must be a standard 5-field cron expression: minute hour day-of-month month day-of-week (e.g. "0 * * * *" for hourly, "*/15 * * * *" for every 15 minutes)

### spec.backup.retentionDays

`int32` · optional (explicit presence)

Delete snapshot objects under the prefix older than this many days
after each run (object age, not object count). Default 14. Zero
keeps every snapshot forever — the bucket's own lifecycle rules then
decide. Pruning runs under the prefix, so two live clusters must
never share one prefix; a restore target deliberately does share the
source's (see `restore`).

- default: `14`
- rule: {"int32":{"gte":0}}

### spec.backup.objectStore

`KubernetesOpenBaoBackupObjectStore` · required

Where the snapshots go. Exactly one store, in that store's own
vocabulary.

- rule: {"required":true}

### spec.backup.objectStore.prefix

`string`

Key prefix inside the bucket or container (a folder for this
vault's snapshots). One prefix per live cluster: retention prunes
under it, and a restore target declares this same prefix to read the
snapshots back — so two live vaults must never share one, while the
source and its restore target deliberately do. Empty = the bucket
root.

### spec.backup.objectStore.s3

`KubernetesOpenBaoS3ObjectStore`

AWS S3 — or ANY S3-compatible store (SeaweedFS, MinIO, Ceph RGW,
...) via the endpoint_url override. Cloudflare R2 has its own arm
(`r2`) that composes the catalog's Cloudflare kinds; this arm still
reaches R2 for a hand-carried endpoint and key pair.

- rule: keyless and access_keys are alternative credential postures — set exactly one
- rule: an S3-compatible endpoint (endpoint_url) authenticates with access_keys — the keyless posture only mints AWS credentials

### spec.backup.objectStore.s3.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.backup.objectStore.s3.region

`string`

AWS region of the bucket. Required for real S3; for S3-compatible
stores use the store's expected value (SeaweedFS and MinIO accept
any; the `r2` arm pins Cloudflare R2's `auto` itself).

### spec.backup.objectStore.s3.endpointUrl

`string | valueFrom`

S3-COMPATIBLE ARM: endpoint URL of the store (e.g.
http://main-s3.object-store.svc.cluster.local:8333 for an
in-cluster KubernetesSeaweedFs — its `s3_endpoint` output, the
default reference). Empty = real AWS S3. For Cloudflare R2 prefer
the `r2` arm, which composes the endpoint from the bucket's account
and jurisdiction.

- references: KubernetesSeaweedFs (`status.outputs.s3_endpoint`)
- rule: endpoint_url must be an http(s) URL (e.g. http://main-s3.object-store.svc.cluster.local:8333)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.backup.objectStore.s3.forcePathStyle

`bool`

Address the bucket as a path (`<endpoint>/<bucket>/...`) instead of
a virtual host (`<bucket>.<endpoint>/...`). Most S3-compatible
stores want this; real AWS S3 does not.

### spec.backup.objectStore.s3.caPem

`string`

PEM CA bundle for verifying a self-signed endpoint_url TLS
certificate (materialized as a Secret the job reads). Public
material, not a credential.

### spec.backup.objectStore.s3.keyless

`bool`

Keyless posture: the backup job's AWS identity (EKS IRSA through
`backup.workload_identity`) authenticates to S3 — no stored keys.
rclone resolves it through the AWS SDK's default chain, which reads
the web-identity token IRSA projects into the pod. Mutually
exclusive with access_keys; requires `backup.workload_identity`.

### spec.backup.objectStore.s3.accessKeys

`KubernetesOpenBaoS3AccessKeys`

Static access keys, materialized as a Kubernetes Secret the job
reads. The declared-credential arm — for S3-compatible stores and
clusters without IRSA.

### spec.backup.objectStore.s3.accessKeys.accessKeyId

`string` · required

Access key ID — the public identifier of the key pair, not a
secret; only the paired secret access key is a credential. For
SeaweedFS or MinIO this is the access key / username.

- rule: {"required":true}

### spec.backup.objectStore.s3.accessKeys.secretAccessKey

`string` · required · sensitive

Secret access key (for SeaweedFS or MinIO: the secret key /
password).

- rule: {"required":true}

### spec.backup.objectStore.gcs

`KubernetesOpenBaoGcsObjectStore`

Google Cloud Storage.

- rule: keyless and service_account_key are alternative credential postures — set exactly one

### spec.backup.objectStore.gcs.bucket

`string | valueFrom` · required

Bucket name. By reference to the bucket resource's `bucket_name`
output, so the store follows the bucket; a literal names a bucket
outside the catalog.

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.backup.objectStore.gcs.keyless

`bool`

Keyless posture: the backup job's GCP identity (GKE Workload
Identity through `backup.workload_identity`) authenticates to GCS —
no stored key. rclone resolves it through Application Default
Credentials, which on GKE is the metadata server's token for the
bound service account. Mutually exclusive with service_account_key;
requires `backup.workload_identity`.

THE IDENTITY NEEDS TWO ROLES ON THE BUCKET: rclone reads the
bucket's attributes before writing, and `roles/storage.objectAdmin`
does not carry `storage.buckets.get`. Grant
`roles/storage.objectAdmin` AND `roles/storage.legacyBucketReader`
(a GcpGcsBucket's `iam_members`, one entry each).

### spec.backup.objectStore.gcs.serviceAccountKey

`string | valueFrom` · sensitive

A GCP service-account key (the JSON key file's content, raw or
base64-encoded — the shape a GcpServiceAccount resource exports as
`key_base64` when declared with a `user_managed_key`), materialized
as a Kubernetes Secret the job reads. The declared-credential arm
for clusters outside GKE backing up to GCS. The same two bucket
roles apply to this account.

- references: GcpServiceAccount (`status.outputs.key_base64`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.key_base64}} -- a bare string does not parse

### spec.backup.objectStore.azureBlob

`KubernetesOpenBaoAzureBlobObjectStore`

Azure Blob Storage.

- rule: set exactly one Azure credential posture: keyless, connection_string, or storage_account + storage_key
- rule: storage_key authenticates a specific account — set storage_account with it
- rule: the keyless posture still needs storage_account — it identifies the storage endpoint

### spec.backup.objectStore.azureBlob.storageAccount

`string`

Storage-account name. Required with storage_key and with keyless
(it identifies the storage endpoint); a connection_string carries
its own.

### spec.backup.objectStore.azureBlob.container

`string` · required

Blob container name.

- rule: {"required":true}

### spec.backup.objectStore.azureBlob.keyless

`bool`

Keyless posture: the backup job's Azure identity (AKS Workload
Identity through `backup.workload_identity`) authenticates to Blob
Storage — no stored secret. rclone resolves it through the Azure
SDK's default credential chain, whose workload-identity step needs
BOTH halves: the ServiceAccount annotation the identity field
renders and the `azure.workload.identity/use: "true"` pod label,
which the module sets on the job. Mutually exclusive with the
declared-credential fields; requires `backup.workload_identity`.

### spec.backup.objectStore.azureBlob.storageKey

`string` · sensitive

Storage-account access key, paired with storage_account.

### spec.backup.objectStore.azureBlob.connectionString

`string` · sensitive

Storage-account connection string — the all-in-one declared
credential. Mutually exclusive with storage_key.

### spec.backup.objectStore.r2

`KubernetesOpenBaoR2ObjectStore`

Cloudflare R2, in R2's own vocabulary: the bucket, the owning
account, the bucket's jurisdiction, and a Cloudflare credential —
each a reference onto the catalog's CloudflareR2Bucket and
CloudflareAccountApiToken by default. The module performs the S3
translation R2 needs (the jurisdiction's endpoint host, region
`auto`, path-style addressing, the token as an S3 key pair);
nothing S3-shaped is typed here.

### spec.backup.objectStore.r2.bucket

`string | valueFrom` · required

Bucket name. By reference to the bucket resource's `bucket_name`
output, so the store follows the bucket; a literal names a bucket
outside the catalog.

- references: CloudflareR2Bucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.backup.objectStore.r2.accountId

`string | valueFrom` · required

The Cloudflare account that owns the bucket (32 hex characters). By
reference to the bucket resource's `account_id` output.

- references: CloudflareR2Bucket (`status.outputs.account_id`)
- rule: account_id is the 32-hex-character Cloudflare account id
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.account_id}} -- a bare string does not parse

### spec.backup.objectStore.r2.jurisdiction

`string | valueFrom`

The bucket's data-residency jurisdiction: `default` (or empty), `eu`,
`fedramp`, or `us`. It selects the S3 host the module composes — a
bucket created in a jurisdiction is unreachable through any other
host — so it must match the bucket exactly; by reference to the
bucket resource's `jurisdiction` output it cannot drift.

- references: CloudflareR2Bucket (`status.outputs.jurisdiction`)
- rule: jurisdiction must be one of "default", "eu", "fedramp", "us" (or empty for default)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareR2Bucket, name: <that resource's name>, fieldPath: status.outputs.jurisdiction}} -- a bare string does not parse

### spec.backup.objectStore.r2.credentials

`KubernetesOpenBaoR2Credentials` · required

The Cloudflare credential, as the S3 key pair R2's S3 API
authenticates. Materialized as the `<name>-backup-credentials`
Secret the job reads; never plaintext in the rendered resource.

- rule: {"required":true}

### spec.backup.objectStore.r2.credentials.accessKeyId

`string | valueFrom` · required

The S3 access key id: the API token's id. By reference to the token
resource's `r2_access_key_id` output.

- references: CloudflareAccountApiToken (`status.outputs.r2_access_key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_access_key_id}} -- a bare string does not parse

### spec.backup.objectStore.r2.credentials.secretAccessKey

`string | valueFrom` · required · sensitive

The S3 secret access key: the SHA-256 of the API token's value. By
reference to the token resource's `r2_secret_access_key` output. Rotates
with the token: a rotated token is a new key pair, and the Secret the
module materializes follows the reference on the next apply.

- references: CloudflareAccountApiToken (`status.outputs.r2_secret_access_key`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareAccountApiToken, name: <that resource's name>, fieldPath: status.outputs.r2_secret_access_key}} -- a bare string does not parse

### spec.backup.workloadIdentity

`KubernetesWorkloadIdentity`

Keyless cloud identity for the backup job's ServiceAccount
(`<name>-backup`, in `namespace`) — annotates it so rclone reaches
S3 (EKS IRSA), Google Cloud Storage (GKE Workload Identity), or
Azure Blob (AKS Workload Identity) without stored keys. Pair with
the store's `keyless: true`. The cloud-side binding names exactly
that ServiceAccount: on GKE a GcpGkeWorkloadIdentityBinding with
`ksa_name` = `<name>-backup` and `ksa_namespace` = this namespace,
one per OpenBao (a restore target is another OpenBao and needs its
own). This is the JOB's identity; the server's identity for KMS
auto-unseal is declared on `auto_unseal` and `service_account`.

### spec.backup.workloadIdentity.gke

`KubernetesWorkloadIdentityGke`

GKE Workload Identity: annotate the ServiceAccount with a GCP service account email.

### spec.backup.workloadIdentity.gke.serviceAccountEmail

`string | valueFrom` · required

GCP service account email, e.g. "dns-manager@my-project.iam.gserviceaccount.com".
Applied as the `iam.gke.io/gcp-service-account` annotation. Accepts a literal
email or a reference to a GcpServiceAccount resource's output.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.backup.workloadIdentity.eks

`KubernetesWorkloadIdentityEksIrsa`

EKS IRSA: annotate the ServiceAccount with an AWS IAM role ARN.

### spec.backup.workloadIdentity.eks.roleArn

`string | valueFrom` · required

AWS IAM role ARN, e.g. "arn:aws:iam::123456789012:role/dns-manager".
Applied as the `eks.amazonaws.com/role-arn` annotation. Accepts a literal ARN
or a reference to an AwsIamRole resource's output.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.backup.workloadIdentity.aks

`KubernetesWorkloadIdentityAks`

Azure AD Workload Identity: annotate the ServiceAccount with an Entra application
(or user-assigned managed identity) client ID.

### spec.backup.workloadIdentity.aks.clientId

`string | valueFrom` · required

Client ID (GUID) of the user-assigned managed identity or Entra application.
Applied as the `azure.workload.identity/client-id` annotation. Accepts a literal
GUID or a reference to an AzureUserAssignedIdentity resource's output.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.backup.workloadIdentity.aks.tenantId

`string` · optional (explicit presence)

Entra tenant ID (GUID). Optional: only needed for cross-tenant scenarios; when
omitted the azure-workload-identity webhook uses its default tenant. Applied as
the `azure.workload.identity/tenant-id` annotation when set.

- rule: {"string":{"uuid":true}}

### spec.backup.auth

`KubernetesOpenBaoBackupAuth`

How the job authenticates to OpenBao: the Kubernetes auth mount and
role from the login recipe above. Defaults name the recipe exactly
as printed; change them only to match an auth mount that already
exists on this cluster.

### spec.backup.auth.mountPath

`string` · optional (explicit presence)

Mount path of the Kubernetes auth method inside OpenBao (the
`-path` of `bao auth enable`). Default `kubernetes`.

- default: `kubernetes`

### spec.backup.auth.role

`string`

The role under that mount the job logs in with. Empty = the module
derives `<name>-backup`, the name the recipe and the exported
`backup_auth_role` output print; set it only to reuse a role that
already exists.

### spec.backup.images

`KubernetesOpenBaoBackupImages`

Image overrides for the two containers each run uses, for mirrors
and air-gapped registries. Unset = the server's own `openbao` image
at the chart's pinned tag, and the official `rclone/rclone` image at
the module's pinned tag.

### spec.backup.images.openbao

`ContainerImage`

The `bao` CLI container (takes and installs snapshots). Unset = the
server's own image at the chart's pinned tag; override only for a
mirror, and keep the tag equal to the server's.

### spec.backup.images.openbao.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.backup.images.openbao.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.backup.images.openbao.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.backup.images.rclone

`ContainerImage`

The rclone container (moves snapshots to and from the store). Unset
= the official `rclone/rclone` image at the module's pinned tag.

### spec.backup.images.rclone.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.backup.images.rclone.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.backup.images.rclone.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.backup.resources

`ContainerResources`

CPU and memory for the job's containers. A snapshot streams to a
scratch volume, so memory does not scale with vault size; these
laboratory defaults suit vaults up to a few hundred megabytes.

### spec.backup.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.backup.resources.limits.cpu

`string`

### spec.backup.resources.limits.memory

`string`

### spec.backup.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.backup.resources.requests.cpu

`string`

### spec.backup.resources.requests.memory

`string`

### spec.restore

`KubernetesOpenBaoRestore`

Restore this cluster from a snapshot in the declared backup store —
the "bad day" declaration: a FRESH cluster with the same `backup`
block as the source (same store, same prefix) and the same seal key,
plus this block. The module renders a one-shot Job that fetches the
named snapshot (or the newest under the prefix) and installs it with
`bao operator raft snapshot restore`; every secret, policy, auth
mount, and token in the snapshot comes back.

REQUIRES AUTO-UNSEAL, SAME KEY: a snapshot is protected by the seal
key of the cluster that took it. With `auto_unseal` on both sides
pointing at the same key, a restore is a single call and the
restored cluster unseals itself. A Shamir cluster restores only by
hand (the snapshot's key shares must be present) — that runbook is
in the component guide, not here.

ONE STEP THE MODULE CANNOT TAKE: the Job authenticates with the
TARGET's initial root token — run `bao operator init` on the fresh
cluster as always, put the returned root token in the Secret
`root_token` names, and the waiting Job proceeds. The token ceases
to exist the moment the source's state lands; delete the Secret
afterwards.

RESTORE MODE: while this block is declared, the backup schedule is
rendered suspended. A fresh target's own hourly backup would
otherwise write snapshots of an EMPTY vault into the shared prefix
while the Job waits for the token, and its retention would prune
the source's snapshots — `latest` would then find the target's own
empty snapshot. After the Job completes, remove this block and apply
again: the finished Job is deleted and the schedule resumes.

ONE-SHOT: every distinct declaration renders a distinctly named Job,
so a restore runs exactly once per declaration and again only when
the declaration changes. The Job is never expired, and a finished
Job must never be deleted by hand: a vanished Job is recreated on
the next apply and restores AGAIN over live data. After a restore
the cluster carries the SOURCE's backup login role, bound to the
source's ServiceAccount name and namespace: a target with the same
name and namespace resumes backups untouched; a renamed one re-runs
the login recipe.

IF THE INSTALL FAILS PART-WAY (a snapshot taken under a different
seal key, a token that is not this cluster's initial root token),
OpenBao seals itself. The Job's log names the cause and the way
out: delete the server pods so they restart and auto-unseal, fix
the cause, then change or re-declare this block to run again.

### spec.restore.snapshotKey

`string`

The object key of the snapshot inside the store, relative to the
bucket or container: `<prefix>/<name>-<UTC timestamp>.snap`, as
written by the source's backup job. Read it from the store's
listing.

- rule: snapshot_key is the snapshot's object key inside the store (e.g. openbao/prod/prod-20260101T020000Z.snap) — name one, or set latest: true

### spec.restore.latest

`bool`

Restore the newest snapshot under the declared prefix. Safe on the
bad day because a cluster in restore mode takes no snapshots of its
own (see the `restore` field) — but the SOURCE's schedule is not
suspended by anything: while the source is still alive, "newest" is
whatever its CronJob wrote last, which may be later than the moment
you meant. Restoring beside a live source (a clone, a migration
rehearsal) names a `snapshot_key` from the store's listing instead.

- rule: latest is a marker — set it to true to restore the newest snapshot, or name a snapshot_key instead

### spec.restore.rootToken

`KubernetesSecretKey` · required

The Secret holding the TARGET's initial root token — the one
`bao operator init` prints on the fresh cluster. Create it after
init (`kubectl create secret generic <name> --from-literal=<key>=<token>`);
the Job waits until it exists — its pod shows
`CreateContainerConfigError` until then, which is the designed
wait for your one step, not a failure. Delete the Secret once the
restore completes: the token stops existing when the source's
state lands.

- rule: root_token names a Secret and the key inside it that holds the fresh cluster's initial root token — set both name and key
- rule: {"required":true}

### spec.restore.rootToken.name

`string`

The name of the Kubernetes Secret.

### spec.restore.rootToken.key

`string`

The key within the Kubernetes Secret.

## Validation Rules

- `spec.backup.requires_raft`: Backups snapshot Raft storage: set server.ha (a single replica is a legal Raft cluster) — dev mode and standalone file storage have no snapshot API.
- `spec.backup.requires_auth_delegator`: The backup job logs in through OpenBao's Kubernetes auth method, which verifies its token with a TokenReview — leave service_account.auth_delegator_enabled on (the default) when backup is declared.
- `spec.restore.requires_backup`: A restore reads from the store declared on backup (bucket, prefix, credentials, identity) — declare backup with the same store the snapshot was written to.
- `spec.restore.requires_auto_unseal`: A declared restore needs auto_unseal with the same seal key the snapshot was taken under — declare the same aws_kms, gcp_kms, azure_key_vault, or transit seal as the source. A Shamir cluster restores by hand; see the component guide.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesOpenBao, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the server runs in. |
| `status.outputs.service` | `string` | The main client Service name (round-robins ALL server pods, including sealed/not-ready ones — by design, so init/unseal can reach them). |
| `status.outputs.internal_service` | `string` | The headless Service (`<name>-internal`) used for peer discovery and Raft cluster addresses. |
| `status.outputs.active_service` | `string` | The active-leader Service (`<name>-active`) — HA mode only, empty otherwise. Points at exactly the elected leader; the right target for write-heavy clients. |
| `status.outputs.ui_service` | `string` | The UI Service name (`<name>-ui`) when ui_enabled, empty otherwise. |
| `status.outputs.api_endpoint` | `string` | In-cluster API endpoint, scheme included (e.g. "http://bao.openbao.svc.cluster.local:8200" — https when TLS is enabled). What secret-consuming addons (external-secrets ClusterSecretStore, cert-manager Vault issuers) should point at. |
| `status.outputs.port` | `string` | API port (8200). |
| `status.outputs.service_account_name` | `string` | The server ServiceAccount name — the identity to bind cloud IAM (auto-unseal KMS access) and OpenBao Kubernetes-auth trust to. |
| `status.outputs.port_forward_command` | `string` | Copy-paste command for reaching the API from a workstation. |
| `status.outputs.backup_service_account_name` | `string` | The backup job's ServiceAccount (`<name>-backup`) — the identity to bind the object store's cloud IAM to (a GcpGkeWorkloadIdentityBinding's `ksa_name`, an IRSA trust policy's subject) and the `bound_service_account_names` of the login recipe. Empty when `backup` is not declared. |
| `status.outputs.backup_policy_name` | `string` | The OpenBao policy the login recipe writes for the backup job (`<name>-backup`; read on `sys/storage/raft/snapshot`). Empty when `backup` is not declared. |
| `status.outputs.backup_auth_role` | `string` | The Kubernetes-auth role the backup job logs in with (`<name>-backup` unless `backup.auth.role` names another). Empty when `backup` is not declared. |
| `status.outputs.backup_cron_job_name` | `string` | The backup CronJob (`<name>-backup`) — `kubectl create job --from=cronjob/<this> <run-name>` triggers a snapshot on demand. Empty when `backup` is not declared. |
| `status.outputs.restore_job_name` | `string` | The restore Job (`<name>-restore-<8 hex>`, the hex hashing the declaration) when `restore` is declared — `kubectl logs job/<this>` is where the restore explains itself. Empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.server.dataStorage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.server.auditStorage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.tls.certSecretName` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.autoUnseal.gcpKms.project` | GcpProject | `status.outputs.project_id` |
| `spec.autoUnseal.gcpKms.keyRing` | GcpKmsKeyRing | `status.outputs.key_ring_name` |
| `spec.autoUnseal.gcpKms.cryptoKey` | GcpKmsKey | `status.outputs.key_name` |
| `spec.autoUnseal.gcpKms.workloadIdentityServiceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.backup.objectStore.s3.endpointUrl` | KubernetesSeaweedFs | `status.outputs.s3_endpoint` |
| `spec.backup.objectStore.gcs.bucket` | GcpGcsBucket | `status.outputs.bucket_name` |
| `spec.backup.objectStore.gcs.serviceAccountKey` | GcpServiceAccount | `status.outputs.key_base64` |
| `spec.backup.objectStore.r2.bucket` | CloudflareR2Bucket | `status.outputs.bucket_name` |
| `spec.backup.objectStore.r2.accountId` | CloudflareR2Bucket | `status.outputs.account_id` |
| `spec.backup.objectStore.r2.jurisdiction` | CloudflareR2Bucket | `status.outputs.jurisdiction` |
| `spec.backup.objectStore.r2.credentials.accessKeyId` | CloudflareAccountApiToken | `status.outputs.r2_access_key_id` |
| `spec.backup.objectStore.r2.credentials.secretAccessKey` | CloudflareAccountApiToken | `status.outputs.r2_secret_access_key` |
| `spec.backup.workloadIdentity.gke.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.backup.workloadIdentity.eks.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.backup.workloadIdentity.aks.clientId` | AzureUserAssignedIdentity | `status.outputs.client_id` |

## See Also

- [Overview](../README.md)
