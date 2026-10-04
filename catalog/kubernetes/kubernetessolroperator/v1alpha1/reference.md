# KubernetesSolrOperator

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesSolrOperatorSpec** installs the Apache Solr Operator —
the Apache Solr project's own operator for running SolrCloud on
Kubernetes — from the official `solr-operator` Helm chart
(https://solr.apache.org/charts). The operator reconciles
`SolrCloud` custom resources (declared with KubernetesSolr) into
running Solr clusters with managed rolling updates, scaling with
replica movement, and backup repositories.

This component installs and configures the ENGINE. Solr clusters
themselves are declared with KubernetesSolr resources — one per
cluster — which this operator reconciles.

ZOOKEEPER: SolrCloud requires ZooKeeper. The chart bundles the
zookeeper-operator as a dependency (installed by default) so a
KubernetesSolr resource can simply declare a provided ZooKeeper
ensemble and the operator provisions it. Disable the bundled
zookeeper-operator only when one already runs in the cluster or
every Solr cluster will connect to an EXTERNAL ensemble.

CRD LIFECYCLE: the modules OWN the operator's CRDs (SolrCloud,
SolrBackup, SolrPrometheusExporter, and ZookeeperCluster for the
bundled dependency). They are derived from the pinned chart at deploy
time, applied keyed by CRD name ahead of the release, kept when the
resource is destroyed (`crds.keep_on_uninstall`), re-adopted on
reinstall, and moved with every `chart_version` bump; a `chart_version`
lower than the CRDs already on the cluster is refused before anything
changes. Uninstalling the operator never cascade-deletes SolrCloud
resources unless the spec says so.

WATCH SCOPE: by default the operator watches ALL namespaces. Set
`watch_namespaces` to fence it to an explicit set.

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
kind: KubernetesSolrOperator
metadata:
  name: solr-op-hack
spec:
  namespace:
    value: solr-op-hack
  createNamespace: true
  chartVersion: 0.9.1
  replicas: 2
  watchNamespaces:
    - search-team-a
    - search-team-b
  # Deliberately the second-install posture: no bundled
  # zookeeper-operator, point the Solr operator at the one already
  # running in the cluster.
  zookeeperOperator:
    install: false
    useExisting: true
  leaderElectionEnabled: true
  metricsEnabled: false
  mtls:
    clientCertSecret:
      value: solr-client-tls
    caCertSecret:
      value: solr-ca-cert
    caCertSecretKey: ca.crt
    insecureSkipVerify: false
    watchForUpdates: false
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: dedicated
      operator: Equal
      value: search
      effect: NoSchedule
  imagePullSecret: mirror-pull-secret
  image:
    repository: mirror.example.com/apache/solr-operator
    tag: v0.9.1
  helmValues: |
    priorityClassName: system-cluster-critical
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.9.1` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.watchNamespaces` | `[]string` |  |  |  |
| `spec.zookeeperOperator` | `KubernetesSolrOperatorZookeeperOperator` |  |  |  |
| `spec.zookeeperOperator.install` | `bool` |  | `true` |  |
| `spec.zookeeperOperator.useExisting` | `bool` |  |  |  |
| `spec.leaderElectionEnabled` | `bool` |  | `true` |  |
| `spec.metricsEnabled` | `bool` |  | `true` |  |
| `spec.mtls` | `KubernetesSolrOperatorMtls` |  |  |  |
| `spec.mtls.clientCertSecret` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.mtls.caCertSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.mtls.caCertSecretKey` | `string` |  | `ca-cert.pem` |  |
| `spec.mtls.insecureSkipVerify` | `bool` |  | `true` |  |
| `spec.mtls.watchForUpdates` | `bool` |  | `true` |  |
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
| `spec.imagePullSecret` | `string` |  |  |  |
| `spec.image` | `KubernetesSolrOperatorImage` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |
| `spec.crds` | `KubernetesSolrOperatorCrds` |  |  |  |
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

Helm chart version to install (e.g. "0.9.1" — chart and operator
versions move together; the chart version has no `v` prefix while
the operator/CRD artifacts carry one). Versions must exist as
SERVED charts in the repository index
(https://solr.apache.org/charts). The operator is pre-1.0: minor
versions can change the CRD API — the module-owned CRDs follow this
pin on every change, so a bump upgrades the schema with the operator.

- default: `0.9.1`

### spec.replicas

`int32` · optional (explicit presence)

Operator replica count. Chart default: 1. Extra replicas are
leader-elected warm standbys.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.watchNamespaces

`[]string`

Watch exactly these namespaces. Empty = the operator watches ALL
namespaces (the chart default). Unlike some operators, the
watched namespaces need only exist by the time SolrCloud
resources appear in them.

### spec.zookeeperOperator

`KubernetesSolrOperatorZookeeperOperator`

The bundled zookeeper-operator dependency. Empty = installed
(the chart default) — the path that makes provided ZooKeeper
ensembles work out of the box.

### spec.zookeeperOperator.install

`bool` · optional (explicit presence)

Install the zookeeper-operator with this release. Default: true.
Set false when a zookeeper-operator already runs in the cluster
(its fixed-name cluster-scoped RBAC conflicts on a second
install) or when every Solr cluster uses an external ensemble.

- default: `true`

### spec.zookeeperOperator.useExisting

`bool`

Tell the Solr operator a zookeeper-operator is present even
though this release does not install one (the chart's `use`
knob). Set together with install=false when the cluster already
runs one.

### spec.leaderElectionEnabled

`bool` · optional (explicit presence)

Leader election between operator replicas. Chart default: true.
Disable only for a single-replica dev install.

- default: `true`

### spec.metricsEnabled

`bool` · optional (explicit presence)

Expose the operator's Prometheus metrics endpoint. Chart
default: true.

- default: `true`

### spec.mtls

`KubernetesSolrOperatorMtls`

Client certificate the operator presents to Solr clusters that
require mutual TLS. Only needed when KubernetesSolr resources
enforce clientAuth on their TLS listeners.

### spec.mtls.clientCertSecret

`string | valueFrom` · required

Secret holding the client certificate the operator presents
(tls.crt/tls.key). Required whenever the mtls block is declared —
an mtls block without a client certificate would render nothing
and silently leave the operator without an identity.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.mtls.caCertSecret

`string | valueFrom`

Secret holding the CA certificate to trust when calling Solr.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.mtls.caCertSecretKey

`string` · optional (explicit presence)

Key within ca_cert_secret holding the CA certificate. Chart
default: "ca-cert.pem".

- default: `ca-cert.pem`

### spec.mtls.insecureSkipVerify

`bool` · optional (explicit presence)

Skip server hostname verification on operator->Solr calls. Chart
default: true (pod-IP calls rarely match certificate SANs).

- default: `true`

### spec.mtls.watchForUpdates

`bool` · optional (explicit presence)

Watch the certificate secret and restart the operator on
rotation. Chart default: true.

- default: `true`

### spec.resources

`ContainerResources`

Operator container resources. Empty = the chart defaults (none
set — the operator is lightweight).

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

### spec.imagePullSecret

`string`

Name of an image-pull secret (in the installation namespace) for
pulling the operator image from a private mirror. The chart
accepts exactly one.

### spec.image

`KubernetesSolrOperatorImage`

Override the operator image (air-gap / private-mirror path).
Empty = the chart defaults (apache/solr-operator at the chart's
matching tag).

### spec.image.repository

`string`

Image repository including registry, e.g.
"my-mirror.example.com/apache/solr-operator". Empty =
"apache/solr-operator".

### spec.image.tag

`string`

Image tag. Empty = the operator tag matching the pinned chart
(`v<chart_version>`).

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (annotations/labels, env vars, sidecar
containers, priority class, service account, the bundled
zookeeper-operator's own values, ...) — never the substitute for
them. Do not put secrets here.

### spec.crds

`KubernetesSolrOperatorCrds`

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

Keep the CRDs (and therefore every SolrCloud, SolrBackup,
SolrPrometheusExporter and ZookeeperCluster in the cluster) when the
resource is destroyed. Default TRUE: deleting the CRDs cascades to
every Solr cluster, a destructive act that must be an explicit false.
A later reinstall re-adopts kept CRDs.

- default: `true`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesSolrOperator, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the operator is installed into. |
| `status.outputs.release_name` | `string` | Helm release name of the operator install (= metadata.name). |
| `status.outputs.deployment_name` | `string` | name of the Solr operator Deployment. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.mtls.clientCertSecret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.mtls.caCertSecret` | KubernetesSecret | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
