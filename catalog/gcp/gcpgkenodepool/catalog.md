# GCP GKE Node Pool

Deploys a GKE node pool — a group of Compute Engine VMs with one shared configuration, attached to a `GcpGkeCluster`. Covers the full pool surface: fixed or autoscaled sizing (per-zone or total bounds), the complete node VM shape (machine type, disks, local SSDs, GPUs), Spot/preemptible capacity models, reservations and Dynamic Workload Scheduler paths, node identity and hardening, Kubernetes scheduling metadata (labels/taints/tags), surge or blue-green upgrade rollouts, per-pool networking, and kubelet/OS tuning. Node pools apply to Standard clusters only — Autopilot clusters manage their own nodes.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **GKE Node Pool** — attached to the referenced cluster, with one managed instance group per zone
- **Sizing** — a fixed node count, cluster-autoscaler management between your bounds (per-zone or total, scale-to-zero capable), or GKE's default
- **Node VMs** — the shared machine shape: machine type, boot disk (optionally CMEK-encrypted), node image, local NVMe SSDs, GPU accelerators with driver installation
- **Scheduling Surface** — Kubernetes node labels and taints, GCE network tags and resource labels
- **Management** — auto-repair and auto-upgrade (both on by default), with surge or blue-green upgrade rollouts

Node pools apply to Standard clusters only — Autopilot clusters manage nodes themselves.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with credentials for the cluster's GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Standard GKE cluster** — deploy a `GcpGkeCluster` first and reference it; the pool resolves the cluster's name and location from its outputs.
- **Zone availability for GPUs** — accelerator models are the most zone-constrained resource in GCP; check `gcloud compute accelerator-types list` before pinning zones.
- **A minimal node service account** (recommended) — a dedicated `GcpServiceAccount` with logging, monitoring, and Artifact Registry read; workload permissions flow through Workload Identity, not node scopes.

## Deploy

### Console

Open the deployment store, find **GCP GKE Node Pool**, and click **Deploy**. The creation wizard walks the decisions in the order a platform engineer designs a pool — cluster attachment, sizing, machine shape, capacity model, identity, scheduling, upgrades. Start from the **On-Demand Autoscaling Pool** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeNodePool
metadata:
  name: general-pool
  org: acme-corp
  env: prod
spec:
  clusterName:
    valueFrom:
      kind: GcpGkeCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  location:
    valueFrom:
      kind: GcpGkeCluster
      name: platform-cluster
      fieldPath: status.outputs.location
  autoscaling:
    minNodes: 0
    maxNodes: 4
  nodeConfig:
    machineType: n2-standard-8
    diskType: pd-balanced
    diskSizeGb: 200
    workloadMetadataMode: GKE_METADATA
```

```shell
planton apply -f gke-node-pool.yaml
```

This creates an autoscaled pool (0-4 nodes per zone — free while idle) of n2-standard-8 nodes with auto-repair and auto-upgrade at their GKE defaults (on). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the cluster and its pools compose through ValueFromRef:

```yaml
spec:
  clusterName:
    valueFrom:
      kind: GcpGkeCluster
      name: platform-cluster
      fieldPath: status.outputs.name
  location:
    valueFrom:
      kind: GcpGkeCluster
      name: platform-cluster
      fieldPath: status.outputs.location
  nodeConfig:
    serviceAccount:
      valueFrom:
        kind: GcpServiceAccount
        name: gke-nodes
        fieldPath: status.outputs.email
    spot: true
    taints:
      - key: capacity
        value: spot
        effect: NO_SCHEDULE
```

The InfraPipeline deploys the cluster and service account first, then attaches the pool with all references resolved.

## Key Configuration

These are the most important decisions when configuring a node pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Sizing** (`nodeCount` XOR `autoscaling`) — autoscaled with a scale-to-zero minimum is the production norm. On REGIONAL clusters, `nodeCount` and per-zone bounds multiply by the zone count (max 4 across 3 zones = 12 nodes); `totalMinNodes`/`totalMaxNodes` cap the absolute number instead.

**Machine type** (`nodeConfig.machineType`, immutable) — GKE's e2-medium default is undersized for real workloads. Size the machine for the pool's purpose: e2-standard for general, n2 for balanced (and gVNIC/TIER_1), c3 for compute-bound, a2/a3/g2 for GPUs.

**Capacity model** (`nodeConfig.spot`, immutable) — Spot is 60-91% cheaper with 30-second preemption notice. Pair it with scale-to-zero autoscaling, `locationPolicy: ANY`, and a taint so only fault-tolerant workloads land.

**Management** (`management.autoRepair` / `autoUpgrade`, both default true) — keep both on; pools on a release channel require auto-upgrade. Pin `version` only with auto-upgrade off.

**Upgrade rollout** (`upgradeSettings`) — surge (default) rolls nodes gradually; blue-green provisions a full parallel node set with a soak window — the strongest rollback story at double capacity in flight.

**Scheduling fence** (`nodeConfig.labels` / `taints`) — labels let workloads choose this pool; taints keep everything else out. Special-purpose pools want both.

**Node tuning** (`nodeConfig.kubeletConfig` / `linuxNodeConfig`) — the full kubelet and OS surface is modeled: CPU/memory/topology managers for latency-critical pinning, soft-eviction thresholds with grace periods, crash-loop backoff caps, image GC by threshold and age, sysctls, cgroup mode, transparent hugepages, and swap with sizing profiles. Set only what you need — unset fields keep GKE defaults.

**Teardown behavior** (`nodeDrainConfig`, `deletionPolicy`) — pace how nodes drain when the pool is deleted (grace periods, PodDisruptionBudget respect with a timeout; note that customized node drain must first be enabled on your project by GCP support), and control what a destroy is allowed to do: PREVENT fails the plan, ABANDON drops the pool from state without touching GKE.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpGkeCluster** | `clusterName` | `status.outputs.name` |
| **GcpGkeCluster** | `location` | `status.outputs.location` |
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** | `nodeConfig.serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** | `nodeConfig.bootDiskKmsKey` | `status.outputs.key_id` |
| **GcpComputeImage** (optional) | `nodeConfig.secondaryBootDisks[].diskImage` | `status.outputs.image_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `node_pool_name` | The pool name as created in GKE | gcloud commands, in-cluster references |
| `node_pool_id` | Fully qualified resource ID (`projects/{p}/locations/{l}/clusters/{c}/nodePools/{n}`) | Dataproc on GKE, API automation |
| `instance_group_urls` | The managed instance groups backing the pool (one per zone) | Instance-group-targeted load-balancer backends |
| `min_nodes` / `max_nodes` | Effective size bounds (autoscaling bounds or the fixed count) | Capacity dashboards |
| `current_node_count` | Nodes per zone at the last deploy (a snapshot — autoscaled pools drift) | Monitoring |
| `location` | The pool's region or zone | Automation |
| `version` | The Kubernetes version on the pool's nodes | Version dashboards, upgrade automation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**On-demand autoscaling pool** — the general-purpose workhorse: autoscaled, on-demand, pd-balanced, GKE_METADATA. Start from the **On-Demand Autoscaling Pool** preset.

**Spot cost-optimized pool** — scale-to-zero Spot capacity with the ANY scale-up policy and a `capacity=spot` taint for fault-tolerant batch. Start from the **Spot Cost-Optimized Pool** preset.

**GPU accelerated pool** — GPU cards with GKE-managed drivers, image streaming for large ML images, and gVNIC for distributed training. Start from the **GPU-Accelerated Pool** preset.

## Works With

- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) — the control plane this pool attaches to, referenced by its `name` and `location` outputs
- [**GCP Service Account**](/infra-catalog/gcp-service-account) — the minimal node identity for production pools
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) — CMEK encryption for node boot disks
