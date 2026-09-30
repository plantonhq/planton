# KubernetesMetricsServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesMetricsServerSpec** installs metrics-server — the cluster's
resource-metrics pipeline — from the official Helm chart (`metrics-server`
at https://kubernetes-sigs.github.io/metrics-server/). It scrapes each
kubelet and serves live CPU/memory usage through the `metrics.k8s.io`
APIService, which is what `kubectl top` reads and what
HorizontalPodAutoscalers need for Resource-type utilization targets —
without metrics-server, HPAs deploy but never receive metric values and
never scale.

ONE INSTALLATION PER CLUSTER: metrics-server registers the cluster-wide
`v1beta1.metrics.k8s.io` APIService, a singleton — a second installation
would fight over it. The Helm release name is therefore fixed to
"metrics-server". Managed clouds differ: GKE and AKS ship metrics-server
built-in as a managed component (do not install this component there —
AKS runs it in kube-system on every cluster); EKS, kind, k3s and most
self-managed clusters need it.

The typed fields below cover the chart's meaningful configuration surface;
`helm_values` remains as the escape hatch for chart values beyond them
(merged last, Helm `-f` semantics, identical on both engines) — a safety
valve, never the primary interface.

## Example

```yaml
# Full-surface test manifest: exercises every typed arm of the spec so the
# offline plan proofs cover what the live lanes may not. Not a realistic
# production shape — see presets for those.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMetricsServer
metadata:
  name: hack-metrics-server
spec:
  namespace:
    value: hack-metrics-server
  createNamespace: true
  chartVersion: "3.13.1"
  replicas: 2
  kubeletInsecureTls: true
  kubeletPreferredAddressTypes:
    - Hostname
    - InternalIP
  metricResolution: 30s
  hostNetwork: true
  apiService:
    create: true
    insecureSkipTlsVerify: false
    caBundle: |
      -----BEGIN CERTIFICATE-----
      MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
      -----END CERTIFICATE-----
  tls:
    type: existing_secret
    existingSecretName:
      value: hack-metrics-server-tls
  resources:
    requests:
      cpu: 100m
      memory: 200Mi
    limits:
      memory: 400Mi
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: node-role.kubernetes.io/control-plane
      operator: Exists
      effect: NoSchedule
  priorityClassName: system-cluster-critical
  podDisruptionBudget: true
  prometheus:
    enabled: true
    serviceMonitor: true
    serviceMonitorInterval: 30s
    serviceMonitorLabels:
      release: kube-prometheus-stack
  image:
    repository: registry.k8s.io/metrics-server/metrics-server
    tag: v0.8.1
  helmValues: |
    deploymentAnnotations:
      example.org/team: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `3.13.1` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.kubeletInsecureTls` | `bool` |  |  |  |
| `spec.kubeletPreferredAddressTypes` | `[]string` |  |  |  |
| `spec.metricResolution` | `string` |  | `15s` |  |
| `spec.hostNetwork` | `bool` |  |  |  |
| `spec.apiService` | `KubernetesMetricsServerApiService` |  |  |  |
| `spec.apiService.create` | `bool` |  | `true` |  |
| `spec.apiService.insecureSkipTlsVerify` | `bool` |  | `true` |  |
| `spec.apiService.caBundle` | `string` |  |  |  |
| `spec.tls` | `KubernetesMetricsServerTls` |  |  |  |
| `spec.tls.type` | `enum` |  | `self_signed` |  |
| `spec.tls.certManagerIssuer` | `CertManagerIssuerRef` |  |  |  |
| `spec.tls.certManagerIssuer.issuer` | `CertManagerNamespacedIssuerRef` |  |  |  |
| `spec.tls.certManagerIssuer.issuer.name` | `string \| valueFrom` | yes |  | KubernetesIssuer (`status.outputs.issuer_name`) |
| `spec.tls.certManagerIssuer.clusterIssuer` | `CertManagerClusterIssuerRef` |  |  |  |
| `spec.tls.certManagerIssuer.clusterIssuer.name` | `string \| valueFrom` | yes |  | KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`) |
| `spec.tls.existingSecretName` | `string \| valueFrom` |  |  | KubernetesSecret (`metadata.name`) |
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
| `spec.priorityClassName` | `string` |  | `system-cluster-critical` |  |
| `spec.podDisruptionBudget` | `bool` |  |  |  |
| `spec.prometheus` | `KubernetesMetricsServerPrometheus` |  |  |  |
| `spec.prometheus.enabled` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitorInterval` | `string` |  | `1m` |  |
| `spec.prometheus.serviceMonitorLabels` | `map<string, string>` |  |  |  |
| `spec.image` | `KubernetesMetricsServerImage` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install metrics-server into ("kube-system" is the
upstream convention — the APIService is cluster infrastructure; a
dedicated "metrics-server" namespace also works). Accepts a literal
namespace name or a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist (it usually does —
kube-system always exists).

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "3.13.1", which ships
metrics-server 0.8.1 — chart and app versions are numbered
independently). Pin deliberately; upgrades re-run the release with the
new chart. Pick versions from the chart repository's index
(`helm search repo`): the served chart is the contract — the upstream
source tree's Chart.yaml can claim a version at a tag that was never
served.

- default: `3.13.1`

### spec.replicas

`int32` · optional (explicit presence)

Replica count. One is standard; run 2 with `pod_disruption_budget` for
HA of the metrics API (the APIService fails over between healthy
replicas).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.kubeletInsecureTls

`bool`

Skip verification of kubelet serving-certificate CAs
(--kubelet-insecure-tls). REQUIRED on clusters whose kubelets serve
self-signed certificates: kind, k3s, kubeadm without kubelet
certificate rotation, many on-prem setups. Leave false on managed
clouds (EKS/AKS kubelet certs verify against the cluster CA). This is
a scrape-path trust decision, not a serving-side one — the APIService
side is `api_service.insecure_skip_tls_verify`.

### spec.kubeletPreferredAddressTypes

`[]string`

Order of address types tried when scraping a kubelet
(--kubelet-preferred-address-types). Upstream default:
[InternalIP, ExternalIP, Hostname]. Reorder on clusters where the
default resolves to unreachable addresses (e.g. Hostname-first on
clusters whose node names resolve internally).

- rule: {"repeated":{"items":{"cel":[{"id":"spec.kubelet_address_type","message":"Each address type must be one of: InternalIP, ExternalIP, Hostname, InternalDNS, ExternalDNS","expression":"this in ['InternalIP', 'ExternalIP', 'Hostname', 'InternalDNS', 'ExternalDNS']"}]}}}

### spec.metricResolution

`string` · optional (explicit presence)

How often kubelets are scraped (--metric-resolution). Upstream
default "15s" — the value HPA freshness is built around; values under
15s burn kubelet CPU for little gain, values over 60s make
autoscaling sluggish.

- default: `15s`
- rule: {"string":{"pattern":"^[0-9]+(s|m)$"}}

### spec.hostNetwork

`bool`

Run metrics-server on the host network. Needed where the API server
cannot reach pod IPs over the overlay (the upstream example: Weave
CNI on EKS).

### spec.apiService

`KubernetesMetricsServerApiService`

The metrics.k8s.io APIService registration.

### spec.apiService.create

`bool` · optional (explicit presence)

Register the v1beta1.metrics.k8s.io APIService with the release.
Chart default: true — without it, kubectl top and HPAs cannot find
the metrics API. Disable only when something else manages the
APIService object.

- default: `true`

### spec.apiService.insecureSkipTlsVerify

`bool` · optional (explicit presence)

Let the API server skip TLS verification when calling metrics-server.
Chart default: true — matches the default self-signed serving
certificate. Set false only when the serving certificate is
verifiable: tls type helm (the chart wires the APIService caBundle
automatically), cert_manager (the CA injector wires it), or
existing_secret with the CA provided via ca_bundle.

- default: `true`

### spec.apiService.caBundle

`string`

PEM-encoded CA bundle the API server uses to verify metrics-server's
serving certificate. Only meaningful with insecure_skip_tls_verify =
false and a serving certificate issued by this CA.

### spec.tls

`KubernetesMetricsServerTls`

How the metrics-server serving certificate (what the API SERVER
verifies when calling the APIService) is provisioned.

- rule: tls type existing_secret requires existing_secret_name — the Secret holding the serving certificate
- rule: cert_manager_issuer is only used with tls type cert_manager — set type accordingly or remove the issuer
- rule: existing_secret_name is only used with tls type existing_secret — set type accordingly or remove the secret reference

### spec.tls.type

`enum` · optional (explicit presence)

Provisioning method. The default self_signed is fine for almost every
cluster (the APIService skips verification by default); pick
cert_manager for a verified chain with automatic renewal.

- default: `self_signed`

Allowed values (use exactly as shown):

- `self_signed` -- metrics-server generates its own self-signed certificate at startup (upstream default; pairs with api_service.insecure_skip_tls_verify).
- `helm` -- Helm generates a self-signed certificate at install time and wires the APIService caBundle automatically (verifiable without cert-manager).
- `cert_manager` -- cert-manager issues and renews the certificate; the CA injector wires the APIService caBundle (requires KubernetesCertManager on the cluster).
- `existing_secret` -- Reuse an existing kubernetes.io/tls Secret you manage.

### spec.tls.certManagerIssuer

`CertManagerIssuerRef`

cert-manager issuer that signs the serving certificate (type
cert_manager). Empty = the chart creates its own self-signed
Issuer + root Certificate chain in the installation namespace.

The grain is the arm: `issuer` (a namespaced Issuer in the installation
namespace) or `cluster_issuer` (a cluster-scoped ClusterIssuer).

- rule: An issuer is required -- choose 'issuer' (a namespaced Issuer in the installation namespace) or 'cluster_issuer' (a cluster-scoped ClusterIssuer)

### spec.tls.certManagerIssuer.issuer

`CertManagerNamespacedIssuerRef`

A namespaced Issuer in the installation namespace, by name.

### spec.tls.certManagerIssuer.issuer.name

`string | valueFrom` · required

Issuer name. Accepts a literal name or a reference to a KubernetesIssuer
resource.

- references: KubernetesIssuer (`status.outputs.issuer_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesIssuer, name: <that resource's name>, fieldPath: status.outputs.issuer_name}} -- a bare string does not parse

### spec.tls.certManagerIssuer.clusterIssuer

`CertManagerClusterIssuerRef`

A cluster-scoped ClusterIssuer, by name.

### spec.tls.certManagerIssuer.clusterIssuer.name

`string | valueFrom` · required

ClusterIssuer name. Accepts a literal name or a reference to a
KubernetesClusterIssuer resource.

- references: KubernetesClusterIssuer (`status.outputs.cluster_issuer_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClusterIssuer, name: <that resource's name>, fieldPath: status.outputs.cluster_issuer_name}} -- a bare string does not parse

### spec.tls.existingSecretName

`string | valueFrom`

Name of the existing kubernetes.io/tls Secret (type existing_secret).

- references: KubernetesSecret (`metadata.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: metadata.name}} -- a bare string does not parse

### spec.resources

`ContainerResources`

Container CPU/memory requests and limits. Chart default: requests
cpu=100m memory=200Mi, no limits — upstream's envelope for clusters
up to 100 nodes (≤70 pods/node). Beyond 100 nodes, upstream guidance
is +1m CPU and +2Mi memory per additional node.

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

Node selector for the metrics-server pods.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the metrics-server pods (e.g. to run on control-plane
nodes).

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

### spec.priorityClassName

`string` · optional (explicit presence)

PriorityClass for the metrics-server pods. Chart default:
system-cluster-critical — the metrics pipeline should outlive workload
evictions (HPAs stop scaling without it).

- default: `system-cluster-critical`

### spec.podDisruptionBudget

`bool`

When true, a PodDisruptionBudget (minAvailable 1) guards
metrics-server — meaningful only with replicas > 1 (with a single
replica it would block every voluntary eviction, wedging node drains).

### spec.prometheus

`KubernetesMetricsServerPrometheus`

Expose metrics-server's OWN Prometheus metrics (scrape performance,
request latencies) and optionally a ServiceMonitor. This is telemetry
ABOUT metrics-server — unrelated to the resource metrics it serves.

- rule: service_monitor requires prometheus metrics to be enabled — the ServiceMonitor would have no metrics endpoint to scrape

### spec.prometheus.enabled

`bool`

Expose metrics-server's own /metrics endpoint through a Service.
Chart default: false.

### spec.prometheus.serviceMonitor

`bool`

Create a ServiceMonitor for scrape discovery. Requires the Prometheus
operator CRDs (e.g. kube-prometheus-stack) on the cluster — the
release FAILS to install without them.

### spec.prometheus.serviceMonitorInterval

`string` · optional (explicit presence)

Scrape interval for the ServiceMonitor. Chart default: "1m".

- default: `1m`

### spec.prometheus.serviceMonitorLabels

`map<string, string>`

Extra labels on the ServiceMonitor — how a Prometheus instance's
serviceMonitorSelector finds it (e.g. {"release": "kube-prometheus-stack"}).

### spec.image

`KubernetesMetricsServerImage`

Image override — the air-gapped/mirror knob. Empty parts = the chart
defaults (registry.k8s.io/metrics-server/metrics-server at the chart's
pinned app version).

### spec.image.repository

`string`

Image repository. Empty = registry.k8s.io/metrics-server/metrics-server.

### spec.image.tag

`string`

Image tag. Empty = the chart's pinned app version (v0.8.1 for chart
3.13.1). Overriding decouples the binary from the chart's tested pair —
prefer bumping chart_version.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics, identical
on both engines). For the chart surface beyond the typed fields —
never the substitute for them. Do not put secrets here.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesMetricsServer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Kubernetes namespace metrics-server was installed into. |
| `status.outputs.release_name` | `string` | Helm release name (fixed "metrics-server" — one installation per cluster). |
| `status.outputs.service_name` | `string` | Name of the metrics-server Service the APIService routes to. |
| `status.outputs.api_service_name` | `string` | Name of the registered APIService ("v1beta1.metrics.k8s.io") — the cluster-wide resource-metrics API kubectl top and HPAs consume. Empty when spec.api_service.create is false. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.tls.certManagerIssuer.issuer.name` | KubernetesIssuer | `status.outputs.issuer_name` |
| `spec.tls.certManagerIssuer.clusterIssuer.name` | KubernetesClusterIssuer | `status.outputs.cluster_issuer_name` |
| `spec.tls.existingSecretName` | KubernetesSecret | `metadata.name` |

## See Also

- [Overview](../README.md)
