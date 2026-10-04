# KubernetesIstio

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesIstioSpec** installs the Istio service-mesh CONTROL PLANE from the
official Helm charts (`base` + `istiod`, plus `cni` + `ztunnel` in ambient mode,
all from https://istio-release.storage.googleapis.com/charts). istiod is the
mesh's brain: it validates mesh configuration, issues workload certificates, and
programs the data plane (sidecar proxies, ambient node proxies, and gateways).

This component deploys NO gateways and NO ingress. Istio implements the
Kubernetes Gateway API natively: creating a KubernetesGateway with
`gateway_class_name: istio` makes istiod provision and program the gateway
deployment automatically, and the route kinds (KubernetesHttpRoute, ...) attach
to it. North-south exposure therefore composes from the Gateway API kinds;
mesh traffic POLICY composes from the typed Istio kinds (KubernetesDestinationRule,
KubernetesPeerAuthentication, KubernetesAuthorizationPolicy, ...), which require
only the CRDs — installed with this component, or standalone via
KubernetesIstioBaseCrds on clusters that use the policy APIs without a mesh.

Data plane choice is first-class: `sidecar` (the classic per-pod proxy, injected
into workload pods) or `ambient` (no sidecars — a per-node ztunnel DaemonSet
carries mTLS/L4, waypoint proxies add L7 where needed; enrollment is per
namespace/pod via the `istio.io/dataplane-mode: ambient` label).

The typed fields below cover the charts' meaningful configuration surface;
`helm_values` remains as the per-release escape hatch for values beyond them
(merged last, Helm `-f` semantics, identical on both engines) — a safety valve,
never the primary interface.

## Example

```yaml
# Full-surface development manifest for the KubernetesIstio module.
#
# Exercises the AMBIENT arm (the larger valid surface: cni + ztunnel install,
# no sidecar_injector — the spec forbids injector settings in ambient mode).
# The sidecar arm (sidecar_injector + cni.enabled) is exercised by the presets
# and the E2E scenarios.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesIstio
metadata:
  name: hack-istio
spec:
  namespace:
    value: istio-system
  create_namespace: true
  version: 1.30.3
  revision: hack
  dataplane_mode: ambient
  istiod:
    autoscale:
      enabled: true
      min_replicas: 2
      max_replicas: 5
      target_cpu_utilization_percent: 75
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: "2"
        memory: 4Gi
    log_level: default:info,validation:debug
    pod_disruption_budget: true
    priority_class_name: system-cluster-critical
    node_selector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: control-plane
        effect: NoSchedule
  mesh_config:
    trust_domain: mesh.example.internal
    outbound_traffic_policy_mode: REGISTRY_ONLY
    access_log_file: /dev/stdout
    cluster_name: hack-cluster
    network: hack-network
    mesh_id: hack-mesh
    enable_prometheus_merge: true
  proxy:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: "2"
        memory: 1Gi
    log_level: warning
    auto_inject: enabled
    cluster_domain: cluster.local
  cni:
    exclude_namespaces:
      - kube-system
      - kube-node-lease
    cni_bin_dir: /opt/cni/bin
    cni_conf_dir: /etc/cni/net.d
    chained: true
  ztunnel:
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: "1"
        memory: 1Gi
    log_level: info
  gateway_defaults:
    service_type: NodePort
  images:
    hub: mirror.example.com/istio
    variant: distroless
    image_pull_secrets:
      - registry-pull-secret
  helm_values:
    istiod: |
      pilot:
        env:
          PILOT_TRACE_SAMPLING: "1.0"
    ztunnel: |
      terminationGracePeriodSeconds: 20
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.version` | `string` |  | `1.30.3` |  |
| `spec.revision` | `string` |  |  |  |
| `spec.dataplaneMode` | `string` |  | `sidecar` |  |
| `spec.istiod` | `KubernetesIstioIstiod` |  |  |  |
| `spec.istiod.replicas` | `int32` |  |  |  |
| `spec.istiod.autoscale` | `KubernetesIstioIstiodAutoscale` |  |  |  |
| `spec.istiod.autoscale.enabled` | `bool` |  | `true` |  |
| `spec.istiod.autoscale.minReplicas` | `int32` |  | `1` |  |
| `spec.istiod.autoscale.maxReplicas` | `int32` |  | `5` |  |
| `spec.istiod.autoscale.targetCpuUtilizationPercent` | `int32` |  | `80` |  |
| `spec.istiod.resources` | `ContainerResources` |  |  |  |
| `spec.istiod.resources.limits` | `CpuMemory` |  |  |  |
| `spec.istiod.resources.limits.cpu` | `string` |  |  |  |
| `spec.istiod.resources.limits.memory` | `string` |  |  |  |
| `spec.istiod.resources.requests` | `CpuMemory` |  |  |  |
| `spec.istiod.resources.requests.cpu` | `string` |  |  |  |
| `spec.istiod.resources.requests.memory` | `string` |  |  |  |
| `spec.istiod.logLevel` | `string` |  | `default:info` |  |
| `spec.istiod.podDisruptionBudget` | `bool` |  | `true` |  |
| `spec.istiod.priorityClassName` | `string` |  |  |  |
| `spec.istiod.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.istiod.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.istiod.tolerations[].key` | `string` |  |  |  |
| `spec.istiod.tolerations[].operator` | `string` |  |  |  |
| `spec.istiod.tolerations[].value` | `string` |  |  |  |
| `spec.istiod.tolerations[].effect` | `string` |  |  |  |
| `spec.istiod.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.meshConfig` | `KubernetesIstioMeshConfig` |  |  |  |
| `spec.meshConfig.trustDomain` | `string` |  | `cluster.local` |  |
| `spec.meshConfig.outboundTrafficPolicyMode` | `string` |  |  |  |
| `spec.meshConfig.accessLogFile` | `string` |  |  |  |
| `spec.meshConfig.clusterName` | `string` |  |  |  |
| `spec.meshConfig.network` | `string` |  |  |  |
| `spec.meshConfig.meshId` | `string` |  |  |  |
| `spec.meshConfig.enablePrometheusMerge` | `bool` |  | `true` |  |
| `spec.proxy` | `KubernetesIstioProxy` |  |  |  |
| `spec.proxy.resources` | `ContainerResources` |  |  |  |
| `spec.proxy.resources.limits` | `CpuMemory` |  |  |  |
| `spec.proxy.resources.limits.cpu` | `string` |  |  |  |
| `spec.proxy.resources.limits.memory` | `string` |  |  |  |
| `spec.proxy.resources.requests` | `CpuMemory` |  |  |  |
| `spec.proxy.resources.requests.cpu` | `string` |  |  |  |
| `spec.proxy.resources.requests.memory` | `string` |  |  |  |
| `spec.proxy.logLevel` | `string` |  |  |  |
| `spec.proxy.autoInject` | `string` |  |  |  |
| `spec.proxy.clusterDomain` | `string` |  | `cluster.local` |  |
| `spec.sidecarInjector` | `KubernetesIstioSidecarInjector` |  |  |  |
| `spec.sidecarInjector.enableNamespacesByDefault` | `bool` |  |  |  |
| `spec.sidecarInjector.rewriteAppHttpProbe` | `bool` |  | `true` |  |
| `spec.cni` | `KubernetesIstioCni` |  |  |  |
| `spec.cni.enabled` | `bool` |  |  |  |
| `spec.cni.excludeNamespaces` | `[]string` |  |  |  |
| `spec.cni.cniBinDir` | `string` |  |  |  |
| `spec.cni.cniConfDir` | `string` |  |  |  |
| `spec.cni.chained` | `bool` |  | `true` |  |
| `spec.ztunnel` | `KubernetesIstioZtunnel` |  |  |  |
| `spec.ztunnel.resources` | `ContainerResources` |  |  |  |
| `spec.ztunnel.resources.limits` | `CpuMemory` |  |  |  |
| `spec.ztunnel.resources.limits.cpu` | `string` |  |  |  |
| `spec.ztunnel.resources.limits.memory` | `string` |  |  |  |
| `spec.ztunnel.resources.requests` | `CpuMemory` |  |  |  |
| `spec.ztunnel.resources.requests.cpu` | `string` |  |  |  |
| `spec.ztunnel.resources.requests.memory` | `string` |  |  |  |
| `spec.ztunnel.logLevel` | `string` |  |  |  |
| `spec.gatewayDefaults` | `KubernetesIstioGatewayDefaults` |  |  |  |
| `spec.gatewayDefaults.serviceType` | `string` |  |  |  |
| `spec.images` | `KubernetesIstioImages` |  |  |  |
| `spec.images.hub` | `string` |  |  |  |
| `spec.images.variant` | `string` |  |  |  |
| `spec.images.imagePullSecrets` | `[]string` |  |  |  |
| `spec.helmValues` | `KubernetesIstioHelmValues` |  |  |  |
| `spec.helmValues.base` | `string` |  |  |  |
| `spec.helmValues.istiod` | `string` |  |  |  |
| `spec.helmValues.cni` | `string` |  |  |  |
| `spec.helmValues.ztunnel` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install the control plane into ("istio-system" by convention —
istiod, and in ambient mode the CNI agent and ztunnel, all live here).
Accepts a literal namespace name or a reference to a KubernetesNamespace
resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton governance
labels) before installing and deleted with the resource. When false, the
namespace must already exist.

### spec.version

`string` · optional (explicit presence)

Helm chart version for every release this component installs (base, istiod,
and in ambient mode cni + ztunnel — Istio versions its charts in lockstep
with the product, e.g. "1.30.3"). Pin deliberately.

IMPORTANT (upgrades): Istio supports sequential, single-minor upgrades only
(1.28 -> 1.29 -> 1.30); skipping minors in place is unsupported. An existing
mesh on an older minor must step through each minor, or run a revisioned
canary control plane (`revision`) and migrate workloads.

- default: `1.30.3`
- rule: {"string":{"pattern":"^[0-9]+\\.[0-9]+\\.[0-9]+$"}}

### spec.revision

`string` · optional (explicit presence)

Control-plane revision name (e.g. "1-30-3"). Empty runs the unnamed
DEFAULT revision — the standard posture. Setting a revision names the
istiod deployment `istiod-<revision>` and scopes injection to workloads
labeled `istio.io/rev: <revision>` (organizations that always run
revisioned control planes). One KubernetesIstio per cluster: the CRDs and
the validation-webhook plumbing are cluster singletons, so side-by-side
canary control planes are deliberately not modeled — upgrade in place,
one minor at a time. Must be a DNS-1123 label (lowercase alphanumerics
and '-').

- rule: {"string":{"maxLen":"56","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.dataplaneMode

`string` · optional (explicit presence)

The mesh's data plane mode. `sidecar` (the default) injects a proxy container
into each workload pod; `ambient` runs sidecar-less (per-node ztunnel for
L4/mTLS + optional waypoints for L7) and additionally installs the istio-cni
node agent and ztunnel DaemonSets. Choose at install time — switching modes
later is a workload-by-workload migration, not a flag flip.

- default: `sidecar`
- rule: {"string":{"in":["sidecar","ambient"]}}

### spec.istiod

`KubernetesIstioIstiod`

istiod (control-plane) tuning: sizing, autoscaling, scheduling, logging.

CRD note: the module applies the Istio CRDs (DestinationRule,
AuthorizationPolicy, ...) itself via server-side apply, OUTSIDE the Helm
release — never Helm-owned. Helm-owned CRDs cannot be adopted from an
existing install, so a cluster running the CRDs-only
KubernetesIstioBaseCrds could never upgrade to the full mesh; module-
owned CRDs are co-ownable by both kinds, making that migration a plain
redeploy. Destroying this component removes the CRDs with everything
else (standard engine semantics — mesh configuration objects cascade).

- rule: replicas and autoscale cannot both be set — the HPA owns the replica count when autoscaling is enabled

### spec.istiod.replicas

`int32` · optional (explicit presence)

Fixed replica count. Chart default: 1. Mutually exclusive with `autoscale`
(the chart's HPA owns the replica count when autoscaling).

- rule: {"int32":{"gte":1}}

### spec.istiod.autoscale

`KubernetesIstioIstiodAutoscale`

Horizontal autoscaling for istiod. The chart enables this by default
(min 1 / max 5 at 80% CPU).

- rule: max_replicas must be greater than or equal to min_replicas

### spec.istiod.autoscale.enabled

`bool` · optional (explicit presence)

Enable the chart-managed HPA. Chart default: true.

- default: `true`

### spec.istiod.autoscale.minReplicas

`int32` · optional (explicit presence)

Minimum replicas. Chart default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.istiod.autoscale.maxReplicas

`int32` · optional (explicit presence)

Maximum replicas. Chart default: 5.

- default: `5`
- rule: {"int32":{"gte":1}}

### spec.istiod.autoscale.targetCpuUtilizationPercent

`int32` · optional (explicit presence)

Target average CPU utilization percentage. Chart default: 80.

- default: `80`
- rule: {"int32":{"lte":100,"gte":1}}

### spec.istiod.resources

`ContainerResources`

istiod container CPU/memory requests and limits. Chart default request:
500m CPU / 2Gi memory.

### spec.istiod.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.istiod.resources.limits.cpu

`string`

### spec.istiod.resources.limits.memory

`string`

### spec.istiod.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.istiod.resources.requests.cpu

`string`

### spec.istiod.resources.requests.memory

`string`

### spec.istiod.logLevel

`string` · optional (explicit presence)

Control-plane log level, per scope as `<scope>:<level>` pairs joined by
commas, or a bare default like "default:info". Chart default: "default:info".

- default: `default:info`

### spec.istiod.podDisruptionBudget

`bool` · optional (explicit presence)

When true (chart default), a PodDisruptionBudget (minAvailable 1) guards the
control plane during node drains — meaningful with more than one replica.

- default: `true`

### spec.istiod.priorityClassName

`string`

PriorityClass for control-plane pods. Set "system-cluster-critical" on
production clusters so istiod is never evicted before the workloads that
depend on it.

### spec.istiod.nodeSelector

`map<string, string>`

Node selector for istiod pods.

### spec.istiod.tolerations

`[]WorkloadToleration`

Tolerations for istiod pods.

### spec.istiod.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.istiod.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.istiod.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.istiod.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.istiod.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.meshConfig

`KubernetesIstioMeshConfig`

Mesh-wide runtime configuration (MeshConfig): trust domain, outbound traffic
policy, access logging, multi-cluster identity. Everything else in MeshConfig
is reachable via helm_values.istiod (meshConfig.*).

### spec.meshConfig.trustDomain

`string` · optional (explicit presence)

The trust domain anchoring every workload identity the mesh issues
(SPIFFE IDs are `spiffe://<trust_domain>/ns/<ns>/sa/<sa>`). Upstream
default: "cluster.local". Set a stable, organization-unique value before
production — changing it later re-identifies every workload.

- default: `cluster.local`

### spec.meshConfig.outboundTrafficPolicyMode

`string` · optional (explicit presence)

What sidecars do with traffic to destinations OUTSIDE the mesh's service
registry. ALLOW_ANY (upstream default) passes it through; REGISTRY_ONLY
blocks it — the egress-lockdown posture, where every external service must
be declared with a KubernetesServiceEntry.

- rule: {"string":{"in":["ALLOW_ANY","REGISTRY_ONLY"]}}

### spec.meshConfig.accessLogFile

`string`

File path for the proxies' access logs (e.g. "/dev/stdout" to enable mesh
access logging everywhere). Empty (upstream default) disables file access
logging; per-workload control belongs to KubernetesTelemetry.

### spec.meshConfig.clusterName

`string`

Cluster name for multi-cluster meshes — must be unique per cluster within
the mesh and consistent across the control plane and data plane installs.
Empty is correct for single-cluster meshes.

### spec.meshConfig.network

`string`

Network name for multi-network meshes (maps this cluster's endpoints into
MeshNetworks). Empty is correct for single-network meshes.

### spec.meshConfig.meshId

`string`

Mesh identifier for mesh federation / telemetry aggregation across meshes.
Upstream defaults it to the trust domain when empty.

### spec.meshConfig.enablePrometheusMerge

`bool` · optional (explicit presence)

When true (upstream default), Prometheus annotations on workloads are merged
with the proxy's own metrics so one scrape endpoint serves both.

- default: `true`

### spec.proxy

`KubernetesIstioProxy`

Defaults for the sidecar proxies injected into workload pods (sidecar mode)
and gateway proxies. Ignored by ambient ztunnel (size it via `ztunnel`).

### spec.proxy.resources

`ContainerResources`

Per-proxy container CPU/memory requests and limits. Chart default:
request 100m/128Mi, limit 2000m/1Gi. These multiply across every injected
pod — size deliberately on large meshes.

### spec.proxy.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.proxy.resources.limits.cpu

`string`

### spec.proxy.resources.limits.memory

`string`

### spec.proxy.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.proxy.resources.requests.cpu

`string`

### spec.proxy.resources.requests.memory

`string`

### spec.proxy.logLevel

`string` · optional (explicit presence)

Proxy log level. Chart default: "warning".

- rule: {"string":{"in":["trace","debug","info","warning","error","critical","off"]}}

### spec.proxy.autoInject

`string` · optional (explicit presence)

The injector's default policy. "enabled" (chart default) injects any pod in
an injection-labeled namespace unless the pod opts out; "disabled" injects
only pods that explicitly opt in via the `sidecar.istio.io/inject: "true"`
annotation.

- rule: {"string":{"in":["enabled","disabled"]}}

### spec.proxy.clusterDomain

`string` · optional (explicit presence)

The cluster's DNS domain. Chart default: "cluster.local" — must match the
cluster's actual kubelet/DNS configuration on the rare clusters that
customize it.

- default: `cluster.local`

### spec.sidecarInjector

`KubernetesIstioSidecarInjector`

Sidecar injection behavior (sidecar mode only).

### spec.sidecarInjector.enableNamespacesByDefault

`bool`

When true, sidecars are injected in ALL namespaces except those labeled
`istio-injection=disabled` (opt-out model). Chart default false: only
namespaces labeled `istio-injection=enabled` are injected (opt-in model).

### spec.sidecarInjector.rewriteAppHttpProbe

`bool` · optional (explicit presence)

When true (chart default), the injector rewrites workload HTTP liveness/
readiness probes to route through the sidecar so kubelet probes keep working
under strict mTLS.

- default: `true`

### spec.cni

`KubernetesIstioCni`

istio-cni node agent tuning. In ambient mode the CNI agent is ALWAYS
installed (it is how traffic reaches ztunnel). In sidecar mode it is opt-in
(`enabled: true`) and replaces the injected privileged init-container with a
node-level agent — required on platforms that forbid NET_ADMIN init
containers (e.g. OpenShift), recommended for tighter pod security everywhere.

### spec.cni.enabled

`bool`

Install the istio-cni node agent in SIDECAR mode (replaces the injected
privileged init-container with a node-level agent). Ambient mode ignores
this flag — the agent always installs there.

### spec.cni.excludeNamespaces

`[]string`

Namespaces the CNI agent never configures traffic redirection in. Chart
default: ["kube-system"].

### spec.cni.cniBinDir

`string` · optional (explicit presence)

Directory of the CNI binaries on nodes. Chart default: /opt/cni/bin.
Platform-specific overrides belong here (e.g. some managed platforms use
/home/kubernetes/bin — see the upstream platform profiles).

### spec.cni.cniConfDir

`string` · optional (explicit presence)

Directory of the CNI config files on nodes. Chart default: /etc/cni/net.d.

### spec.cni.chained

`bool` · optional (explicit presence)

Deploy the Istio CNI config as a chained plugin (chart default true).
Platforms that reject chained plugins (e.g. OpenShift) set false for a
standalone config file.

- default: `true`

### spec.ztunnel

`KubernetesIstioZtunnel`

ztunnel (ambient node proxy) tuning. Only meaningful in ambient mode.

### spec.ztunnel.resources

`ContainerResources`

ztunnel container CPU/memory requests and limits. Chart default request:
200m CPU / 512Mi memory (sized for ~200k-pod clusters; scale with cluster
size and connection volume).

### spec.ztunnel.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.ztunnel.resources.limits.cpu

`string`

### spec.ztunnel.resources.limits.memory

`string`

### spec.ztunnel.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.ztunnel.resources.requests.cpu

`string`

### spec.ztunnel.resources.requests.memory

`string`

### spec.ztunnel.logLevel

`string` · optional (explicit presence)

ztunnel log level. Chart default: "info".

- rule: {"string":{"in":["trace","debug","info","warn","error"]}}

### spec.gatewayDefaults

`KubernetesIstioGatewayDefaults`

Defaults applied to the gateway deployments istiod auto-provisions for
Gateway API Gateways with `gateway_class_name: istio`.

### spec.gatewayDefaults.serviceType

`string` · optional (explicit presence)

Service type for gateways provisioned from the `istio` GatewayClass.
Upstream default: LoadBalancer. Set NodePort or ClusterIP on clusters
without a cloud load-balancer controller (bare metal, kind, k3s) — a
per-Gateway override also exists via the Gateway's `infrastructure`
parameters.

- rule: {"string":{"in":["ClusterIP","NodePort","LoadBalancer"]}}

### spec.images

`KubernetesIstioImages`

Image source overrides for every Istio image (pilot, proxy, cni, ztunnel) —
the air-gapped/mirror and variant knobs.

### spec.images.hub

`string`

Registry/hub serving all Istio images (pilot, proxyv2, install-cni,
ztunnel) — the air-gapped/mirror knob. Empty = docker.io/istio.

### spec.images.variant

`string` · optional (explicit presence)

Image variant: "distroless" (recommended hardening; ambient's own profile
default) or "debug". Empty = the release's default variant.

- rule: {"string":{"in":["debug","distroless"]}}

### spec.images.imagePullSecrets

`[]string`

Names of image-pull Secrets (in the install namespace) attached to every
Istio ServiceAccount — required when `hub` points at a private registry.

### spec.helmValues

`KubernetesIstioHelmValues`

Per-release escape hatches: additional chart values as YAML documents,
merged LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For chart surface beyond the typed fields —
never the substitute for them. Do not put secrets here.

### spec.helmValues.base

`string`

Extra values for the `base` release (CRDs, validation webhook plumbing).

### spec.helmValues.istiod

`string`

Extra values for the `istiod` release — the main chart: meshConfig.*,
pilot.env, sidecarInjectorWebhook.*, and everything else istiod.

### spec.helmValues.cni

`string`

Extra values for the `cni` release (installed in ambient mode or when
cni.enabled).

### spec.helmValues.ztunnel

`string`

Extra values for the `ztunnel` release (installed in ambient mode).

## Validation Rules

- `istio.ztunnel_requires_ambient`: ztunnel settings apply only when dataplane_mode is 'ambient' — the ztunnel release is not installed in sidecar mode
- `istio.sidecar_injector_requires_sidecar_mode`: sidecar_injector settings apply only when dataplane_mode is 'sidecar' — ambient workloads are not injected

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesIstio, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the control plane is installed in (e.g. "istio-system"). |
| `status.outputs.istiod_service_name` | `string` | Name of the istiod Service in that namespace ("istiod" for the default revision, "istiod-<revision>" for a named revision) — the discovery address data-plane proxies and remote clusters connect to. |
| `status.outputs.revision` | `string` | The control-plane revision installed ("default" when no revision is named). Workloads pin to a revisioned control plane with the `istio.io/rev` label. |
| `status.outputs.gateway_class_name` | `string` | Name of the GatewayClass istiod serves ("istio"). Create a KubernetesGateway with this gateway_class_name and istiod provisions and programs the gateway deployment automatically — the composition seam for north-south exposure. |
| `status.outputs.trust_domain` | `string` | The mesh's trust domain (the identity root of every workload certificate, e.g. "cluster.local") — the prefix of principal strings in KubernetesAuthorizationPolicy rules. |
| `status.outputs.dataplane_mode` | `string` | The data plane mode the mesh was installed with ("sidecar" or "ambient") — tells composed resources whether workloads enroll via sidecar injection labels or the ambient dataplane-mode label. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
