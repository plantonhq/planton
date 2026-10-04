# KubernetesOpenSearchOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesOpenSearchOperatorSpec** installs the OpenSearch
Kubernetes Operator — the opensearch-project's operator for running
OpenSearch (the Apache-2.0 search and analytics engine) on
Kubernetes — from the official `opensearch-operator` Helm chart
(https://opensearch-project.github.io/opensearch-k8s-operator/).
The operator reconciles `OpenSearchCluster` custom resources
(declared with KubernetesOpenSearch) into running search clusters
with managed TLS, security bootstrap, rolling upgrades and
Dashboards.

This component installs and configures the ENGINE. Search clusters
themselves are declared with KubernetesOpenSearch resources — one
per cluster — which this operator reconciles.

CHART/IMAGE PINNING: pick `chart_version` from the SERVED repository
index. Chart 2.8.0 pairs with the stable operator 2.8.0 image; the
newer served charts (2.8.3+, 3.0.x) default their manager image to a
PRERELEASE tag (verified against the served charts' own appVersion)
and additionally bundle next-generation `opensearch.org` API-group
CRDs the stable operator does not serve — pin those lines only after
upstream cuts a stable 3.x operator release.

NAMING: keep the resource name at 27 characters or fewer. The chart
derives long child names from its fullname (the longest,
"<fullname>-controller-manager-metrics-service", adds 36 characters),
and Kubernetes caps object names at 63 — verified live: the chart
truncates the fullname but NOT the names built from it, so an
over-long name fails at the API server mid-install. The modules pin
the fullname to the resource name, which is what makes the 27-char
budget hold.

CRD LIFECYCLE: the chart TEMPLATES the OpenSearch CRDs as
release-owned resources (no keep-on-uninstall knob upstream). The
modules therefore OWN the CRDs: derived from the pinned chart at
deploy time, applied keyed by CRD name ahead of the release, kept
when the resource is destroyed (`crds.keep_on_uninstall`), re-adopted
on reinstall, and moved with every `chart_version` bump; a
`chart_version` lower than the CRDs already on the cluster is refused
before anything changes. Uninstalling the operator never
cascade-deletes OpenSearchCluster resources and their data unless the
spec says so.

WATCH SCOPE: by default the operator watches ALL namespaces
(cluster-wide RBAC). Set `watch_namespace` to restrict it to one
namespace; pair with `use_role_bindings` to avoid cluster-wide RBAC
entirely on shared clusters.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface development manifest: exercises every typed field so the
# offline plan proofs cover arms the live lanes exclude. Values are
# deliberately non-default where a default exists, so rendered-values
# inspection shows every mapping.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearchOperator
metadata:
  name: os-op-hack
spec:
  namespace:
    value: opensearch-op
  createNamespace: true
  chartVersion: 2.8.0
  # watchNamespace + useRoleBindings together (spec CEL requires the
  # pairing): the fully namespace-scoped posture.
  watchNamespace: opensearch-op
  useRoleBindings: true
  logLevel: debug
  dnsBase: cluster.local
  parallelRecoveryEnabled: false
  pprofEndpointsEnabled: true
  kubeRbacProxyEnabled: false
  resources:
    requests:
      cpu: 150m
      memory: 384Mi
    limits:
      cpu: 300m
      memory: 512Mi
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: dedicated
      operator: Equal
      value: search
      effect: NoSchedule
  imagePullSecrets:
    - mirror-pull-secret
  image:
    repository: mirror.example.com/opensearchproject/opensearch-operator
    tag: 2.8.0
  helmValues: |
    podLabels:
      team: search-platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `2.8.0` |  |
| `spec.watchNamespace` | `string` |  |  |  |
| `spec.useRoleBindings` | `bool` |  |  |  |
| `spec.logLevel` | `string` |  | `info` |  |
| `spec.dnsBase` | `string` |  | `cluster.local` |  |
| `spec.parallelRecoveryEnabled` | `bool` |  | `true` |  |
| `spec.pprofEndpointsEnabled` | `bool` |  |  |  |
| `spec.kubeRbacProxyEnabled` | `bool` |  | `true` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.image` | `KubernetesOpenSearchOperatorImage` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |
| `spec.crds` | `KubernetesOpenSearchOperatorCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |

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

Helm chart version to install (e.g. "2.8.0"). Versions must exist
as SERVED charts in the repository index
(https://opensearch-project.github.io/opensearch-k8s-operator/).
The chart's default manager image tag is its appVersion — the
served 2.8.3/2.8.4/3.0.x charts all default to a PRERELEASE
operator image (verified against the served charts), so 2.8.0 is
the newest stable pairing. A version bump is a license and
API-group re-check: the 3.x line migrates the CRDs from
`opensearch.opster.io` to `opensearch.org`.

- default: `2.8.0`

### spec.watchNamespace

`string`

Restrict the operator to watch a single namespace instead of the
cluster-wide default. When set, OpenSearchCluster resources in
other namespaces are ignored by this install. Empty = watch all
namespaces (the chart default).

### spec.useRoleBindings

`bool`

Use namespace-scoped RoleBindings instead of ClusterRoleBindings
for the operator's RBAC. Chart default: false. Only valid together
with `watch_namespace` — a cluster-wide operator cannot run on
namespace-scoped permissions (the chart grants roles only in the
watched namespace).

### spec.logLevel

`string` · optional (explicit presence)

Operator manager log level: debug, info (the chart default), warn,
or error.

- default: `info`
- rule: log level must be debug, info, warn, or error

### spec.dnsBase

`string` · optional (explicit presence)

Kubernetes cluster DNS domain the operator bakes into generated
certificates and discovery addresses. Chart default:
"cluster.local". Set only on clusters with a non-default DNS
domain — a mismatch produces TLS certificates whose SANs do not
match the service DNS names nodes advertise.

- default: `cluster.local`

### spec.parallelRecoveryEnabled

`bool` · optional (explicit presence)

Recover cluster pods in parallel instead of one at a time after
failures. Chart default: true (upstream marks it experimental —
disable if recovery behaves unexpectedly).

- default: `true`

### spec.pprofEndpointsEnabled

`bool`

Enable the Go pprof debug endpoints on the manager (port 6060).
Chart default: false. Debugging aid only — never enable in
production.

### spec.kubeRbacProxyEnabled

`bool` · optional (explicit presence)

Deploy the kube-rbac-proxy sidecar that shields the operator's
metrics endpoint behind Kubernetes RBAC. Chart default: true.
KNOW THIS: the chart's own default sidecar image
(gcr.io/kubebuilder/kube-rbac-proxy) was deleted upstream and can
never be pulled (verified live) — the modules re-point it at the
maintainer's quay.io repository at the chart's pinned tag, so the
default posture actually installs. Override the image via
helm_values for air-gapped mirrors.

- default: `true`

### spec.resources

`ContainerResources`

Operator manager container resources. Empty = the chart defaults
(requests 100m/350Mi, limits 200m/500Mi).

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
pulling the manager image from a private mirror.

### spec.image

`KubernetesOpenSearchOperatorImage`

Override the operator manager image (air-gap / private-mirror
path). Empty = the chart defaults
(opensearchproject/opensearch-operator at the chart's appVersion).

### spec.image.repository

`string`

Image repository including registry, e.g.
"my-mirror.example.com/opensearchproject/opensearch-operator".
Empty = "opensearchproject/opensearch-operator".

### spec.image.tag

`string`

Image tag. Empty = the chart's appVersion (the operator release
paired with the pinned chart).

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (pod annotations/labels, priority class, security contexts,
extra env, probe tuning, kube-rbac-proxy image/resources, service
account name, ...) — never the substitute for them. Do not put
secrets here. Two keys are off limits: `installCRDs` is re-pinned
by the modules after this merge (the module owns the CRD lifecycle
— a Helm-owned install would cascade-delete every OpenSearchCluster
on uninstall), and `nameOverride`/`fullnameOverride` break the
exported deployment_name output, which derives from the chart's
default naming.

### spec.crds

`KubernetesOpenSearchOperatorCrds`

CRD installation lifecycle. Both default to true; unset means the
module owns the CRDs and keeps them.

### spec.crds.install

`bool` · optional (explicit presence)

Derive and apply the CRDs from the pinned chart ahead of the release.
Default TRUE. Set false ONLY when the CRDs are owned elsewhere (a
GitOps-managed bundle): the release still installs with CRDs skipped,
and with the CRDs absent the operator cannot start, so this is a
bring-your-own-CRDs arm, never a lighter install.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Keep the CRDs (and therefore every OpenSearchCluster and every
OpenSearch user, role, tenant, template and policy object in the
cluster) when the resource is destroyed. Default TRUE: deleting the
CRDs cascades to every search cluster and its data, a destructive act
that must be an explicit false. A later reinstall re-adopts kept CRDs.

- default: `true`

## Validation Rules

- `spec.role_bindings_require_watch_namespace`: use_role_bindings requires watch_namespace — namespace-scoped RBAC cannot serve a cluster-wide operator

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesOpenSearchOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator is installed into. |
| `status.outputs.release_name` | `string` | Helm release name of the operator install (= metadata.name). |
| `status.outputs.deployment_name` | `string` | name of the operator controller-manager Deployment (the chart's fixed component name within the release). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
