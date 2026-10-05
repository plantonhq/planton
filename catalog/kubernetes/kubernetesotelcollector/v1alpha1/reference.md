# KubernetesOtelCollector

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesOtelCollectorSpec** declares one OpenTelemetry Collector —
the telemetry pipeline that receives, processes and exports logs,
metrics and traces — as an `OpenTelemetryCollector` custom resource
reconciled by the OpenTelemetry Operator (deploy
KubernetesOtelOperator first).

THE PIPELINE IS THE PRODUCT: `config_yaml` carries the collector's
own configuration document (receivers → processors → exporters wired
into service pipelines) on the collector's OWN open contract — the
OpenTelemetry component registry is unbounded by design, so this is
the upstream grain, not an escape hatch. The presets carry the
high-demand pipelines ready to remix: cluster logs → Loki (daemonset
mode — this kind is how logs reach a KubernetesLoki), traces →
Tempo, and OTLP fan-in → both.

MODES: `deployment` (default) for a scalable gateway/fan-in
collector; `daemonset` for per-node collection (log files, host and
kubelet metrics — pair with `volumes` hostPath mounts to reach
/var/log); `statefulset` when the target allocator or persistent
queues need stable identities; `sidecar` registers the CR for the
operator's per-pod injection (pods opt in with the
`sidecar.opentelemetry.io/inject` annotation) — no standalone
workload is created.

SECRETS: never inline credentials in `config_yaml`. Load them as
environment variables from existing Secrets (`env_from_secrets`) and
reference them in the config as `${env:VAR_NAME}` — the collector
expands them at start, so tokens never land in the rendered
ConfigMap.

PERMISSIONS: receivers that read cluster state (k8s_events,
kubeletstats, k8s_cluster, filelog with kubernetes enrichment) need
RBAC beyond the default ServiceAccount — compose a
KubernetesServiceAccount + KubernetesRbac and set `service_account`.

The operator injects the collector image when `image` is empty and
derives Service ports from the declared receivers — the exported
OTLP endpoints assume the standard `otlp` receiver on 4317/4318
(every preset declares it).

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the collector CR expressed at once — statefulset
# mode with the operator-managed autoscaler (CPU + memory targets), a
# custom image, an existing ServiceAccount, plain env vars plus the
# secret-env credential path (referenced in config as ${env:...} — the
# secret-safe contract), emptyDir and secret volumes with mounts, an
# extra UDP Service port the operator cannot infer, resources sized with
# the memory_limiter, and full scheduling — while the spec's validation
# rules stay satisfied (autoscaler excludes replicas; scaling only in
# workload modes).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOtelCollector
metadata:
  name: otel-gateway-full
spec:
  namespace:
    value: observability
  createNamespace: true
  mode: statefulset
  configYaml: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318
      syslog:
        udp:
          listen_address: 0.0.0.0:54527
        protocol: rfc5424
    processors:
      memory_limiter:
        check_interval: 1s
        limit_mib: 400
        spike_limit_mib: 100
      batch: {}
    exporters:
      otlphttp:
        endpoint: http://tempo.observability.svc.cluster.local:4318
        headers:
          authorization: "Bearer ${env:TEMPO_BEARER_TOKEN}"
    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [memory_limiter, batch]
          exporters: [otlphttp]
        logs:
          receivers: [syslog]
          processors: [memory_limiter, batch]
          exporters: [otlphttp]
  autoscaler:
    minReplicas: 2
    maxReplicas: 5
    targetCpuUtilization: 75
    targetMemoryUtilization: 80
  image: mirror.example.com/otel/custom-collector:0.156.0
  serviceAccount: otel-gateway-sa
  env:
    GOMAXPROCS: "2"
  envFromSecrets:
    - tempo-credentials
  volumes:
    - name: scratch
      mountPath: /var/lib/otelcol
      emptyDir:
        sizeLimit: 1Gi
    - name: extra-ca
      mountPath: /etc/ssl/extra
      readOnly: true
      secret:
        name: extra-ca-bundle
        key: ca.crt
  additionalPorts:
    - name: syslog
      port: 54527
      protocol: UDP
  resources:
    requests:
      cpu: 200m
      memory: 512Mi
    limits:
      memory: 512Mi
  scheduling:
    nodeSelector:
      workload: observability
    tolerations:
      - key: dedicated
        operator: Equal
        value: observability
        effect: NoSchedule
    priorityClassName: high-priority
  podSecurityContext:
    runAsNonRoot: true
    fsGroup: 10001
    fsGroupChangePolicy: OnRootMismatch
    seccompProfile:
      type: RuntimeDefault
  serviceMonitorEnabled: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.mode` | `enum` |  | `deployment` |  |
| `spec.configYaml` | `string` | yes |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.autoscaler` | `KubernetesOtelCollectorAutoscaler` |  |  |  |
| `spec.autoscaler.minReplicas` | `int32` |  | `1` |  |
| `spec.autoscaler.maxReplicas` | `int32` |  |  |  |
| `spec.autoscaler.targetCpuUtilization` | `int32` |  | `80` |  |
| `spec.autoscaler.targetMemoryUtilization` | `int32` |  |  |  |
| `spec.image` | `string` |  |  |  |
| `spec.serviceAccount` | `string` |  |  |  |
| `spec.env` | `map<string, string>` |  |  |  |
| `spec.envFromSecrets` | `[]string` |  |  |  |
| `spec.volumes` | `[]VolumeMount` |  |  |  |
| `spec.volumes[].name` | `string` | yes |  |  |
| `spec.volumes[].mountPath` | `string` | yes |  |  |
| `spec.volumes[].readOnly` | `bool` |  |  |  |
| `spec.volumes[].subPath` | `string` |  |  |  |
| `spec.volumes[].configMap` | `ConfigMapVolumeSource` |  |  |  |
| `spec.volumes[].configMap.name` | `string` | yes |  |  |
| `spec.volumes[].configMap.key` | `string` |  |  |  |
| `spec.volumes[].configMap.path` | `string` |  |  |  |
| `spec.volumes[].configMap.defaultMode` | `int32` |  |  |  |
| `spec.volumes[].secret` | `SecretVolumeSource` |  |  |  |
| `spec.volumes[].secret.name` | `string` | yes |  |  |
| `spec.volumes[].secret.key` | `string` |  |  |  |
| `spec.volumes[].secret.path` | `string` |  |  |  |
| `spec.volumes[].secret.defaultMode` | `int32` |  |  |  |
| `spec.volumes[].hostPath` | `HostPathVolumeSource` |  |  |  |
| `spec.volumes[].hostPath.path` | `string` | yes |  |  |
| `spec.volumes[].hostPath.type` | `string` |  |  |  |
| `spec.volumes[].emptyDir` | `EmptyDirVolumeSource` |  |  |  |
| `spec.volumes[].emptyDir.medium` | `string` |  |  |  |
| `spec.volumes[].emptyDir.sizeLimit` | `string` |  |  |  |
| `spec.volumes[].pvc` | `PvcVolumeSource` |  |  |  |
| `spec.volumes[].pvc.claimName` | `string` | yes |  |  |
| `spec.volumes[].pvc.readOnly` | `bool` |  |  |  |
| `spec.volumes[].serviceAccountToken` | `ServiceAccountTokenVolumeSource` |  |  |  |
| `spec.volumes[].serviceAccountToken.audience` | `string` | yes |  |  |
| `spec.volumes[].serviceAccountToken.expirationSeconds` | `int64` |  |  |  |
| `spec.volumes[].serviceAccountToken.path` | `string` |  |  |  |
| `spec.additionalPorts` | `[]KubernetesOtelCollectorPort` |  |  |  |
| `spec.additionalPorts[].name` | `string` | yes |  |  |
| `spec.additionalPorts[].port` | `int32` |  |  |  |
| `spec.additionalPorts[].protocol` | `string` |  | `TCP` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.scheduling` | `KubernetesOtelCollectorScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.podSecurityContext` | `WorkloadPodSecurityContext` |  |  |  |
| `spec.podSecurityContext.runAsUser` | `int64` |  |  |  |
| `spec.podSecurityContext.runAsGroup` | `int64` |  |  |  |
| `spec.podSecurityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.podSecurityContext.fsGroup` | `int64` |  |  |  |
| `spec.podSecurityContext.fsGroupChangePolicy` | `string` |  |  |  |
| `spec.podSecurityContext.supplementalGroups` | `[]int64` |  |  |  |
| `spec.podSecurityContext.sysctls` | `[]WorkloadSysctl` |  |  |  |
| `spec.podSecurityContext.sysctls[].name` | `string` | yes |  |  |
| `spec.podSecurityContext.sysctls[].value` | `string` | yes |  |  |
| `spec.podSecurityContext.seccompProfile` | `WorkloadSeccompProfile` |  |  |  |
| `spec.podSecurityContext.seccompProfile.type` | `string` | yes |  |  |
| `spec.podSecurityContext.seccompProfile.localhostProfile` | `string` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to deploy into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before deploying and deleted with the resource.
When false, the namespace must already exist.

### spec.mode

`enum` · optional (explicit presence)

How the collector runs. Empty = deployment.

- default: `deployment`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `kubernetes_otel_collector_mode_unspecified` -- Unspecified. Defaults to deployment.
- `deployment` -- A Deployment — the scalable gateway/fan-in shape (apps push to it, it fans out to backends). The default.
- `daemonset` -- A DaemonSet — one collector per node, for log files and host/kubelet metrics that only exist node-locally.
- `statefulset` -- A StatefulSet — stable identities, for the target allocator and persistent sending queues.
- `sidecar` -- No standalone workload: the operator injects this collector as a sidecar into pods annotated `sidecar.opentelemetry.io/inject`.

### spec.configYaml

`string` · required

The collector configuration document (YAML): receivers,
processors, exporters, connectors, extensions and the service
pipelines wiring them together. Validated for shape by the
operator's admission webhook at apply time. See the SECRETS note
above — credentials ride environment variables, never this
document.

- rule: {"required":true}

### spec.replicas

`int32` · optional (explicit presence)

Number of collector replicas. Deployment/statefulset modes only —
daemonset runs one per node and sidecar runs inside the target
pods. Ignored when `autoscaler` is set.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.autoscaler

`KubernetesOtelCollectorAutoscaler`

Horizontal autoscaling (the operator manages an HPA for the
collector). Deployment/statefulset modes only.

- rule: max_replicas must be at least min_replicas

### spec.autoscaler.minReplicas

`int32` · optional (explicit presence)

Minimum replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.autoscaler.maxReplicas

`int32`

Maximum replicas.

- rule: {"int32":{"gte":1}}

### spec.autoscaler.targetCpuUtilization

`int32` · optional (explicit presence)

Target average CPU utilization percent across replicas.

- default: `80`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.autoscaler.targetMemoryUtilization

`int32` · optional (explicit presence)

Target average memory utilization percent across replicas. Empty =
CPU-only scaling.

- rule: {"int32":{"lte":100,"gte":1}}

### spec.image

`string`

Collector image. Empty = the operator's default (the
opentelemetry-collector-k8s distribution at the operator's paired
version — override fleet-wide via the operator kind's
`default_collector_image`). Set for a custom distribution with
extra components (e.g. a builder image carrying a vendor
exporter).

### spec.serviceAccount

`string`

Existing ServiceAccount to run as. Empty = the operator creates a
default one. Set when the pipeline reads cluster state (see the
PERMISSIONS note above).

### spec.env

`map<string, string>`

Plain environment variables for the collector container
(referenced in config as `${env:NAME}`).

### spec.envFromSecrets

`[]string`

Names of existing Secrets loaded WHOLE as environment variables
(each key becomes a variable) — the credential path for exporters;
see the SECRETS note above.

### spec.volumes

`[]VolumeMount`

Volumes mounted into the collector container. The daemonset
log-collection pattern mounts /var/log/pods (hostPath, read-only);
checkpoint state for the filelog receiver rides a hostPath or
emptyDir.

### spec.volumes[].name

`string` · required

Name of the volume mount. Must be unique within the container.
Used to correlate with the volume definition.

- rule: {"required":true}

### spec.volumes[].mountPath

`string` · required

Path within the container at which the volume should be mounted.
Must be an absolute path.

- rule: {"required":true}

### spec.volumes[].readOnly

`bool`

Whether the volume should be mounted read-only.
Default is false.

### spec.volumes[].subPath

`string`

Path within the volume from which the container's volume should be mounted.
Defaults to "" (volume's root).
Useful for mounting a subdirectory of a volume.

### spec.volumes[].configMap

`ConfigMapVolumeSource`

ConfigMap volume source.
Use this to mount a ConfigMap as a file or directory.

### spec.volumes[].configMap.name

`string` · required

Name of the ConfigMap to mount.
Can reference a ConfigMap defined in spec.config_maps or an existing one in the namespace.

- rule: {"required":true}

### spec.volumes[].configMap.key

`string`

Specific key from the ConfigMap to mount as a single file.
If not specified, all keys are mounted as files in the directory.

### spec.volumes[].configMap.path

`string`

If key is specified, this is the filename to use for the mounted file.
Defaults to the key name if not specified.
Example: key="config" path="app.yaml" mounts the "config" key as "app.yaml"

### spec.volumes[].configMap.defaultMode

`int32`

Mode bits to use on created files. Must be a value between 0 and 0777.
Defaults to 0644.
Use 0755 (493 in decimal) for executable scripts.

### spec.volumes[].secret

`SecretVolumeSource`

Secret volume source.
Use this to mount a Secret as a file or directory.

### spec.volumes[].secret.name

`string` · required

Name of the Secret to mount.

- rule: {"required":true}

### spec.volumes[].secret.key

`string`

Specific key from the Secret to mount as a single file.
If not specified, all keys are mounted as files in the directory.

### spec.volumes[].secret.path

`string`

If key is specified, this is the filename to use for the mounted file.
Defaults to the key name if not specified.

### spec.volumes[].secret.defaultMode

`int32`

Mode bits to use on created files. Must be a value between 0 and 0777.
Defaults to 0644.

### spec.volumes[].hostPath

`HostPathVolumeSource`

HostPath volume source.
Use this to mount a file or directory from the host node's filesystem.
Common for DaemonSets that need access to node-level resources.

### spec.volumes[].hostPath.path

`string` · required

Path on the host to mount.

- rule: {"required":true}

### spec.volumes[].hostPath.type

`string`

Type of the host path.
Valid values:
  "" - Empty string (default) means no check is performed before mounting
  "DirectoryOrCreate" - Create directory if it doesn't exist
  "Directory" - Directory must exist
  "FileOrCreate" - Create file if it doesn't exist
  "File" - File must exist
  "Socket" - UNIX socket must exist
  "CharDevice" - Character device must exist
  "BlockDevice" - Block device must exist

- rule: Type must be one of: "", "DirectoryOrCreate", "Directory", "FileOrCreate", "File", "Socket", "CharDevice", "BlockDevice"

### spec.volumes[].emptyDir

`EmptyDirVolumeSource`

EmptyDir volume source.
Use this for temporary storage that is erased when the pod is removed.
Useful for scratch space, caching, or sharing data between containers.

### spec.volumes[].emptyDir.medium

`string`

Medium for the empty directory.
"" (default) uses the node's default medium (typically disk).
"Memory" uses a tmpfs (RAM-backed filesystem).

Memory-backed volumes are faster but:
- Count against container memory limits
- Are lost on node restart
- Should have sizeLimit set to prevent OOM

- rule: Medium must be either "" or "Memory"

### spec.volumes[].emptyDir.sizeLimit

`string`

Size limit for the empty directory.
Format: Kubernetes quantity (e.g., "1Gi", "500Mi").
Only strictly enforced when medium is "Memory".
For disk-backed volumes, this is a best-effort limit.

### spec.volumes[].pvc

`PvcVolumeSource`

PersistentVolumeClaim volume source.
Use this to mount an existing PVC.
For StatefulSets, this can reference a volumeClaimTemplate.

### spec.volumes[].pvc.claimName

`string` · required

Name of the PersistentVolumeClaim to mount.
For StatefulSets, this can be the name of a volumeClaimTemplate.

- rule: {"required":true}

### spec.volumes[].pvc.readOnly

`bool`

Whether the PVC should be mounted read-only.
Default is false.

### spec.volumes[].serviceAccountToken

`ServiceAccountTokenVolumeSource`

Projected ServiceAccount token volume source.
Use this to mount a short-lived, audience-bound identity token that the
kubelet issues for the pod's ServiceAccount and rotates automatically.

### spec.volumes[].serviceAccountToken.audience

`string` · required

Intended audience of the token. The receiving service must identify
itself with this audience when verifying the token; a token minted for a
different audience is rejected. Required: an audience-less token would be
replayable against any service in the cluster.

- rule: {"required":true}

### spec.volumes[].serviceAccountToken.expirationSeconds

`int64`

Requested lifetime of the token in seconds. The kubelet starts rotating
the token when it passes 80% of its lifetime or 24 hours, whichever is
shorter. Defaults to 3600 (1 hour). The Kubernetes API enforces a
minimum of 600 (10 minutes).

- rule: Expiration must be at least 600 seconds (the Kubernetes API minimum)

### spec.volumes[].serviceAccountToken.path

`string`

Filename for the token relative to the mount path.
Defaults to "token".

### spec.additionalPorts

`[]KubernetesOtelCollectorPort`

Extra Service ports beyond the ones the operator derives from the
declared receivers (it knows the standard components' ports —
OTLP, jaeger, zipkin, prometheus). Needed only for receivers the
operator cannot infer.

### spec.additionalPorts[].name

`string` · required

Port name (lowercase RFC-1123 label, e.g. "syslog").

- rule: {"required":true,"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.additionalPorts[].port

`int32`

Port number.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.additionalPorts[].protocol

`string` · optional (explicit presence)

Protocol: "TCP" (default) or "UDP".

- default: `TCP`
- rule: {"string":{"in":["TCP","UDP"]}}

### spec.resources

`ContainerResources`

CPU and memory for the collector container. Empty = no
requests/limits. Size memory with the memory_limiter processor's
limit — the collector sheds load instead of OOMing when they
agree.

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

### spec.scheduling

`KubernetesOtelCollectorScheduling`

Scheduling for the collector pods (deployment/daemonset/
statefulset modes). In sidecar mode the collector runs inside the
TARGET pods, whose scheduling this CR does not control —
tolerations and priority_class_name are rejected there (mirrored
from the CRD's own admission rules; see the message-level CEL).

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the collector pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the collector pods. A daemonset that reads every
node's logs tolerates every NoSchedule taint (`operator: Exists`,
`effect: NoSchedule`, no key): control-plane nodes and tainted pools
alike (build or GPU pools keep other pods off with a taint), so no
node's logs are skipped. A control-plane-only toleration silently
misses every other tainted node.

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

### spec.scheduling.priorityClassName

`string`

Priority class name for the collector pods.

### spec.podSecurityContext

`WorkloadPodSecurityContext`

Pod-level security context for the collector pods. The daemonset
log-collection pattern typically needs `run_as_user: 0` —
container runtimes write pod log files readable only by root, and
the default collector image runs as a non-root user that cannot
open them (the filelog receiver then reports permission errors
instead of shipping logs).

### spec.podSecurityContext.runAsUser

`int64` · optional (explicit presence)

UID all container processes run as unless overridden per container.

### spec.podSecurityContext.runAsGroup

`int64` · optional (explicit presence)

Primary GID all container processes run as unless overridden per container.

### spec.podSecurityContext.runAsNonRoot

`bool` · optional (explicit presence)

Refuse to start any container whose effective user is root.

### spec.podSecurityContext.fsGroup

`int64` · optional (explicit presence)

GID that owns mounted volumes and is added to every container's supplemental
groups — the standard fix for "permission denied" on persistent volumes written
by non-root apps.

### spec.podSecurityContext.fsGroupChangePolicy

`string`

When volume ownership is re-chowned to fs_group: "Always" (default) or
"OnRootMismatch" (skip the recursive chown when the root already matches —
dramatically faster pod starts on large volumes).

- rule: fsGroupChangePolicy must be either "Always" or "OnRootMismatch"

### spec.podSecurityContext.supplementalGroups

`[]int64`

Additional group IDs applied to all container processes.

### spec.podSecurityContext.sysctls

`[]WorkloadSysctl`

Kernel parameters set for the pod. Only safe sysctls (or those the cluster
administrator has allow-listed on the kubelet) are admitted.

### spec.podSecurityContext.sysctls[].name

`string` · required

Sysctl name, e.g. "net.core.somaxconn".

- rule: {"required":true}

### spec.podSecurityContext.sysctls[].value

`string` · required

Sysctl value, e.g. "1024".

- rule: {"required":true}

### spec.podSecurityContext.seccompProfile

`WorkloadSeccompProfile`

Pod-wide seccomp profile; containers may override with their own.

- rule: localhost_profile is required when type is "Localhost" and must be empty otherwise

### spec.podSecurityContext.seccompProfile.type

`string` · required

Profile type: "RuntimeDefault" (the container runtime's default filter — the
recommended baseline), "Unconfined" (no filtering), or "Localhost" (a profile
file installed on the node, named via localhost_profile).

- rule: Seccomp profile type must be one of "RuntimeDefault", "Unconfined", or "Localhost"
- rule: {"required":true}

### spec.podSecurityContext.seccompProfile.localhostProfile

`string`

Path of the profile file relative to the node's seccomp profile root. Required
when (and only meaningful when) type is "Localhost".

### spec.serviceMonitorEnabled

`bool`

When true, the operator creates a monitor for the collector's own
metrics, so Prometheus sees what the collector is doing: records
accepted and sent, sends that failed, and how full its sending queue
is (the first sign that a backend is refusing or away). A
ServiceMonitor on the collector's monitoring Service (port 8888) in
deployment, daemonset and statefulset modes; a PodMonitor in sidecar
mode. The collector serves these metrics either way; this only tells
Prometheus to read them. Needs the Prometheus operator's CRDs on the
cluster when the OTel operator starts: it looks for them once, at
start, so install the monitoring stack first (KubernetesOtelOperator
`depends_on` the KubernetesKubePrometheusStack). Default false.

## Validation Rules

- `spec.mode.sidecar.excludes_scheduling`: sidecar mode injects the collector into target pods — tolerations and priority_class_name have no standalone pod to apply to (the CRD rejects them at admission)
- `spec.mode.daemonset_sidecar.exclude_scaling`: replicas and autoscaler apply only to deployment/statefulset modes — a daemonset runs one collector per node and a sidecar runs inside the target pods
- `spec.autoscaler.excludes_replicas`: autoscaler manages the replica count — a non-default replicas must not be set with it

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesOtelCollector, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the collector runs in. |
| `status.outputs.collector_name` | `string` | name of the OpenTelemetryCollector custom resource (= metadata.name). The operator derives every child name from it. |
| `status.outputs.service` | `string` | name of the collector Service (`<name>-collector`) the operator creates, carrying the ports derived from the declared receivers. Empty in sidecar mode (no standalone workload exists). |
| `status.outputs.otlp_grpc_endpoint` | `string` | in-cluster OTLP gRPC ingest endpoint (`<service>:4317`) — where applications send telemetry. Valid when the config declares the standard `otlp` receiver (every preset does). Empty in sidecar mode. |
| `status.outputs.otlp_http_endpoint` | `string` | in-cluster OTLP HTTP ingest endpoint (`http://<service>:4318`). Valid when the config declares the standard `otlp` receiver. Empty in sidecar mode. |
| `status.outputs.headless_service` | `string` | name of the headless Service (`<name>-collector-headless`) — per-pod addressing for load-balancing-aware gRPC clients. Empty in sidecar mode. |
| `status.outputs.monitoring_service` | `string` | name of the monitoring Service (`<name>-collector-monitoring`, port 8888) — the collector's own metrics. Empty in sidecar mode. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesPlantonPlatform | `spec.observability.otlpHttpEndpoint` | `status.outputs.otlp_http_endpoint` |

## See Also

- [Overview](../README.md)
