# KubernetesAltinityOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesAltinityOperatorSpec** installs the Altinity ClickHouse
operator — the Apache-2.0 operator for running ClickHouse (the
columnar OLAP database) on Kubernetes — from the official
`altinity-clickhouse-operator` Helm chart
(https://docs.altinity.com/clickhouse-operator/). The operator
reconciles `ClickHouseInstallation` (CHI) and
`ClickHouseKeeperInstallation` (CHK) custom resources into running
clusters with generated server configuration, rolling restarts and
per-host StatefulSets.

This component installs and configures the ENGINE. ClickHouse
clusters themselves are declared with KubernetesClickHouse
resources — one per cluster — which this operator reconciles.

CHART/IMAGE PINNING: pick `chart_version` from the SERVED repository
index (https://docs.altinity.com/clickhouse-operator/index.yaml).
Chart versions track operator releases one-to-one (chart 0.27.2 =
operator image 0.27.2); all default images at the pinned chart
verified pullable on Docker Hub.

CRD LIFECYCLE: the chart ships the four CRDs in its `crds/`
directory — Helm installs them on first install and NEVER deletes
them on uninstall (keep-on-uninstall is inherent; removing the
operator never cascade-deletes ClickHouseInstallation resources or
their data). CRD UPGRADES are handled by the chart's pre-upgrade
hook job (`crd_hook`), which server-side-applies the CRDs on every
install and upgrade — leave it enabled or chart upgrades silently
run new operators against old schemas.

NAMING: keep the resource name at 39 characters or fewer. The
modules pin the chart fullname to the resource name, and the
longest generated child name (`<fullname>-keeper-templatesd-files`)
adds 24 characters to it against the Kubernetes 63-character cap.

WATCH SCOPE: by default the operator watches ONLY its own
namespace. Set `watch_namespaces` to widen the scope — every
namespace that will hold KubernetesClickHouse resources must be
covered, or use [".*"] to watch the whole cluster.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — a safety valve, never the primary interface.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAltinityOperator
metadata:
  name: altinity-operator
spec:
  namespace:
    value: altinity-operator
  createNamespace: true
  watchNamespaces:
    - ".*"
  operatorCredentials:
    username: clickhouse_operator
    password:
      value: ch0p-Str0ng-Passw0rd-2026
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.27.2` |  |
| `spec.watchNamespaces` | `[]string` |  |  |  |
| `spec.namespaceScopedRbac` | `bool` |  |  |  |
| `spec.operatorCredentials` | `KubernetesAltinityOperatorCredentials` |  |  |  |
| `spec.operatorCredentials.username` | `string` |  | `clickhouse_operator` |  |
| `spec.operatorCredentials.password` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.metrics` | `KubernetesAltinityOperatorMetrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  | `true` |  |
| `spec.metrics.resources` | `ContainerResources` |  |  |  |
| `spec.metrics.resources.limits` | `CpuMemory` |  |  |  |
| `spec.metrics.resources.limits.cpu` | `string` |  |  |  |
| `spec.metrics.resources.limits.memory` | `string` |  |  |  |
| `spec.metrics.resources.requests` | `CpuMemory` |  |  |  |
| `spec.metrics.resources.requests.cpu` | `string` |  |  |  |
| `spec.metrics.resources.requests.memory` | `string` |  |  |  |
| `spec.crdHook` | `KubernetesAltinityOperatorCrdHook` |  |  |  |
| `spec.crdHook.enabled` | `bool` |  | `true` |  |
| `spec.crdHook.image` | `ContainerImage` |  |  |  |
| `spec.crdHook.image.repo` | `string` |  |  |  |
| `spec.crdHook.image.tag` | `string` |  |  |  |
| `spec.crdHook.image.pullSecretName` | `string` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.image` | `ContainerImage` |  |  |  |
| `spec.image.repo` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.pullSecretName` | `string` |  |  |  |
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

Helm chart version to install (e.g. "0.27.2"). Versions must
exist as SERVED charts in the repository index
(https://docs.altinity.com/clickhouse-operator/index.yaml); the
chart version and the operator image tag move in lockstep.

- default: `0.27.2`

### spec.watchNamespaces

`[]string`

Namespaces the operator watches for ClickHouseInstallation /
ClickHouseKeeperInstallation resources. Entries are regular
expressions (exact names work as-is). Empty (the chart default) =
the operator watches ONLY its own namespace; [".*"] = watch all
namespaces. Every namespace that will hold KubernetesClickHouse
resources must be covered here.

### spec.namespaceScopedRbac

`bool`

Create namespace-scoped Roles/RoleBindings instead of cluster-wide
RBAC (chart `rbac.namespaceScoped`). Chart default: false. Only
sound when the operator watches its own namespace alone — a
namespace-scoped operator cannot manage clusters it can watch but
not touch.

### spec.operatorCredentials

`KubernetesAltinityOperatorCredentials`

Credentials the operator itself uses to connect to every managed
ClickHouse instance (host management, schema propagation, metrics
scraping). The chart provisions them as a Secret and injects the
matching user into every managed cluster. UNSET IS UNSAFE FOR
PRODUCTION: the chart's defaults are the publicly documented
clickhouse_operator / clickhouse_operator_password — always set a
real password outside of throwaway environments.

### spec.operatorCredentials.username

`string` · optional (explicit presence)

User name. Default "clickhouse_operator".

- default: `clickhouse_operator`

### spec.operatorCredentials.password

`string | valueFrom` · required · sensitive

Password. Accepts a literal value or a reference to another
resource's output. Stored in the chart-managed Secret; the
operator injects the matching user into every managed cluster
restricted to the operator pod's address.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.metrics

`KubernetesAltinityOperatorMetrics`

The metrics-exporter sidecar that serves Prometheus metrics for
every managed cluster (port 8888). Enabled by default upstream.

### spec.metrics.enabled

`bool` · optional (explicit presence)

Run the metrics-exporter sidecar. Default true (the upstream
default); disabling it removes Prometheus metrics for every
managed cluster.

- default: `true`

### spec.metrics.resources

`ContainerResources`

Sidecar container resources. Empty = the chart defaults (no
requests/limits). The exporter's memory grows with the number and
size of managed clusters.

### spec.metrics.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.metrics.resources.limits.cpu

`string`

### spec.metrics.resources.limits.memory

`string`

### spec.metrics.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.metrics.resources.requests.cpu

`string`

### spec.metrics.resources.requests.memory

`string`

### spec.crdHook

`KubernetesAltinityOperatorCrdHook`

The pre-install/pre-upgrade hook job that server-side-applies the
four CRDs (chart `crdHook`). Enabled by default upstream — leave
it on so chart upgrades carry CRD schema changes. KNOW THIS: the
hook's default image is bitnami/kubectl:latest — pullable today
(verified live) but frozen since Bitnami's 2025 public-catalog
retirement, so it will silently age; override `image` for
air-gapped mirrors or a maintained kubectl build.

### spec.crdHook.enabled

`bool` · optional (explicit presence)

Run the hook. Default true (the upstream default). When disabled,
Helm still installs the CRDs on FIRST install (native `crds/`
handling) but upgrades stop carrying CRD schema changes — only
disable if CRD lifecycle is managed elsewhere.

- default: `true`

### spec.crdHook.image

`ContainerImage`

Override the hook's kubectl image. Empty = the chart default
(bitnami/kubectl:latest — see the KNOW THIS note on `crd_hook`).

### spec.crdHook.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.crdHook.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.crdHook.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.resources

`ContainerResources`

Operator container resources. Empty = the chart defaults (no
requests/limits).

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

### spec.serviceMonitorEnabled

`bool`

Create a ServiceMonitor for Prometheus Operator scraping (both the
per-cluster metrics endpoint on 8888 and the operator's own on
9999). Chart default: false. Requires the Prometheus Operator CRDs
on the cluster — enabling it without them fails the install.

### spec.nodeSelector

`map<string, string>`

Node selector for the operator pod.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the operator pod.

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

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the installation namespace) for
pulling the operator images from a private mirror.

### spec.image

`ContainerImage`

Override the operator image (air-gap / private-mirror path).
Empty = the chart defaults (altinity/clickhouse-operator at the
chart's version).

### spec.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (pod annotations/labels, security contexts, extra env,
probe tuning, topology spread, the embedded operator config files,
Grafana dashboard ConfigMaps, ...) — never the substitute for
them. Do not put secrets here. One key is off limits:
`fullnameOverride` is pinned by the modules to the resource name —
overriding it breaks the naming budget and the exported outputs.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAltinityOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator is installed into. |
| `status.outputs.release_name` | `string` | Helm release name of the operator install (= metadata.name). |
| `status.outputs.deployment_name` | `string` | name of the operator Deployment (= the chart fullname, which the modules pin to the resource name). |
| `status.outputs.credentials_secret_name` | `string` | name of the chart-managed Secret holding the operator's ClickHouse credentials (fields username/password), e.g. <name>. |
| `status.outputs.metrics_endpoint` | `string` | in-cluster metrics endpoint serving Prometheus metrics for every managed cluster, e.g. http://<name>-metrics.<ns>.svc.cluster.local:8888/metrics. Empty when the metrics exporter is disabled. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
