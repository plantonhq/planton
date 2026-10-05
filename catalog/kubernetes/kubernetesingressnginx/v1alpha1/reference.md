# KubernetesIngressNginx

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesIngressNginxSpec** installs the ingress-nginx controller — the
cluster's HTTP(S) entry point — from the official Helm chart
(`ingress-nginx` at https://kubernetes.github.io/ingress-nginx). The
controller watches Ingress resources of its IngressClass and programs an
embedded NGINX to route external traffic to Services.

This component installs the CONTROLLER only. Routing rules are separate
first-class resources: create KubernetesIngress objects referencing this
controller's `ingress_class_name` output. TLS certificates come from
cert-manager (KubernetesCertManager + issuers) via each Ingress's TLS
secret, or cluster-wide via `default_tls_certificate` below.

MULTIPLE CONTROLLERS PER CLUSTER are first-class (the upstream pattern for
public + internal traffic splits): each instance gets its own resource
with a DISTINCT `ingress_class.name` — Helm release naming, controller
resource names, and leader-election identity all derive from
`metadata.name`, so instances never collide. What the host cloud
provisions for the controller's LoadBalancer Service is driven entirely
by `service.annotations` (NLB vs ALB on AWS, internal LB on GCP/Azure,
...) — see the per-cloud recipes in the kind's README.

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
kind: KubernetesIngressNginx
metadata:
  name: hack-ingress-nginx
spec:
  namespace:
    value: hack-ingress-nginx
  createNamespace: true
  chartVersion: "4.15.1"
  ingressClass:
    name: hack-nginx
    isDefaultClass: false
    watchIngressWithoutClass: false
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
    targetCpuUtilizationPercent: 60
    targetMemoryUtilizationPercent: 70
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      memory: 256Mi
  service:
    type: load_balancer
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-type: external
      service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
    externalTrafficPolicy: local
    loadBalancerSourceRanges:
      - 10.0.0.0/8
      - 203.0.113.0/24
    loadBalancerClass: service.k8s.aws/nlb
    httpNodePort: 30080
    httpsNodePort: 30443
    internal:
      enabled: true
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-scheme: internal
  nginxConfig:
    proxy-body-size: 16m
    use-forwarded-headers: "true"
    worker-shutdown-timeout: 240s
  allowSnippetAnnotations: true
  defaultTlsCertificate:
    secretName:
      value: hack-wildcard-tls
    namespace: hack-ingress-nginx
  defaultBackend:
    enabled: true
    replicas: 1
    image: registry.k8s.io/defaultbackend-amd64:1.5
    resources:
      requests:
        cpu: 10m
        memory: 20Mi
  admissionWebhooks:
    enabled: true
    failurePolicy: ignore
    timeoutSeconds: 20
  metrics:
    enabled: true
    serviceMonitor: true
    serviceMonitorInterval: 45s
    serviceMonitorLabels:
      release: kube-prometheus-stack
  tcpServices:
    "5432": default/postgres:5432
  udpServices:
    "53": kube-system/kube-dns:53
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: dedicated
      operator: Equal
      value: ingress
      effect: NoSchedule
  priorityClassName: system-cluster-critical
  imageRegistry: registry.k8s.io
  helmValues: |
    controller:
      enableTopologyAwareRouting: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `4.15.1` |  |
| `spec.ingressClass` | `KubernetesIngressNginxIngressClass` |  |  |  |
| `spec.ingressClass.name` | `string` |  | `nginx` |  |
| `spec.ingressClass.isDefaultClass` | `bool` |  |  |  |
| `spec.ingressClass.controllerValue` | `string` |  |  |  |
| `spec.ingressClass.watchIngressWithoutClass` | `bool` |  |  |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.autoscaling` | `KubernetesIngressNginxAutoscaling` |  |  |  |
| `spec.autoscaling.enabled` | `bool` |  |  |  |
| `spec.autoscaling.minReplicas` | `int32` |  | `1` |  |
| `spec.autoscaling.maxReplicas` | `int32` |  | `11` |  |
| `spec.autoscaling.targetCpuUtilizationPercent` | `int32` |  | `50` |  |
| `spec.autoscaling.targetMemoryUtilizationPercent` | `int32` |  | `50` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.service` | `KubernetesIngressNginxService` |  |  |  |
| `spec.service.type` | `enum` |  | `load_balancer` |  |
| `spec.service.annotations` | `map<string, string>` |  |  |  |
| `spec.service.externalTrafficPolicy` | `enum` |  | `cluster` |  |
| `spec.service.loadBalancerSourceRanges` | `[]string` |  |  |  |
| `spec.service.loadBalancerClass` | `string` |  |  |  |
| `spec.service.enableHttp` | `bool` |  | `true` |  |
| `spec.service.enableHttps` | `bool` |  | `true` |  |
| `spec.service.httpNodePort` | `int32` |  |  |  |
| `spec.service.httpsNodePort` | `int32` |  |  |  |
| `spec.service.internal` | `KubernetesIngressNginxInternalService` |  |  |  |
| `spec.service.internal.enabled` | `bool` |  |  |  |
| `spec.service.internal.annotations` | `map<string, string>` |  |  |  |
| `spec.controllerKind` | `enum` |  | `deployment` |  |
| `spec.hostNetwork` | `bool` |  |  |  |
| `spec.hostPorts` | `bool` |  |  |  |
| `spec.nginxConfig` | `map<string, string>` |  |  |  |
| `spec.allowSnippetAnnotations` | `bool` |  |  |  |
| `spec.defaultTlsCertificate` | `KubernetesIngressNginxDefaultTlsCertificate` |  |  |  |
| `spec.defaultTlsCertificate.secretName` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.defaultTlsCertificate.namespace` | `string` |  |  |  |
| `spec.defaultBackend` | `KubernetesIngressNginxDefaultBackend` |  |  |  |
| `spec.defaultBackend.enabled` | `bool` |  |  |  |
| `spec.defaultBackend.replicas` | `int32` |  | `1` |  |
| `spec.defaultBackend.image` | `string` |  |  |  |
| `spec.defaultBackend.resources` | `ContainerResources` |  |  |  |
| `spec.defaultBackend.resources.limits` | `CpuMemory` |  |  |  |
| `spec.defaultBackend.resources.limits.cpu` | `string` |  |  |  |
| `spec.defaultBackend.resources.limits.memory` | `string` |  |  |  |
| `spec.defaultBackend.resources.requests` | `CpuMemory` |  |  |  |
| `spec.defaultBackend.resources.requests.cpu` | `string` |  |  |  |
| `spec.defaultBackend.resources.requests.memory` | `string` |  |  |  |
| `spec.admissionWebhooks` | `KubernetesIngressNginxAdmissionWebhooks` |  |  |  |
| `spec.admissionWebhooks.enabled` | `bool` |  | `true` |  |
| `spec.admissionWebhooks.failurePolicy` | `enum` |  | `fail` |  |
| `spec.admissionWebhooks.timeoutSeconds` | `int32` |  |  |  |
| `spec.metrics` | `KubernetesIngressNginxMetrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  |  |  |
| `spec.metrics.serviceMonitor` | `bool` |  |  |  |
| `spec.metrics.serviceMonitorInterval` | `string` |  | `30s` |  |
| `spec.metrics.serviceMonitorLabels` | `map<string, string>` |  |  |  |
| `spec.tcpServices` | `map<string, string>` |  |  |  |
| `spec.udpServices` | `map<string, string>` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.priorityClassName` | `string` |  |  |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install the controller into ("ingress-nginx" by
convention; a per-instance namespace like "ingress-nginx-internal" for
additional controllers). Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "4.15.1", which ships controller
v1.15.1 — chart and controller versions are released together but
numbered independently). Pin deliberately; upgrades re-run the release
with the new chart. Pick versions from the chart repository's index
(`helm search repo`): the served chart is the contract — the upstream
source tree's Chart.yaml can claim a version at a tag that was never
served.

- default: `4.15.1`

### spec.ingressClass

`KubernetesIngressNginxIngressClass`

The IngressClass this controller owns — how Ingress resources select
it. Defaults produce the cluster's standard "nginx" class.

### spec.ingressClass.name

`string` · optional (explicit presence)

IngressClass name — what KubernetesIngress resources put in
`ingress_class_name` to select this controller. Default "nginx".
MUST be unique per controller instance on a cluster: a second
controller needs its own class (e.g. "nginx-internal").
IngressClasses are immutable — changing the name replaces the class.

- default: `nginx`
- rule: {"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.ingressClass.isDefaultClass

`bool`

Mark this class as the cluster default: Ingress resources created
WITHOUT an ingressClassName get assigned to it. At most one default
class per cluster — the API server rejects Ingress creation when
several classes claim default.

### spec.ingressClass.controllerValue

`string`

The controller identifier inside the class (`spec.controller` of the
IngressClass, also passed to the controller as --controller-class).
Empty = derived: the chart default "k8s.io/ingress-nginx" for class
name "nginx", otherwise "k8s.io/<class-name>" so additional
controllers isolate automatically. Set explicitly only to adopt an
existing class vocabulary.

### spec.ingressClass.watchIngressWithoutClass

`bool`

Also reconcile Ingress resources that specify NO ingress class at all
(legacy objects predating IngressClass). Off by default; enable on at
most one controller per cluster.

### spec.replicas

`int32` · optional (explicit presence)

Controller replica count. Production entry points run 2+ for zero-drop
rollouts and node-failure tolerance; pair with `autoscaling` for
traffic-following capacity. With more than one replica (or autoscaling
min_replicas > 1) the chart automatically guards the controller with a
PodDisruptionBudget (minAvailable 1) — there is no separate toggle.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.autoscaling

`KubernetesIngressNginxAutoscaling`

Horizontal autoscaling for the controller Deployment (chart-managed
HPA). Requires metrics-server on the cluster for the utilization
targets to receive values. When enabled, `replicas` is ignored — the
HPA owns the replica count between min and max.

- rule: autoscaling min_replicas cannot exceed max_replicas

### spec.autoscaling.enabled

`bool`

Enable the chart's HPA. Requires metrics-server
(KubernetesMetricsServer) for utilization metrics to flow.

### spec.autoscaling.minReplicas

`int32` · optional (explicit presence)

Minimum replicas. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.autoscaling.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas. Chart default: 11.

- default: `11`
- rule: {"int32":{"gte":1}}

### spec.autoscaling.targetCpuUtilizationPercent

`int32` · optional (explicit presence)

Target average CPU utilization percentage. Chart default: 50.

- default: `50`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.autoscaling.targetMemoryUtilizationPercent

`int32` · optional (explicit presence)

Target average memory utilization percentage. Chart default: 50.

- default: `50`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.resources

`ContainerResources`

Controller container CPU/memory requests and limits. Chart default:
requests cpu=100m memory=90Mi, no limits (upstream recommends leaving
CPU unlimited — throttling an entry point converts load spikes into
cluster-wide latency).

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

### spec.service

`KubernetesIngressNginxService`

The controller Service — the cluster's traffic entry and THE cloud
integration surface. On managed clouds the `annotations` map is how the
environment provisions and shapes the load balancer.

### spec.service.type

`enum` · optional (explicit presence)

Service type. Default LoadBalancer — the host cloud provisions the
entry LB. Use node_port on bare metal (or with host_ports/DaemonSet),
cluster_ip for internal-only controllers fronted by something else.

- default: `load_balancer`

Allowed values (use exactly as shown):

- `load_balancer` -- Cloud load balancer (default) — the managed-cloud entry pattern.
- `node_port` -- Node ports only — bare-metal/edge, or behind an external LB you manage.
- `cluster_ip` -- Cluster-internal only — for east-west or mesh-fronted setups.

### spec.service.annotations

`map<string, string>`

Annotations on the controller Service — HOW THE HOST CLOUD SHAPES THE
LOAD BALANCER. This is the environment-injection surface for this
component; the controller itself never calls cloud APIs. Examples:
AWS NLB: {"service.beta.kubernetes.io/aws-load-balancer-type": "external",
          "service.beta.kubernetes.io/aws-load-balancer-nlb-target-type": "ip"};
GCP internal: {"networking.gke.io/load-balancer-type": "Internal"};
Azure internal: {"service.beta.kubernetes.io/azure-load-balancer-internal": "true"}.
Full per-cloud recipes in the kind's README.

The AWS recipe above requires the AWS Load Balancer Controller
installed in the cluster — the "external" type family is its
vocabulary, and without it the Service never receives an address (no
error surfaces; the annotations simply have no reader, and this
module's readiness wait then times out loudly). On clusters without
that controller, EKS's built-in cloud controller answers
{"service.beta.kubernetes.io/aws-load-balancer-type": "nlb"} instead —
both paths verified live.

### spec.service.externalTrafficPolicy

`enum` · optional (explicit presence)

External traffic policy. `local` preserves client source IPs and
avoids an extra hop — the usual production choice for LoadBalancer
controllers (health-check semantics differ; see README).

- default: `cluster`

Allowed values (use exactly as shown):

- `cluster` -- Kubernetes default: traffic may hop nodes (source IP is SNATed).
- `local` -- Deliver only to local controller pods, preserving the client source IP (the standard choice for an ingress controller behind a cloud LB).

### spec.service.loadBalancerSourceRanges

`[]string`

Restrict which source CIDRs may reach the load balancer
(`loadBalancerSourceRanges`). Empty = open to all sources.

- rule: {"repeated":{"items":{"cel":[{"id":"spec.service.source_range_cidr","message":"Each source range must be a CIDR, e.g. '10.0.0.0/8' or '203.0.113.4/32'","expression":"this.matches('^([0-9]{1,3}\\\\.){3}[0-9]{1,3}/[0-9]{1,2}$') || this.matches('^[0-9a-fA-F:]+/[0-9]{1,3}$')"}]}}}

### spec.service.loadBalancerClass

`string`

LoadBalancer implementation class (`loadBalancerClass`) — selects a
non-default LB controller where the cloud offers several (e.g.
"service.k8s.aws/nlb" for the AWS Load Balancer Controller). Empty =
the cloud's default implementation.

### spec.service.enableHttp

`bool` · optional (explicit presence)

Serve plain HTTP (port 80). Chart default: true. Disable for
HTTPS-only entry points (clients on port 80 then get connection
refused rather than a redirect — prefer NGINX's ssl-redirect when a
redirect is wanted).

- default: `true`

### spec.service.enableHttps

`bool` · optional (explicit presence)

Serve HTTPS (port 443). Chart default: true.

- default: `true`

### spec.service.httpNodePort

`int32` · optional (explicit presence)

Fixed HTTP node port (type node_port or load_balancer). Leave unset
to let the service controller allocate one from the cluster's
node-port range.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.service.httpsNodePort

`int32` · optional (explicit presence)

Fixed HTTPS node port. Leave unset to let the service controller
allocate one from the node-port range.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.service.internal

`KubernetesIngressNginxInternalService`

ALSO create the chart's second, INTERNAL load-balancer Service
alongside the external one — the single-controller dual-LB pattern
(same pods answer both a public and a private address). For fully
separate public/internal stacks, prefer two controller instances with
distinct ingress classes.

- rule: The internal controller Service requires at least one annotation (the chart refuses to render it bare — the cloud's internal-LB annotation is what makes it internal, e.g. service.beta.kubernetes.io/azure-load-balancer-internal: "true")

### spec.service.internal.enabled

`bool`

Create the internal Service.

### spec.service.internal.annotations

`map<string, string>`

Annotations for the internal Service — REQUIRED by the chart when
enabled, and this is where the cloud's internal-LB annotation goes
(e.g. {"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal"}).

### spec.controllerKind

`enum` · optional (explicit presence)

Run the controller as a Deployment (default) or a DaemonSet. DaemonSet
is the bare-metal/edge pattern — one controller per node, typically
paired with `host_network` or `host_ports` and a `node_port` or
headless service instead of a cloud LB.

- default: `deployment`

Allowed values (use exactly as shown):

- `deployment` -- Standard replicated Deployment (default; pairs with a LoadBalancer Service).
- `daemon_set` -- One controller pod per (selected) node — the bare-metal/edge pattern.

### spec.hostNetwork

`bool`

Run the controller pods on the host network (bare-metal pattern —
binds node IPs directly; the module sets dnsPolicy to
ClusterFirstWithHostNet so in-cluster name resolution keeps working).
Mutually exclusive with `host_ports` — host networking already exposes
every listener on the node.

### spec.hostPorts

`bool`

Expose the controller's HTTP/HTTPS listeners as hostPorts 80/443 on
each node running a controller pod (the CNI-friendly alternative to
`host_network` for bare-metal). Typically combined with
controller_kind=daemon_set.

### spec.nginxConfig

`map<string, string>`

Global NGINX tuning applied through the controller's ConfigMap —
upstream's own key/value vocabulary (proxy-body-size,
worker-shutdown-timeout, use-forwarded-headers, ssl-protocols, ...).
Keys and values are passed through verbatim; the authoritative key
list is the upstream ConfigMap reference. Kept upstream-shaped
deliberately — inventing a typed mirror of ~200 tuning keys would
drift from the controller's own documentation.

### spec.allowSnippetAnnotations

`bool`

Allow `*-snippet` annotations on Ingress resources (raw NGINX config
injection). Upstream default false since CVE-2021-25742: snippets let
any Ingress author run arbitrary NGINX directives in the shared
controller — enable only on clusters where every Ingress author is
trusted at controller level.

### spec.defaultTlsCertificate

`KubernetesIngressNginxDefaultTlsCertificate`

Cluster-wide default TLS certificate, served on HTTPS requests that
match no Ingress TLS block (and on the default backend). The
cert-manager seam: point it at a KubernetesCertificate's secret output
for an auto-renewed wildcard default.

### spec.defaultTlsCertificate.secretName

`string | valueFrom` · required

Name of the kubernetes.io/tls Secret holding the default certificate.
The cert-manager seam: reference a KubernetesCertificate's secret
output for an auto-renewed default.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.defaultTlsCertificate.namespace

`string`

Namespace of the Secret. Empty = the controller's installation
namespace.

### spec.defaultBackend

`KubernetesIngressNginxDefaultBackend`

Serve a catch-all backend for requests matching no Ingress rule
(chart's defaultBackend — returns 404/healthz). Off upstream by
default; the controller itself answers 404 without it. Enable when a
branded error page or separate error telemetry is wanted.

### spec.defaultBackend.enabled

`bool`

Deploy the default backend.

### spec.defaultBackend.replicas

`int32` · optional (explicit presence)

Default backend replica count. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.defaultBackend.image

`string`

Custom image serving the error responses (e.g. a branded error page).
Empty = the chart's registry.k8s.io/defaultbackend-amd64 image.

### spec.defaultBackend.resources

`ContainerResources`

Default backend container CPU/memory requests and limits.

### spec.defaultBackend.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.defaultBackend.resources.limits.cpu

`string`

### spec.defaultBackend.resources.limits.memory

`string`

### spec.defaultBackend.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.defaultBackend.resources.requests.cpu

`string`

### spec.defaultBackend.resources.requests.memory

`string`

### spec.admissionWebhooks

`KubernetesIngressNginxAdmissionWebhooks`

The admission webhook validating Ingress resources before they reach
the controller (rejects broken NGINX config at kubectl-apply time
instead of taking down the entry point). Enabled upstream by default.

### spec.admissionWebhooks.enabled

`bool` · optional (explicit presence)

Run the validating webhook. Chart default: true. Disable only where
hook Jobs are forbidden (the certgen patch Job is how the webhook
bootstraps its certificate).

- default: `true`

### spec.admissionWebhooks.failurePolicy

`enum` · optional (explicit presence)

What the API server does with Ingress writes when the webhook cannot
be reached.

- default: `fail`

Allowed values (use exactly as shown):

- `fail` -- Reject Ingress writes when the webhook is unavailable (chart default — safe: a broken Ingress cannot slip in during webhook downtime).
- `ignore` -- Admit Ingress writes when the webhook is unavailable (availability over safety — a bad Ingress can then break NGINX config reloads).

### spec.admissionWebhooks.timeoutSeconds

`int32` · optional (explicit presence)

Seconds the API server waits on the webhook before applying
failure_policy. Chart leaves the Kubernetes default (10s).

- rule: {"int32":{"lte":30,"gte":1}}

### spec.metrics

`KubernetesIngressNginxMetrics`

Prometheus metrics exposure for the controller (port 10254) and
ServiceMonitor creation for scrape discovery.

- rule: service_monitor requires metrics to be enabled — the ServiceMonitor would have no metrics endpoint to scrape

### spec.metrics.enabled

`bool`

Expose controller metrics (port 10254) and create the metrics Service.
Chart default: false.

### spec.metrics.serviceMonitor

`bool`

Create a ServiceMonitor for scrape discovery. Requires the Prometheus
operator CRDs (e.g. kube-prometheus-stack) on the cluster — the
release FAILS to install without them.

### spec.metrics.serviceMonitorInterval

`string` · optional (explicit presence)

Scrape interval for the ServiceMonitor. Chart default: "30s".

- default: `30s`

### spec.metrics.serviceMonitorLabels

`map<string, string>`

Extra labels on the ServiceMonitor — how a Prometheus instance's
serviceMonitorSelector finds it (e.g. {"release": "kube-prometheus-stack"}).

### spec.tcpServices

`map<string, string>`

Expose raw TCP services through the controller: map of external port →
"namespace/service:port" upstream target (upstream's
exposing-tcp-udp-services mechanism). The module publishes each port on
the controller Service automatically.

- rule: {"map":{"keys":{"cel":[{"id":"spec.tcp_services.port_key","message":"TCP service keys must be port numbers between 1 and 65535, e.g. \"5432\"","expression":"this.matches('^[0-9]+$') && int(this) >= 1 && int(this) <= 65535"}]},"values":{"cel":[{"id":"spec.tcp_services.target_format","message":"Each TCP service target must be 'namespace/service:port', e.g. 'default/postgres:5432'","expression":"this.matches('^[a-z0-9-]+/[a-z0-9-]+:[0-9]+(::PROXY)?$')"}]}}}

### spec.udpServices

`map<string, string>`

Expose raw UDP services through the controller: map of external port →
"namespace/service:port" upstream target.

- rule: {"map":{"keys":{"cel":[{"id":"spec.udp_services.port_key","message":"UDP service keys must be port numbers between 1 and 65535, e.g. \"53\"","expression":"this.matches('^[0-9]+$') && int(this) >= 1 && int(this) <= 65535"}]},"values":{"cel":[{"id":"spec.udp_services.target_format","message":"Each UDP service target must be 'namespace/service:port', e.g. 'kube-system/kube-dns:53'","expression":"this.matches('^[a-z0-9-]+/[a-z0-9-]+:[0-9]+$')"}]}}}

### spec.nodeSelector

`map<string, string>`

Node selector for controller pods. Chart default:
{"kubernetes.io/os": "linux"}.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for controller pods (e.g. a dedicated ingress node pool's
taint).

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

`string`

PriorityClass for controller pods. An entry point should usually
outrank ordinary workloads — consider a high-priority class (or
KubernetesPriorityClass) on production clusters.

### spec.imageRegistry

`string`

Registry serving the chart's images (controller, kube-webhook-certgen,
defaultbackend) — the air-gapped/mirror knob (`global.image.registry`).
Empty = registry.k8s.io.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics, identical
on both engines). For the chart surface beyond the typed fields —
never the substitute for them. Do not put secrets here.

## Validation Rules

- `spec.host_network_xor_host_ports`: host_network and host_ports are alternatives — host networking already binds every listener on the node, so hostPorts would be redundant. Choose one.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesIngressNginx, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Kubernetes namespace the controller was installed into. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name; controller resources are named "<release>-controller"). |
| `status.outputs.ingress_class_name` | `string` | Name of the IngressClass this controller owns — what KubernetesIngress resources reference in `ingress_class_name` to route through this controller. |
| `status.outputs.controller_service_name` | `string` | Name of the controller's external Service (the traffic entry point; type per spec.service.type). External-dns and manual DNS records point at this Service's address. |
| `status.outputs.internal_service_name` | `string` | Name of the controller's internal Service — populated only when spec.service.internal.enabled is true. |
| `status.outputs.load_balancer_ip` | `string` | External IP address of the controller's LoadBalancer, once the host cloud provisions it (GCP/Azure populate an IP). Empty on clusters without a cloud LB controller (e.g. kind) and on providers that populate a hostname instead. |
| `status.outputs.load_balancer_hostname` | `string` | External hostname of the controller's LoadBalancer (AWS ELB/NLB populate a DNS name). Empty when the provider populates an IP, and on clusters without a cloud LB controller. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.defaultTlsCertificate.secretName` | KubernetesCertificate | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
