# KubernetesHarbor

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesHarborSpec** installs Harbor — the CNCF-graduated
container registry that stores, signs, and scans OCI artifacts —
from the official `harbor` chart (https://helm.goharbor.io,
chart 1.19.x = Harbor 2.15.x).

Harbor renders as a set of cooperating components: core (API +
auth), portal (web UI), registry (the OCI distribution backend,
with its registryctl sidecar deployment), jobservice (replication,
GC, scan jobs), the Trivy vulnerability scanner (on by default),
an optional Prometheus exporter, and an nginx front door that
terminates client traffic for every exposure mode this component
models. State lives in PostgreSQL, Redis, and the artifact storage
backend — each with an in-cluster arm for evaluation and an
external arm for production composition.

CREDENTIALS ARE NEVER THE CHART DEFAULTS: the chart's documented
defaults (`Harbor12345` admin password, `changeit` database
password, `not-a-secure-key` encryption key,
`harbor_registry_user/harbor_registry_password` internal registry
credential) are public knowledge and never ship. Every internal
credential is generated per-install and delivered through
module-owned Secrets; the admin password is exported as the
`admin_password_secret` output — the credential handle.

EXPOSURE COMPOSES (never embedded): this component always deploys
the chart's nginx proxy front door behind a ClusterIP, NodePort,
or LoadBalancer Service. The chart's `ingress` and Gateway API
`route` exposure types are deliberately not modeled — north-south
exposure composes from the catalog's exposure kinds pointed at the
exported front-door Service.

## Example

```yaml
# Full-surface development manifest — exercises every module-rendered arm
# the kind-cluster lanes exclude (composed external PostgreSQL/Redis,
# S3-compatible artifact storage with declared credentials + the
# disable-redirect posture, LoadBalancer exposure with TLS from a
# cert Secret, internal TLS from per-component Secrets, the air-gap
# image mirror, Trivy air-gap knobs, metrics + ServiceMonitor, the
# outbound proxy, and the escape hatch).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHarbor
metadata:
  name: harbor-dev
spec:
  namespace:
    value: harbor-dev
  createNamespace: true
  externalUrl: https://harbor.dev.example.com
  expose:
    serviceType: LoadBalancer
    tls:
      enabled: true
      certSecretName:
        value: harbor-dev-tls
    serviceAnnotations:
      service.beta.kubernetes.io/aws-load-balancer-type: nlb
    sourceRanges:
      - 10.0.0.0/8
  database:
    external:
      host:
        value: harbor-pg-rw
      username: harbor
      passwordSecretName:
        value: harbor-pg-app
      coreDatabase: registry
      sslMode: require
  cache:
    external:
      addr:
        value: harbor-valkey:6379
      password: dev-only-placeholder
  storage:
    s3:
      bucket: harbor-dev-artifacts
      region: us-east-1
      endpoint:
        value: http://harbor-objects-s3:8333
      credentials:
        accessKey: devonlyaccesskey
        secretKey: dev-only-placeholder-secret
      disableRedirect: true
      secure: false
  trivy:
    enabled: true
    replicas: 1
    skipUpdate: true
    offlineScan: true
    severity: HIGH,CRITICAL
    githubToken: dev-only-placeholder-token
  core:
    replicas: 2
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: 1000m
        memory: 2Gi
  portal:
    replicas: 2
  registry:
    replicas: 2
  jobservice:
    replicas: 2
    maxJobWorkers: 20
    logDiskSize: 2Gi
  nginx:
    replicas: 2
  internalTls:
    enabled: true
    certSecrets:
      coreSecretName: harbor-dev-core-tls
      jobserviceSecretName: harbor-dev-js-tls
      registrySecretName: harbor-dev-reg-tls
      portalSecretName: harbor-dev-portal-tls
      trivySecretName: harbor-dev-trivy-tls
  metrics:
    enabled: true
    serviceMonitorEnabled: true
    serviceMonitorLabels:
      release: monitoring
  cacheLayer:
    enabled: true
    expireHours: 48
  outboundProxy:
    httpProxy: http://proxy.internal:3128
    httpsProxy: http://proxy.internal:3128
  logLevel: warning
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  caBundleSecretName: harbor-dev-storage-ca
  updateStrategy: Recreate
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: registry
        effect: NoSchedule
  helmValues: |
    core:
      podAnnotations:
        example.planton.ai/full-surface: "true"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.19.1` |  |
| `spec.externalUrl` | `string` |  |  |  |
| `spec.expose` | `KubernetesHarborExpose` |  |  |  |
| `spec.expose.serviceType` | `string` |  | `ClusterIP` |  |
| `spec.expose.tls` | `KubernetesHarborExposeTls` |  |  |  |
| `spec.expose.tls.enabled` | `bool` |  |  |  |
| `spec.expose.tls.certSecretName` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.expose.nodePorts` | `KubernetesHarborNodePorts` |  |  |  |
| `spec.expose.nodePorts.http` | `int32` |  | `30002` |  |
| `spec.expose.nodePorts.https` | `int32` |  | `30003` |  |
| `spec.expose.serviceAnnotations` | `map<string, string>` |  |  |  |
| `spec.expose.sourceRanges` | `[]string` |  |  |  |
| `spec.expose.loadBalancerIp` | `string` |  |  |  |
| `spec.adminAuth` | `KubernetesHarborAdminAuth` |  |  |  |
| `spec.adminAuth.existingSecretName` | `string` |  |  |  |
| `spec.adminAuth.existingSecretKey` | `string` |  | `HARBOR_ADMIN_PASSWORD` |  |
| `spec.database` | `KubernetesHarborDatabase` | yes |  |  |
| `spec.database.internal` | `KubernetesHarborInternalDatabase` |  |  |  |
| `spec.database.internal.resources` | `ContainerResources` |  |  |  |
| `spec.database.internal.resources.limits` | `CpuMemory` |  |  |  |
| `spec.database.internal.resources.limits.cpu` | `string` |  |  |  |
| `spec.database.internal.resources.limits.memory` | `string` |  |  |  |
| `spec.database.internal.resources.requests` | `CpuMemory` |  |  |  |
| `spec.database.internal.resources.requests.cpu` | `string` |  |  |  |
| `spec.database.internal.resources.requests.memory` | `string` |  |  |  |
| `spec.database.internal.diskSize` | `string` |  | `1Gi` |  |
| `spec.database.internal.storageClass` | `string` |  |  |  |
| `spec.database.internal.shmSizeLimit` | `string` |  | `512Mi` |  |
| `spec.database.external` | `KubernetesHarborExternalDatabase` |  |  |  |
| `spec.database.external.host` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.rw_service`) |
| `spec.database.external.port` | `int32` |  | `5432` |  |
| `spec.database.external.username` | `string` | yes |  |  |
| `spec.database.external.passwordSecretName` | `string \| valueFrom` | yes |  | KubernetesPostgres (`status.outputs.password_secret.name`) |
| `spec.database.external.coreDatabase` | `string` |  | `registry` |  |
| `spec.database.external.sslMode` | `string` |  | `disable` |  |
| `spec.cache` | `KubernetesHarborCache` | yes |  |  |
| `spec.cache.internal` | `KubernetesHarborInternalRedis` |  |  |  |
| `spec.cache.internal.resources` | `ContainerResources` |  |  |  |
| `spec.cache.internal.resources.limits` | `CpuMemory` |  |  |  |
| `spec.cache.internal.resources.limits.cpu` | `string` |  |  |  |
| `spec.cache.internal.resources.limits.memory` | `string` |  |  |  |
| `spec.cache.internal.resources.requests` | `CpuMemory` |  |  |  |
| `spec.cache.internal.resources.requests.cpu` | `string` |  |  |  |
| `spec.cache.internal.resources.requests.memory` | `string` |  |  |  |
| `spec.cache.internal.diskSize` | `string` |  | `1Gi` |  |
| `spec.cache.internal.storageClass` | `string` |  |  |  |
| `spec.cache.external` | `KubernetesHarborExternalRedis` |  |  |  |
| `spec.cache.external.addr` | `string \| valueFrom` | yes |  | KubernetesValkey (`status.outputs.kube_endpoint`) |
| `spec.cache.external.sentinelMasterSet` | `string` |  |  |  |
| `spec.cache.external.username` | `string` |  |  |  |
| `spec.cache.external.password` | `string` (sensitive) |  |  |  |
| `spec.cache.external.existingSecretName` | `string` |  |  |  |
| `spec.cache.external.tlsEnabled` | `bool` |  |  |  |
| `spec.storage` | `KubernetesHarborArtifactStorage` | yes |  |  |
| `spec.storage.filesystem` | `KubernetesHarborFilesystemStorage` |  |  |  |
| `spec.storage.filesystem.diskSize` | `string` |  | `5Gi` |  |
| `spec.storage.filesystem.storageClass` | `string` |  |  |  |
| `spec.storage.filesystem.accessMode` | `string` |  | `ReadWriteOnce` |  |
| `spec.storage.s3` | `KubernetesHarborS3Storage` |  |  |  |
| `spec.storage.s3.bucket` | `string` | yes |  |  |
| `spec.storage.s3.region` | `string` | yes |  |  |
| `spec.storage.s3.endpoint` | `string \| valueFrom` |  |  | KubernetesSeaweedFs (`status.outputs.s3_endpoint`) |
| `spec.storage.s3.credentials` | `KubernetesHarborS3Credentials` |  |  |  |
| `spec.storage.s3.credentials.accessKey` | `string` |  |  |  |
| `spec.storage.s3.credentials.secretKey` | `string` (sensitive) |  |  |  |
| `spec.storage.s3.credentials.existingSecretName` | `string` |  |  |  |
| `spec.storage.s3.disableRedirect` | `bool` |  |  |  |
| `spec.storage.s3.encrypt` | `bool` |  |  |  |
| `spec.storage.s3.secure` | `bool` |  | `true` |  |
| `spec.storage.s3.skipVerify` | `bool` |  |  |  |
| `spec.storage.s3.rootDirectory` | `string` |  |  |  |
| `spec.storage.s3.storageClass` | `string` |  |  |  |
| `spec.storage.gcs` | `KubernetesHarborGcsStorage` |  |  |  |
| `spec.storage.gcs.bucket` | `string` | yes |  |  |
| `spec.storage.gcs.useWorkloadIdentity` | `bool` |  |  |  |
| `spec.storage.gcs.keyData` | `string` (sensitive) |  |  |  |
| `spec.storage.gcs.existingSecretName` | `string` |  |  |  |
| `spec.storage.gcs.rootDirectory` | `string` |  |  |  |
| `spec.storage.gcs.chunkSize` | `int32` |  |  |  |
| `spec.storage.azure` | `KubernetesHarborAzureStorage` |  |  |  |
| `spec.storage.azure.accountName` | `string` | yes |  |  |
| `spec.storage.azure.container` | `string` | yes |  |  |
| `spec.storage.azure.accountKey` | `string` (sensitive) |  |  |  |
| `spec.storage.azure.existingSecretName` | `string` |  |  |  |
| `spec.storage.azure.realm` | `string` |  | `core.windows.net` |  |
| `spec.trivy` | `KubernetesHarborTrivy` |  |  |  |
| `spec.trivy.enabled` | `bool` |  | `true` |  |
| `spec.trivy.replicas` | `int32` |  | `1` |  |
| `spec.trivy.resources` | `ContainerResources` |  |  |  |
| `spec.trivy.resources.limits` | `CpuMemory` |  |  |  |
| `spec.trivy.resources.limits.cpu` | `string` |  |  |  |
| `spec.trivy.resources.limits.memory` | `string` |  |  |  |
| `spec.trivy.resources.requests` | `CpuMemory` |  |  |  |
| `spec.trivy.resources.requests.cpu` | `string` |  |  |  |
| `spec.trivy.resources.requests.memory` | `string` |  |  |  |
| `spec.trivy.diskSize` | `string` |  | `5Gi` |  |
| `spec.trivy.skipUpdate` | `bool` |  |  |  |
| `spec.trivy.skipJavaDbUpdate` | `bool` |  |  |  |
| `spec.trivy.offlineScan` | `bool` |  |  |  |
| `spec.trivy.dbRepositories` | `[]string` |  |  |  |
| `spec.trivy.javaDbRepositories` | `[]string` |  |  |  |
| `spec.trivy.githubToken` | `string` (sensitive) |  |  |  |
| `spec.trivy.severity` | `string` |  |  |  |
| `spec.trivy.ignoreUnfixed` | `bool` |  |  |  |
| `spec.trivy.timeout` | `string` |  |  |  |
| `spec.core` | `KubernetesHarborComponent` |  |  |  |
| `spec.core.replicas` | `int32` |  | `1` |  |
| `spec.core.resources` | `ContainerResources` |  |  |  |
| `spec.core.resources.limits` | `CpuMemory` |  |  |  |
| `spec.core.resources.limits.cpu` | `string` |  |  |  |
| `spec.core.resources.limits.memory` | `string` |  |  |  |
| `spec.core.resources.requests` | `CpuMemory` |  |  |  |
| `spec.core.resources.requests.cpu` | `string` |  |  |  |
| `spec.core.resources.requests.memory` | `string` |  |  |  |
| `spec.portal` | `KubernetesHarborComponent` |  |  |  |
| `spec.portal.replicas` | `int32` |  | `1` |  |
| `spec.portal.resources` | `ContainerResources` |  |  |  |
| `spec.portal.resources.limits` | `CpuMemory` |  |  |  |
| `spec.portal.resources.limits.cpu` | `string` |  |  |  |
| `spec.portal.resources.limits.memory` | `string` |  |  |  |
| `spec.portal.resources.requests` | `CpuMemory` |  |  |  |
| `spec.portal.resources.requests.cpu` | `string` |  |  |  |
| `spec.portal.resources.requests.memory` | `string` |  |  |  |
| `spec.registry` | `KubernetesHarborComponent` |  |  |  |
| `spec.registry.replicas` | `int32` |  | `1` |  |
| `spec.registry.resources` | `ContainerResources` |  |  |  |
| `spec.registry.resources.limits` | `CpuMemory` |  |  |  |
| `spec.registry.resources.limits.cpu` | `string` |  |  |  |
| `spec.registry.resources.limits.memory` | `string` |  |  |  |
| `spec.registry.resources.requests` | `CpuMemory` |  |  |  |
| `spec.registry.resources.requests.cpu` | `string` |  |  |  |
| `spec.registry.resources.requests.memory` | `string` |  |  |  |
| `spec.jobservice` | `KubernetesHarborJobservice` |  |  |  |
| `spec.jobservice.replicas` | `int32` |  | `1` |  |
| `spec.jobservice.resources` | `ContainerResources` |  |  |  |
| `spec.jobservice.resources.limits` | `CpuMemory` |  |  |  |
| `spec.jobservice.resources.limits.cpu` | `string` |  |  |  |
| `spec.jobservice.resources.limits.memory` | `string` |  |  |  |
| `spec.jobservice.resources.requests` | `CpuMemory` |  |  |  |
| `spec.jobservice.resources.requests.cpu` | `string` |  |  |  |
| `spec.jobservice.resources.requests.memory` | `string` |  |  |  |
| `spec.jobservice.maxJobWorkers` | `int32` |  | `10` |  |
| `spec.jobservice.logDiskSize` | `string` |  | `1Gi` |  |
| `spec.nginx` | `KubernetesHarborComponent` |  |  |  |
| `spec.nginx.replicas` | `int32` |  | `1` |  |
| `spec.nginx.resources` | `ContainerResources` |  |  |  |
| `spec.nginx.resources.limits` | `CpuMemory` |  |  |  |
| `spec.nginx.resources.limits.cpu` | `string` |  |  |  |
| `spec.nginx.resources.limits.memory` | `string` |  |  |  |
| `spec.nginx.resources.requests` | `CpuMemory` |  |  |  |
| `spec.nginx.resources.requests.cpu` | `string` |  |  |  |
| `spec.nginx.resources.requests.memory` | `string` |  |  |  |
| `spec.internalTls` | `KubernetesHarborInternalTls` |  |  |  |
| `spec.internalTls.enabled` | `bool` |  |  |  |
| `spec.internalTls.certSecrets` | `KubernetesHarborInternalTlsSecrets` |  |  |  |
| `spec.internalTls.certSecrets.coreSecretName` | `string` | yes |  |  |
| `spec.internalTls.certSecrets.jobserviceSecretName` | `string` | yes |  |  |
| `spec.internalTls.certSecrets.registrySecretName` | `string` | yes |  |  |
| `spec.internalTls.certSecrets.portalSecretName` | `string` | yes |  |  |
| `spec.internalTls.certSecrets.trivySecretName` | `string` |  |  |  |
| `spec.internalTls.strongSslCiphers` | `bool` |  |  |  |
| `spec.metrics` | `KubernetesHarborMetrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  |  |  |
| `spec.metrics.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.metrics.serviceMonitorInterval` | `string` |  |  |  |
| `spec.metrics.serviceMonitorLabels` | `map<string, string>` |  |  |  |
| `spec.cacheLayer` | `KubernetesHarborCacheLayer` |  |  |  |
| `spec.cacheLayer.enabled` | `bool` |  |  |  |
| `spec.cacheLayer.expireHours` | `int32` |  | `24` |  |
| `spec.outboundProxy` | `KubernetesHarborOutboundProxy` |  |  |  |
| `spec.outboundProxy.httpProxy` | `string` |  |  |  |
| `spec.outboundProxy.httpsProxy` | `string` |  |  |  |
| `spec.outboundProxy.noProxy` | `string` |  |  |  |
| `spec.logLevel` | `string` |  | `info` |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.caBundleSecretName` | `string` |  |  |  |
| `spec.keepVolumesOnUninstall` | `bool` |  | `true` |  |
| `spec.updateStrategy` | `string` |  | `RollingUpdate` |  |
| `spec.scheduling` | `KubernetesHarborScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. NOTE the external
database/cache credential Secrets are read from this same
namespace (a Kubernetes constraint: secretKeyRef and the chart's
template-time Secret reads are namespace-local) — co-locate
Harbor with its database or replicate the credential Secret.

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

Helm chart version to install (e.g. "1.19.1" = Harbor 2.15.1).
Versions must exist in the SERVED index at
https://helm.goharbor.io.

- default: `1.19.1`

### spec.externalUrl

`string`

The URL clients use to reach this Harbor —
`protocol://host[:port]`, no trailing slash. THIS IS LOAD-BEARING
FOR PUSHES AND PULLS, not a display setting: Harbor embeds it in
the token-service URL returned to every OCI client, so `docker
login/push/pull` FAIL AUTH when the address the client dialed
does not match this value. Set it to the address the exposure in
front of this component actually serves (the composed ingress /
gateway hostname, the LoadBalancer address, or
`http://localhost:<port>` for port-forward access). An `https://`
URL with `expose.tls` disabled is legitimate exactly when TLS
terminates in front of Harbor (composed ingress or cloud LB).

- rule: external_url must be protocol://host[:port] with no trailing slash (e.g. https://harbor.example.com).

### spec.expose

`KubernetesHarborExpose`

How the front door is exposed inside the cluster. Empty = a
ClusterIP Service (compose exposure kinds in front of it).

### spec.expose.serviceType

`string` · optional (explicit presence)

Service type for the front door. Empty = "ClusterIP".
The Service is named after this resource (metadata.name) — the
chart's default fixed name "harbor" would collide between two
installs in one namespace.

- default: `ClusterIP`
- rule: service_type must be one of: ClusterIP, NodePort, LoadBalancer.

### spec.expose.tls

`KubernetesHarborExposeTls`

TLS termination at the nginx front door.

### spec.expose.tls.enabled

`bool`

Terminate TLS at nginx. When false, nginx serves plain HTTP —
pair with TLS termination in the composed exposure in front, or
accept `--insecure-registry` clients on a lab install.

### spec.expose.tls.certSecretName

`string | valueFrom`

Name of an existing kubernetes.io/tls Secret (keys `tls.crt` /
`tls.key`) — the cert-manager seam: accepts a literal or a
reference to a KubernetesCertificate pointed at the front-door
hostname. EMPTY WITH TLS ENABLED = the chart auto-generates a
self-signed certificate — and regenerates it ON EVERY APPLY
(helm `genCA` carries no state), so clients that pinned the
previous cert break; auto mode is for labs only.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.expose.nodePorts

`KubernetesHarborNodePorts`

NodePort numbers when service_type is NodePort. Empty = the
chart defaults (30002 http, 30003 https).

### spec.expose.nodePorts.http

`int32` · optional (explicit presence)

HTTP node port. Empty = 30002 (chart default).

- default: `30002`
- rule: {"int32":{"lte":32767,"gte":30000}}

### spec.expose.nodePorts.https

`int32` · optional (explicit presence)

HTTPS node port. Empty = 30003 (chart default).

- default: `30003`
- rule: {"int32":{"lte":32767,"gte":30000}}

### spec.expose.serviceAnnotations

`map<string, string>`

Annotations for the front-door Service — THE CLOUD
LOAD-BALANCER SURFACE when service_type is LoadBalancer (NLB
class/scheme on EKS, internal-LB flags on GKE/AKS, any
controller-specific tuning).

### spec.expose.sourceRanges

`[]string`

LoadBalancer-only: restrict client CIDRs
(loadBalancerSourceRanges).

### spec.expose.loadBalancerIp

`string` · optional (explicit presence)

LoadBalancer-only: request a specific LB IP where the cloud
supports assigning one.

### spec.adminAuth

`KubernetesHarborAdminAuth`

Admin ("admin" user) password source. Unset = the module
GENERATES a random password and exports it via the
`admin_password_secret` output — the recommended posture.

### spec.adminAuth.existingSecretName

`string` · optional (explicit presence)

Name of an existing Secret holding the admin password. Empty =
the module generates a random password into a module-owned
Secret (`<name>-admin-auth`) and exports it as the
admin_password_secret output.

### spec.adminAuth.existingSecretKey

`string` · optional (explicit presence)

Key within that Secret. Empty = "HARBOR_ADMIN_PASSWORD" (the
chart's contract key).

- default: `HARBOR_ADMIN_PASSWORD`

### spec.database

`KubernetesHarborDatabase` · required

The PostgreSQL database holding Harbor's metadata (projects,
users, artifacts, scan results). Exactly one arm.

- rule: {"required":true}
- rule: Exactly one database engine is required: internal or external.

### spec.database.internal

`KubernetesHarborInternalDatabase`

The chart's in-cluster single-node PostgreSQL StatefulSet.
EVALUATION-GRADE BY UPSTREAM'S OWN POSITION: one replica, no
failover, no backups — fine for labs and small teams, never
for production registries. The superuser password is
module-generated (the chart's documented default never
ships).

### spec.database.internal.resources

`ContainerResources`

CPU/memory for the PostgreSQL container. The chart ships no
defaults; these are modest laboratory defaults.

### spec.database.internal.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.database.internal.resources.limits.cpu

`string`

### spec.database.internal.resources.limits.memory

`string`

### spec.database.internal.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.database.internal.resources.requests.cpu

`string`

### spec.database.internal.resources.requests.memory

`string`

### spec.database.internal.diskSize

`string` · optional (explicit presence)

Data volume size (StatefulSet template PVC — immutable after
creation, and NEVER deleted by uninstall). Empty = "1Gi" (chart
default); size for artifact metadata growth on anything beyond
a lab.

- default: `1Gi`
- rule: disk_size must be a Kubernetes quantity (e.g. 10Gi).

### spec.database.internal.storageClass

`string` · optional (explicit presence)

StorageClass for the data volume. Empty = the cluster default.

### spec.database.internal.shmSizeLimit

`string` · optional (explicit presence)

/dev/shm size limit — PostgreSQL uses shared memory for its
shared_buffers. Empty = "512Mi" (chart default).

- default: `512Mi`

### spec.database.external

`KubernetesHarborExternalDatabase`

An external PostgreSQL — the production arm. A
KubernetesPostgres resource composes naturally (its outputs
are the field defaults); any reachable PostgreSQL works.

### spec.database.external.host

`string | valueFrom` · required

PostgreSQL host — a Service name (same namespace) or a full
FQDN. Accepts a literal or a reference to a KubernetesPostgres
resource (its read-write Service — always the current primary).

- references: KubernetesPostgres (`status.outputs.rw_service`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.rw_service}} -- a bare string does not parse

### spec.database.external.port

`int32` · optional (explicit presence)

PostgreSQL port. Empty = 5432.

- default: `5432`
- rule: {"int32":{"lte":65535,"gt":0}}

### spec.database.external.username

`string` · required

Database username (on a KubernetesPostgres this is the bootstrap
owner role — ownership covers everything Harbor's migrations
need).

- rule: {"required":true}

### spec.database.external.passwordSecretName

`string | valueFrom` · required

Name of the Secret holding the user's password. THE CHART'S
CONTRACT PINS THE KEY: the password must sit under the key
`password` — which is EXACTLY the key a KubernetesPostgres
application Secret uses, so its credential Secret composes
as-is (the default below). For any other Secret, place the
password under a `password` key. Read at install time from the
install namespace.

- references: KubernetesPostgres (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesPostgres, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.database.external.coreDatabase

`string` · optional (explicit presence)

Database that holds Harbor's schema (must exist — on a
KubernetesPostgres declare it at bootstrap via initdb). Empty =
"registry" (chart default).

- default: `registry`

### spec.database.external.sslMode

`string` · optional (explicit presence)

Postgres sslmode for the connection. Empty = "disable" — the
in-cluster plaintext default; set verify-full for external or
managed databases.

- default: `disable`
- rule: ssl_mode must be one of: disable, require, verify-ca, verify-full.

### spec.cache

`KubernetesHarborCache` · required

The Redis cache backing sessions, job queues, and manifest
caching. Exactly one arm.

- rule: {"required":true}
- rule: Exactly one cache engine is required: internal or external.

### spec.cache.internal

`KubernetesHarborInternalRedis`

The chart's in-cluster single-node Redis StatefulSet —
unauthenticated, no replication, evaluation-grade by
upstream's own position.

### spec.cache.internal.resources

`ContainerResources`

CPU/memory for the Redis container. The chart ships no
defaults; these are modest laboratory defaults.

### spec.cache.internal.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.cache.internal.resources.limits.cpu

`string`

### spec.cache.internal.resources.limits.memory

`string`

### spec.cache.internal.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.cache.internal.resources.requests.cpu

`string`

### spec.cache.internal.resources.requests.memory

`string`

### spec.cache.internal.diskSize

`string` · optional (explicit presence)

Data volume size (StatefulSet template PVC — never deleted by
uninstall). Empty = "1Gi" (chart default).

- default: `1Gi`
- rule: disk_size must be a Kubernetes quantity (e.g. 1Gi).

### spec.cache.internal.storageClass

`string` · optional (explicit presence)

StorageClass for the data volume. Empty = the cluster default.

### spec.cache.external

`KubernetesHarborExternalRedis`

An external Redis-compatible endpoint — the production arm. A
KubernetesValkey resource composes naturally for the address
(Valkey speaks the Redis protocol).

- rule: Declare the Redis credential once: password (module-materialized Secret) or existing_secret_name, not both.

### spec.cache.external.addr

`string | valueFrom` · required

Redis address as `host:port` — or, for Sentinel, a
comma-separated `host:port` list of the sentinels (set
sentinel_master_set alongside). Accepts a literal or a
reference to a KubernetesValkey resource (its in-cluster
endpoint output).

- references: KubernetesValkey (`status.outputs.kube_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.kube_endpoint}} -- a bare string does not parse

### spec.cache.external.sentinelMasterSet

`string` · optional (explicit presence)

The Sentinel master-set name — required when addr lists
sentinels, empty for a plain Redis endpoint.

### spec.cache.external.username

`string` · optional (explicit presence)

Redis username when ACLs are in play (on a KubernetesValkey this
is one of its declared ACL usernames). Empty = the default user.

### spec.cache.external.password

`string` · optional (explicit presence) · sensitive

The user's password. Declare it here and the module materializes
it into a module-owned Secret (`<name>-redis-auth`, the chart's
REDIS_PASSWORD/REDIS_USERNAME contract keys) before the release
— the password never rides rendered Helm values. Mutually
exclusive with existing_secret_name. Empty with no
existing_secret_name = an unauthenticated Redis.

### spec.cache.external.existingSecretName

`string` · optional (explicit presence)

Name of an existing Secret holding the credential under the
chart's contract keys: `REDIS_PASSWORD` (and `REDIS_USERNAME`
for ACL users). Read at install time from the install
namespace. Mutually exclusive with password. NOTE for
KubernetesValkey composition: Valkey's generated auth Secret
keys passwords BY USERNAME, not by this contract — bridge with
a Secret carrying REDIS_PASSWORD, or declare the same password
on both sides.

### spec.cache.external.tlsEnabled

`bool`

Use TLS (rediss://) for the connection. Server-authentication
only — the chart does not support mTLS to Redis; a self-signed
server rides ca_bundle_secret_name at the spec level.

### spec.storage

`KubernetesHarborArtifactStorage` · required

Where the registry stores artifact blobs. Exactly one backend.

- rule: {"required":true}
- rule: Exactly one storage backend is required: filesystem, s3, gcs, or azure.

### spec.storage.filesystem

`KubernetesHarborFilesystemStorage`

A PersistentVolumeClaim — the zero-dependency arm. Fine for a
lab or a single-node registry; object storage is the
production posture (HA, unbounded growth, lifecycle
policies).

### spec.storage.filesystem.diskSize

`string` · optional (explicit presence)

Artifact volume size. Empty = "5Gi" (chart default) — size for
real image traffic; the volume holds every layer of every
artifact.

- default: `5Gi`
- rule: disk_size must be a Kubernetes quantity (e.g. 100Gi).

### spec.storage.filesystem.storageClass

`string` · optional (explicit presence)

StorageClass for the artifact volume. Empty = the cluster default.

### spec.storage.filesystem.accessMode

`string` · optional (explicit presence)

PVC access mode. Empty = "ReadWriteOnce". ReadWriteMany is
REQUIRED to run more than one registry replica on this backend
(enforced at validation) and needs a storage class that
supports it (NFS/EFS/Filestore class).

- default: `ReadWriteOnce`
- rule: access_mode must be ReadWriteOnce or ReadWriteMany.

### spec.storage.s3

`KubernetesHarborS3Storage`

S3 or any S3-compatible endpoint (a KubernetesSeaweedFs
composes naturally via endpoint; MinIO, R2, GCS-interop all
work the same way).

### spec.storage.s3.bucket

`string` · required

Bucket name. The bucket must already exist.

- rule: {"required":true}

### spec.storage.s3.region

`string` · required

AWS region. For non-AWS S3-compatible endpoints the driver
still requires a value — any well-formed region string works
(e.g. "us-east-1").

- rule: {"required":true}

### spec.storage.s3.endpoint

`string | valueFrom`

Endpoint URL for S3-COMPATIBLE stores — set for anything that
is not AWS S3 itself. Accepts a literal or a reference to a
KubernetesSeaweedFs resource (its S3 endpoint output).

- references: KubernetesSeaweedFs (`status.outputs.s3_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.storage.s3.credentials

`KubernetesHarborS3Credentials`

Credentials. Unset = the AMBIENT chain (EKS IRSA / Pod
Identity, node roles): the registry's AWS SDK resolves
credentials from the pod environment — annotate/bind the
component ServiceAccounts out of band.

- rule: Declare S3 credentials once: access_key/secret_key (module-materialized Secret) or existing_secret_name, not both.
- rule: access_key and secret_key must be declared together.

### spec.storage.s3.credentials.accessKey

`string` · optional (explicit presence)

Access key ID. Declared credentials are materialized into a
module-owned Secret (`<name>-storage-auth`, the chart's
REGISTRY_STORAGE_S3_ACCESSKEY/SECRETKEY contract keys) before
the release — nothing credential-bearing rides rendered values.
Mutually exclusive with existing_secret_name.

### spec.storage.s3.credentials.secretKey

`string` · optional (explicit presence) · sensitive

Secret access key (paired with access_key).

### spec.storage.s3.credentials.existingSecretName

`string` · optional (explicit presence)

Name of an existing Secret carrying the chart's contract keys
REGISTRY_STORAGE_S3_ACCESSKEY and REGISTRY_STORAGE_S3_SECRETKEY.
Mutually exclusive with access_key/secret_key.

### spec.storage.s3.disableRedirect

`bool`

Disable client redirects to the backend: the registry serves
blob bytes itself instead of redirecting clients to presigned
storage URLs. REQUIRED for in-cluster S3 stores whose endpoint
clients cannot reach (SeaweedFS, MinIO) — a redirect would hand
clients an unreachable URL.

### spec.storage.s3.encrypt

`bool`

Server-side encryption for stored blobs (AWS S3).

### spec.storage.s3.secure

`bool` · optional (explicit presence)

Use HTTPS to the endpoint. Set false for plain-HTTP in-cluster stores.

- default: `true`

### spec.storage.s3.skipVerify

`bool`

Skip TLS certificate verification (self-signed endpoints).

### spec.storage.s3.rootDirectory

`string` · optional (explicit presence)

Key prefix within the bucket. Empty = the bucket root.

### spec.storage.s3.storageClass

`string` · optional (explicit presence)

S3 storage class for written blobs (e.g. STANDARD, REDUCED_REDUNDANCY).

### spec.storage.gcs

`KubernetesHarborGcsStorage`

Google Cloud Storage.

- rule: Declare GCS credentials once: use_workload_identity (keyless), key_data (module-materialized Secret), or existing_secret_name.

### spec.storage.gcs.bucket

`string` · required

Bucket name. The bucket must already exist.

- rule: {"required":true}

### spec.storage.gcs.useWorkloadIdentity

`bool`

Authenticate via GKE Workload Identity — the KEYLESS arm
(recommended on GKE): no key material anywhere; bind the
registry's Kubernetes ServiceAccount to a GCP service account
with bucket access out of band. Mutually exclusive with
key_data / existing_secret_name.

### spec.storage.gcs.keyData

`string` · optional (explicit presence) · sensitive

Base64-encoded service-account key JSON. Declared keys are
materialized into a module-owned Secret (`<name>-storage-auth`,
the chart's GCS_KEY_DATA contract key) before the release.
Mutually exclusive with use_workload_identity.

### spec.storage.gcs.existingSecretName

`string` · optional (explicit presence)

Name of an existing Secret carrying the chart's contract key
GCS_KEY_DATA. Mutually exclusive with use_workload_identity and
key_data.

### spec.storage.gcs.rootDirectory

`string` · optional (explicit presence)

Key prefix within the bucket. Empty = the bucket root.

### spec.storage.gcs.chunkSize

`int32` · optional (explicit presence)

Upload chunk size in bytes. Empty = the driver default (5242880).

### spec.storage.azure

`KubernetesHarborAzureStorage`

Azure Blob Storage.

- rule: Declare EXACTLY one Azure credential arm: account_key (module-materialized Secret) or existing_secret_name — the registry's azure driver has no ambient credential chain.

### spec.storage.azure.accountName

`string` · required

Storage account name.

- rule: {"required":true}

### spec.storage.azure.container

`string` · required

Blob container name. The container must already exist.

- rule: {"required":true}

### spec.storage.azure.accountKey

`string` · optional (explicit presence) · sensitive

Storage account key. Declared keys are materialized into a
module-owned Secret (`<name>-storage-auth`, the chart's
AZURE_STORAGE_ACCESS_KEY contract key) before the release.
Mutually exclusive with existing_secret_name.

### spec.storage.azure.existingSecretName

`string` · optional (explicit presence)

Name of an existing Secret carrying the chart's contract key
AZURE_STORAGE_ACCESS_KEY. Mutually exclusive with account_key.

### spec.storage.azure.realm

`string` · optional (explicit presence)

Azure cloud realm. Empty = core.windows.net (public cloud).

- default: `core.windows.net`

### spec.trivy

`KubernetesHarborTrivy`

The Trivy vulnerability scanner. Unset = ENABLED with the
chart's defaults (Trivy is Harbor's built-in scanner and the
chart ships it on) — set `enabled: false` to run scannerless.

### spec.trivy.enabled

`bool` · optional (explicit presence)

Deploy Trivy. Empty = TRUE (chart truth — Trivy is Harbor's
built-in scanner).

- default: `true`

### spec.trivy.replicas

`int32` · optional (explicit presence)

Trivy replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.trivy.resources

`ContainerResources`

CPU/memory. Empty = the chart's own defaults (requests
200m/512Mi, limits 1/1Gi) — the one component the chart sizes
itself.

### spec.trivy.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.trivy.resources.limits.cpu

`string`

### spec.trivy.resources.limits.memory

`string`

### spec.trivy.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.trivy.resources.requests.cpu

`string`

### spec.trivy.resources.requests.memory

`string`

### spec.trivy.diskSize

`string` · optional (explicit presence)

Vulnerability-database cache volume size. Empty = "5Gi" (chart
default).

- default: `5Gi`

### spec.trivy.skipUpdate

`bool`

KNOW THIS (internet-reaching runtime behavior): Trivy DOWNLOADS
its vulnerability database from an OCI registry at scan time
and refreshes it every ~12 hours — a fresh install needs
outbound access to the db_repositories below (or the
outbound_proxy). In an AIR-GAPPED cluster set skip_update AND
skip_java_db_update to true and pre-load the database files
onto the cache volume (`/home/scanner/.cache/trivy/...`)
yourself — with skip_update and no pre-loaded DB, every scan
fails.

### spec.trivy.skipJavaDbUpdate

`bool`

Skip the Java vulnerability DB download (see skip_update).

### spec.trivy.offlineScan

`bool`

Prevent Trivy from making API requests to identify
dependencies during scans (air-gap posture; DB downloads are
governed separately by skip_update).

### spec.trivy.dbRepositories

`[]string`

OCI repositories to pull the vulnerability DB from, in priority
order. Empty = the chart defaults (mirror.gcr.io/aquasec/trivy-db,
then ghcr.io/aquasecurity/trivy-db).

### spec.trivy.javaDbRepositories

`[]string`

OCI repositories for the Java vulnerability DB, in priority
order. Empty = the chart defaults.

### spec.trivy.githubToken

`string` · optional (explicit presence) · sensitive

GitHub token to raise the anonymous DB-download rate limit.
KNOW THIS: the chart accepts this only as a plain value that it
renders into its own Secret — it rides the release values
(sensitive-marked, never printed in plans) unlike this
component's other credentials, which travel via pre-created
Secrets. Leave empty unless you hit rate limits.

### spec.trivy.severity

`string` · optional (explicit presence)

Comma-separated severities to report. Empty = the scanner
default (UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL).

### spec.trivy.ignoreUnfixed

`bool`

Report only vulnerabilities with a published fix.

### spec.trivy.timeout

`string` · optional (explicit presence)

Scan timeout. Empty = "5m0s" (chart default).

### spec.core

`KubernetesHarborComponent`

Sizing for the core component (API server, auth, webhooks).
First-boot schema migrations run inside core's startup window —
upstream budgets the startup probe at 360 × 10s.

### spec.core.replicas

`int32` · optional (explicit presence)

Pod replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.core.resources

`ContainerResources`

CPU and memory for the component container.

### spec.core.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.core.resources.limits.cpu

`string`

### spec.core.resources.limits.memory

`string`

### spec.core.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.core.resources.requests.cpu

`string`

### spec.core.resources.requests.memory

`string`

### spec.portal

`KubernetesHarborComponent`

Sizing for the portal component (web UI).

### spec.portal.replicas

`int32` · optional (explicit presence)

Pod replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.portal.resources

`ContainerResources`

CPU and memory for the component container.

### spec.portal.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.portal.resources.limits.cpu

`string`

### spec.portal.resources.limits.memory

`string`

### spec.portal.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.portal.resources.requests.cpu

`string`

### spec.portal.resources.requests.memory

`string`

### spec.registry

`KubernetesHarborComponent`

Sizing for the registry component (the OCI distribution backend
+ its registryctl controller container). More than one replica
on the `filesystem` storage backend requires the PVC access mode
to be ReadWriteMany — enforced at validation.

### spec.registry.replicas

`int32` · optional (explicit presence)

Pod replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.registry.resources

`ContainerResources`

CPU and memory for the component container.

### spec.registry.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.registry.resources.limits.cpu

`string`

### spec.registry.resources.limits.memory

`string`

### spec.registry.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.registry.resources.requests.cpu

`string`

### spec.registry.resources.requests.memory

`string`

### spec.jobservice

`KubernetesHarborJobservice`

Sizing and job tuning for the jobservice component.

### spec.jobservice.replicas

`int32` · optional (explicit presence)

Pod replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.jobservice.resources

`ContainerResources`

CPU and memory for the jobservice container.

### spec.jobservice.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.jobservice.resources.limits.cpu

`string`

### spec.jobservice.resources.limits.memory

`string`

### spec.jobservice.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.jobservice.resources.requests.cpu

`string`

### spec.jobservice.resources.requests.memory

`string`

### spec.jobservice.maxJobWorkers

`int32` · optional (explicit presence)

Maximum concurrent job workers (replication, GC, scans). Empty
= 10 (chart default).

- default: `10`
- rule: {"int32":{"gte":1}}

### spec.jobservice.logDiskSize

`string` · optional (explicit presence)

Job-log volume size (job logs default to file storage on a
PVC). Empty = "1Gi" (chart default).

- default: `1Gi`

### spec.nginx

`KubernetesHarborComponent`

Sizing for the nginx front door (always deployed — it terminates
client traffic for every exposure mode this component models).

### spec.nginx.replicas

`int32` · optional (explicit presence)

Pod replicas. Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.nginx.resources

`ContainerResources`

CPU and memory for the component container.

### spec.nginx.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.nginx.resources.limits.cpu

`string`

### spec.nginx.resources.limits.memory

`string`

### spec.nginx.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.nginx.resources.requests.cpu

`string`

### spec.nginx.resources.requests.memory

`string`

### spec.internalTls

`KubernetesHarborInternalTls`

TLS between Harbor's own components (core ↔ registry ↔
jobservice ↔ portal ↔ trivy). Off by default — the usual
in-cluster posture.

### spec.internalTls.enabled

`bool`

Serve HTTPS between components.

### spec.internalTls.certSecrets

`KubernetesHarborInternalTlsSecrets`

Per-component certificate Secrets (kubernetes.io/tls shape plus
a `ca.crt` key) — the cert-manager seam. EMPTY = the chart
auto-generates an internal CA and per-component certs — and
REGENERATES THEM ON EVERY APPLY (helm genCA carries no state),
rolling every component each time; auto mode is for labs only.

### spec.internalTls.certSecrets.coreSecretName

`string` · required

- rule: {"required":true}

### spec.internalTls.certSecrets.jobserviceSecretName

`string` · required

- rule: {"required":true}

### spec.internalTls.certSecrets.registrySecretName

`string` · required

- rule: {"required":true}

### spec.internalTls.certSecrets.portalSecretName

`string` · required

- rule: {"required":true}

### spec.internalTls.certSecrets.trivySecretName

`string` · optional (explicit presence)

Required only when Trivy is enabled; ignored otherwise.

### spec.internalTls.strongSslCiphers

`bool`

Restrict components to strong SSL ciphers.

### spec.metrics

`KubernetesHarborMetrics`

Prometheus metrics for core/registry/jobservice plus the
dedicated exporter Deployment, and the optional ServiceMonitor.

### spec.metrics.enabled

`bool`

Expose /metrics on core, registry, and jobservice, and deploy
the dedicated harbor-exporter.

### spec.metrics.serviceMonitorEnabled

`bool`

Create a ServiceMonitor (requires the monitoring.coreos.com
CRDs — a KubernetesKubePrometheusStack composes).

### spec.metrics.serviceMonitorInterval

`string` · optional (explicit presence)

ServiceMonitor scrape interval. Empty = the Prometheus default.

### spec.metrics.serviceMonitorLabels

`map<string, string>`

Extra labels for the ServiceMonitor (Prometheus selector matching).

### spec.cacheLayer

`KubernetesHarborCacheLayer`

The optional Redis-backed manifest/metadata cache layer
(improves high-concurrency pull performance). Off by default
(chart truth).

### spec.cacheLayer.enabled

`bool`

Cache project/repository/artifact/manifest metadata in Redis —
improves high-concurrency pull performance. Off by default
(chart truth).

### spec.cacheLayer.expireHours

`int32` · optional (explicit presence)

Cache TTL in hours. Empty = 24 (chart default).

- default: `24`
- rule: {"int32":{"gte":1}}

### spec.outboundProxy

`KubernetesHarborOutboundProxy`

Outbound proxy for Trivy database updates and replication to
registries that cannot be reached directly.

### spec.outboundProxy.httpProxy

`string` · optional (explicit presence)

HTTP proxy URL.

### spec.outboundProxy.httpsProxy

`string` · optional (explicit presence)

HTTPS proxy URL.

### spec.outboundProxy.noProxy

`string` · optional (explicit presence)

Extra no-proxy entries appended to the chart's built-in
component list (every Harbor component Service is always
excluded). Empty = "127.0.0.1,localhost,.local,.internal"
(chart default).

### spec.logLevel

`string` · optional (explicit presence)

Component log level. Empty = "info".

- default: `info`
- rule: log_level must be one of: debug, info, warning, error, fatal.

### spec.imageRegistry

`string` · optional (explicit presence)

Registry mirror for AIR-GAPPED installs: rewrites the registry
host of every component image (the chart pins all components to
`docker.io/goharbor/<component>:v<appVersion>`; this replaces
`docker.io` — e.g. "mirror.example.com" pulls
`mirror.example.com/goharbor/harbor-core:...`). The image
REPOSITORY PATHS and TAGS stay chart-truth; mirror the
goharbor/* images under the same paths.

### spec.imagePullSecrets

`[]string`

Names of existing imagePullSecrets to attach to every component
pod (for the mirror above or Docker Hub rate limits). The
Secrets must already exist in the install namespace.

### spec.caBundleSecretName

`string` · optional (explicit presence)

Name of an existing Secret carrying a `ca.crt` key, injected
into the trust store of core/jobservice/registry/trivy — for
artifact storage or replication endpoints serving self-signed
certificates.

### spec.keepVolumesOnUninstall

`bool` · optional (explicit presence)

Keep the registry and jobservice PVCs when this resource is
destroyed. Empty = TRUE (chart truth: the chart stamps
`helm.sh/resource-policy: keep` on both PVCs by default, so
artifact blobs and job logs survive uninstall for a reinstall to
adopt). KNOW THIS: the INTERNAL database and Redis volumes are
StatefulSet-template PVCs that Helm NEVER deletes regardless of
this setting — plan on sweeping them explicitly when retiring an
install for good.

- default: `true`

### spec.updateStrategy

`string` · optional (explicit presence)

Rollout strategy for the components with persistent volumes
(registry, jobservice). Empty = "RollingUpdate". Set "Recreate"
when their PVC access mode is ReadWriteOnce and the storage
class cannot multi-attach — a rolling update would wedge the new
pod on volume attach.

- default: `RollingUpdate`
- rule: update_strategy must be RollingUpdate or Recreate.

### spec.scheduling

`KubernetesHarborScheduling`

Pod scheduling constraints, applied to EVERY Harbor component
(core, portal, registry, jobservice, trivy, nginx, exporter, and
the internal database/redis when those arms are active) — split
placement would separate components that share volumes and
fail over together. Per-component placement rides helm_values.

### spec.scheduling.nodeSelector

`map<string, string>`

Schedule onto nodes carrying these labels.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.helmValues

`string`

Advanced escape hatch: raw Helm values merged LAST (Helm `-f`
semantics) over everything this spec renders — later keys win.
Use it for surfaces deliberately not modeled (per-component
probes/placement, swift/oss storage backends, GDPR knobs, trace
export, registry middleware). The module re-pins
`fullnameOverride` and the exposure Service name after the
merge. YAML document as a string.

## Validation Rules

- `spec.registry.filesystem_multi_replica`: More than one registry replica on the filesystem backend requires storage.filesystem.access_mode: ReadWriteMany — with ReadWriteOnce each replica would need the same volume attached to a different node. Use an object-storage backend (s3/gcs/azure) to scale the registry.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesHarbor, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace Harbor is installed into. |
| `status.outputs.expose_service` | `string` | The front-door Service (nginx) — the ONE service clients and composed exposure kinds point at; it proxies UI, API, and OCI traffic. Named after this resource. |
| `status.outputs.kube_endpoint` | `string` | In-cluster URL of the front door, e.g. `http://<expose_service>.<namespace>.svc.cluster.local:80` (https/443 when expose.tls is enabled). |
| `status.outputs.external_url` | `string` | The declared external URL — the address OCI clients must use (token-service auth is bound to it). |
| `status.outputs.core_service` | `string` | Core (API) Service name. |
| `status.outputs.portal_service` | `string` | Portal (web UI) Service name. |
| `status.outputs.registry_service` | `string` | Registry (OCI distribution) Service name. |
| `status.outputs.jobservice_service` | `string` | Jobservice Service name. |
| `status.outputs.trivy_service` | `string` | Trivy Service name — empty when the scanner is disabled. |
| `status.outputs.database_service` | `string` | Internal database Service name — set only on the internal database arm (empty when composing an external PostgreSQL). |
| `status.outputs.redis_service` | `string` | Internal Redis Service name — set only on the internal cache arm. |
| `status.outputs.admin_username` | `string` | Admin username — always "admin" (a Harbor constant). |
| `status.outputs.admin_password_secret` | `KubernetesSecretKey` | The Secret holding the admin password: the module-owned `<name>-admin-auth` (key HARBOR_ADMIN_PASSWORD) on the generated arm, or the user's declared Secret/key. Retrieve: `kubectl get secret <name> -n <namespace> -o jsonpath='{.data.<key>}' \| base64 -d`. |
| `status.outputs.admin_password_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.admin_password_secret.key` | `string` | The key within the Kubernetes Secret. |
| `status.outputs.port_forward_command` | `string` | Copy-paste command to reach the UI/API from a workstation. NOTE pushes/pulls through this tunnel only authenticate when external_url matches the forwarded address (see the spec's external_url comment). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.expose.tls.certSecretName` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.database.external.host` | KubernetesPostgres | `status.outputs.rw_service` |
| `spec.database.external.passwordSecretName` | KubernetesPostgres | `status.outputs.password_secret.name` |
| `spec.cache.external.addr` | KubernetesValkey | `status.outputs.kube_endpoint` |
| `spec.storage.s3.endpoint` | KubernetesSeaweedFs | `status.outputs.s3_endpoint` |

## See Also

- [Overview](../README.md)
