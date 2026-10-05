# KubernetesRayCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesRayClusterSpec** declares a Ray cluster — the Kubernetes
CR (`rayclusters.ray.io/v1`) that the KubeRay operator reconciles
into a head pod, worker groups, and the Services that Ray clients,
jobs, and dashboards connect to. PREREQUISITE: a
KubernetesKubeRayOperator whose watch scope covers this namespace
(cluster-wide with the operator's defaults).

WHAT A RAY CLUSTER IS: the shared distributed-compute service that
ML and data teams (and their agents) connect to — notebooks and
applications dial the CLIENT port (10001), jobs submit through the
DASHBOARD port (8265, the Job Submission API), and Ray's own
processes coordinate through the GCS port (6379) on the head.

STATE TRUTH, READ THIS FIRST: the head pod's GCS (Global Control
Store) holds the cluster's control state IN MEMORY. Without GCS
fault tolerance, deleting or losing the HEAD pod loses every job,
actor, and worker registration — the operator rebuilds an empty
cluster. Enable `gcs_fault_tolerance` (state in an external
Redis-protocol store — compose a KubernetesValkey) for head-loss
survival; worker pods are always expendable.

VERSION/IMAGE LOCKSTEP: `ray_version` must match the Ray inside
`image` — the operator reads the version to shape its commands but
runs the image as given; a mismatch fails at runtime, not at apply.
This spec derives the image from ray_version by default so the
lockstep cannot drift unless you override image deliberately.

AUTHENTICATION (Ray ≥ 2.52 / KubeRay ≥ 1.6): with `auth.mode`
"token", the operator provisions a bearer token Secret and every
API surface (dashboard, job submission, client) requires it —
exported as the credential handle. "disabled" is the legacy
open-cluster posture: anyone who can reach the dashboard port can
run arbitrary code — never expose it beyond the cluster.

NAME BUDGET: keep this resource's name at 40 characters or fewer —
the operator derives `<name>-head-svc` and per-group pod names
(`<name>-<group>-worker-…`), and Kubernetes names cap at 63. Both
modules fail loudly past the budget; keep worker group names short
for the same reason.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the cluster expressed at once — a custom image,
# a production head posture (tasks off the head, extra ray start
# params), two worker groups (a CPU group with autoscaling bounds and a
# GPU group with extended resource limits and accelerator scheduling),
# the in-tree autoscaler, GCS fault tolerance against an external
# Redis-protocol store with a password Secret, and explicit token auth.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRayCluster
metadata:
  name: ml-ray-full
spec:
  namespace:
    value: ml-platform
  createNamespace: true
  rayVersion: 2.52.0
  image: rayproject/ray:2.52.0
  head:
    resources:
      requests:
        cpu: "1"
        memory: 4Gi
      limits:
        cpu: "2"
        memory: 8Gi
    scheduleTasksOnHead: false
    rayStartParams:
      object-store-memory: "1000000000"
    scheduling:
      nodeSelector:
        workload: ray-head
  workerGroups:
    - name: cpu
      replicas: 2
      minReplicas: 0
      maxReplicas: 10
      resources:
        requests:
          cpu: "1"
          memory: 2Gi
        limits:
          cpu: "2"
          memory: 4Gi
    - name: gpu
      replicas: 1
      minReplicas: 0
      maxReplicas: 4
      resources:
        requests:
          cpu: "4"
          memory: 16Gi
        limits:
          cpu: "8"
          memory: 32Gi
      extraResourceLimits:
        nvidia.com/gpu: "1"
      scheduling:
        nodeSelector:
          accelerator: nvidia
        tolerations:
          - key: nvidia.com/gpu
            operator: Exists
            effect: NoSchedule
  autoscaling:
    enabled: true
    idleTimeoutSeconds: 120
    upscalingMode: Aggressive
  gcsFaultTolerance:
    enabled: true
    redisAddress:
      value: state-valkey.ml-platform.svc.cluster.local:6379
    redisPasswordSecret:
      name:
        value: state-valkey-auth
      key: default
    externalStorageNamespace: ml-ray
  auth:
    mode: token
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.rayVersion` | `string` | yes |  |  |
| `spec.image` | `string` |  |  |  |
| `spec.head` | `KubernetesRayClusterHead` | yes |  |  |
| `spec.head.resources` | `ContainerResources` | yes |  |  |
| `spec.head.resources.limits` | `CpuMemory` |  |  |  |
| `spec.head.resources.limits.cpu` | `string` |  |  |  |
| `spec.head.resources.limits.memory` | `string` |  |  |  |
| `spec.head.resources.requests` | `CpuMemory` |  |  |  |
| `spec.head.resources.requests.cpu` | `string` |  |  |  |
| `spec.head.resources.requests.memory` | `string` |  |  |  |
| `spec.head.scheduleTasksOnHead` | `bool` |  | `false` |  |
| `spec.head.rayStartParams` | `map<string, string>` |  |  |  |
| `spec.head.scheduling` | `KubernetesRayClusterScheduling` |  |  |  |
| `spec.head.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.head.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.head.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.head.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.head.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.head.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.head.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.head.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.workerGroups` | `[]KubernetesRayClusterWorkerGroup` |  |  |  |
| `spec.workerGroups[].name` | `string` | yes |  |  |
| `spec.workerGroups[].replicas` | `int32` |  | `1` |  |
| `spec.workerGroups[].minReplicas` | `int32` |  |  |  |
| `spec.workerGroups[].maxReplicas` | `int32` |  |  |  |
| `spec.workerGroups[].resources` | `ContainerResources` | yes |  |  |
| `spec.workerGroups[].resources.limits` | `CpuMemory` |  |  |  |
| `spec.workerGroups[].resources.limits.cpu` | `string` |  |  |  |
| `spec.workerGroups[].resources.limits.memory` | `string` |  |  |  |
| `spec.workerGroups[].resources.requests` | `CpuMemory` |  |  |  |
| `spec.workerGroups[].resources.requests.cpu` | `string` |  |  |  |
| `spec.workerGroups[].resources.requests.memory` | `string` |  |  |  |
| `spec.workerGroups[].rayStartParams` | `map<string, string>` |  |  |  |
| `spec.workerGroups[].scheduling` | `KubernetesRayClusterScheduling` |  |  |  |
| `spec.workerGroups[].scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.workerGroups[].scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.workerGroups[].scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.workerGroups[].scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.workerGroups[].scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.workerGroups[].scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.workerGroups[].scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.workerGroups[].scheduling.priorityClassName` | `string` |  |  |  |
| `spec.workerGroups[].extraResourceLimits` | `map<string, string>` |  |  |  |
| `spec.autoscaling` | `KubernetesRayClusterAutoscaling` |  |  |  |
| `spec.autoscaling.enabled` | `bool` |  |  |  |
| `spec.autoscaling.idleTimeoutSeconds` | `int32` |  |  |  |
| `spec.autoscaling.upscalingMode` | `string` |  |  |  |
| `spec.autoscaling.resources` | `ContainerResources` |  |  |  |
| `spec.autoscaling.resources.limits` | `CpuMemory` |  |  |  |
| `spec.autoscaling.resources.limits.cpu` | `string` |  |  |  |
| `spec.autoscaling.resources.limits.memory` | `string` |  |  |  |
| `spec.autoscaling.resources.requests` | `CpuMemory` |  |  |  |
| `spec.autoscaling.resources.requests.cpu` | `string` |  |  |  |
| `spec.autoscaling.resources.requests.memory` | `string` |  |  |  |
| `spec.gcsFaultTolerance` | `KubernetesRayClusterGcsFaultTolerance` |  |  |  |
| `spec.gcsFaultTolerance.enabled` | `bool` |  |  |  |
| `spec.gcsFaultTolerance.redisAddress` | `string \| valueFrom` |  |  | KubernetesValkey (`status.outputs.kube_endpoint`) |
| `spec.gcsFaultTolerance.redisPasswordSecret` | `KubernetesRayClusterSecretSelector` |  |  |  |
| `spec.gcsFaultTolerance.redisPasswordSecret.name` | `string \| valueFrom` | yes |  | KubernetesValkey (`status.outputs.password_secret.name`) |
| `spec.gcsFaultTolerance.redisPasswordSecret.key` | `string` | yes |  |  |
| `spec.gcsFaultTolerance.externalStorageNamespace` | `string` |  |  |  |
| `spec.auth` | `KubernetesRayClusterAuth` |  |  |  |
| `spec.auth.mode` | `string` |  | `token` |  |
| `spec.auth.existingTokenSecretName` | `string` |  |  |  |
| `spec.suspend` | `bool` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to deploy into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. Must be inside the
KubernetesKubeRayOperator's watch scope. NOTE the GCS
fault-tolerance credential Secrets ride secretKeyRefs, readable
only from this same namespace — co-locate the cluster with its
Valkey or replicate the Secret.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the CR and deleted with the resource.
When false, the namespace must already exist.

### spec.rayVersion

`string` · required

Ray version running in the cluster (e.g. "2.52.0" — the version
KubeRay v1.6.x ships samples for). Drives the default image
(`rayproject/ray:<ray_version>`) and the operator's command
shaping. Must match the Ray inside a custom image exactly.

- rule: {"required":true}

### spec.image

`string`

Container image for every Ray node. Empty = `rayproject/ray:
<ray_version>` — the official image, version-locked to
ray_version by construction. Set for custom images (your
dependencies baked in; CUDA variants like
`rayproject/ray:2.52.0-gpu`) and keep the Ray inside identical
to ray_version.

### spec.head

`KubernetesRayClusterHead` · required

The head node — REQUIRED. Runs the GCS, the dashboard, the job
submission server, and the autoscaler sidecar when autoscaling
is on.

- rule: {"required":true}

### spec.head.resources

`ContainerResources` · required

CPU and memory for the head container. REQUIRED: an unsized head
is the classic Ray outage (GCS + dashboard + scheduler share it;
upstream guidance starts production heads at 4 CPU / 8Gi —
scale with cluster size; labs run smaller consciously).

- rule: {"required":true}

### spec.head.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.head.resources.limits.cpu

`string`

### spec.head.resources.limits.memory

`string`

### spec.head.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.head.resources.requests.cpu

`string`

### spec.head.resources.requests.memory

`string`

### spec.head.scheduleTasksOnHead

`bool` · optional (explicit presence)

Schedule Ray TASKS onto the head? Empty = false: the modules
render `num-cpus: "0"` so the scheduler keeps application work
off the head — the production posture (a task-loaded head
starves the GCS and destabilizes the whole cluster). Set true
for single-node labs where the head is the only capacity.

- default: `false`

### spec.head.rayStartParams

`map<string, string>`

Extra `ray start` parameters for the head, by flag name without
the leading dashes (e.g. "object-store-memory":
"1000000000"). The modules own num-cpus (from
schedule_tasks_on_head) and dashboard-host — entries here must
not repeat them.

### spec.head.scheduling

`KubernetesRayClusterScheduling`

Scheduling for the head pod. Pin the head to stable capacity —
without GCS fault tolerance, losing its node loses the cluster's
state (see the spec-level STATE TRUTH).

### spec.head.scheduling.nodeSelector

`map<string, string>`

Node selector for the pods.

### spec.head.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes (GPU taints live here).

### spec.head.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.head.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.head.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.head.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.head.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.head.scheduling.priorityClassName

`string`

Priority class name for the pods.

### spec.workerGroups

`[]KubernetesRayClusterWorkerGroup`

Worker groups — homogeneous sets of worker pods (same resources,
same Ray start parameters). A CPU group and a GPU group is the
classic two-group shape. A cluster with no worker groups is
legal: tasks run on the head (labs only — production heads
should stay unloaded).

- rule: Worker group sizing must order min_replicas <= replicas <= max_replicas.

### spec.workerGroups[].name

`string` · required

Group name (lowercase DNS label; kept short — it lands inside
every worker pod name).

- rule: {"required":true,"string":{"pattern":"^[a-z0-9]([-a-z0-9]{0,22}[a-z0-9])?$"}}

### spec.workerGroups[].replicas

`int32` · optional (explicit presence)

Worker pods in the group. With autoscaling enabled this is the
INITIAL size only — the autoscaler owns it afterwards (between
min_replicas and max_replicas).

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.workerGroups[].minReplicas

`int32` · optional (explicit presence)

Autoscaling floor for the group. Empty = 0 (the group can scale
to nothing when idle).

- rule: {"int32":{"gte":0}}

### spec.workerGroups[].maxReplicas

`int32` · optional (explicit presence)

Autoscaling ceiling for the group. Empty = replicas (a fixed
group). Meaningful only with spec.autoscaling enabled.

- rule: {"int32":{"gte":0}}

### spec.workerGroups[].resources

`ContainerResources` · required

CPU and memory for each worker container — REQUIRED (Ray
schedules against declared capacity; unsized workers make its
scheduler blind). Accelerators are declared separately in
extra_resource_limits.

- rule: {"required":true}

### spec.workerGroups[].resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.workerGroups[].resources.limits.cpu

`string`

### spec.workerGroups[].resources.limits.memory

`string`

### spec.workerGroups[].resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.workerGroups[].resources.requests.cpu

`string`

### spec.workerGroups[].resources.requests.memory

`string`

### spec.workerGroups[].rayStartParams

`map<string, string>`

Extra `ray start` parameters for this group's workers, by flag
name without the leading dashes.

### spec.workerGroups[].scheduling

`KubernetesRayClusterScheduling`

Scheduling for this group's pods (GPU node selectors and
tolerations live here).

### spec.workerGroups[].scheduling.nodeSelector

`map<string, string>`

Node selector for the pods.

### spec.workerGroups[].scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes (GPU taints live here).

### spec.workerGroups[].scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.workerGroups[].scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.workerGroups[].scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.workerGroups[].scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.workerGroups[].scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.workerGroups[].scheduling.priorityClassName

`string`

Priority class name for the pods.

### spec.workerGroups[].extraResourceLimits

`map<string, string>`

Extended resource limits for each worker container, by resource
key — the accelerator surface ("nvidia.com/gpu": "1",
"amd.com/gpu", TPU keys). Ray discovers accelerators from these
container limits and schedules tasks against them; pair with the
group's node selector/tolerations so the pods land on
accelerator nodes.

### spec.autoscaling

`KubernetesRayClusterAutoscaling`

Ray's own autoscaler: the operator injects an autoscaler sidecar
into the head pod that scales each worker group between its
min_replicas and max_replicas based on the Ray scheduler's
demand (queued tasks/actors — application-aware, unlike HPA).
Worker `replicas` become the initial size; the autoscaler owns
them afterwards.

### spec.autoscaling.enabled

`bool`

Inject the autoscaler sidecar into the head pod.

### spec.autoscaling.idleTimeoutSeconds

`int32` · optional (explicit presence)

Seconds a worker stays idle before the autoscaler reclaims it.
Empty = 60 (the upstream default).

- rule: {"int32":{"gt":0}}

### spec.autoscaling.upscalingMode

`string`

How aggressively to add capacity: "Default" (empty — respond to
demand as it arrives), "Aggressive" (upscale speculatively), or
"Conservative" (rate-limited upscaling).

- rule: upscaling_mode must be "Default", "Aggressive", "Conservative", or empty.

### spec.autoscaling.resources

`ContainerResources`

CPU and memory for the autoscaler sidecar container. Empty = the
operator defaults.

### spec.autoscaling.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.autoscaling.resources.limits.cpu

`string`

### spec.autoscaling.resources.limits.memory

`string`

### spec.autoscaling.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.autoscaling.resources.requests.cpu

`string`

### spec.autoscaling.resources.requests.memory

`string`

### spec.gcsFaultTolerance

`KubernetesRayClusterGcsFaultTolerance`

GCS fault tolerance: the head's control state lives in an
external Redis-protocol store instead of head-pod memory, so a
replaced head RECOVERS the cluster (jobs, actors, workers)
instead of rebuilding an empty one. Compose a KubernetesValkey —
the FK defaults wire its endpoint and credentials.

- rule: GCS fault tolerance needs the external store's endpoint — set redis_address (host:port, or a KubernetesValkey reference).

### spec.gcsFaultTolerance.enabled

`bool`

Enable GCS fault tolerance.

### spec.gcsFaultTolerance.redisAddress

`string | valueFrom`

Redis-protocol endpoint as `host:port`. Accepts a literal or a
reference to a KubernetesValkey resource (its write Service
endpoint — always the primary).

- references: KubernetesValkey (`status.outputs.kube_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.kube_endpoint}} -- a bare string does not parse

### spec.gcsFaultTolerance.redisPasswordSecret

`KubernetesRayClusterSecretSelector`

Password for the store, read from a Secret. On a
KubernetesValkey: its module-materialized `<name>-auth` Secret
(the FK default wires the name; the key is the ACL username —
"default" unless you declared users). Leave unset only for an
auth-less store.

### spec.gcsFaultTolerance.redisPasswordSecret.name

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesValkey resource (its `<name>-auth` Secret).

- references: KubernetesValkey (`status.outputs.password_secret.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesValkey, name: <that resource's name>, fieldPath: status.outputs.password_secret.name}} -- a bare string does not parse

### spec.gcsFaultTolerance.redisPasswordSecret.key

`string` · required

Key within the Secret (on a KubernetesValkey auth Secret: the
ACL username — "default" unless you declared users).

- rule: {"required":true}

### spec.gcsFaultTolerance.externalStorageNamespace

`string`

Storage namespace inside the store — lets several Ray clusters
share one Valkey without key collisions. Empty = the operator
derives one from the cluster's UID (safe; set it explicitly only
when you need the state to survive delete-and-recreate of the
CR itself).

### spec.auth

`KubernetesRayClusterAuth`

API authentication. Empty = token mode ON (the secure default
for this catalog): the operator provisions a bearer-token Secret
(exported as the credential handle) and the dashboard, job
submission, and client surfaces all require it. Set mode
"disabled" ONLY for fenced lab clusters — an open Ray cluster
executes arbitrary code for anyone who can reach port 8265.

### spec.auth.mode

`string` · optional (explicit presence)

"token" (empty default): the operator provisions a bearer-token
Secret NAMED EXACTLY AFTER THIS RESOURCE (key `auth_token` —
the operator's own naming, source-verified; exported as the
credential handle) and every API surface requires it.
"disabled": the legacy open cluster — anyone reaching the
dashboard port runs arbitrary code; fenced labs only.

- default: `token`
- rule: Auth mode must be "token" or "disabled".

### spec.auth.existingTokenSecretName

`string`

Bring-your-own token: name of an existing Secret with the token
under key `auth_token`. Empty = the operator generates one per
cluster (the normal path).

### spec.suspend

`bool`

Suspend the cluster: the operator deletes head and worker PODS
but keeps the declaration and (with GCS fault tolerance) the
external state. Un-suspend to resume. The idle-GPU cost lever.

## Validation Rules

- `spec.worker_groups.unique_names`: Worker group names must be unique — the operator derives pod names per group and duplicate groups collide.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesRayCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the cluster runs in. |
| `status.outputs.head_service` | `string` | the head Service name (`<name>-head-svc`) — every endpoint below rides it. |
| `status.outputs.client_endpoint` | `string` | in-cluster CLIENT endpoint (`<head_service>.<namespace>.svc.cluster. local:10001`) — what ray.init("ray://…") dials from notebooks and applications. |
| `status.outputs.dashboard_endpoint` | `string` | in-cluster DASHBOARD/Job-API endpoint (`…:8265`) — the Job Submission API and the web dashboard. Authenticated in token mode. |
| `status.outputs.gcs_endpoint` | `string` | in-cluster GCS endpoint (`…:6379`) — Ray's own coordination port; what `ray start --address` joins. |
| `status.outputs.auth_token_secret` | `KubernetesSecretKey` | the bearer-token Secret (key `auth_token`) in token auth mode — the credential handle for the dashboard/job/client APIs. Unset when auth is disabled. |
| `status.outputs.auth_token_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.auth_token_secret.key` | `string` | The key within the Kubernetes Secret. |
| `status.outputs.port_forward_command` | `string` | kubectl port-forward one-liner for reaching the dashboard from a workstation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.gcsFaultTolerance.redisAddress` | KubernetesValkey | `status.outputs.kube_endpoint` |
| `spec.gcsFaultTolerance.redisPasswordSecret.name` | KubernetesValkey | `status.outputs.password_secret.name` |

## See Also

- [Overview](../README.md)
