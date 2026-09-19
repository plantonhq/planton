# DigitalOceanKubernetesCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanKubernetesClusterSpec models the full digitalocean_kubernetes_cluster
resource surface: version/region/VPC placement, the inline default node pool
(with labels, taints, tags, and autoscaling), HA control plane, surge and auto
upgrades, maintenance policy, control-plane firewall, pod/service/worker
subnets, isolated workers, SSO, cluster-autoscaler tuning, registry
integration, kubeconfig expiry, destroy-time cleanup, and the managed addon
toggles (routing agent, GPU device plugins and DRA drivers, RDMA, CoreDNS
autoscaler, P2P OCI registry).

Additional node pools beyond the inline default pool are separate
DigitalOceanKubernetesNodePool resources.

## Example

```yaml
# Example DigitalOceanKubernetesCluster manifests.
#
# Deploy with: planton apply -f manifest.yaml
#
# The first document is the smallest real cluster (one node, everything else
# at DigitalOcean defaults). The second exercises the full surface: HA
# control plane, maintenance policy, control-plane firewall, custom pod and
# service subnets, cluster-autoscaler tuning, SSO, managed addons, and a
# default node pool with labels, a taint, tags, and autoscaling.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanKubernetesCluster
metadata:
  name: example-doks-minimal
spec:
  clusterName: example-doks-minimal
  region: nyc3
  kubernetesVersion: "1.35"
  vpc:
    value: b5648f9e-a28a-4760-bb87-b2fad07ae295
  defaultNodePool:
    size: s-1vcpu-2gb
    nodeCount: 1
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanKubernetesCluster
metadata:
  name: example-doks-full
spec:
  clusterName: example-doks-full
  region: nyc3
  kubernetesVersion: "1.35"
  vpc:
    value: b5648f9e-a28a-4760-bb87-b2fad07ae295
  highlyAvailable: true
  autoUpgrade: true
  surgeUpgrade: true
  registryIntegration: true
  maintenancePolicy:
    day: sunday
    startTime: "02:00"
  controlPlaneFirewall:
    enabled: true
    allowedAddresses:
      - "203.0.113.5/32"
      - "198.51.100.0/24"
  clusterSubnet: "10.100.0.0/16"
  serviceSubnet: "10.200.0.0/16"
  kubeconfigExpireSeconds: 3600
  clusterAutoscalerConfiguration:
    scaleDownUtilizationThreshold: 0.5
    scaleDownUnneededTime: "1m30s"
    expanders:
      - least-waste
      - random
  sso:
    enabled: true
    required: false
    issuerUrl: "https://sso.example.com"
    clientId: "example-client-id"
  routingAgent:
    enabled: true
  corednsAutoscaler:
    enabled: true
  tags:
    - env:example
  defaultNodePool:
    size: s-2vcpu-4gb
    # No nodeCount: with autoScale on, the pool starts at minNodes and the
    # autoscaler owns the count (a stated count would fight it).
    autoScale: true
    minNodes: 2
    maxNodes: 5
    labels:
      workload: general
    taints:
      - key: dedicated
        value: platform
        effect: NoSchedule
    tags:
      - pool:default
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.clusterName` | `string` | yes |  |  |
| `spec.region` | `enum` | yes |  |  |
| `spec.kubernetesVersion` | `string` | yes |  |  |
| `spec.vpc` | `string \| valueFrom` | yes |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.highlyAvailable` | `bool` |  | `false` |  |
| `spec.autoUpgrade` | `bool` |  |  |  |
| `spec.registryIntegration` | `bool` |  |  |  |
| `spec.tags` | `[]string` |  |  |  |
| `spec.defaultNodePool` | `DigitalOceanKubernetesClusterDefaultNodePool` | yes |  |  |
| `spec.defaultNodePool.size` | `string` | yes |  |  |
| `spec.defaultNodePool.nodeCount` | `uint32` |  |  |  |
| `spec.defaultNodePool.autoScale` | `bool` |  |  |  |
| `spec.defaultNodePool.minNodes` | `uint32` |  |  |  |
| `spec.defaultNodePool.maxNodes` | `uint32` |  |  |  |
| `spec.defaultNodePool.labels` | `map<string, string>` |  |  |  |
| `spec.defaultNodePool.taints` | `[]DigitalOceanKubernetesClusterNodePoolTaint` |  |  |  |
| `spec.defaultNodePool.taints[].key` | `string` | yes |  |  |
| `spec.defaultNodePool.taints[].value` | `string` |  |  |  |
| `spec.defaultNodePool.taints[].effect` | `string` | yes |  |  |
| `spec.defaultNodePool.tags` | `[]string` |  |  |  |
| `spec.defaultNodePool.gpuPartitionMode` | `string` |  |  |  |
| `spec.surgeUpgrade` | `bool` |  | `true` |  |
| `spec.maintenancePolicy` | `DigitalOceanKubernetesClusterMaintenancePolicy` |  |  |  |
| `spec.maintenancePolicy.day` | `string` | yes |  |  |
| `spec.maintenancePolicy.startTime` | `string` | yes |  |  |
| `spec.controlPlaneFirewall` | `DigitalOceanKubernetesClusterControlPlaneFirewall` |  |  |  |
| `spec.controlPlaneFirewall.enabled` | `bool` | yes |  |  |
| `spec.controlPlaneFirewall.allowedAddresses` | `[]string` |  |  |  |
| `spec.clusterSubnet` | `string` |  |  |  |
| `spec.serviceSubnet` | `string` |  |  |  |
| `spec.workerSubnetUuid` | `string` |  |  |  |
| `spec.isolatedWorkers` | `bool` |  |  |  |
| `spec.destroyAllAssociatedResources` | `bool` |  |  |  |
| `spec.kubeconfigExpireSeconds` | `uint32` |  |  |  |
| `spec.clusterAutoscalerConfiguration` | `DigitalOceanKubernetesClusterAutoscalerConfiguration` |  |  |  |
| `spec.clusterAutoscalerConfiguration.scaleDownUtilizationThreshold` | `double` |  |  |  |
| `spec.clusterAutoscalerConfiguration.scaleDownUnneededTime` | `string` |  |  |  |
| `spec.clusterAutoscalerConfiguration.expanders` | `[]string` |  |  |  |
| `spec.sso` | `DigitalOceanKubernetesClusterSso` |  |  |  |
| `spec.sso.enabled` | `bool` | yes |  |  |
| `spec.sso.required` | `bool` |  |  |  |
| `spec.sso.issuerUrl` | `string` |  |  |  |
| `spec.sso.clientId` | `string` |  |  |  |
| `spec.routingAgent` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.routingAgent.enabled` | `bool` | yes |  |  |
| `spec.p2pOciRegistryPlugin` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.p2pOciRegistryPlugin.enabled` | `bool` | yes |  |  |
| `spec.amdGpuDevicePlugin` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.amdGpuDevicePlugin.enabled` | `bool` | yes |  |  |
| `spec.amdGpuDraDriver` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.amdGpuDraDriver.enabled` | `bool` | yes |  |  |
| `spec.amdGpuDeviceMetricsExporterPlugin` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.amdGpuDeviceMetricsExporterPlugin.enabled` | `bool` | yes |  |  |
| `spec.nvidiaGpuDevicePlugin` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.nvidiaGpuDevicePlugin.enabled` | `bool` | yes |  |  |
| `spec.nvidiaGpuDraDriver` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.nvidiaGpuDraDriver.enabled` | `bool` | yes |  |  |
| `spec.rdmaSharedDevicePlugin` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.rdmaSharedDevicePlugin.enabled` | `bool` | yes |  |  |
| `spec.corednsAutoscaler` | `DigitalOceanKubernetesClusterFeatureToggle` |  |  |  |
| `spec.corednsAutoscaler.enabled` | `bool` | yes |  |  |

## Field Details

### spec.clusterName

`string` · required

A human-readable name for the Kubernetes cluster. This name is the
cluster's identifier in DigitalOcean.

- rule: {"required":true}

### spec.region

`enum` · required

The DigitalOcean region where the cluster's control plane and nodes are
provisioned. Cannot be changed after creation.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.kubernetesVersion

`string` · required

The Kubernetes version to create the cluster at, as DigitalOcean offers
it TODAY: either a minor prefix ("1.35" -- DigitalOcean resolves it to the
current patch) or a full slug ("1.35.7-do.5"). Prefer the prefix: patch
slugs are retired every few weeks and a create naming a retired slug
fails with 422, while a minor stays creatable for its whole support
window. The live offer list is GET /v2/kubernetes/options (or `doctl
kubernetes options versions`); on 2026-09-16 it was 1.34, 1.35, 1.36.
This is the creation pin only: patch upgrades ride auto_upgrade, and
both provisioners ignore later drift on this field because DigitalOcean
recreates the whole cluster when the configured version is lower than
the live one.

- rule: {"required":true}

### spec.vpc

`string | valueFrom` · required

Reference to the DigitalOcean VPC where the cluster will reside. The
cluster consumes the VPC's ID (a DigitalOcean UUID), so a reference
resolves to the DigitalOceanVpc's exported vpc_id output rather than its
metadata name. Cannot be changed after creation.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.highlyAvailable

`bool`

Whether to run a highly available control plane (multiple replicas,
additional cost). HA is one-way: once enabled it cannot be turned off.
Unset sends an explicit false, which is deliberate: newer DOKS versions
default HA on server-side, and an explicit false keeps the cheaper
single-replica control plane unless HA is asked for.

- default: `false`

### spec.autoUpgrade

`bool`

Whether the cluster automatically upgrades to new patch releases inside
the maintenance window. Minor/major upgrades are never automatic.

### spec.registryIntegration

`bool`

Whether to integrate the account's DigitalOcean Container Registry
(DOCR): pulls from the private registry work cluster-wide without
imagePullSecrets. This is an account-level association the API never
reports back, so it survives as configuration only.

### spec.tags

`[]string`

Tags applied to the cluster in DigitalOcean, in addition to the standard
Planton labels both provisioners always apply. Never include
"terraform:default-node-pool" -- the provider manages that tag itself to
mark the inline pool.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.defaultNodePool

`DigitalOceanKubernetesClusterDefaultNodePool` · required

The cluster's inline default node pool. Its name is synthesized by the
provisioners; additional pools are separate DigitalOceanKubernetesNodePool
resources.

- rule: {"required":true}
- rule: auto_scale requires min_nodes >= 1 and max_nodes >= min_nodes
- rule: node_count is required when auto_scale is off and must be left unset when auto_scale is on (the pool starts at min_nodes and the autoscaler owns the count)

### spec.defaultNodePool.size

`string` · required

The slug identifier for the Droplet size of each node (e.g.
"s-2vcpu-4gb"). Changing it replaces the whole cluster.

- rule: {"required":true}

### spec.defaultNodePool.nodeCount

`uint32`

The fixed number of nodes in the pool. Required when auto_scale is off;
must be left unset when auto_scale is on -- the pool then starts at
min_nodes and DigitalOcean's autoscaler owns the count from there, and
both provisioners send no count at all (a stated count would be written
back from the live pool on every read and re-applied on every update,
fighting the autoscaler).

### spec.defaultNodePool.autoScale

`bool`

Whether DigitalOcean's cluster-autoscaler manages this pool's node count
between min_nodes and max_nodes.

### spec.defaultNodePool.minNodes

`uint32`

Minimum node count when auto_scale is enabled.

### spec.defaultNodePool.maxNodes

`uint32`

Maximum node count when auto_scale is enabled.

### spec.defaultNodePool.labels

`map<string, string>`

(Optional) Kubernetes labels applied to every node in the pool, in
addition to the standard Planton labels both provisioners always apply.

### spec.defaultNodePool.taints

`[]DigitalOceanKubernetesClusterNodePoolTaint`

(Optional) Kubernetes taints applied to every node in the pool.

### spec.defaultNodePool.taints[].key

`string` · required

Taint key, e.g. "dedicated".

- rule: {"required":true}

### spec.defaultNodePool.taints[].value

`string`

(Optional) Taint value, e.g. "gpu-workloads". Kubernetes allows
valueless taints, so empty is legal; the provisioners always send the
value (possibly empty), which is all the provider's required leaf asks.

### spec.defaultNodePool.taints[].effect

`string` · required

Taint effect. One of NoSchedule, PreferNoSchedule, NoExecute
(case-sensitive, exactly as Kubernetes spells them).

- rule: {"required":true,"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.defaultNodePool.tags

`[]string`

(Optional) DigitalOcean tags applied to the pool's Droplets, in addition
to the cluster's tags.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.defaultNodePool.gpuPartitionMode

`string`

(Optional) GPU partitioning mode for AMD GPU Droplet sizes. Changing it
replaces the whole cluster.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AMD_PARTITION_MODE_SPX_NPS1","AMD_PARTITION_MODE_DPX_NPS2"]}}

### spec.surgeUpgrade

`bool` · optional (explicit presence)

Whether upgrades temporarily provision surge nodes to minimize downtime.
DigitalOcean defaults this ON, so unset defers to that default; set it to
false only to explicitly disable surge upgrades.

- default: `true`

### spec.maintenancePolicy

`DigitalOceanKubernetesClusterMaintenancePolicy`

(Optional) Weekly window in which DigitalOcean applies automatic
maintenance and auto-upgrades. When unset, DigitalOcean picks a window.

### spec.maintenancePolicy.day

`string` · required

Day of the week, or "any" to let DigitalOcean pick the day.
Case-insensitive.

- rule: {"required":true,"string":{"pattern":"^(?i)(any|monday|tuesday|wednesday|thursday|friday|saturday|sunday)$"}}

### spec.maintenancePolicy.startTime

`string` · required

Start of the window in UTC, 24-hour "HH:MM", for example "02:00".

- rule: {"required":true,"string":{"pattern":"^([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.controlPlaneFirewall

`DigitalOceanKubernetesClusterControlPlaneFirewall`

(Optional) Firewall in front of the cluster's public API server endpoint.
When unset, the API server accepts connections from anywhere.

### spec.controlPlaneFirewall.enabled

`bool` · required · optional (explicit presence)

Whether the firewall is enforced. Keeping enabled explicit (rather than
inferring it from a non-empty address list) allows staging rules while
disabled and matches the provider's block, whose enabled leaf is
required. optional gives the bool presence so an explicit false is valid.

- rule: {"required":true}

### spec.controlPlaneFirewall.allowedAddresses

`[]string`

Source addresses allowed to reach the API server, as plain IPs or CIDR
blocks. The provider validates nothing here; catching a typo before
apply beats a lockout after.

- rule: {"repeated":{"items":{"cel":[{"id":"ip_or_cidr","message":"must be an IP address or CIDR block","expression":"this.isIp() || this.isIpPrefix()"}]}}}

### spec.clusterSubnet

`string`

(Optional) CIDR block for pod IPs. When unset, DigitalOcean assigns one.
Cannot be changed after creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipWithPrefixlen":true}}

### spec.serviceSubnet

`string`

(Optional) CIDR block for service ClusterIPs. When unset, DigitalOcean
assigns one. Cannot be changed after creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"ipWithPrefixlen":true}}

### spec.workerSubnetUuid

`string`

(Optional) UUID of the DigitalOcean-managed subnet to place worker nodes
in. Literal only: worker subnets are DigitalOcean-assigned network
slices, not a Planton-managed kind. Cannot be changed after creation.

### spec.isolatedWorkers

`bool`

(Optional) Whether worker nodes are blocked from public internet access
(control plane connectivity is routed internally). Cannot be changed
after creation.

### spec.destroyAllAssociatedResources

`bool`

(Optional) When true, destroying the cluster also deletes the load
balancers, volumes, and volume snapshots it created. DANGEROUS and
destroy-time only: it never affects the running cluster, and the API
never reports it back.

### spec.kubeconfigExpireSeconds

`uint32`

(Optional) Validity of the provisioner-fetched kubeconfig credentials in
seconds. 0 (unset) means DigitalOcean's 7-day default.

### spec.clusterAutoscalerConfiguration

`DigitalOceanKubernetesClusterAutoscalerConfiguration`

(Optional) Tuning for the DigitalOcean-managed cluster-autoscaler.

### spec.clusterAutoscalerConfiguration.scaleDownUtilizationThreshold

`double` · optional (explicit presence)

(Optional) Node utilization fraction (0.0-1.0) below which a node is
eligible for scale-down. Optional so an unset value defers to
DigitalOcean's default rather than sending 0.

- rule: {"double":{"lte":1,"gte":0}}

### spec.clusterAutoscalerConfiguration.scaleDownUnneededTime

`string`

(Optional) How long a node must stay unneeded before scale-down, as a Go
duration string, e.g. "1m30s".

### spec.clusterAutoscalerConfiguration.expanders

`[]string`

(Optional) Expander strategies the autoscaler uses to pick which pool to
grow, in priority order (e.g. "least-waste", "random", "priority").

### spec.sso

`DigitalOceanKubernetesClusterSso`

(Optional) OpenID Connect single sign-on for the cluster's Kubernetes
API.

### spec.sso.enabled

`bool` · required · optional (explicit presence)

Whether SSO is enabled. optional gives the bool presence so an explicit
false is valid.

- rule: {"required":true}

### spec.sso.required

`bool`

(Optional) Whether SSO is mandatory for all cluster access.

### spec.sso.issuerUrl

`string`

(Optional) OIDC issuer URL of the identity provider.

### spec.sso.clientId

`string`

(Optional) OAuth client ID registered with the identity provider.

### spec.routingAgent

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) The DigitalOcean routing agent, required for some advanced
networking features. Unset leaves the addon at DigitalOcean's default.

### spec.routingAgent.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.p2pOciRegistryPlugin

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) Peer-to-peer OCI registry mirror addon for faster image pulls
across nodes. Unset leaves the addon at DigitalOcean's default.
Requires kubernetes_version 1.36.0-do.2 or later: on an older version
DigitalOcean rejects the whole cluster create with a validation 422
("p2p-oci-registry is only supported on DOKS v1.36.0-do.2 or later")
and creates nothing.

### spec.p2pOciRegistryPlugin.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.amdGpuDevicePlugin

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) AMD GPU device plugin addon (mutually exclusive with the AMD
DRA driver). Unset leaves the addon at DigitalOcean's default.

### spec.amdGpuDevicePlugin.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.amdGpuDraDriver

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) AMD GPU dynamic-resource-allocation driver addon (mutually
exclusive with the AMD device plugin). Unset leaves the addon at
DigitalOcean's default.

### spec.amdGpuDraDriver.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.amdGpuDeviceMetricsExporterPlugin

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) AMD GPU device metrics exporter addon. Unset leaves the addon
at DigitalOcean's default.

### spec.amdGpuDeviceMetricsExporterPlugin.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.nvidiaGpuDevicePlugin

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) NVIDIA GPU device plugin addon (mutually exclusive with the
NVIDIA DRA driver). Unset leaves the addon at DigitalOcean's default.

### spec.nvidiaGpuDevicePlugin.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.nvidiaGpuDraDriver

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) NVIDIA GPU dynamic-resource-allocation driver addon (mutually
exclusive with the NVIDIA device plugin). Unset leaves the addon at
DigitalOcean's default.

### spec.nvidiaGpuDraDriver.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.rdmaSharedDevicePlugin

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) RDMA shared device plugin addon for high-performance
networking on supported GPU droplets. Unset leaves the addon at
DigitalOcean's default.

### spec.rdmaSharedDevicePlugin.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

### spec.corednsAutoscaler

`DigitalOceanKubernetesClusterFeatureToggle`

(Optional) CoreDNS horizontal autoscaler addon. Unset leaves the addon
at DigitalOcean's default, which depends on the Kubernetes version: off
through 1.35, on from 1.36. Set it explicitly when the cluster's
behavior must not change across a version upgrade.

### spec.corednsAutoscaler.enabled

`bool` · required · optional (explicit presence)

Whether the addon is enabled. optional gives the bool presence so an
explicit false (assert OFF) is valid, not just true.

- rule: {"required":true}

## Validation Rules

- `amd_gpu_plugin_xor_dra`: amd_gpu_device_plugin and amd_gpu_dra_driver are mutually exclusive
- `nvidia_gpu_plugin_xor_dra`: nvidia_gpu_device_plugin and nvidia_gpu_dra_driver are mutually exclusive

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanKubernetesCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | The unique identifier (UUID) of the created Kubernetes cluster. |
| `status.outputs.kubeconfig` | `string` | The raw kubeconfig YAML for accessing the cluster (not base64-encoded); write it to a file and point KUBECONFIG at it. Contains admin credentials -- treat as a secret. |
| `status.outputs.api_server_endpoint` | `string` | The endpoint URL of the Kubernetes API server for the cluster. |
| `status.outputs.urn` | `string` | The uniform resource name of the cluster ("do:kubernetes:<cluster_id>"), used when attaching the cluster to a DigitalOcean project. |
| `status.outputs.ipv4_address` | `string` | The public IPv4 address of the cluster's control plane, when DigitalOcean reports one. Clusters created today report NONE: the API server sits behind DigitalOcean's own front end and is reachable only by the api_server_endpoint hostname (measured on a single-replica 1.35 cluster, not just on HA clusters). Both provisioners export the value verbatim, so expect an empty string; anything that needs the control plane's address -- allowlists, health probes -- should use api_server_endpoint. |
| `status.outputs.default_node_pool_id` | `string` | The unique identifier (UUID) of the cluster's inline default node pool. |
| `status.outputs.cluster_subnet` | `string` | The CIDR block from which pod IPs are assigned. |
| `status.outputs.service_subnet` | `string` | The CIDR block from which service ClusterIPs are assigned. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpc` | DigitalOceanVpc | `status.outputs.vpc_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanDatabaseFirewall | `spec.kubernetesClusterIds` | `status.outputs.cluster_id` |
| DigitalOceanFirewall | `spec.inboundRules[].sourceKubernetesIds` | `status.outputs.cluster_id` |
| DigitalOceanFirewall | `spec.outboundRules[].destinationKubernetesIds` | `status.outputs.cluster_id` |
| DigitalOceanKubernetesNodePool | `spec.cluster` | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
