# Kubernetes Ray Cluster

## When NOT to Use This

**One resource is ONE Ray cluster** — the declaration of the
RayCluster CR (`rayclusters.ray.io/v1`) that the KubeRay operator
reconciles into a head pod, worker groups, and the Services that Ray
clients, jobs, and dashboards connect to.

Not the right kind when:

- **The operator is missing** — a KubernetesKubeRayOperator whose
  watch scope covers this namespace (cluster-wide with the operator's
  defaults) is the PREREQUISITE. Nothing reconciles this declaration
  without it.
- **You want the operator** — that is KubernetesKubeRayOperator; this
  resource declares one cluster against it.

## The state truth, read this first

The head pod's GCS (Global Control Store) holds the cluster's control
state IN MEMORY. Without GCS fault tolerance, deleting or losing the
HEAD pod loses every job, actor, and worker registration — the
operator rebuilds an empty cluster. Enable `gcs_fault_tolerance` to
put that state in an external Redis-protocol store instead, so a
replaced head RECOVERS the cluster: compose a KubernetesValkey — the
foreign-key defaults wire its endpoint and its auth Secret. The
credential rides a secretKeyRef, readable only from this same
namespace: co-locate the cluster with its Valkey or replicate the
Secret. Worker pods are always expendable.

## Version and image move in lockstep

`ray_version` must match the Ray inside `image` — the operator reads
the version to shape its commands but runs the image as given; a
mismatch fails at runtime, not at apply. This spec derives the image
from `ray_version` by default (`rayproject/ray:<ray_version>`), so
the lockstep cannot drift unless you override `image` deliberately —
custom images (your dependencies baked in, CUDA variants like
`rayproject/ray:2.52.0-gpu`) must keep the Ray inside identical to
the declared version.

## Token auth is the default — because an open Ray cluster runs your code

With `auth.mode` empty or `token` (the catalog default, Ray ≥ 2.52 /
KubeRay ≥ 1.6), the operator provisions a bearer-token Secret (key
`auth_token`, exported as the `auth_token_secret` credential handle)
and every API surface — dashboard, job submission, client — requires
it. The operator's own default is DISABLED, so the modules render the
auth block ALWAYS rather than leaving it to the CR default: an absent
block would silently deploy the legacy open cluster, where anyone who
can reach the dashboard port (8265) runs arbitrary code. Set
`disabled` only for fenced lab clusters.

## The unloaded head is the production posture

With `head.schedule_tasks_on_head` empty (= false), the modules
render `num-cpus: "0"` so the Ray scheduler keeps application work
off the head — a task-loaded head starves the GCS and destabilizes
the whole cluster. The head is REQUIRED and its resources are
required too: an unsized head is the classic Ray outage (GCS,
dashboard, and scheduler share it; upstream guidance starts
production heads at 4 CPU / 8Gi and scales with cluster size). Set
`schedule_tasks_on_head: true` only for single-node labs where the
head is the only capacity — and pin production heads to stable nodes,
because without GCS fault tolerance losing the head's node loses the
cluster's state.

## Accelerators ride `extra_resource_limits` — limits only

Worker groups declare accelerators through `extra_resource_limits`
(`nvidia.com/gpu: "1"`, `amd.com/gpu`, TPU keys), which land in the
container LIMITS only: Kubernetes rejects requests-without-limits for
extended resources, and Ray discovers accelerators from the container
limits. Pair the group's node selector and tolerations so the pods
land on accelerator nodes. With autoscaling enabled, a GPU group can
start at zero and materialize only when a task asks for GPUs — Ray's
own autoscaler scales each group between its bounds on the
scheduler's demand (queued tasks and actors — application-aware,
unlike HPA); `replicas` is the initial size only.

## The 40-character name budget

Keep this resource's name at 40 characters or fewer: the operator
derives `<name>-head-svc` and per-group worker pod names
(`<name>-<group>-worker-…`), and Kubernetes caps names at 63. Both
modules fail loudly past the budget; keep worker group names short
for the same reason.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to deploy into — must be inside the
  operator's watch scope (literal or a KubernetesNamespace reference)
- **`spec.ray_version`**: the Ray version (e.g. `2.52.0` — the
  version KubeRay v1.6.x ships samples for); drives the default image
- **`spec.head`** with **`head.resources`**: the head node and its
  sizing (see the unloaded-head posture above)

### Common

- **`spec.worker_groups`**: homogeneous sets of worker pods (a CPU
  group and a GPU group is the classic two-group shape); each group's
  `resources` is required — Ray schedules against declared capacity
- **`spec.autoscaling`**: the in-tree autoscaler sidecar
  (`idle_timeout_seconds` default 60, `upscaling_mode`
  Default/Aggressive/Conservative)
- **`spec.gcs_fault_tolerance`**: the head-loss survival arm (see the
  state truth above); `external_storage_namespace` lets several
  clusters share one store — empty is safe (derived from the CR's
  UID); set it only when state must survive delete-and-recreate of
  the CR itself
- **`spec.auth`**: `token` (default) or `disabled`;
  `existing_token_secret_name` brings your own token Secret (key
  `auth_token`)
- **`spec.image`**: custom image, version-locked to `ray_version`
- **`spec.suspend`**: the idle-GPU cost lever — the operator deletes
  head and worker PODS but keeps the declaration and (with GCS fault
  tolerance) the external state; un-suspend to resume

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `head_service` | The head Service (`<name>-head-svc`, the operator's naming contract) — every endpoint below rides it |
| `client_endpoint` | In-cluster CLIENT endpoint (port 10001) — what `ray.init("ray://…")` dials from notebooks and applications |
| `dashboard_endpoint` | In-cluster DASHBOARD/Job-API endpoint (port 8265) — the Job Submission API and the web dashboard; authenticated in token mode |
| `gcs_endpoint` | In-cluster GCS endpoint (port 6379) — Ray's own coordination port; what `ray start --address` joins |
| `auth_token_secret` | The bearer-token Secret (key `auth_token`) in token auth mode — the credential handle; unset when auth is disabled |
| `port_forward_command` | kubectl port-forward one-liner for reaching the dashboard from a workstation |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **KubernetesKubeRayOperator is the prerequisite** — its watch scope
  must cover this namespace.
- **`gcs_fault_tolerance.redis_address`** is a foreign key (default
  kind KubernetesValkey, field path `status.outputs.kube_endpoint` —
  its write Service endpoint, always the primary), and
  `redis_password_secret.name` defaults to the same Valkey's
  module-materialized auth Secret (the key is the ACL username —
  `default` unless you declared users). Deploy the Valkey first, in
  this same namespace — the credential secretKeyRef cannot cross
  namespaces.
- **The apply is deliberately non-blocking**: cluster readiness
  depends on the operator (multi-GB Ray image pulls, autoscaler
  sidecar injection, GCS startup) — the CR applies and the operator
  reconciles; nothing blocks on a controller.

## Examples

### Lab: single node, tasks on the head

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRayCluster
metadata:
  name: ray-lab
spec:
  namespace:
    value: ml-lab
  createNamespace: true
  rayVersion: 2.52.0
  head:
    resources:
      requests:
        cpu: "1"
        memory: 4Gi
      limits:
        cpu: "2"
        memory: 8Gi
    scheduleTasksOnHead: true
```

### Production: autoscaling CPU + GPU groups, GCS fault tolerance

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesRayCluster
metadata:
  name: ml-ray
spec:
  namespace:
    value: ml-platform
  createNamespace: true
  rayVersion: 2.52.0
  head:
    resources:
      requests:
        cpu: "2"
        memory: 8Gi
      limits:
        cpu: "4"
        memory: 16Gi
    scheduleTasksOnHead: false
  workerGroups:
    - name: cpu
      replicas: 2
      minReplicas: 0
      maxReplicas: 10
      resources:
        requests:
          cpu: "2"
          memory: 8Gi
        limits:
          cpu: "4"
          memory: 16Gi
    - name: gpu
      replicas: 0
      minReplicas: 0
      maxReplicas: 4
      resources:
        requests:
          cpu: "4"
          memory: 24Gi
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
    idleTimeoutSeconds: 300
  gcsFaultTolerance:
    enabled: true
    redisAddress:
      valueFrom:
        name: ray-state
    redisPasswordSecret:
      name:
        valueFrom:
          name: ray-state
      key: default
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
