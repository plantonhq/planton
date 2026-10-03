# Kubernetes Cilium

## When NOT to Use This

**One installation per cluster.** Cilium is the node dataplane — the agent
DaemonSet, the operator, and the generated CNI configuration are cluster
singletons, so the Helm release name is fixed to `cilium`. Check whether the
cluster already runs Cilium (or another CNI you intend to keep) before
adding this component.

Also not the right kind when:

- **You only need Kubernetes NetworkPolicy on a managed cloud whose native
  CNI already enforces it** — some managed offerings ship policy enforcement
  without Cilium; adding a second enforcement layer buys complexity, not
  security.
- **You want an ingress/gateway product, not a CNI** — Cilium includes a
  Gateway API implementation, but installing the cluster dataplane just for
  north-south routing is the wrong trade; a dedicated Gateway controller is
  lighter.
- **The cluster was created WITH a default CNI and you cannot chain** —
  running Cilium as the primary CNI requires the cluster to be created
  without one (kind: `disableDefaultCNI: true`; AKS: `--network-plugin
  none`; kubeadm: no CNI addon). If the incumbent must stay, use CNI
  chaining instead (see below).

## Overview

**KubernetesCilium** installs Cilium — the eBPF-based networking,
network-security, and observability engine — from the official Helm chart
(`cilium` at `https://helm.cilium.io`). Cilium is the cluster's CNI: it
wires pod networking, enforces NetworkPolicy (standard Kubernetes policies
plus Cilium's own L7-aware policies), can replace kube-proxy entirely with
eBPF service load-balancing, and streams flow-level observability through
Hubble.

**The load-bearing choice — two ways to run it:**

1. **Primary CNI** (default): Cilium owns pod networking. The cluster must
   be created WITHOUT another CNI; nodes sit NotReady until Cilium installs
   — that is by design, and the install itself is what unblocks scheduling.
2. **CNI chaining** (`cni.chaining_mode`): Cilium runs ON TOP of an existing
   CNI — the incumbent keeps IPAM and basic routing while Cilium attaches
   eBPF programs for policy enforcement, load-balancing, and observability.
   This is the no-rip-and-replace path (e.g. EKS with the AWS VPC CNI via
   `aws-cni`).

The typed spec covers the chart's meaningful configuration surface, with a
`helm_values` escape hatch (merged last, Helm `-f` semantics, identical on
both engines) for anything beyond it.

**Key design points:**

- **The release name is fixed to `cilium`** — one dataplane per cluster is
  an upstream constraint, so the release name never derives from
  `metadata.name`. The chart names its workloads with fixed names too
  (DaemonSet `cilium`, Deployment `cilium-operator`).
- **Install into `kube-system`** — the upstream convention; the agent is
  cluster infrastructure and several chart defaults assume it
  (`create_namespace` stays false — kube-system always exists).
- **Cross-field rules are CEL-enforced in the spec**, so misconfigurations
  fail at validation time instead of producing a half-working dataplane:
  `gateway_api` requires `kube_proxy_replacement` (the operator otherwise
  disables Gateway API support with only a log warning);
  `kube_proxy_replacement` requires `k8s_service_host` (without kube-proxy
  the agent cannot resolve the in-cluster API server address); chaining
  requires `cni.exclusive: false` (exclusive mode would rename the very CNI
  configuration the chain depends on); at most one cloud integration arm.
- **The install waits for the whole dataplane** — both engines install
  atomically with a 600-second timeout (not the usual 300): the agent must
  roll out on every node, and on a fresh cluster nodes transition
  NotReady→Ready only as Cilium wires each one.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: installation namespace (`kube-system` is the
  upstream convention) — literal or a KubernetesNamespace reference

### Cluster identity and chart

- **`spec.chart_version`**: pinned chart version (default `1.19.6`; Cilium
  chart and app versions move together)
- **`spec.cluster_name`**: the name this cluster carries in Hubble flows and
  any future Cluster Mesh (default `default`; keep it unique across clusters
  you may ever mesh)

### Networking posture

- **`spec.ipam`**: who hands out pod IPs — `cluster-pool` (chart default;
  the operator carves per-node CIDRs from `cluster_pool_ipv4_pod_cidrs`),
  `kubernetes` (use the node's Kubernetes-assigned PodCIDR — the right mode
  on kind and kubeadm), `eni` (AWS ENI, pairs with `cloud.aws_eni`),
  `azure`, `alibabacloud`, `multi-pool`, or `delegated-plugin` (chaining)
- **`spec.routing`**: `tunnel` (encapsulation — vxlan or geneve — works on
  any network; chart default) or `native` (route pod CIDRs on the fabric;
  set `ipv4_native_routing_cidr`, optionally `auto_direct_node_routes` on a
  shared L2 segment)
- **`spec.kube_proxy_replacement` + `spec.k8s_service_host` /
  `k8s_service_port`**: replace kube-proxy with eBPF service
  load-balancing. The cluster must be created without kube-proxy, and the
  agent needs the API server's real address before any service
  load-balancing exists
- **`spec.cni`**: chaining mode (`aws-cni`, `flannel`, `generic-veth`,
  `portmap`), chaining target, and `exclusive` (must be false when
  chaining)
- **`spec.cloud`**: at most one of `aws_eni` (ENI datapath as primary CNI),
  `aks_byocni` (AKS clusters created with `--network-plugin none`), `gke`
  (GKE node-image/routing integration)

### Security and observability

- **`spec.policy_enforcement_mode`**: `default` (enforce only where policies
  select pods), `always` (default-deny), or `never` (observe only)
- **`spec.hubble`**: flow observability — enabled in the agent by default;
  add `relay` (cluster-wide aggregation), `ui` (service map; requires
  relay), `metrics` (metric families like `dns`, `drop`, `tcp`, `flow`,
  `http`), and `metrics_service_monitor`
- **`spec.encryption`**: transparent pod-to-pod encryption — `wireguard`
  (automatic key management) or `ipsec`; `node_encryption` (WireGuard only)
- **`spec.gateway_api`**: Cilium's Gateway API implementation (creates the
  `cilium` GatewayClass; requires `kube_proxy_replacement` and the Gateway
  API CRDs on the cluster)
- **`spec.bandwidth_manager`**: enforce pod egress-bandwidth annotations in
  eBPF; optional BBR congestion control
- **`spec.prometheus`**: agent + operator `/metrics` endpoints and optional
  ServiceMonitors (require the Prometheus operator CRDs — the release fails
  to install without them)

### Sizing

- **`spec.operator.replicas`**: chart default 2 (HA, leader-elected, spread
  by pod anti-affinity) — run 1 on single-node clusters or the rollout never
  settles
- **`spec.agent_resources` / `spec.operator.resources`**: container
  resources (empty = chart defaults, which set none — size deliberately for
  production)
- **`spec.helm_values`**: escape hatch for the chart's long tail (BGP
  control plane, L2 announcements, Cluster Mesh, ingress controller, image
  overrides, ...) — never the primary interface

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Installation namespace |
| `release_name` | Helm release name (always `cilium`) |
| `cluster_name` | Cluster identity Cilium runs under — the name in Hubble flows and any future Cluster Mesh |
| `hubble_relay_service_name` | `hubble-relay` when `hubble.relay` is enabled; empty otherwise |
| `hubble_ui_service_name` | `hubble-ui` when `hubble.ui` is enabled; empty otherwise — port-forward it to open the service map |
| `gateway_class_name` | `cilium` when `gateway_api` is enabled; empty otherwise |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind KubernetesNamespace,
  field path `spec.name`) — though for the conventional `kube-system`
  install a literal `value: kube-system` is the norm.
- **With `gateway_api: true`**, KubernetesGateway resources reference the
  registered GatewayClass through this component's `gateway_class_name`
  output. Deploy KubernetesGatewayApiCrds first — Cilium implements the
  Gateway API but does not ship its CRDs.
- **With `prometheus.service_monitor` or `hubble.metrics_service_monitor`**,
  the Prometheus operator CRDs must already be on the cluster or the release
  fails to install.
- **NetworkPolicy-dependent workloads** need no reference: once Cilium is
  the enforcing CNI, standard NetworkPolicy resources on the cluster start
  being enforced.

## Examples

### Minimal (kind / single-node dev cluster, primary CNI)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCilium
metadata:
  name: cilium
spec:
  namespace:
    value: kube-system
  ipam:
    mode: kubernetes # the control plane already allocates per-node PodCIDRs
  operator:
    replicas: 1 # the chart's 2-replica default cannot schedule on one node
```

### Advanced (kube-proxy-free primary CNI with full observability)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCilium
metadata:
  name: cilium
spec:
  namespace:
    value: kube-system
  clusterName: prod-us-east
  ipam:
    mode: cluster-pool
    clusterPoolIpv4PodCidrs:
      - 10.42.0.0/16
    clusterPoolIpv4MaskSize: 24
  routing:
    mode: tunnel
    tunnelProtocol: vxlan
  kubeProxyReplacement: true
  k8sServiceHost: api.prod-us-east.example.com
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
      - http
  prometheus:
    enabled: true
    serviceMonitor: true # requires the Prometheus operator CRDs
  encryption:
    enabled: true
    type: wireguard
  bandwidthManager:
    enabled: true
  policyEnforcementMode: default
  operator:
    replicas: 2
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        memory: 256Mi
  agentResources:
    requests:
      cpu: 200m
      memory: 512Mi
    limits:
      memory: 1Gi
```

### EKS chaining (keep the AWS VPC CNI, add policy + observability)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCilium
metadata:
  name: cilium
spec:
  namespace:
    value: kube-system
  cni:
    chainingMode: aws-cni # the incumbent keeps IPAM and interface wiring
    exclusive: false      # mandatory with chaining (CEL-enforced)
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
