# Azure AKS Node Pool

Adds a standalone node pool to an existing Azure Kubernetes Service (AKS) cluster. The pool is deliberately separate from the cluster Infra Component so application capacity can scale, price (Spot), upgrade, and taint independently of the control plane and its built-in system pool. Each pool is the unit of compute shape — general, memory-optimized, GPU, Spot, or Windows — with a fully independent lifecycle: scale, upgrade, rotate, or delete a pool without touching the cluster or its sibling pools.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **AKS Agent Pool** -- a `containerservice.AgentPool` attached to the referenced cluster, with the VM size, fixed count or autoscaling bounds, OS family/image, and pool role (System / User) you choose
- **Pricing posture** -- on-demand (default) or Spot with eviction policy and max price; Spot pools arrive pre-tainted and skip surge upgrades
- **Placement** -- availability zones, optional node/pod subnets, instance-level public IPs, and dedicated-placement ARM IDs (proximity / host / capacity reservation groups)
- **Scheduling handles** -- node labels and taints so workloads select (or are repelled from) this pool
- **Optional tuning** -- surge upgrade settings, temporary rotation name, kubelet/Linux OS/node-network profiles, Windows outbound NAT
- **Azure Tags** -- pool-level tags on the scale set and its disks, plus Planton-derived resource tags

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AKS cluster** -- this pool attaches by ARM ID. Provide it directly or reference an AzureAksCluster Infra Component via ValueFromRef (`status.outputs.cluster_id`). The pool inherits the cluster's region and resource group.
- **Quota** -- the chosen VM size needs available vCPU quota in the cluster's region (Azure Portal → Quotas → Compute).
- **Windows pools** -- the parent cluster must already carry a Windows profile (admin credentials).
- **Subnet headroom (optional)** -- when placing the pool in its own subnet on flat Azure CNI, size for nodes plus max-pods IPs.

## Deploy

### Console

Open the deployment store, find **Azure AKS Node Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the pool configuration steps. Start from the **On-Demand General Purpose Node Pool** preset in the [Presets](#presets) tab for a production-ready user pool, or **Spot Cost-Optimized Node Pool** for fault-tolerant batch capacity.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureAksNodePool
metadata:
  name: prod-apps-pool
  org: acme-corp
  env: prod
spec:
  kubernetesClusterId:
    value: "/subscriptions/.../managedClusters/prod-aks"
  name: apps
  vmSize: Standard_D4s_v5
  mode: USER
  osType: LINUX
  autoScalingEnabled: true
  minCount: 0
  maxCount: 10
  zones: ["1", "2", "3"]
```

```shell
planton apply -f azure-aks-node-pool.yaml
```

This creates a zone-spread, autoscaled User pool that can scale to zero when idle. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the pool to the cluster deployed in the same InfraPipeline:

```yaml
spec:
  kubernetesClusterId:
    valueFrom:
      kind: AzureAksCluster
      name: production-aks
      fieldPath: status.outputs.cluster_id
  name: apps
  vmSize: Standard_D8s_v5
  mode: USER
  autoScalingEnabled: true
  minCount: 2
  maxCount: 20
  zones: ["1", "2", "3"]
  vnetSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: aks-apps
      fieldPath: status.outputs.subnet_id
```

The InfraPipeline resolves the dependency graph, deploys the cluster (and subnet) first, then provisions the pool with the resolved values.

## Key Configuration

These are the most important decisions when configuring an AKS node pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cluster attachment** -- `kubernetesClusterId` is the only way in. The pool has no `region` or `resourceGroup` of its own; it inherits both from the cluster. Name, OS family, and the cluster FK are create-time facts — changing any replaces the pool.

**Role and OS** -- `mode: USER` is the normal choice for application capacity. `mode: SYSTEM` pools must be Linux and on-demand (ARM rejects Windows or Spot system pools). `osSku` options follow `osType` so a Windows image cannot be expressed on a Linux pool.

**Scale** -- exactly one scale mode applies: `nodeCount` (fixed, including 0 to park the pool) or `autoScalingEnabled` with `minCount`/`maxCount`. User pools may scale to zero; that is the modern cost posture for idle capacity.

**Economics** -- leave `priority` unset for on-demand. `priority: SPOT` enables eviction policy and max price (`-1` pays up to on-demand). Spot pools skip surge upgrades and arrive with Azure's Spot taint — workloads need the matching toleration.

**Scheduling** -- `nodeLabels` and `nodeTaints` (`key=value:Effect`) dedicate the pool. Prefer one clear purpose label over many overlapping selectors.

**Rotation** -- set `temporaryNameForRotation` so ForceNew pool property changes (VM size, zones, OS disk type) rotate in place instead of recreating the pool.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureAksCluster** | `kubernetesClusterId` | `status.outputs.cluster_id` |
| **AzureSubnet** | `vnetSubnetId`, `podSubnetId` | `status.outputs.subnet_id` |
| **AzurePublicIpPrefix** | `nodePublicIpPrefixId` | `status.outputs.public_ip_prefix_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `node_pool_id` | Azure Resource Manager ID of the agent pool | Diagnostics, RBAC scoping, external references |
| `node_pool_name` | Name of the pool (matches `spec.name`) | kubectl / scripting (`agentpool` label) |
| `node_image_version` | The node image version currently running | Upgrade auditing |

## Common Patterns

**On-demand application capacity** -- a User pool of general-purpose VMs, autoscaled across three zones. The production default: guaranteed capacity, zone-resilient, and the autoscaler owns the count. Start from the **On-Demand General Purpose Node Pool** preset.

**Spot batch and CI capacity** -- `priority: SPOT` with `minCount: 0` trades evictability for a 30-90% discount. Right for fault-tolerant, stateless, or queue-driven work only: workloads need the toleration for Azure's Spot taint, and eviction can take nodes at any time. Start from the **Spot Cost-Optimized Node Pool** preset.

**Special-hardware pool reserved by taint** -- a GPU (or Windows) pool with a purpose label and a `NoSchedule` taint, so expensive nodes never fill with ordinary pods. Pods opt in with a toleration plus nodeSelector. Start from the **GPU Node Pool Reserved by Taint** preset; the same label-plus-taint pattern applies to Windows pools.

**Canary node upgrades pool by pool** -- pools may lag the control plane by up to two minor versions via `orchestratorVersion`. Upgrade one pool, soak, then roll the rest — the seam AKS gives you for de-risking node upgrades without a second cluster.

## Works With

- [**Azure AKS Cluster**](/infra-catalog/azure-aks-cluster) -- the parent cluster this pool attaches to by ARM ID; the pool inherits its region and resource group
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- optional dedicated node subnet (`vnetSubnetId`) or pod subnet (`podSubnetId`) for segmenting pools on Azure CNI
- [**Azure Public IP Prefix**](/infra-catalog/azure-public-ip-prefix) -- allocates node public IPs from one allowlistable CIDR when `nodePublicIpEnabled` is set
