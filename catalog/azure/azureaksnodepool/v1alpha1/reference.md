# AzureAksNodePool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureAksNodePoolSpec** defines the configuration for creating an AKS
node pool: a scale set of worker nodes attached to an existing AKS
cluster.

Node pools are the unit of compute shape in AKS. The cluster itself
carries only the mandatory default (system) pool; every workload-shaped
pool -- general compute, memory-optimized, GPU, spot, Windows -- is one
of these resources, referencing its cluster by ARM ID. Pools have fully
independent lifecycles: scale, upgrade, rotate, or delete a pool without
touching the cluster or its siblings.

The pool is an ARM child of the cluster: the cluster's ID carries the
resource group and cluster name, and the modules derive both from it
rather than modeling redundant fields that could contradict the
referenced cluster.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAksNodePool
metadata:
  name: production-app-pool
  labels:
    environment: production
    team: platform
spec:
  kubernetesClusterId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/production-rg/providers/Microsoft.ContainerService/managedClusters/production-aks-cluster
  name: general
  vmSize: Standard_D8s_v5
  autoScalingEnabled: true
  minCount: 2
  maxCount: 10
  zones:
    - "1"
    - "2"
    - "3"
  mode: USER
  nodeLabels:
    workload: application
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.kubernetesClusterId` | `string \| valueFrom` | yes |  | AzureAksCluster (`status.outputs.cluster_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.vmSize` | `string` | yes | `Standard_D4s_v5` |  |
| `spec.mode` | `enum` |  |  |  |
| `spec.osType` | `enum` |  |  |  |
| `spec.osSku` | `enum` |  |  |  |
| `spec.nodeCount` | `int32` |  |  |  |
| `spec.autoScalingEnabled` | `bool` |  |  |  |
| `spec.minCount` | `int32` |  |  |  |
| `spec.maxCount` | `int32` |  |  |  |
| `spec.maxPods` | `int32` |  |  |  |
| `spec.priority` | `enum` |  |  |  |
| `spec.evictionPolicy` | `enum` |  |  |  |
| `spec.spotMaxPrice` | `double` |  |  |  |
| `spec.nodeLabels` | `map<string, string>` |  |  |  |
| `spec.nodeTaints` | `[]string` |  |  |  |
| `spec.zones` | `[]string` |  |  |  |
| `spec.vnetSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.podSubnetId` | `string \| valueFrom` |  |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.orchestratorVersion` | `string` |  |  |  |
| `spec.osDiskSizeGb` | `int32` |  |  |  |
| `spec.osDiskType` | `enum` |  |  |  |
| `spec.kubeletDiskType` | `enum` |  |  |  |
| `spec.fipsEnabled` | `bool` |  |  |  |
| `spec.hostEncryptionEnabled` | `bool` |  |  |  |
| `spec.nodePublicIpEnabled` | `bool` |  |  |  |
| `spec.nodePublicIpPrefixId` | `string \| valueFrom` |  |  | AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`) |
| `spec.gpuInstance` | `enum` |  |  |  |
| `spec.gpuDriver` | `enum` |  |  |  |
| `spec.proximityPlacementGroupId` | `string` |  |  |  |
| `spec.hostGroupId` | `string` |  |  |  |
| `spec.capacityReservationGroupId` | `string` |  |  |  |
| `spec.scaleDownMode` | `enum` |  |  |  |
| `spec.snapshotId` | `string` |  |  |  |
| `spec.workloadRuntime` | `enum` |  |  |  |
| `spec.ultraSsdEnabled` | `bool` |  |  |  |
| `spec.temporaryNameForRotation` | `string` |  |  |  |
| `spec.upgradeSettings` | `AzureAksNodePoolUpgradeSettings` |  |  |  |
| `spec.upgradeSettings.maxSurge` | `string` |  |  |  |
| `spec.upgradeSettings.maxUnavailable` | `string` |  |  |  |
| `spec.upgradeSettings.drainTimeoutInMinutes` | `int32` |  |  |  |
| `spec.upgradeSettings.nodeSoakDurationInMinutes` | `int32` |  |  |  |
| `spec.upgradeSettings.undrainableNodeBehavior` | `enum` |  |  |  |
| `spec.kubeletConfig` | `AzureAksNodePoolKubeletConfig` |  |  |  |
| `spec.kubeletConfig.cpuManagerPolicy` | `enum` |  |  |  |
| `spec.kubeletConfig.cpuCfsQuotaEnabled` | `bool` |  | `true` |  |
| `spec.kubeletConfig.cpuCfsQuotaPeriod` | `string` |  |  |  |
| `spec.kubeletConfig.imageGcHighThreshold` | `int32` |  |  |  |
| `spec.kubeletConfig.imageGcLowThreshold` | `int32` |  |  |  |
| `spec.kubeletConfig.topologyManagerPolicy` | `enum` |  |  |  |
| `spec.kubeletConfig.allowedUnsafeSysctls` | `[]string` |  |  |  |
| `spec.kubeletConfig.containerLogMaxSizeMb` | `int32` |  |  |  |
| `spec.kubeletConfig.containerLogMaxFiles` | `int32` |  |  |  |
| `spec.kubeletConfig.podMaxPid` | `int32` |  |  |  |
| `spec.linuxOsConfig` | `AzureAksNodePoolLinuxOsConfig` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig` | `AzureAksNodePoolSysctlConfig` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.fsAioMaxNr` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.fsFileMax` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.fsInotifyMaxUserWatches` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.fsNrOpen` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.kernelThreadsMax` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netCoreNetdevMaxBacklog` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netCoreOptmemMax` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netCoreRmemDefault` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netCoreRmemMax` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netCoreSomaxconn` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netCoreWmemDefault` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netCoreWmemMax` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMin` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMax` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh1` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh2` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh3` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4TcpFinTimeout` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveIntvl` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveProbes` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveTime` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4TcpMaxSynBacklog` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4TcpMaxTwBuckets` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netIpv4TcpTwReuse` | `bool` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackBuckets` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackMax` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.vmMaxMapCount` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.vmSwappiness` | `int32` |  |  |  |
| `spec.linuxOsConfig.sysctlConfig.vmVfsCachePressure` | `int32` |  |  |  |
| `spec.linuxOsConfig.transparentHugePage` | `enum` |  |  |  |
| `spec.linuxOsConfig.transparentHugePageDefrag` | `enum` |  |  |  |
| `spec.linuxOsConfig.swapFileSizeMb` | `int32` |  |  |  |
| `spec.nodeNetworkProfile` | `AzureAksNodePoolNodeNetworkProfile` |  |  |  |
| `spec.nodeNetworkProfile.allowedHostPorts` | `[]AzureAksNodePoolAllowedHostPorts` |  |  |  |
| `spec.nodeNetworkProfile.allowedHostPorts[].portStart` | `int32` |  |  |  |
| `spec.nodeNetworkProfile.allowedHostPorts[].portEnd` | `int32` |  |  |  |
| `spec.nodeNetworkProfile.allowedHostPorts[].protocol` | `enum` |  |  |  |
| `spec.nodeNetworkProfile.applicationSecurityGroupIds` | `[]string` |  |  |  |
| `spec.nodeNetworkProfile.nodePublicIpTags` | `map<string, string>` |  |  |  |
| `spec.windowsProfile` | `AzureAksNodePoolWindowsProfile` |  |  |  |
| `spec.windowsProfile.outboundNatEnabled` | `bool` |  | `true` |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.kubernetesClusterId

`string | valueFrom` · required

The parent AKS cluster, by ARM ID.
Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ContainerService/managedClusters/{name}
Changing the cluster replaces the pool.

- references: AzureAksCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureAksCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.name

`string` · required

Agent-pool name: 1-12 lowercase letters and numbers, starting with a
letter (Windows pools: at most 6 characters). Changing the name
normally replaces the pool -- set temporary_name_for_rotation to
rotate through a stand-in instead.

- rule: Node pool names are 1-12 lowercase letters and numbers and start with a letter
- rule: {"required":true}

### spec.vmSize

`string` · required

Azure VM size for the pool's nodes, e.g. "Standard_D4s_v5" (general),
"Standard_E8s_v5" (memory-optimized), "Standard_NC24ads_A100_v4"
(GPU). Changing the size rotates the pool (see
temporary_name_for_rotation).

- default: `Standard_D4s_v5`
- rule: {"required":true}

### spec.mode

`enum`

The pool's purpose. Unspecified applies Azure's default (USER): runs
application workloads, may scale to zero, may be Windows or spot.
SYSTEM pools host cluster-critical pods -- they must be Linux,
on-demand (never spot), and keep at least one node.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_mode_unspecified` -- Not specified: Azure's default (User).
- `SYSTEM` -- Hosts cluster-critical system pods. Linux and on-demand only; keeps at least one node.
- `USER` -- Runs application workloads. May scale to zero, run Windows, or use spot capacity.

### spec.osType

`enum`

Node OS. Unspecified applies Azure's default (LINUX). WINDOWS
requires the cluster to carry a windows_profile, and the pool must
be USER mode.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_os_type_unspecified` -- Not specified: Azure's default (Linux).
- `LINUX` -- Linux nodes.
- `WINDOWS` -- Windows Server nodes -- requires the cluster's windows_profile and USER mode.

### spec.osSku

`enum`

Node OS image. Unspecified applies Azure's default for the OS type
(Ubuntu for Linux, Windows Server 2022 for Windows). AZURE_LINUX is
Microsoft's minimal container-host distro; version-pinned values pin
the OS major independent of Kubernetes version.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_os_sku_unspecified` -- Not specified: Azure's default image for the pool's OS type (currently Ubuntu for Linux, Windows Server 2022 for Windows).
- `UBUNTU` -- Ubuntu, version following AKS's default for the Kubernetes version.
- `UBUNTU_2204` -- Ubuntu 22.04 LTS, pinned.
- `UBUNTU_2404` -- Ubuntu 24.04 LTS, pinned.
- `AZURE_LINUX` -- Azure Linux (Microsoft's minimal container-host distro), version following AKS's default.
- `AZURE_LINUX_3` -- Azure Linux 3, pinned.
- `WINDOWS_2019` -- Windows Server 2019 (retired for new pools on Kubernetes >= 1.33).
- `WINDOWS_2022` -- Windows Server 2022.

### spec.nodeCount

`int32`

Fixed node count (0-1000) when autoscaling is off, or the initial
count when it is on. USER pools may sit at 0 (parked); SYSTEM pools
need at least 1. With autoscaling enabled, leave unset to let the
autoscaler own the count from the start.

- rule: node_count must be between 0 and 1000

### spec.autoScalingEnabled

`bool`

Whether the cluster autoscaler manages this pool's node count
between min_count and max_count. Tune cluster-wide autoscaler
behavior through the cluster's auto_scaler_profile.

### spec.minCount

`int32`

Minimum node count for autoscaling (0-1000; USER pools may scale to
zero). Requires auto_scaling_enabled.

- rule: min_count must be between 0 and 1000

### spec.maxCount

`int32`

Maximum node count for autoscaling (0-1000). Requires
auto_scaling_enabled.

- rule: max_count must be between 0 and 1000

### spec.maxPods

`int32`

Maximum pods per node. Unset applies Azure's plugin-dependent
default (250 for Azure CNI overlay, 30 for traditional Azure CNI).
Set at pool creation; raising it later rotates the pool.

### spec.priority

`enum`

VM priority. Unspecified applies Azure's default (REGULAR:
on-demand). SPOT draws from Azure's spare capacity at 30-90%
discount but can be evicted at any time -- fault-tolerant,
stateless, or batch workloads only. Spot pools must be USER mode and
cannot change priority in place.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_priority_unspecified` -- Not specified: Azure's default (Regular).
- `REGULAR` -- On-demand VMs -- guaranteed capacity.
- `SPOT` -- Spot VMs -- 30-90% cheaper, evictable at any time. USER pools only.

### spec.evictionPolicy

`enum`

What eviction does to a SPOT node. Unspecified applies Azure's
default (EVICTION_DELETE): the VM is deleted and billing stops.
EVICTION_DEALLOCATE stops the VM but keeps its disks -- faster
return when capacity comes back, but disks bill while stopped.
Only valid for SPOT pools; fixed at creation.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_eviction_policy_unspecified` -- Not specified: Azure's default (Delete).
- `EVICTION_DELETE` -- Evicted VMs are deleted -- billing stops entirely.
- `EVICTION_DEALLOCATE` -- Evicted VMs are stopped with disks kept -- faster return, disks keep billing.

### spec.spotMaxPrice

`double`

Ceiling price per node-hour in US dollars (up to 5 decimals), e.g.
0.27113. Leave unset for -1: pay up to the on-demand price and never
be evicted on price (the setting nearly everyone wants -- capacity
eviction still applies). Only valid for SPOT pools; fixed at
creation.

### spec.nodeLabels

`map<string, string>`

Kubernetes labels applied to this pool's nodes, for scheduling
(nodeSelector/affinity), e.g. {"workload": "gpu"}.

### spec.nodeTaints

`[]string`

Kubernetes taints applied to this pool's nodes, each
"key=value:Effect", e.g. "sku=gpu:NoSchedule" -- the standard way to
reserve special-hardware pools for the workloads that need them.
Spot pools automatically carry
"kubernetes.azure.com/scalesetpriority=spot:NoSchedule" in addition.

- rule: {"repeated":{"items":{"string":{"pattern":"^[^=]+=[^:]*:(NoSchedule|PreferNoSchedule|NoExecute)$"}}}}

### spec.zones

`[]string`

Availability zones to spread nodes across, e.g. ["1", "2", "3"].
Leave empty for regional (non-zonal) placement. Changing zones
rotates the pool.

- rule: {"repeated":{"items":{"string":{"in":["1","2","3"]}}}}

### spec.vnetSubnetId

`string | valueFrom`

The subnet this pool's nodes deploy into. Leave unset to inherit the
cluster's node subnet -- correct for nearly every pool. Reference a
different AzureSubnet to segment pools across subnets (e.g. a
dedicated subnet for an internet-exposed pool). Changing it replaces
the pool. A node pool is an ARM child of its cluster
(managedClusters/{cluster}/agentPools/{pool}) and lives there on a
diagram; the subnet is where its nodes attach, so the reference is
access, not placement -- otherwise a pool on its own subnet would be
drawn outside the cluster it belongs to.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.podSubnetId

`string | valueFrom`

A separate subnet for POD IPs (traditional Azure CNI with dynamic
pod IP allocation). Only meaningful alongside vnet_subnet_id on
clusters using non-overlay Azure CNI. Access, not placement, like
vnet_subnet_id.

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.orchestratorVersion

`string`

Kubernetes version for this pool's nodes. Unset follows the control
plane. Pools may lag the control plane by up to two minor versions
-- useful for canarying node upgrades pool by pool.

### spec.osDiskSizeGb

`int32`

OS disk size in GiB. Unset applies Azure's VM-size-dependent
default.

### spec.osDiskType

`enum`

OS disk placement. Unspecified applies Azure's default (MANAGED): a
persistent managed disk. EPHEMERAL places the OS disk on node-local
storage -- faster and free whenever the VM size's cache disk fits
the image; nodes are cattle, their OS disks need no durability.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_os_disk_type_unspecified` -- Not specified: Azure's default (Managed).
- `MANAGED` -- Persistent managed disk -- survives node restarts, billed separately.
- `EPHEMERAL` -- Node-local ephemeral disk -- faster, free, right for stateless nodes whenever the VM size's cache disk fits the OS image.

### spec.kubeletDiskType

`enum`

Where kubelet state (image layers, emptyDir volumes) lives.
Unspecified applies Azure's default (the OS disk). TEMPORARY uses
the VM's temp disk for higher IOPS on image-churn-heavy nodes.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_kubelet_disk_type_unspecified` -- Not specified: Azure's default (the OS disk).
- `OS` -- Kubelet state on the OS disk.
- `TEMPORARY` -- Kubelet state on the VM's temporary disk (higher IOPS for image churn; contents lost on deallocation, which kubelet state tolerates).

### spec.fipsEnabled

`bool`

Whether nodes get FIPS 140-2 validated OS images (compliance
environments). Changing it rotates the pool.

### spec.hostEncryptionEnabled

`bool`

Whether host-based encryption is enabled: data on the node's temp
disks and disk caches is encrypted at rest. Requires the
EncryptionAtHost feature on the subscription. Changing it rotates
the pool.

### spec.nodePublicIpEnabled

`bool`

Whether each node gets its own public IP -- niche direct-node-
ingress patterns (game servers, agent fleets). Egress normally flows
through the cluster's outbound path instead.

### spec.nodePublicIpPrefixId

`string | valueFrom`

Public IP prefix to allocate node public IPs from, so node IPs come
from one known, allowlistable CIDR. Requires node_public_ip_enabled.

- references: AzurePublicIpPrefix (`status.outputs.public_ip_prefix_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIpPrefix, name: <that resource's name>, fieldPath: status.outputs.public_ip_prefix_id}} -- a bare string does not parse

### spec.gpuInstance

`enum`

GPU Multi-Instance GPU (MIG) profile for A100 sizes, partitioning
each physical GPU into isolated slices (e.g. MIG1G = 7 slices).
Only for MIG-capable VM sizes. Changing it rotates the pool.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_gpu_instance_unspecified` -- Not specified: no MIG partitioning.
- `MIG1G` -- Seven 1g.5gb slices per GPU.
- `MIG2G` -- Three 2g.10gb slices per GPU.
- `MIG3G` -- Two 3g.20gb slices per GPU.
- `MIG4G` -- One 4g.20gb slice per GPU.
- `MIG7G` -- One 7g.40gb slice (the whole GPU as a single MIG device).

### spec.gpuDriver

`enum`

Whether AKS installs the NVIDIA GPU driver on GPU nodes. Unspecified
applies Azure's default (install when the VM size has a GPU). NONE
skips installation for teams running the GPU operator themselves.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_gpu_driver_unspecified` -- Not specified: Azure's default (install the driver on GPU sizes).
- `INSTALL` -- AKS installs the NVIDIA driver.
- `NONE` -- AKS skips driver installation -- for self-managed GPU operators.

### spec.proximityPlacementGroupId

`string`

ARM id of a Proximity Placement Group to co-locate nodes for minimal
inter-node latency (HPC). Plain ARM id (no Planton kind). Changing
it rotates the pool.

### spec.hostGroupId

`string`

ARM id of a Dedicated Host Group to place nodes on your isolated
physical hosts (compliance isolation). Plain ARM id. Changing it
rotates the pool.

### spec.capacityReservationGroupId

`string`

ARM id of a Capacity Reservation Group to draw guaranteed compute
capacity from. Plain ARM id. Changing it rotates the pool.

### spec.scaleDownMode

`enum`

What scale-down does with removed nodes. Unspecified applies Azure's
default (DELETE): nodes are deleted and stop billing. DEALLOCATE
stops them but keeps disks for faster scale-up at storage cost.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_scale_down_mode_unspecified` -- Not specified: Azure's default (Delete).
- `DELETE` -- Removed nodes are deleted -- billing stops entirely.
- `DEALLOCATE` -- Removed nodes are stopped (deallocated) -- compute billing stops but disks persist for faster scale-up.

### spec.snapshotId

`string`

ARM id of a node-pool snapshot to source this pool's configuration
from (replicating a known-good pool config across clusters). Plain
ARM id.

### spec.workloadRuntime

`enum`

Container runtime class. Unspecified runs standard containers
(OCIContainer, Azure's default). KATA_MSHV_VM_ISOLATION runs each
pod in a lightweight utility VM for kernel isolation; WASM_WASI runs
WebAssembly workloads (deprecated upstream -- prefer the Spin
operator).

Allowed values (use exactly as shown):

- `azure_aks_node_pool_workload_runtime_unspecified` -- Not specified: Azure's default (OCIContainer).
- `OCI_CONTAINER` -- Standard OCI containers.
- `KATA_MSHV_VM_ISOLATION` -- Kata Containers on Microsoft Hyper-V: each pod in a lightweight utility VM for kernel-level isolation.
- `WASM_WASI` -- WebAssembly (WASI) workloads. Deprecated upstream -- prefer the Spin operator for new WASM workloads.

### spec.ultraSsdEnabled

`bool`

Whether nodes may use Ultra SSD (zone-pinned, highest-IOPS) data
disks. Requires zones. Changing it rotates the pool.

### spec.temporaryNameForRotation

`string`

Stand-in pool name AKS uses to rotate this pool through otherwise
replace-forcing changes (vm_size, os_disk_type, fips_enabled...):
a temporary pool with this name carries the workloads while the real
pool is rebuilt. Same format as name. Set it proactively on
production pools.

- rule: Node pool names are 1-12 lowercase letters and numbers and start with a letter

### spec.upgradeSettings

`AzureAksNodePoolUpgradeSettings`

How node upgrades roll through the pool: surge or unavailability
sizing, drain behavior, and soak time between nodes.

- rule: Set at most one of max_surge or max_unavailable -- they are mutually exclusive rollout strategies

### spec.upgradeSettings.maxSurge

`string`

Extra nodes added during an upgrade, as a count ("2") or percentage
("10%" -- AKS's recommended default). Mutually exclusive with
max_unavailable.

### spec.upgradeSettings.maxUnavailable

`string`

Nodes that may be unavailable during an upgrade, as a count or
percentage -- upgrades in place without surge cost, at reduced
capacity. Mutually exclusive with max_surge.

### spec.upgradeSettings.drainTimeoutInMinutes

`int32`

Minutes to wait for a node to drain before giving up (honoring pod
disruption budgets). Unset keeps Azure's default (30).

### spec.upgradeSettings.nodeSoakDurationInMinutes

`int32`

Minutes to soak (wait) after each upgraded node before the next
(0-30). Unset upgrades continuously.

- rule: node_soak_duration_in_minutes must be between 0 and 30

### spec.upgradeSettings.undrainableNodeBehavior

`enum`

What happens to a node that will not drain (PDB-blocked).
Unspecified keeps Azure's default (the upgrade fails). CORDON
quarantines the node and proceeds; SCHEDULE lets pods return to it.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_undrainable_node_behavior_unspecified` -- Not specified: Azure's default -- the upgrade errors on an undrainable node.
- `CORDON` -- Cordon the undrainable node into quarantine and continue.
- `SCHEDULE` -- Leave the node schedulable and continue.

### spec.kubeletConfig

`AzureAksNodePoolKubeletConfig`

Kubelet tuning for this pool's nodes -- CPU manager policy, image GC
thresholds, container log rotation, pid limits. Unset fields keep
AKS defaults. Changing kubelet config rotates the pool.

### spec.kubeletConfig.cpuManagerPolicy

`enum`

CPU manager policy. Unspecified keeps AKS's default ("none").
STATIC gives Guaranteed-QoS pods exclusive cores -- latency-
sensitive workloads (trading, real-time media).

Allowed values (use exactly as shown):

- `azure_aks_node_pool_cpu_manager_policy_unspecified` -- Not specified: kubelet default ("none") -- shared CPU pool.
- `CPU_MANAGER_NONE` -- Shared CPU pool for all pods.
- `CPU_MANAGER_STATIC` -- Exclusive core pinning for Guaranteed-QoS pods with integer CPU requests.

### spec.kubeletConfig.cpuCfsQuotaEnabled

`bool` · optional (explicit presence)

Whether CPU CFS quota enforcement applies to containers with CPU
limits. Azure's default is true; disabling trades limit enforcement
for throttling-sensitive latency.

- default: `true`

### spec.kubeletConfig.cpuCfsQuotaPeriod

`string`

CPU CFS quota period, e.g. "100ms" (the kubelet default). Shorter
periods smooth throttling for latency-sensitive workloads.

### spec.kubeletConfig.imageGcHighThreshold

`int32`

Disk usage percentage that triggers image garbage collection
(0-100). Unset keeps the kubelet default (85).

- rule: image_gc_high_threshold is a percentage between 0 and 100

### spec.kubeletConfig.imageGcLowThreshold

`int32`

Disk usage percentage image GC frees down to (0-100). Unset keeps
the kubelet default (80).

- rule: image_gc_low_threshold is a percentage between 0 and 100

### spec.kubeletConfig.topologyManagerPolicy

`enum`

NUMA topology alignment policy for pod resources. Unspecified keeps
the kubelet default ("none").

Allowed values (use exactly as shown):

- `azure_aks_node_pool_topology_manager_policy_unspecified` -- Not specified: kubelet default ("none") -- no NUMA alignment.
- `TOPOLOGY_NONE` -- No alignment.
- `BEST_EFFORT` -- Prefer aligned placement, admit regardless.
- `RESTRICTED` -- Admit only pods whose preferred alignment is achievable.
- `SINGLE_NUMA_NODE` -- Admit only pods placeable on a single NUMA node.

### spec.kubeletConfig.allowedUnsafeSysctls

`[]string`

Unsafe sysctls (or patterns like "net.*") pods may set via their
security context. Empty means none -- the safe default.

### spec.kubeletConfig.containerLogMaxSizeMb

`int32`

Maximum container log file size in MB before rotation.

### spec.kubeletConfig.containerLogMaxFiles

`int32`

Maximum rotated container log files kept per container (>= 2).

- rule: container_log_max_files must be at least 2 (the active file plus one rotation)

### spec.kubeletConfig.podMaxPid

`int32`

Maximum processes per pod. Unset keeps the kubelet default
(unlimited, -1).

### spec.linuxOsConfig

`AzureAksNodePoolLinuxOsConfig`

Linux kernel and OS tuning -- sysctl values, transparent huge pages,
swap file. Linux pools only. Unset fields keep AKS defaults.
Changing OS config rotates the pool.

### spec.linuxOsConfig.sysctlConfig

`AzureAksNodePoolSysctlConfig`

Kernel sysctl overrides. Only the sysctls AKS allows are modeled;
each carries the ARM-enforced range in its validation.

### spec.linuxOsConfig.sysctlConfig.fsAioMaxNr

`int32`

fs.aio-max-nr (65536-6553500): max concurrent async I/O requests.

- rule: fs_aio_max_nr must be between 65536 and 6553500

### spec.linuxOsConfig.sysctlConfig.fsFileMax

`int32`

fs.file-max (8192-12000500): system-wide open file handle limit.

- rule: fs_file_max must be between 8192 and 12000500

### spec.linuxOsConfig.sysctlConfig.fsInotifyMaxUserWatches

`int32`

fs.inotify.max_user_watches (781250-2097152): inotify watch limit --
raise for file-watching-heavy workloads (IDEs, log shippers).

- rule: fs_inotify_max_user_watches must be between 781250 and 2097152

### spec.linuxOsConfig.sysctlConfig.fsNrOpen

`int32`

fs.nr_open (8192-20000500): per-process open file limit.

- rule: fs_nr_open must be between 8192 and 20000500

### spec.linuxOsConfig.sysctlConfig.kernelThreadsMax

`int32`

kernel.threads-max (20-513785): system-wide thread limit.

- rule: kernel_threads_max must be between 20 and 513785

### spec.linuxOsConfig.sysctlConfig.netCoreNetdevMaxBacklog

`int32`

net.core.netdev_max_backlog (1000-3240000): NIC ingress queue
length.

- rule: net_core_netdev_max_backlog must be between 1000 and 3240000

### spec.linuxOsConfig.sysctlConfig.netCoreOptmemMax

`int32`

net.core.optmem_max (20480-4194304): per-socket ancillary buffer
max.

- rule: net_core_optmem_max must be between 20480 and 4194304

### spec.linuxOsConfig.sysctlConfig.netCoreRmemDefault

`int32`

net.core.rmem_default (212992-134217728): default socket receive
buffer.

- rule: net_core_rmem_default must be between 212992 and 134217728

### spec.linuxOsConfig.sysctlConfig.netCoreRmemMax

`int32`

net.core.rmem_max (212992-134217728): max socket receive buffer.

- rule: net_core_rmem_max must be between 212992 and 134217728

### spec.linuxOsConfig.sysctlConfig.netCoreSomaxconn

`int32`

net.core.somaxconn (4096-3240000): listen backlog limit.

- rule: net_core_somaxconn must be between 4096 and 3240000

### spec.linuxOsConfig.sysctlConfig.netCoreWmemDefault

`int32`

net.core.wmem_default (212992-134217728): default socket send
buffer.

- rule: net_core_wmem_default must be between 212992 and 134217728

### spec.linuxOsConfig.sysctlConfig.netCoreWmemMax

`int32`

net.core.wmem_max (212992-134217728): max socket send buffer.

- rule: net_core_wmem_max must be between 212992 and 134217728

### spec.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMin

`int32`

net.ipv4.ip_local_port_range minimum (1024-60999).

- rule: net_ipv4_ip_local_port_range_min must be between 1024 and 60999

### spec.linuxOsConfig.sysctlConfig.netIpv4IpLocalPortRangeMax

`int32`

net.ipv4.ip_local_port_range maximum (32768-65535).

- rule: net_ipv4_ip_local_port_range_max must be between 32768 and 65535

### spec.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh1

`int32`

net.ipv4.neigh.default.gc_thresh1 (128-80000): ARP cache soft floor.

- rule: net_ipv4_neigh_default_gc_thresh1 must be between 128 and 80000

### spec.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh2

`int32`

net.ipv4.neigh.default.gc_thresh2 (512-90000): ARP cache soft
ceiling.

- rule: net_ipv4_neigh_default_gc_thresh2 must be between 512 and 90000

### spec.linuxOsConfig.sysctlConfig.netIpv4NeighDefaultGcThresh3

`int32`

net.ipv4.neigh.default.gc_thresh3 (1024-100000): ARP cache hard
limit.

- rule: net_ipv4_neigh_default_gc_thresh3 must be between 1024 and 100000

### spec.linuxOsConfig.sysctlConfig.netIpv4TcpFinTimeout

`int32`

net.ipv4.tcp_fin_timeout (5-120): FIN-WAIT-2 hold seconds.

- rule: net_ipv4_tcp_fin_timeout must be between 5 and 120

### spec.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveIntvl

`int32`

net.ipv4.tcp_keepalive_intvl (10-90): keepalive probe interval.

- rule: net_ipv4_tcp_keepalive_intvl must be between 10 and 90

### spec.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveProbes

`int32`

net.ipv4.tcp_keepalive_probes (1-15): unanswered probes before drop.

- rule: net_ipv4_tcp_keepalive_probes must be between 1 and 15

### spec.linuxOsConfig.sysctlConfig.netIpv4TcpKeepaliveTime

`int32`

net.ipv4.tcp_keepalive_time (30-432000): idle seconds before
keepalives start.

- rule: net_ipv4_tcp_keepalive_time must be between 30 and 432000

### spec.linuxOsConfig.sysctlConfig.netIpv4TcpMaxSynBacklog

`int32`

net.ipv4.tcp_max_syn_backlog (128-3240000): half-open connection
queue.

- rule: net_ipv4_tcp_max_syn_backlog must be between 128 and 3240000

### spec.linuxOsConfig.sysctlConfig.netIpv4TcpMaxTwBuckets

`int32`

net.ipv4.tcp_max_tw_buckets (8000-1440000): TIME-WAIT socket cap.

- rule: net_ipv4_tcp_max_tw_buckets must be between 8000 and 1440000

### spec.linuxOsConfig.sysctlConfig.netIpv4TcpTwReuse

`bool`

net.ipv4.tcp_tw_reuse: allow reusing TIME-WAIT sockets for new
outbound connections.

### spec.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackBuckets

`int32`

net.netfilter.nf_conntrack_buckets (65536-524288): conntrack hash
size.

- rule: net_netfilter_nf_conntrack_buckets must be between 65536 and 524288

### spec.linuxOsConfig.sysctlConfig.netNetfilterNfConntrackMax

`int32`

net.netfilter.nf_conntrack_max (131072-2097152): tracked connection
cap -- raise for high-connection-count proxies.

- rule: net_netfilter_nf_conntrack_max must be between 131072 and 2097152

### spec.linuxOsConfig.sysctlConfig.vmMaxMapCount

`int32`

vm.max_map_count (65530-262144): memory-map areas per process --
Elasticsearch famously needs 262144.

- rule: vm_max_map_count must be between 65530 and 262144

### spec.linuxOsConfig.sysctlConfig.vmSwappiness

`int32`

vm.swappiness (0-100): kernel swap eagerness.

- rule: vm_swappiness must be between 0 and 100

### spec.linuxOsConfig.sysctlConfig.vmVfsCachePressure

`int32`

vm.vfs_cache_pressure (0-100): dentry/inode cache reclaim pressure.

- rule: vm_vfs_cache_pressure must be between 0 and 100

### spec.linuxOsConfig.transparentHugePage

`enum`

Transparent Huge Pages mode. Unspecified keeps the OS default
("always"). Databases with sparse access patterns often want MADVISE
or NEVER.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_transparent_huge_page_unspecified` -- Not specified: the OS default ("always").
- `THP_ALWAYS` -- THP for all memory regions.
- `THP_MADVISE` -- THP only for madvise(MADV_HUGEPAGE) regions.
- `THP_NEVER` -- THP disabled.

### spec.linuxOsConfig.transparentHugePageDefrag

`enum`

Transparent Huge Pages defrag behavior. Unspecified keeps the OS
default ("madvise").

Allowed values (use exactly as shown):

- `azure_aks_node_pool_transparent_huge_page_defrag_unspecified` -- Not specified: the OS default ("madvise").
- `DEFRAG_ALWAYS` -- Synchronous defrag on every THP allocation.
- `DEFRAG_DEFER` -- Defer defrag to kswapd.
- `DEFRAG_DEFER_MADVISE` -- Defer generally, defrag synchronously for madvise regions.
- `DEFRAG_MADVISE` -- Defrag synchronously only for madvise regions.
- `DEFRAG_NEVER` -- Never defrag.

### spec.linuxOsConfig.swapFileSizeMb

`int32`

Swap file size in MB on each node. Unset means no swap -- the
Kubernetes-recommended default.

### spec.nodeNetworkProfile

`AzureAksNodePoolNodeNetworkProfile`

Node-level network hardening: allowed host ports, application
security groups for nodes, and tags on node public IPs.

### spec.nodeNetworkProfile.allowedHostPorts

`[]AzureAksNodePoolAllowedHostPorts`

Host port ranges pods may bind on the node -- each entry opens a
port range/protocol in the node's network security rules.

### spec.nodeNetworkProfile.allowedHostPorts[].portStart

`int32`

Start of the port range (1-65535).

- rule: port_start must be between 1 and 65535

### spec.nodeNetworkProfile.allowedHostPorts[].portEnd

`int32`

End of the port range (1-65535, >= port_start).

- rule: port_end must be between 1 and 65535

### spec.nodeNetworkProfile.allowedHostPorts[].protocol

`enum`

Protocol for the range.

Allowed values (use exactly as shown):

- `azure_aks_node_pool_host_port_protocol_unspecified` -- Not specified.
- `TCP` -- TCP.
- `UDP` -- UDP.

### spec.nodeNetworkProfile.applicationSecurityGroupIds

`[]string`

ARM ids of Application Security Groups the pool's nodes join, so NSG
rules can target "this pool's nodes" as a group. Plain ARM ids.

### spec.nodeNetworkProfile.nodePublicIpTags

`map<string, string>`

Azure tags applied to the node PUBLIC IPs (with
node_public_ip_enabled), e.g. routing preference tags. Set at pool
creation.

### spec.windowsProfile

`AzureAksNodePoolWindowsProfile`

Windows-pool networking: outbound NAT control. Windows pools only;
fixed at creation.

### spec.windowsProfile.outboundNatEnabled

`bool` · optional (explicit presence)

Whether Windows nodes get outbound NAT provided by the cluster.
Azure's default is true; disable only when routing Windows egress
through your own NAT arrangement. Fixed at creation.

- default: `true`

### spec.tags

`map<string, string>`

Free-form tags applied to the pool's VM scale set, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins. Updatable in place.

## Validation Rules

- `aks_node_pool_autoscaling_bounds`: With auto_scaling_enabled, set max_count (min <= max); without it, leave min_count and max_count unset
- `aks_node_pool_system_is_linux_ondemand`: SYSTEM pools must be Linux and on-demand: leave os_type unset (or LINUX) and priority unset (or REGULAR)
- `aks_node_pool_eviction_requires_spot`: eviction_policy is only valid for SPOT pools (priority SPOT)
- `aks_node_pool_spot_price_requires_spot`: spot_max_price is only valid for SPOT pools (priority SPOT)
- `aks_node_pool_spot_price_valid`: spot_max_price is either -1 (pay up to the on-demand price) or a positive US-dollar amount, e.g. 0.27113
- `aks_node_pool_spot_no_surge`: Spot pools do not support upgrade_settings max_surge or max_unavailable (azurerm's contract)
- `aks_node_pool_ip_prefix_requires_public_ip`: node_public_ip_prefix_id requires node_public_ip_enabled to be true
- `aks_node_pool_os_sku_matches_os_type`: os_sku must match os_type: Ubuntu/Azure Linux SKUs for Linux pools, Windows SKUs for Windows pools
- `aks_node_pool_windows_name_length`: Windows node pool names are at most 6 characters
- `aks_node_pool_windows_profile_requires_windows`: windows_profile is only valid for Windows pools (os_type WINDOWS)
- `aks_node_pool_linux_os_config_requires_linux`: linux_os_config is only valid for Linux pools

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureAksNodePool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.node_pool_id` | `string` | The Azure Resource Manager ID of the agent pool. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ContainerService/managedClusters/{cluster}/agentPools/{name} |
| `status.outputs.node_pool_name` | `string` | The name of the node pool (the value used in Kubernetes node labels like "kubernetes.azure.com/agentpool" for scheduling). |
| `status.outputs.node_image_version` | `string` | The node image version the pool is actually running (e.g. "AKSUbuntu-2204gen2containerd-202502.03.0") -- useful for auditing node-OS patch currency across pools. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kubernetesClusterId` | AzureAksCluster | `status.outputs.cluster_id` |
| `spec.vnetSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.podSubnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.nodePublicIpPrefixId` | AzurePublicIpPrefix | `status.outputs.public_ip_prefix_id` |

## See Also

- [Overview](../README.md)
