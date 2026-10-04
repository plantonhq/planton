# KubernetesGhaRunnerScaleSetController

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesGhaRunnerScaleSetControllerSpec** installs the GitHub
Actions Runner Scale Set controller — GitHub's official
actions-runner-controller (ARC) manager — from the official
`gha-runner-scale-set-controller` chart (OCI,
ghcr.io/actions/actions-runner-controller-charts).

This component installs the ENGINE. Runner fleets themselves are
declared with KubernetesGhaRunnerScaleSet resources — one per
repository/organization/enterprise registration — which this
controller reconciles into listener pods and ephemeral runner pods.

ONE CONTROLLER PER CLUSTER is the sane default: the controller
watches all namespaces, so every runner scale set on the cluster is
served by it. Fence the watch with `flags.watch_single_namespace`
only for hard multi-tenancy — each fenced controller then needs its
runner scale sets to reference it explicitly (the scale set kind's
`controller_service_account` field).

CRD LIFECYCLE: the chart ships the actions.github.com CRDs
(AutoscalingRunnerSet, EphemeralRunner, ...) in its crds/ directory —
Helm installs them on first install and NEVER removes them:
destroying the controller keeps the CRDs (and any declared runner
scale sets' objects) on the cluster (verified live). Runner scale
sets stop reconciling without a controller, so still destroy
KubernetesGhaRunnerScaleSet resources first; a later controller
install adopts the kept CRDs cleanly (they carry no release
ownership metadata).

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# typed block rendered at once (image override, all flags incl. the
# structured rate limiter, metrics, scheduling, escape hatch).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-full
spec:
  namespace:
    value: arc-system
  createNamespace: true
  chartVersion: 0.14.2
  replicas: 2
  image:
    repo: mirror.example.com/gha-rs-controller
    tag: 0.14.2
    pullSecretName: mirror-pull
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: "1"
      memory: 512Mi
  flags:
    logLevel: info
    logFormat: json
    watchSingleNamespace: ci-runners
    runnerMaxConcurrentReconciles: 4
    updateStrategy: eventual
    excludeLabelPropagationPrefixes:
      - argocd.argoproj.io/instance
    k8sClientRateLimiterQps: 30
    k8sClientRateLimiterBurst: 60
    rateLimiter: typed_rate_limiter
    healthProbeBindAddress: ":8081"
    priorityClassName: system-cluster-critical
  metrics:
    controllerManagerAddr: ":8080"
    listenerAddr: ":8080"
    listenerEndpoint: /metrics
  imagePullSecrets:
    - extra-pull
  scheduling:
    nodeSelector:
      role: platform
    tolerations:
      - key: platform
        operator: Exists
        effect: NoSchedule
  helmValues: |
    podLabels:
      team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.14.2` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.image` | `ContainerImage` |  |  |  |
| `spec.image.repo` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.pullSecretName` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.flags` | `KubernetesGhaRunnerScaleSetControllerFlags` |  |  |  |
| `spec.flags.logLevel` | `string` |  |  |  |
| `spec.flags.logFormat` | `string` |  |  |  |
| `spec.flags.watchSingleNamespace` | `string` |  |  |  |
| `spec.flags.runnerMaxConcurrentReconciles` | `int32` |  | `2` |  |
| `spec.flags.updateStrategy` | `string` |  |  |  |
| `spec.flags.excludeLabelPropagationPrefixes` | `[]string` |  |  |  |
| `spec.flags.k8sClientRateLimiterQps` | `int32` |  |  |  |
| `spec.flags.k8sClientRateLimiterBurst` | `int32` |  |  |  |
| `spec.flags.rateLimiter` | `string` |  |  |  |
| `spec.flags.healthProbeBindAddress` | `string` |  |  |  |
| `spec.flags.priorityClassName` | `string` |  |  |  |
| `spec.metrics` | `KubernetesGhaRunnerScaleSetControllerMetrics` |  |  |  |
| `spec.metrics.controllerManagerAddr` | `string` | yes |  |  |
| `spec.metrics.listenerAddr` | `string` | yes |  |  |
| `spec.metrics.listenerEndpoint` | `string` | yes |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesGhaRunnerScaleSetControllerScheduling` |  |  |  |
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
reference to a KubernetesNamespace resource.

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

Helm chart version to install (e.g. "0.14.2" — chart and
controller image move in lockstep). Versions must exist as SERVED
charts in the OCI registry
(ghcr.io/actions/actions-runner-controller-charts).

- default: `0.14.2`

### spec.replicas

`int32` · optional (explicit presence)

Number of controller replicas. Empty = 1. With more than one, the
chart enables leader election automatically — extra replicas are
hot standbys, not workload shards.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.image

`ContainerImage`

Override the controller image (air-gap / private-mirror path).
Empty = ghcr.io/actions/gha-runner-scale-set-controller at the
chart's appVersion.

### spec.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.resources

`ContainerResources`

CPU and memory for the controller container. Empty = no
requests/limits (the chart default).

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

### spec.flags

`KubernetesGhaRunnerScaleSetControllerFlags`

Controller behavior flags.

### spec.flags.logLevel

`string` · optional (explicit presence)

Log level. Empty = "debug" (the chart default — noisy; production
clusters usually set "info").

- rule: {"string":{"in":["","debug","info","warn","error"]}}

### spec.flags.logFormat

`string` · optional (explicit presence)

Log format. Empty = "text" (the chart default).

- rule: {"string":{"in":["","text","json"]}}

### spec.flags.watchSingleNamespace

`string`

Restrict the controller to watch ONE namespace instead of the
whole cluster. Runner scale sets outside it are ignored — and
each scale set install must then name this controller's
ServiceAccount explicitly (the scale set kind's
`controller_service_account`). Empty = watch all namespaces.

- rule: {"string":{"pattern":"^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.flags.runnerMaxConcurrentReconciles

`int32` · optional (explicit presence)

Maximum concurrent reconciles of the EphemeralRunner controller.
Empty = 2 (the chart default). Raising it improves runner
throughput at the cost of API-server and GitHub-API load.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.flags.updateStrategy

`string` · optional (explicit presence)

How upgrades behave while jobs run: `immediate` (the chart
default — recreate listeners at once; may briefly overprovision
runners) or `eventual` (tear down, then wait for running jobs to
finish before recreating — no overprovisioning, slower rollout).

- rule: {"string":{"in":["","immediate","eventual"]}}

### spec.flags.excludeLabelPropagationPrefixes

`[]string`

Label PREFIXES excluded from propagation to listener/runner
resources (e.g. "argocd.argoproj.io/instance" when a GitOps tool
labels the install and must not see those labels come back on
children).

### spec.flags.k8sClientRateLimiterQps

`int32` · optional (explicit presence)

Kubernetes API client rate-limiter QPS for the controller. Empty
= the controller default.

- rule: {"int32":{"gte":1}}

### spec.flags.k8sClientRateLimiterBurst

`int32` · optional (explicit presence)

Kubernetes API client rate-limiter burst for the controller.
Empty = the controller default.

- rule: {"int32":{"gte":1}}

### spec.flags.rateLimiter

`string` · optional (explicit presence)

Workqueue rate limiter: `bucket_rate_limiter` (the default —
per-item exponential backoff plus a global token bucket) or
`typed_rate_limiter` (per-item backoff only).

- rule: {"string":{"in":["","bucket_rate_limiter","typed_rate_limiter"]}}

### spec.flags.healthProbeBindAddress

`string`

Bind address for the health-probe endpoint (e.g. ":8081"). When
set, the chart adds liveness/readiness probes to the controller
pod. Empty = probes disabled (the chart default).

### spec.flags.priorityClassName

`string`

Priority class name for the controller pod. Use
"system-cluster-critical" so the runner control plane survives
node-pressure evictions.

### spec.metrics

`KubernetesGhaRunnerScaleSetControllerMetrics`

Prometheus metrics for the controller manager AND the listener
pods it creates. Empty = metrics disabled (the chart default —
the chart passes empty metrics flags).

### spec.metrics.controllerManagerAddr

`string` · required

Metrics bind address of the controller manager (e.g. ":8080").

- rule: {"required":true}

### spec.metrics.listenerAddr

`string` · required

Metrics bind address of each listener pod (e.g. ":8080").

- rule: {"required":true}

### spec.metrics.listenerEndpoint

`string` · required

Metrics HTTP path on the listener (e.g. "/metrics").

- rule: {"required":true}

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the install namespace) for
pulling the controller image from a private mirror. Also passed
to listener pods for pulling the listener image.

### spec.scheduling

`KubernetesGhaRunnerScaleSetControllerScheduling`

Scheduling for the controller pod.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the controller pod.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the controller pod.

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

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (env entries, volumes, pprof, security
contexts, ...) — never the substitute for them. Do not put
secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGhaRunnerScaleSetController, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the controller runs in. |
| `status.outputs.release_name` | `string` | Helm release name (equals metadata.name). |
| `status.outputs.service_account_name` | `string` | Name of the controller's ServiceAccount — what a KubernetesGhaRunnerScaleSet references in `controller_service_account` when this controller watches a single namespace. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
