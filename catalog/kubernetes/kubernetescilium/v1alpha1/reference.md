# KubernetesCilium

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesCiliumSpec** installs Cilium — the eBPF-based networking,
network-security, and observability engine — from the official Helm chart
(`cilium` at https://helm.cilium.io). Cilium is the cluster's CNI: it wires
pod networking, enforces NetworkPolicy (standard Kubernetes policies plus
Cilium's own L7-aware policies), can replace kube-proxy entirely with eBPF
service load-balancing, and streams flow-level observability through
Hubble.

ONE INSTALLATION PER CLUSTER: Cilium is the node dataplane — the agent
DaemonSet, operator, and generated CNI configuration are cluster
singletons, so the Helm release name is fixed to "cilium". Install it into
"kube-system" (the upstream convention; the agent is cluster
infrastructure).

TWO WAYS TO RUN IT (the load-bearing choice):

1. PRIMARY CNI (default): Cilium owns pod networking. The cluster must be
   created WITHOUT another CNI (kind: `disableDefaultCNI: true`; AKS:
   BYOCNI via `aks_byocni`; self-managed kubeadm: no CNI addon). Nodes stay
   NotReady until Cilium installs — that is by design.

2. CNI CHAINING (`cni.chaining_mode`): Cilium runs ON TOP of an existing
   CNI — the incumbent keeps doing IPAM and basic routing while Cilium
   attaches eBPF programs for policy enforcement, load-balancing, and
   observability. This is the no-rip-and-replace path for clusters whose
   networking must stay (e.g. EKS with the AWS VPC CNI via `aws-cni`).

The typed fields below cover the chart's meaningful configuration surface;
`helm_values` remains as the escape hatch for the chart's long tail
(merged last, Helm `-f` semantics, identical on both engines) — a safety
valve, never the primary interface.

## Example

```yaml
# Full-surface test manifest: exercises every typed arm of the spec so the
# offline plan proofs cover what the live lanes may not. Kind-style posture
# (primary CNI on a cluster created with disableDefaultCNI +
# kubeProxyMode none); cloud and cni chaining stay unset — they are
# mutually-exclusive postures with this one. Not a realistic production
# shape — see presets for those.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCilium
metadata:
  name: hack-cilium
spec:
  namespace:
    value: kube-system
  # kube-system always exists — the module must not try to create it.
  createNamespace: false
  chartVersion: "1.19.6"
  clusterName: hack-kind
  ipam:
    # kind's control plane already allocates per-node PodCIDRs — use them
    # (cluster-pool knobs only apply to cluster-pool mode, so they stay
    # unset here).
    mode: kubernetes
  routing:
    # tunnel mode: the native-routing knobs (ipv4NativeRoutingCidr,
    # autoDirectNodeRoutes) are CEL-fenced to native mode and stay unset.
    mode: tunnel
    tunnelProtocol: vxlan
  # Replace kube-proxy with eBPF — the cluster is created with
  # kubeProxyMode none, so the agent needs the API server address before
  # any service load-balancing exists (the control-plane node's container
  # name on kind).
  kubeProxyReplacement: true
  k8sServiceHost: hack-cilium-control-plane
  k8sServicePort: 6443
  hubble:
    enabled: true
    relay: true
    ui: true
    metrics:
      - dns
      - drop
      - tcp
      - flow
      - icmp
      - http
    # No Prometheus operator CRDs on the hack cluster — the release would
    # fail to install with a ServiceMonitor.
    metricsServiceMonitor: false
  encryption:
    enabled: true
    type: wireguard
    nodeEncryption: true
  policyEnforcementMode: default
  # Requires kubeProxyReplacement (set above) and the Gateway API CRDs on
  # the cluster.
  gatewayApi: true
  bandwidthManager:
    enabled: true
    bbr: true
  operator:
    # Single-node kind cluster: the chart-default 2 replicas cannot both
    # schedule under pod anti-affinity and the rollout never settles.
    replicas: 1
    resources:
      requests:
        cpu: 50m
        memory: 128Mi
      limits:
        memory: 256Mi
  agentResources:
    requests:
      cpu: 100m
      memory: 512Mi
    limits:
      memory: 1Gi
  prometheus:
    enabled: true
    # Same CRD constraint as hubble.metricsServiceMonitor.
    serviceMonitor: false
  helmValues: |
    debug:
      enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.19.6` |  |
| `spec.clusterName` | `string` |  | `default` |  |
| `spec.ipam` | `KubernetesCiliumIpam` |  |  |  |
| `spec.ipam.mode` | `string` |  | `cluster-pool` |  |
| `spec.ipam.clusterPoolIpv4PodCidrs` | `[]string` |  |  |  |
| `spec.ipam.clusterPoolIpv4MaskSize` | `int32` |  |  |  |
| `spec.routing` | `KubernetesCiliumRouting` |  |  |  |
| `spec.routing.mode` | `string` |  | `tunnel` |  |
| `spec.routing.tunnelProtocol` | `string` |  |  |  |
| `spec.routing.ipv4NativeRoutingCidr` | `string` |  |  |  |
| `spec.routing.autoDirectNodeRoutes` | `bool` |  |  |  |
| `spec.kubeProxyReplacement` | `bool` |  |  |  |
| `spec.k8sServiceHost` | `string` |  |  |  |
| `spec.k8sServicePort` | `int32` |  |  |  |
| `spec.cni` | `KubernetesCiliumCni` |  |  |  |
| `spec.cni.chainingMode` | `string` |  |  |  |
| `spec.cni.chainingTarget` | `string` |  |  |  |
| `spec.cni.exclusive` | `bool` |  | `true` |  |
| `spec.cloud` | `KubernetesCiliumCloudIntegration` |  |  |  |
| `spec.cloud.awsEni` | `bool` |  |  |  |
| `spec.cloud.aksByocni` | `bool` |  |  |  |
| `spec.cloud.gke` | `bool` |  |  |  |
| `spec.hubble` | `KubernetesCiliumHubble` |  |  |  |
| `spec.hubble.enabled` | `bool` |  | `true` |  |
| `spec.hubble.relay` | `bool` |  |  |  |
| `spec.hubble.ui` | `bool` |  |  |  |
| `spec.hubble.metrics` | `[]string` |  |  |  |
| `spec.hubble.metricsServiceMonitor` | `bool` |  |  |  |
| `spec.encryption` | `KubernetesCiliumEncryption` |  |  |  |
| `spec.encryption.enabled` | `bool` |  |  |  |
| `spec.encryption.type` | `string` |  | `wireguard` |  |
| `spec.encryption.nodeEncryption` | `bool` |  |  |  |
| `spec.policyEnforcementMode` | `string` |  | `default` |  |
| `spec.gatewayApi` | `bool` |  |  |  |
| `spec.bandwidthManager` | `KubernetesCiliumBandwidthManager` |  |  |  |
| `spec.bandwidthManager.enabled` | `bool` |  |  |  |
| `spec.bandwidthManager.bbr` | `bool` |  |  |  |
| `spec.operator` | `KubernetesCiliumOperator` |  |  |  |
| `spec.operator.replicas` | `int32` |  | `2` |  |
| `spec.operator.resources` | `ContainerResources` |  |  |  |
| `spec.operator.resources.limits` | `CpuMemory` |  |  |  |
| `spec.operator.resources.limits.cpu` | `string` |  |  |  |
| `spec.operator.resources.limits.memory` | `string` |  |  |  |
| `spec.operator.resources.requests` | `CpuMemory` |  |  |  |
| `spec.operator.resources.requests.cpu` | `string` |  |  |  |
| `spec.operator.resources.requests.memory` | `string` |  |  |  |
| `spec.agentResources` | `ContainerResources` |  |  |  |
| `spec.agentResources.limits` | `CpuMemory` |  |  |  |
| `spec.agentResources.limits.cpu` | `string` |  |  |  |
| `spec.agentResources.limits.memory` | `string` |  |  |  |
| `spec.agentResources.requests` | `CpuMemory` |  |  |  |
| `spec.agentResources.requests.cpu` | `string` |  |  |  |
| `spec.agentResources.requests.memory` | `string` |  |  |  |
| `spec.prometheus` | `KubernetesCiliumPrometheus` |  |  |  |
| `spec.prometheus.enabled` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install Cilium into ("kube-system" is the upstream
convention — the agent is cluster infrastructure and several chart
defaults assume it). Accepts a literal namespace name or a reference to
a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist (kube-system always does).

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "1.19.6" — Cilium chart and app
versions move together). Pin deliberately; upgrades re-run the release
with the new chart. Pick versions from the chart repository's index
(`helm search repo`): the served chart is the contract — the upstream
source tree's Chart.yaml can claim a version at a tag that was never
served.

- default: `1.19.6`

### spec.clusterName

`string` · optional (explicit presence)

Cluster identity for Cilium: the name distinguishes this cluster in
multi-cluster (Cluster Mesh) setups and appears in Hubble flow data.
Chart default: "default". Keep it unique across clusters you may ever
mesh together.

- default: `default`
- rule: {"string":{"maxLen":"32","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.ipam

`KubernetesCiliumIpam`

IP Address Management: who hands out pod IPs.

### spec.ipam.mode

`string` · optional (explicit presence)

IPAM mode. "cluster-pool" (chart default): Cilium's operator carves
per-node pod CIDRs from cluster_pool_ipv4_pod_cidrs. "kubernetes":
use the node's Kubernetes-assigned PodCIDR (the right mode on kind and
kubeadm clusters, whose CIDRs the control plane already allocates).
"eni": AWS ENI mode (pods get VPC IPs; pairs with cloud.aws_eni).
"azure": Azure IPAM. "alibabacloud": Alibaba ENI. "multi-pool":
multiple user-defined pools. "delegated-plugin": IPAM delegated to the
chained CNI (chaining setups).

- default: `cluster-pool`
- rule: ipam mode must be one of 'cluster-pool', 'kubernetes', 'eni', 'azure', 'alibabacloud', 'multi-pool', or 'delegated-plugin'

### spec.ipam.clusterPoolIpv4PodCidrs

`[]string`

IPv4 CIDRs the operator carves per-node pod ranges from (cluster-pool
mode only). Chart default: ["10.0.0.0/8"]. Must not overlap the node
network or Service CIDR.

- rule: {"repeated":{"items":{"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$"}}}}

### spec.ipam.clusterPoolIpv4MaskSize

`int32` · optional (explicit presence)

Per-node mask size carved out of the cluster pool (cluster-pool mode
only). Chart default: 24 (254 pod IPs per node).

- rule: {"int32":{"lte":30,"gte":8}}

### spec.routing

`KubernetesCiliumRouting`

Datapath routing between nodes: tunnel (encapsulation, works
everywhere) or native (routes pod CIDRs on the underlying network —
faster, needs a routable fabric).

- rule: ipv4_native_routing_cidr and auto_direct_node_routes only apply when routing mode is 'native' — remove them or switch the mode

### spec.routing.mode

`string` · optional (explicit presence)

Routing mode. "tunnel" (chart default): encapsulate pod traffic between
nodes — works on any network. "native": route pod CIDRs directly on the
underlying fabric — lower overhead, requires the network to carry pod
CIDRs (set ipv4_native_routing_cidr, and auto_direct_node_routes on a
shared L2 segment).

- default: `tunnel`
- rule: routing mode must be either 'tunnel' or 'native'

### spec.routing.tunnelProtocol

`string` · optional (explicit presence)

Encapsulation protocol in tunnel mode. Chart default: "vxlan";
"geneve" is the alternative (required by some datapath features).

- rule: tunnel_protocol must be either 'vxlan' or 'geneve'

### spec.routing.ipv4NativeRoutingCidr

`string`

CIDR within which traffic is NOT masqueraded in native-routing mode —
the pod CIDR range the fabric can route back.

- rule: ipv4_native_routing_cidr must be an IPv4 CIDR (e.g. 10.0.0.0/8)

### spec.routing.autoDirectNodeRoutes

`bool`

In native mode, have each agent install direct routes to the pod CIDRs
of other nodes on the same L2 segment (autoDirectNodeRoutes).

### spec.kubeProxyReplacement

`bool`

Replace kube-proxy with Cilium's eBPF service load-balancing
(kubeProxyReplacement). The cluster should then be created WITHOUT
kube-proxy (kind: `kubeProxyMode: none`; kubeadm:
`--skip-phases=addon/kube-proxy`). When replacing kube-proxy, also set
k8s_service_host/k8s_service_port so the agent can reach the API server
before any service load-balancing exists.

### spec.k8sServiceHost

`string`

API server address the agent uses BEFORE service load-balancing is up —
required with kube_proxy_replacement (there is no kube-proxy to resolve
the in-cluster kubernetes.default Service). On kind this is the
control-plane node's container name; on managed clouds the API endpoint
hostname.

### spec.k8sServicePort

`int32` · optional (explicit presence)

API server port paired with k8s_service_host. Chart treats empty as
unset; 6443 is the common value.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.cni

`KubernetesCiliumCni`

CNI installation behavior — including CHAINING MODE, the arm that runs
Cilium on top of an existing CNI instead of replacing it.

- rule: cni.exclusive must be false when chaining_mode is set — exclusive mode renames the very CNI configuration the chain depends on

### spec.cni.chainingMode

`string` · optional (explicit presence)

Chain Cilium onto an existing CNI instead of replacing it. The
incumbent CNI keeps IPAM and interface wiring; Cilium attaches eBPF for
policy, load-balancing, and observability. Values (upstream closed
set): "none" (no chaining — primary CNI), "aws-cni" (EKS with the AWS
VPC CNI), "flannel", "generic-veth" (any veth-based CNI),
"portmap" (add hostPort support via the portmap plugin).

- rule: cni chaining_mode must be one of 'none', 'aws-cni', 'flannel', 'generic-veth', or 'portmap'

### spec.cni.chainingTarget

`string`

CNI network name to chain into (implies generic-veth chaining): the
agent watches for a CNI network with this name and layers Cilium onto
its configuration. aws-cni chaining implies target "aws-cni" — no need
to set it.

### spec.cni.exclusive

`bool` · optional (explicit presence)

Make Cilium take ownership of /etc/cni/net.d, renaming non-Cilium CNI
configurations so no pod can schedule through another CNI during agent
downtime. Chart default: true. MUST be false in chaining mode — the
chained CNI's configuration has to survive.

- default: `true`

### spec.cloud

`KubernetesCiliumCloudIntegration`

Cloud-provider datapath integrations (AWS ENI, Azure BYOCNI, GKE).
Exactly the environment-injection surface: the same spec deploys to any
cluster, these arms adapt the datapath to the host cloud.

- rule: at most one cloud integration (aws_eni, aks_byocni, gke) can be enabled — they configure mutually exclusive datapaths

### spec.cloud.awsEni

`bool`

AWS ENI datapath (EKS/self-managed on EC2 as PRIMARY CNI): pods receive
VPC-routable IPs from ENIs Cilium manages. Pair with ipam mode "eni".
The agent uses the node instance role (or IRSA) for EC2 ENI calls. To
instead keep the AWS VPC CNI and chain Cilium on top, use
cni.chaining_mode "aws-cni" and leave this off.

### spec.cloud.aksByocni

`bool`

Azure AKS BYOCNI (aksbyocni): the supported way to run Cilium as the
primary CNI on AKS clusters created with `--network-plugin none`.
Enables the Azure-compatible encapsulation defaults. (The legacy
azure-IPAM integration with service-principal credentials is
deliberately not typed — reach it via helm_values if you must.)

### spec.cloud.gke

`bool`

GKE datapath integration (gke.enabled): configures the agent for GKE
node images and routing (pairs with ipam mode "kubernetes" and native
routing per upstream's GKE guide).

### spec.hubble

`KubernetesCiliumHubble`

Hubble — Cilium's flow-level observability layer (relay, UI, metrics).
Enabled by default upstream.

- rule: hubble.ui requires hubble.relay — the UI reads flows exclusively through the relay service
- rule: hubble.relay/ui/metrics require hubble.enabled — with Hubble off in the agent there are no flows to serve
- rule: hubble.metrics_service_monitor requires at least one entry in hubble.metrics — the ServiceMonitor would have no endpoint to scrape

### spec.hubble.enabled

`bool` · optional (explicit presence)

Enable Hubble in the agent (chart default: true). Disabling removes
flow observability entirely.

- default: `true`

### spec.hubble.relay

`bool`

Deploy hubble-relay: the cluster-wide flow aggregation service the UI
and hubble CLI talk to.

### spec.hubble.ui

`bool`

Deploy the Hubble UI (service-map web console). Requires relay.

### spec.hubble.metrics

`[]string`

Hubble metric families to export (e.g. "dns", "drop", "tcp", "flow",
"icmp", "http" — entries accept upstream's option syntax like
"dns:query;ignoreAAAA"). Empty = Hubble metrics disabled (the chart
default).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.hubble.metricsServiceMonitor

`bool`

Create a ServiceMonitor for the Hubble metrics endpoint. Requires the
Prometheus operator CRDs on the cluster — the release FAILS to install
without them. Only meaningful when metrics is non-empty.

### spec.encryption

`KubernetesCiliumEncryption`

Transparent encryption of pod-to-pod traffic (WireGuard or IPsec).

- rule: node_encryption is only effective with encryption type 'wireguard' (the upstream constraint) — remove it or switch the type
- rule: node_encryption requires encryption.enabled — there is no node-to-node encryption without the encryption datapath on

### spec.encryption.enabled

`bool`

Enable transparent network encryption.

### spec.encryption.type

`string` · optional (explicit presence)

Encryption method: "wireguard" (kernel WireGuard, key management
automatic — the low-friction choice) or "ipsec" (requires a
pre-created key Secret, supports key rotation policies). Upstream also
accepts "ztunnel" for Istio ambient interop — deliberately not typed;
reach it via helm_values with an ambient mesh.

WireGuard rides the NODE kernel's module: verify the node OS ships it
(mainstream cloud images do; container-VM and older distro kernels may
not) — on a node without the module the agent fails to start rather
than silently sending plaintext.

- default: `wireguard`
- rule: encryption type must be either 'wireguard' or 'ipsec'

### spec.encryption.nodeEncryption

`bool`

Also encrypt pure node-to-node traffic (not just pod traffic).
WireGuard only.

### spec.policyEnforcementMode

`string` · optional (explicit presence)

NetworkPolicy enforcement posture (policyEnforcementMode). "default"
enforces only where policies select pods; "always" denies everything
not explicitly allowed; "never" disables enforcement (observability
only). Closed upstream enum.

- default: `default`
- rule: policy_enforcement_mode must be one of 'default', 'always', or 'never'

### spec.gatewayApi

`bool`

Cilium's Gateway API implementation: istiod-style north-south routing
with Envoy embedded in the agent. Creates the "cilium" GatewayClass.
REQUIRES kube_proxy_replacement — the operator disables Gateway API
support (with only a log warning) when kube-proxy replacement is off,
so this spec makes the dependency loud instead of silent. Requires the
Gateway API CRDs on the cluster (KubernetesGatewayApiCrds).

### spec.bandwidthManager

`KubernetesCiliumBandwidthManager`

eBPF bandwidth manager: enforces pod egress bandwidth annotations and
enables BBR congestion control for pods.

- rule: bbr requires the bandwidth manager to be enabled

### spec.bandwidthManager.enabled

`bool`

Enforce pod egress bandwidth (kubernetes.io/egress-bandwidth
annotations) in eBPF instead of noisy token-bucket qdiscs.

### spec.bandwidthManager.bbr

`bool`

Use BBR congestion control for pods (requires a 5.18+ kernel per
upstream guidance).

### spec.operator

`KubernetesCiliumOperator`

cilium-operator sizing. The operator handles IPAM allocation and CRD
lifecycle; chart default is 2 replicas for HA (they lead-elect). On
single-node clusters run 1 — two replicas cannot both schedule with
pod anti-affinity.

### spec.operator.replicas

`int32` · optional (explicit presence)

Operator replicas. Chart default: 2 (HA via leader election, spread by
pod anti-affinity). Run 1 on single-node clusters — the second replica
cannot schedule and the rollout never settles.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.operator.resources

`ContainerResources`

Operator container resources. Empty = chart defaults.

### spec.operator.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.operator.resources.limits.cpu

`string`

### spec.operator.resources.limits.memory

`string`

### spec.operator.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.operator.resources.requests.cpu

`string`

### spec.operator.resources.requests.memory

`string`

### spec.agentResources

`ContainerResources`

Agent (cilium DaemonSet) container resources. Empty = chart defaults
(no requests/limits — size deliberately for production).

### spec.agentResources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.agentResources.limits.cpu

`string`

### spec.agentResources.limits.memory

`string`

### spec.agentResources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.agentResources.requests.cpu

`string`

### spec.agentResources.requests.memory

`string`

### spec.prometheus

`KubernetesCiliumPrometheus`

Cilium's own Prometheus telemetry (agent + operator metrics and
optional ServiceMonitors). Hubble METRICS are separate — see
hubble.metrics.

- rule: service_monitor requires prometheus metrics to be enabled — the ServiceMonitor would have no metrics endpoint to scrape

### spec.prometheus.enabled

`bool`

Expose agent and operator /metrics endpoints.

### spec.prometheus.serviceMonitor

`bool`

Create ServiceMonitors for scrape discovery. Requires the Prometheus
operator CRDs on the cluster — the release FAILS to install without
them.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics, identical
on both engines). For the chart surface beyond the typed fields (BGP
control plane, L2 announcements, Cluster Mesh, ingress controller,
per-image overrides, ...) — never the substitute for them. Do not put
secrets here.

## Validation Rules

- `spec.gateway_api_requires_kpr`: gateway_api requires kube_proxy_replacement — Cilium's operator disables Gateway API support when kube-proxy replacement is off (it only logs a warning, so the gateway would silently never program)
- `spec.kpr_requires_api_server_address`: kube_proxy_replacement requires k8s_service_host — without kube-proxy the agent has no service load-balancer yet and cannot resolve the in-cluster API server address

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesCilium, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace Cilium was installed into (the resolved spec.namespace). |
| `status.outputs.release_name` | `string` | Helm release name — fixed "cilium" (one dataplane per cluster). |
| `status.outputs.cluster_name` | `string` | Cluster identity Cilium runs under (the resolved spec.cluster_name) — the name this cluster is known by in Hubble flows and any future Cluster Mesh. |
| `status.outputs.hubble_relay_service_name` | `string` | Name of the hubble-relay Service (fixed "hubble-relay" by the chart) when hubble.relay is enabled; empty otherwise. The address the hubble CLI and UI read flows from. |
| `status.outputs.hubble_ui_service_name` | `string` | Name of the hubble-ui Service (fixed "hubble-ui" by the chart) when hubble.ui is enabled; empty otherwise — port-forward it to open the service map. |
| `status.outputs.gateway_class_name` | `string` | Name of the GatewayClass Cilium's Gateway API implementation registers (fixed "cilium" by the chart) when gateway_api is enabled; empty otherwise. KubernetesGateway resources reference it via gateway_class_name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
