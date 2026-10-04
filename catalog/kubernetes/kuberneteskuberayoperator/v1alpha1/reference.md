# KubernetesKubeRayOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesKubeRayOperatorSpec** deploys the KubeRay operator —
the controller that turns KubernetesRayCluster declarations (and,
unmodeled here, RayJob/RayService CRs authored directly) into
running Ray clusters — from the official `kuberay-operator` Helm
chart (https://ray-project.github.io/kuberay-helm).

GRAIN: one operator per cluster is the normal posture (it watches
every namespace by default and runs leader election out of the
box). Ray clusters themselves are the many-per-cluster workload —
declare them as KubernetesRayCluster resources.

CRD LIFECYCLE: the chart ships its three ray.io CRDs (rayclusters,
rayjobs, rayservices) in its `crds/` directory — Helm installs
them once, NEVER upgrades them on chart upgrades, and LEAVES them
(and every Ray declaration) on uninstall. Apply the new release's
CRD files manually when a chart bump changes them. NOTE the CRDs
are large (~1MB each): they install server-side.

NO ADMISSION WEBHOOK: the operator validates in its reconcile
loop — no webhook, no certificate machinery, no cert-manager
dependency. A bad RayCluster surfaces on the CR's status
conditions, not as an admission rejection.

NAMING: the chart hardcodes its fullname, name, AND service
account to "kuberay-operator" by default; the modules re-pin all
three to this resource's name so instances stay distinguishable.
Keep the name at 47 characters or fewer — the longest derived
child name suffixes "-leader-election" (16 chars) and Kubernetes
names cap at 63. Both engines fail loudly over the budget.

The typed fields cover the chart's meaningful surface;
`helm_values` remains the escape hatch (merged last, Helm `-f`
semantics, identical on both engines) — operator env-var feature
flags, single-namespace RBAC shapes, log-file mounts — a safety
valve, never the primary interface.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the operator expressed at once — the fenced
# watch posture, an explicit leader-election restatement, a batch
# scheduler, feature-gate overrides (exercising the list-replacement
# merge over the chart defaults), metrics + ServiceMonitor, resources,
# the image-registry mirror, and an escape-hatch document (logging
# encoder) — while the module's fullname re-pin stays assertable in the
# rendered plan.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKubeRayOperator
metadata:
  name: kuberay-operator-full
spec:
  namespace:
    value: ray-system
  createNamespace: true
  chartVersion: 1.6.2
  watchNamespaces:
    - ml-team-a
    - ml-team-b
  leaderElectionEnabled: true
  batchScheduler: yunikorn
  featureGates:
    - name: RayServiceIncrementalUpgrade
      enabled: true
  metricsEnabled: true
  serviceMonitorEnabled: true
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 1Gi
  imageRegistry: mirror.example.com
  helmValues: |
    logging:
      stdoutEncoder: console
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.6.2` |  |
| `spec.watchNamespaces` | `[]string` |  |  |  |
| `spec.leaderElectionEnabled` | `bool` |  | `true` |  |
| `spec.batchScheduler` | `string` |  |  |  |
| `spec.featureGates` | `[]KubernetesKubeRayOperatorFeatureGate` |  |  |  |
| `spec.featureGates[].name` | `string` | yes |  |  |
| `spec.featureGates[].enabled` | `bool` |  |  |  |
| `spec.metricsEnabled` | `bool` |  | `true` |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install the operator into. Accepts a literal
namespace name or a reference to a KubernetesNamespace resource.

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

Helm chart version to install (e.g. "1.6.2" — chart 1.6.2 pairs
with operator v1.6.2). Versions must exist as SERVED charts in
the repository index
(https://ray-project.github.io/kuberay-helm). NOTE chart
upgrades never upgrade the ray.io CRDs (`crds/` directory
posture) — apply new CRD files manually when a bump changes
them.

- default: `1.6.2`

### spec.watchNamespaces

`[]string`

Namespaces the operator watches for Ray CRs. Empty = every
namespace (the normal one-operator-per-cluster posture). With a
list set, Ray declarations OUTSIDE these namespaces are ignored
without an error — the fenced multi-tenant posture. The chart
scopes the per-namespace reconcile RBAC to the same list.

### spec.leaderElectionEnabled

`bool` · optional (explicit presence)

Leader election. Empty = true (the chart default — safe for
single replicas and required for standbys). Disable only in
constrained RBAC environments that cannot grant lease
permissions.

- default: `true`

### spec.batchScheduler

`string`

Integrate a batch scheduler for gang-scheduling Ray pods:
"volcano", "yunikorn", or "scheduler-plugins". Empty = none (the
default kube-scheduler places pods individually). The named
scheduler must already run on the cluster — the operator only
emits its scheduling directives.

- rule: batch_scheduler must be "volcano", "yunikorn", "scheduler-plugins", or empty (default scheduler).

### spec.featureGates

`[]KubernetesKubeRayOperatorFeatureGate`

Operator feature gates, exactly as the pinned release names them.
At chart 1.6.2 the defaults are: RayClusterStatusConditions,
RayJobDeletionPolicy, and RayMultiHostIndexing ON;
RayServiceIncrementalUpgrade and RayCronJob OFF. Only list gates
you are deliberately flipping — the modules merge your entries
over the full default set (Helm would otherwise REPLACE the list
and silently flip unlisted defaults).

### spec.featureGates[].name

`string` · required

Gate name as the pinned release documents it.

- rule: {"required":true}

### spec.featureGates[].enabled

`bool`

Desired state.

### spec.metricsEnabled

`bool` · optional (explicit presence)

Emit operator control-plane metrics (port 8080). Empty = true
(the chart default).

- default: `true`

### spec.serviceMonitorEnabled

`bool`

Render a ServiceMonitor for Prometheus discovery of the
operator's metrics. Requires the monitoring.coreos.com CRDs on
the cluster (deploy KubernetesKubePrometheusStack first) — the
install FAILS without them.

### spec.resources

`ContainerResources`

CPU and memory for the operator container. Empty = the chart
defaults (100m CPU / 512Mi limits — upstream sizes ~500MB per
500 managed Ray pods; scale memory with fleet size).

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

### spec.imageRegistry

`string`

Registry that replaces the registry part of the operator image
(`quay.io/kuberay/operator`) — the air-gap/private-mirror path.
This does NOT rewrite the Ray images clusters run — those ride
each KubernetesRayCluster's own image field.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (operator env feature flags, logging
encoders, single-namespace RBAC, the Kubernetes-proxy dialing
mode) — never the substitute for them.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesKubeRayOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it, so child names derive from it. |
| `status.outputs.watched_namespaces` | `[]string` | namespaces the operator watches for Ray CRs. Empty means cluster-wide — RayCluster declarations reconcile anywhere. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
