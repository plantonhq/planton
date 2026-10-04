# KubernetesSolr

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesSolrSpec** declares an Apache SolrCloud cluster as a
`SolrCloud` custom resource reconciled by the Apache Solr Operator
(KubernetesSolrOperator, the registry prerequisite).

The operator manages the node StatefulSet, ZooKeeper wiring,
managed rolling updates that keep shards available, scale
operations that move replicas off departing pods, TLS, basic-auth
security bootstrap, and registered backup repositories.

ZOOKEEPER: every SolrCloud needs an ensemble. The default (empty
`zookeeper`) provisions a 3-node ensemble through the
zookeeper-operator the KubernetesSolrOperator chart bundles;
`zookeeper.external` connects to an ensemble you already run
instead.

STORAGE: ephemeral (emptyDir) is the operator default — data is
LOST on pod eviction. Declare `storage.persistent` for anything
beyond throwaway experiments.

SECURITY: `security.authentication_type: basic` bootstraps
basic-auth with operator-generated credentials in a Secret (see
outputs); no credential ever appears in this spec.

EXPOSURE: in-cluster access rides the common Service (see stack
outputs). The `external` block models the operator's own
Ingress/ExternalDNS exposure; composing a KubernetesIngress or
Gateway API route over the common service handle is equally valid
and keeps exposure as a first-class graph node.

## Example

```yaml
# Full-surface manifest for offline module proofs (tofu validate/plan and
# pulumi preview). Exercises every typed block coherently: a provided
# ZooKeeper ensemble with persistence and tuned resources, persistent node
# storage, JVM/logging/GC tuning, node resources with scheduling
# constraints, keystore TLS, basic-auth security, one S3 and one shared-
# volume backup repository, the operator's Ingress exposure arm, managed
# rolling updates with int-and-percentage budgets, and explicit scaling
# flags. Both engines must render identical SolrCloud CRs from it.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolr
metadata:
  name: solr-hack
spec:
  namespace:
    value: solr-hack-ns
  createNamespace: true
  replicas: 3
  version: 9.10.0
  zookeeper:
    provided:
      replicas: 3
      chroot: /solr
      persistence:
        size: 5Gi
        storageClass:
          value: fast-ssd
      resources:
        limits:
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 250m
          memory: 512Mi
  storage:
    persistent:
      size: 20Gi
      storageClass:
        value: fast-ssd
      reclaimPolicy: Delete
  javaMem: -Xms1g -Xmx1g
  solrOpts: -Dsolr.autoSoftCommit.maxTime=10000
  logLevel: WARN
  gcTune: -XX:+UseG1GC -XX:MaxGCPauseMillis=200
  resources:
    limits:
      cpu: 2000m
      memory: 3Gi
    requests:
      cpu: 500m
      memory: 2Gi
  security:
    authenticationType: basic
    probesRequireAuth: true
  tls:
    pkcs12Secret:
      name: solr-hack-keystore
      key: keystore.p12
    keystorePasswordSecret:
      name: solr-hack-keystore-pass
      key: password
    clientAuth: Want
    verifyClientHostname: true
  backupRepositories:
    - name: s3-backups
      s3:
        region: us-west-2
        bucket: solr-hack-backups
        baseLocation: /clusters/solr-hack
        endpoint: http://minio.minio-system.svc:9000
        credentials:
          accessKeyIdSecret:
            name: solr-hack-s3-creds
            key: access-key-id
          secretAccessKeySecret:
            name: solr-hack-s3-creds
            key: secret-access-key
    - name: shared-volume
      volume:
        pvcClaimName: solr-hack-backup-pvc
        directory: solr-hack
  solrModules:
    - analytics
    - ltr
  additionalLibs:
    - /opt/solr/custom-libs
  updateStrategy:
    method: Managed
    maxPodsUnavailable: "2"
    maxShardReplicasUnavailable: 25%
    restartSchedule: "@every 168h"
  availability:
    pdbEnabled: true
  scaling:
    vacatePodsOnScaleDown: true
    populatePodsOnScaleUp: false
  external:
    method: Ingress
    domainName: search.example.com
    useExternalAddress: true
    hideNodes: false
  podPort: 8983
  nodeSelector:
    disktype: ssd
  tolerations:
    - key: dedicated
      operator: Equal
      value: search
      effect: NoSchedule
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.replicas` | `int32` |  | `3` |  |
| `spec.version` | `string` | yes |  |  |
| `spec.imageRepository` | `string` |  |  |  |
| `spec.zookeeper` | `KubernetesSolrZookeeper` |  |  |  |
| `spec.zookeeper.provided` | `KubernetesSolrProvidedZookeeper` |  |  |  |
| `spec.zookeeper.provided.replicas` | `int32` |  | `3` |  |
| `spec.zookeeper.provided.persistence` | `KubernetesSolrProvidedZookeeperPersistence` |  |  |  |
| `spec.zookeeper.provided.persistence.size` | `string` |  |  |  |
| `spec.zookeeper.provided.persistence.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.zookeeper.provided.resources` | `ContainerResources` |  |  |  |
| `spec.zookeeper.provided.resources.limits` | `CpuMemory` |  |  |  |
| `spec.zookeeper.provided.resources.limits.cpu` | `string` |  |  |  |
| `spec.zookeeper.provided.resources.limits.memory` | `string` |  |  |  |
| `spec.zookeeper.provided.resources.requests` | `CpuMemory` |  |  |  |
| `spec.zookeeper.provided.resources.requests.cpu` | `string` |  |  |  |
| `spec.zookeeper.provided.resources.requests.memory` | `string` |  |  |  |
| `spec.zookeeper.provided.chroot` | `string` |  | `/` |  |
| `spec.zookeeper.external` | `KubernetesSolrExternalZookeeper` |  |  |  |
| `spec.zookeeper.external.connectionString` | `string` | yes |  |  |
| `spec.zookeeper.external.chroot` | `string` |  | `/` |  |
| `spec.storage` | `KubernetesSolrStorage` |  |  |  |
| `spec.storage.persistent` | `KubernetesSolrPersistentStorage` |  |  |  |
| `spec.storage.persistent.size` | `string` | yes |  |  |
| `spec.storage.persistent.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.storage.persistent.reclaimPolicy` | `string` |  | `Retain` |  |
| `spec.storage.ephemeral` | `KubernetesSolrEphemeralStorage` |  |  |  |
| `spec.storage.ephemeral.sizeLimit` | `string` |  |  |  |
| `spec.javaMem` | `string` |  |  |  |
| `spec.solrOpts` | `string` |  |  |  |
| `spec.logLevel` | `string` |  | `INFO` |  |
| `spec.gcTune` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.security` | `KubernetesSolrSecurity` |  |  |  |
| `spec.security.authenticationType` | `string` |  |  |  |
| `spec.security.basicAuthSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.security.probesRequireAuth` | `bool` |  |  |  |
| `spec.security.bootstrapSecurityJson` | `KubernetesSolrSecretKeyRef` |  |  |  |
| `spec.security.bootstrapSecurityJson.name` | `string` | yes |  |  |
| `spec.security.bootstrapSecurityJson.key` | `string` | yes |  |  |
| `spec.tls` | `KubernetesSolrTls` |  |  |  |
| `spec.tls.pkcs12Secret` | `KubernetesSolrSecretKeyRef` | yes |  |  |
| `spec.tls.pkcs12Secret.name` | `string` | yes |  |  |
| `spec.tls.pkcs12Secret.key` | `string` | yes |  |  |
| `spec.tls.keystorePasswordSecret` | `KubernetesSolrSecretKeyRef` | yes |  |  |
| `spec.tls.keystorePasswordSecret.name` | `string` | yes |  |  |
| `spec.tls.keystorePasswordSecret.key` | `string` | yes |  |  |
| `spec.tls.truststoreSecret` | `KubernetesSolrSecretKeyRef` |  |  |  |
| `spec.tls.truststoreSecret.name` | `string` | yes |  |  |
| `spec.tls.truststoreSecret.key` | `string` | yes |  |  |
| `spec.tls.truststorePasswordSecret` | `KubernetesSolrSecretKeyRef` |  |  |  |
| `spec.tls.truststorePasswordSecret.name` | `string` | yes |  |  |
| `spec.tls.truststorePasswordSecret.key` | `string` | yes |  |  |
| `spec.tls.clientAuth` | `string` |  | `None` |  |
| `spec.tls.verifyClientHostname` | `bool` |  |  |  |
| `spec.backupRepositories` | `[]KubernetesSolrBackupRepository` |  |  |  |
| `spec.backupRepositories[].name` | `string` | yes |  |  |
| `spec.backupRepositories[].s3` | `KubernetesSolrS3Repository` |  |  |  |
| `spec.backupRepositories[].s3.region` | `string` | yes |  |  |
| `spec.backupRepositories[].s3.bucket` | `string` | yes |  |  |
| `spec.backupRepositories[].s3.baseLocation` | `string` |  |  |  |
| `spec.backupRepositories[].s3.endpoint` | `string` |  |  |  |
| `spec.backupRepositories[].s3.credentials` | `KubernetesSolrS3Credentials` |  |  |  |
| `spec.backupRepositories[].s3.credentials.accessKeyIdSecret` | `KubernetesSolrSecretKeyRef` |  |  |  |
| `spec.backupRepositories[].s3.credentials.accessKeyIdSecret.name` | `string` | yes |  |  |
| `spec.backupRepositories[].s3.credentials.accessKeyIdSecret.key` | `string` | yes |  |  |
| `spec.backupRepositories[].s3.credentials.secretAccessKeySecret` | `KubernetesSolrSecretKeyRef` |  |  |  |
| `spec.backupRepositories[].s3.credentials.secretAccessKeySecret.name` | `string` | yes |  |  |
| `spec.backupRepositories[].s3.credentials.secretAccessKeySecret.key` | `string` | yes |  |  |
| `spec.backupRepositories[].gcs` | `KubernetesSolrGcsRepository` |  |  |  |
| `spec.backupRepositories[].gcs.bucket` | `string` | yes |  |  |
| `spec.backupRepositories[].gcs.gcsCredentialSecret` | `KubernetesSolrSecretKeyRef` |  |  |  |
| `spec.backupRepositories[].gcs.gcsCredentialSecret.name` | `string` | yes |  |  |
| `spec.backupRepositories[].gcs.gcsCredentialSecret.key` | `string` | yes |  |  |
| `spec.backupRepositories[].gcs.baseLocation` | `string` |  |  |  |
| `spec.backupRepositories[].volume` | `KubernetesSolrVolumeRepository` |  |  |  |
| `spec.backupRepositories[].volume.pvcClaimName` | `string` | yes |  |  |
| `spec.backupRepositories[].volume.directory` | `string` |  |  |  |
| `spec.solrModules` | `[]string` |  |  |  |
| `spec.additionalLibs` | `[]string` |  |  |  |
| `spec.updateStrategy` | `KubernetesSolrUpdateStrategy` |  |  |  |
| `spec.updateStrategy.method` | `string` |  | `Managed` |  |
| `spec.updateStrategy.maxPodsUnavailable` | `string` |  |  |  |
| `spec.updateStrategy.maxShardReplicasUnavailable` | `string` |  |  |  |
| `spec.updateStrategy.restartSchedule` | `string` |  |  |  |
| `spec.availability` | `KubernetesSolrAvailability` |  |  |  |
| `spec.availability.pdbEnabled` | `bool` |  | `true` |  |
| `spec.scaling` | `KubernetesSolrScaling` |  |  |  |
| `spec.scaling.vacatePodsOnScaleDown` | `bool` |  | `true` |  |
| `spec.scaling.populatePodsOnScaleUp` | `bool` |  | `true` |  |
| `spec.external` | `KubernetesSolrExternal` |  |  |  |
| `spec.external.method` | `string` | yes |  |  |
| `spec.external.domainName` | `string` | yes |  |  |
| `spec.external.useExternalAddress` | `bool` |  |  |  |
| `spec.external.hideCommon` | `bool` |  |  |  |
| `spec.external.hideNodes` | `bool` |  |  |  |
| `spec.podPort` | `int32` |  | `8983` |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to create the cluster in. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the cluster and deleted with the
resource. When false, the namespace must already exist.

### spec.replicas

`int32` · optional (explicit presence)

Number of Solr nodes. Operator default: 3. One node works for
development; production shard/replica placement needs 3+.

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.version

`string` · required

Solr version to run — the image tag of the official `solr`
image, e.g. "9.10.0". Check the operator's compatibility matrix
before pinning a new major line.

- rule: {"required":true}

### spec.imageRepository

`string`

Override the Solr image repository (air-gap / custom-image
path). Empty = "solr" (Docker Hub official). The tag always
comes from `version`.

### spec.zookeeper

`KubernetesSolrZookeeper`

ZooKeeper wiring. Empty = a provided 3-node ensemble managed by
the bundled zookeeper-operator.

### spec.zookeeper.provided

`KubernetesSolrProvidedZookeeper`

An ensemble provisioned and managed by the zookeeper-operator
(the KubernetesSolrOperator chart must have it installed —
its default).

### spec.zookeeper.provided.replicas

`int32` · optional (explicit presence)

Ensemble size. Default: 3 (the ZooKeeper quorum minimum for
production; 1 works for development).

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.zookeeper.provided.persistence

`KubernetesSolrProvidedZookeeperPersistence`

Persistent storage for the ensemble. Empty = a small PVC on the
default StorageClass (the zookeeper-operator default).

### spec.zookeeper.provided.persistence.size

`string`

Volume size per ZooKeeper member, e.g. "5Gi". Empty = the
zookeeper-operator default (20Gi).

### spec.zookeeper.provided.persistence.storageClass

`string | valueFrom`

StorageClass for the ensemble's PVCs. Accepts a literal class
name or a reference to a KubernetesStorageClass resource. Empty
= the cluster default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.zookeeper.provided.resources

`ContainerResources`

ZooKeeper container resources. Empty = zookeeper-operator
defaults.

### spec.zookeeper.provided.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.zookeeper.provided.resources.limits.cpu

`string`

### spec.zookeeper.provided.resources.limits.memory

`string`

### spec.zookeeper.provided.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.zookeeper.provided.resources.requests.cpu

`string`

### spec.zookeeper.provided.resources.requests.memory

`string`

### spec.zookeeper.provided.chroot

`string` · optional (explicit presence)

The chroot Solr uses within the ensemble. Default "/".

- default: `/`

### spec.zookeeper.external

`KubernetesSolrExternalZookeeper`

Connect to an ensemble running outside the operator's
management.

### spec.zookeeper.external.connectionString

`string` · required

The ensemble connection string reachable from inside the
Kubernetes cluster, e.g. "zk-0.zk:2181,zk-1.zk:2181".

- rule: {"required":true}

### spec.zookeeper.external.chroot

`string` · optional (explicit presence)

The chroot Solr uses within the ensemble. Default "/".

- default: `/`

### spec.storage

`KubernetesSolrStorage`

Data storage. Empty = ephemeral emptyDir (operator default) —
data is LOST when a pod leaves its node. Declare `persistent`
for real workloads.

### spec.storage.persistent

`KubernetesSolrPersistentStorage`

PVC-backed storage — data survives pod loss.

### spec.storage.persistent.size

`string` · required

Volume size per Solr node, e.g. "20Gi".

- rule: {"required":true}

### spec.storage.persistent.storageClass

`string | valueFrom`

StorageClass for the node PVCs. Accepts a literal class name or
a reference to a KubernetesStorageClass resource. Empty = the
cluster default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.storage.persistent.reclaimPolicy

`string` · optional (explicit presence)

What happens to the PVCs when the SolrCloud is deleted: Retain
(operator default — data outlives the resource) or Delete.

- default: `Retain`
- rule: reclaim policy must be Retain or Delete

### spec.storage.ephemeral

`KubernetesSolrEphemeralStorage`

emptyDir storage — data is LOST when a pod leaves its node.

### spec.storage.ephemeral.sizeLimit

`string`

Optional cap on the emptyDir size, e.g. "10Gi". Empty = no
limit.

### spec.javaMem

`string`

JVM heap for Solr nodes, e.g. "-Xms1g -Xmx1g". Operator default:
"-Xms300m -Xmx300m" — size it to roughly half the container
memory.

### spec.solrOpts

`string`

Extra system properties appended to SOLR_OPTS,
e.g. "-Dsolr.autoSoftCommit.maxTime=10000".

### spec.logLevel

`string` · optional (explicit presence)

Solr log level. Operator default: INFO.

- default: `INFO`
- rule: log level must be ERROR, WARN, INFO, DEBUG, or TRACE

### spec.gcTune

`string`

GC tuning flags exported via GC_TUNE, e.g.
"-XX:+UseG1GC -XX:MaxGCPauseMillis=200".

### spec.resources

`ContainerResources`

Solr node container resources. Empty = no requests/limits (set
them for any real deployment; pair memory with `java_mem`).

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.security

`KubernetesSolrSecurity`

Basic-auth security bootstrap.

### spec.security.authenticationType

`string` · optional (explicit presence)

Authentication plugin: "basic" is the one the operator manages.
Empty = security disabled (open cluster — development only).

- rule: authentication_type must be basic

### spec.security.basicAuthSecret

`string | valueFrom`

Existing kubernetes.io/basic-auth Secret with the credentials
the OPERATOR uses against secured pods. Empty = the operator
bootstraps security.json plus admin/solr/k8s-oper users and
writes their credentials to `<name>-solrcloud-basic-auth`
(see outputs). If you later rotate that password through
Solr's security API, update the Secret too — the operator locks
itself out otherwise (upstream contract).

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.security.probesRequireAuth

`bool`

Probe endpoints require authentication too. Operator default:
false (probes stay open).

### spec.security.bootstrapSecurityJson

`KubernetesSolrSecretKeyRef`

Bring-your-own bootstrap security.json from a Secret key —
advanced setups only; the operator applies it once and never
updates it.

### spec.security.bootstrapSecurityJson.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.security.bootstrapSecurityJson.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.tls

`KubernetesSolrTls`

Server TLS for Solr's HTTP listener (keystore-based, the
operator's model). When set, all endpoints switch to https.

### spec.tls.pkcs12Secret

`KubernetesSolrSecretKeyRef` · required

Secret + key holding a PKCS#12 keystore (e.g. produced by a
KubernetesCertificate with a pkcs12 keystore output).

- rule: {"required":true}

### spec.tls.pkcs12Secret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.tls.pkcs12Secret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.tls.keystorePasswordSecret

`KubernetesSolrSecretKeyRef` · required

Secret + key holding the keystore password (JVMs refuse
password-less PKCS#12).

- rule: {"required":true}

### spec.tls.keystorePasswordSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.tls.keystorePasswordSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.tls.truststoreSecret

`KubernetesSolrSecretKeyRef`

Optional separate truststore (PKCS#12). Empty = the keystore
doubles as the truststore.

### spec.tls.truststoreSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.tls.truststoreSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.tls.truststorePasswordSecret

`KubernetesSolrSecretKeyRef`

Optional truststore password. Empty = the keystore password.

### spec.tls.truststorePasswordSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.tls.truststorePasswordSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.tls.clientAuth

`string` · optional (explicit presence)

Client-certificate demand on the server listener: None (operator
default), Want, or Need. Need requires the operator's own mTLS
client identity (KubernetesSolrOperator `mtls`) or probes and
reconciliation calls fail.

- default: `None`
- rule: client_auth must be None, Want, or Need

### spec.tls.verifyClientHostname

`bool`

Verify client hostnames during the TLS handshake.

### spec.backupRepositories

`[]KubernetesSolrBackupRepository`

Backup repositories registered on the cluster — the targets
SolrBackup operations (operational verbs, run against the Solr
API or as SolrBackup CRs) write to. Each name must be unique.

- rule: declare a backend: s3, gcs, or volume

### spec.backupRepositories[].name

`string` · required

Repository name (referenced by backup operations). Alphanumeric
with interior dashes/underscores, max 100 characters — the CRD's
own pattern (broader than a DNS label; the operator accepts
uppercase and underscores here).

- rule: {"required":true,"string":{"maxLen":"100","pattern":"^[a-zA-Z0-9]([-_a-zA-Z0-9]*[a-zA-Z0-9])?$"}}

### spec.backupRepositories[].s3

`KubernetesSolrS3Repository`

Amazon S3 (or an S3-compatible endpoint).

### spec.backupRepositories[].s3.region

`string` · required

AWS region of the bucket, e.g. "us-west-2". Required even for
S3-compatible endpoints (any value works there).

- rule: {"required":true}

### spec.backupRepositories[].s3.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.backupRepositories[].s3.baseLocation

`string`

Chroot within the bucket. Empty = "/".

### spec.backupRepositories[].s3.endpoint

`string`

Full endpoint URL for S3-compatible stores (MinIO, ...). Empty =
AWS S3.

### spec.backupRepositories[].s3.credentials

`KubernetesSolrS3Credentials`

Declared credentials read from existing Secrets. Empty = the
nodes' ambient identity (EKS: an IRSA-bound ServiceAccount on
the Solr pods — the keyless path).

### spec.backupRepositories[].s3.credentials.accessKeyIdSecret

`KubernetesSolrSecretKeyRef`

Secret + key holding the AWS access key ID.

### spec.backupRepositories[].s3.credentials.accessKeyIdSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.backupRepositories[].s3.credentials.accessKeyIdSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.backupRepositories[].s3.credentials.secretAccessKeySecret

`KubernetesSolrSecretKeyRef`

Secret + key holding the AWS secret access key.

### spec.backupRepositories[].s3.credentials.secretAccessKeySecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.backupRepositories[].s3.credentials.secretAccessKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.backupRepositories[].gcs

`KubernetesSolrGcsRepository`

Google Cloud Storage.

### spec.backupRepositories[].gcs.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.backupRepositories[].gcs.gcsCredentialSecret

`KubernetesSolrSecretKeyRef`

Secret + key holding a Google service-account key JSON. Empty =
ambient identity (GKE Workload Identity on the Solr pods).

### spec.backupRepositories[].gcs.gcsCredentialSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.backupRepositories[].gcs.gcsCredentialSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.backupRepositories[].gcs.baseLocation

`string`

Chroot within the bucket. Empty = "/".

### spec.backupRepositories[].volume

`KubernetesSolrVolumeRepository`

A shared volume (must be multi-writer: ReadWriteMany PVC or
NFS).

### spec.backupRepositories[].volume.pvcClaimName

`string` · required

Name of an existing ReadWriteMany PVC mounted to every Solr node
for backup data.

- rule: {"required":true}

### spec.backupRepositories[].volume.directory

`string`

Directory within the volume. Empty = the cluster name.

### spec.solrModules

`[]string`

Solr modules to load at startup (e.g. "analytics", "ltr").
Modules required by a declared backup repository load
automatically.

### spec.additionalLibs

`[]string`

Extra classpath directories inside the image (custom plugin
jars baked into a derived image).

### spec.updateStrategy

`KubernetesSolrUpdateStrategy`

How rolling updates execute. Empty = the operator's Managed
strategy (keeps shards available, the safe default).

### spec.updateStrategy.method

`string` · optional (explicit presence)

Managed (operator default — shard-aware parallel restarts),
StatefulSet (plain ordinal order), or Manual (never restart —
you delete pods).

- default: `Managed`
- rule: method must be Managed, StatefulSet, or Manual

### spec.updateStrategy.maxPodsUnavailable

`string`

Managed only: max pods updating at once — integer ("2") or
percentage ("25%", the operator default).

### spec.updateStrategy.maxShardReplicasUnavailable

`string`

Managed only: max replicas of any one shard down at once —
integer (operator default 1) or percentage.

### spec.updateStrategy.restartSchedule

`string`

Restart the cluster on a schedule (CRON syntax, e.g.
"@every 168h").

### spec.availability

`KubernetesSolrAvailability`

Cluster-wide PodDisruptionBudget. Empty = enabled (operator
default).

### spec.availability.pdbEnabled

`bool` · optional (explicit presence)

Create the cluster-wide PodDisruptionBudget. Operator default:
true.

- default: `true`

### spec.scaling

`KubernetesSolrScaling`

Scale-operation data handling. Empty = operator defaults (both
true — replicas are moved, not dropped).

### spec.scaling.vacatePodsOnScaleDown

`bool` · optional (explicit presence)

Move replicas off pods before scale-down deletes them. Operator
default: true.

- default: `true`

### spec.scaling.populatePodsOnScaleUp

`bool` · optional (explicit presence)

Move replicas onto new pods after scale-up (Solr 9.3+). Operator
default: true.

- default: `true`

### spec.external

`KubernetesSolrExternal`

The operator's own external exposure (Ingress or ExternalDNS
with per-node addressability — what SolrJ/CloudSolrClient needs
to reach individual nodes from outside). For simple HTTP access
compose a KubernetesIngress over the common service instead.

### spec.external.method

`string` · required

Ingress (per-node paths through an ingress controller) or
ExternalDNS (per-node DNS records).

- rule: method must be Ingress or ExternalDNS
- rule: {"required":true}

### spec.external.domainName

`string` · required

Domain the cluster is addressed under,
e.g. "search.example.com" — nodes become
`<ns>-<name>-solrcloud-*.<domain>`.

- rule: {"required":true}

### spec.external.useExternalAddress

`bool`

Solr advertises its EXTERNAL addresses to clients (needed for
CloudSolrClient outside the cluster). Operator default: false.

### spec.external.hideCommon

`bool`

Skip exposing the common (all-nodes) service.

### spec.external.hideNodes

`bool`

Skip exposing individual node services.

### spec.podPort

`int32` · optional (explicit presence)

The port Solr listens on (operator default 8983). The common
Service in front of the nodes exposes the operator's own default
(80, or 443 with TLS) — deliberately not modeled; the exported
endpoint carries the effective port.

- default: `8983`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.nodeSelector

`map<string, string>`

Node selector for Solr pods.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for Solr pods.

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesSolr, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the cluster runs in. |
| `status.outputs.cluster_name` | `string` | name of the SolrCloud resource (= metadata.name). |
| `status.outputs.common_service_name` | `string` | name of the common Service fronting all Solr nodes (the operator names it `<name>-solrcloud-common`). |
| `status.outputs.internal_endpoint` | `string` | in-cluster base URL of the cluster, e.g. http://main-solrcloud-common.search.svc.cluster.local (the common service listens on port 80, or 443 with TLS). |
| `status.outputs.basic_auth_secret_name` | `string` | name of the operator-generated basic-auth Secret (`<name>-solrcloud-basic-auth`, kubernetes.io/basic-auth with username/password fields — the operator's READ-ONLY k8s-oper user). The admin and solr users' passwords live in the sibling `<name>-solrcloud-security-bootstrap` Secret (one key per user); use its `admin` key for administrative API calls. Empty when security is disabled or a user-provided basic_auth_secret is in play. |
| `status.outputs.zookeeper_connection_string` | `string` | ZooKeeper connection string the cluster uses (host:port/chroot). |
| `status.outputs.port_forward_command` | `string` | command to port-forward the common service to a developer laptop, e.g. kubectl port-forward svc/main-solrcloud-common -n search 8983:80 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.zookeeper.provided.persistence.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.storage.persistent.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.security.basicAuthSecret` | KubernetesSecret | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
