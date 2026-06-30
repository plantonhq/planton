# OCI Container Engine Node Pool: Design Rationale and Research

## Introduction

The OciContainerEngineNodePool component manages OKE worker nodes — the compute instances that run Kubernetes pods. While the OciContainerEngineCluster component manages the control plane (API server, etcd, scheduler), the node pool is where workloads actually execute. Getting the abstraction right here determines whether users can express their compute requirements declaratively across a wide range of scenarios — from a 3-node dev pool to a 100-node production pool with GPU nodes, preemptible batch workers, and ARM cost-optimized instances — all composed via the same API surface.

The spec surface (13 top-level fields, 8 nested messages, 1 enum) mirrors OKE's node pool API with minimal abstraction. This document explains the design decisions that shaped the component.

## Why Cluster and Node Pool Are Separate Resources

This decision is documented in the OciContainerEngineCluster `docs/README.md` and summarized here for context:

1. **Different lifecycles.** Clusters are created once and rarely changed. Node pools are scaled, upgraded, replaced, and deleted frequently.
2. **Heterogeneous compute.** A production cluster typically has multiple node pools with different shapes, sizes, and policies (general-purpose, GPU, ARM, preemptible). A combined "cluster + nodes" resource cannot express this.
3. **Matches OKE's API.** Separate Terraform resources (`oci_containerengine_cluster`, `oci_containerengine_node_pool`). Separate Pulumi resources. Combining them would fight the provider API.
4. **Industry precedent.** AWS EKS (cluster + managed node group), GCP GKE (cluster + node pool), Azure AKS (cluster + agent pool) all separate control plane from compute.
5. **Infra-chart composability.** The OKE Environment infra chart composes one cluster with multiple node pools. Separate resources make this composition natural.

## Why Placement Configs Are a Repeated List

The `nodeConfigDetails.placementConfigs` field is `repeated PlacementConfig` — a list of placement configurations, not a map keyed by availability domain or subnet.

### Why Not a Map?

A map keyed by AD name (e.g., `placementConfigs: { "Uocm:PHX-AD-1": { subnetId: ... } }`) was considered. The list was chosen because:

1. **Per-placement heterogeneity.** Each placement config can have different settings: different subnets, different fault domain constraints, different capacity reservations, and different preemptible configurations. A map key would need to encode all of this, or the value would need all the same fields — making it equivalent to a list entry.

2. **Multiple entries for the same AD.** It is valid in OKE to have two placement configs for the same AD with different subnets or different fault domain constraints. A map keyed by AD would not support this.

3. **Matches the provider API.** Both the Terraform resource (`placement_configs` is a list of objects) and the Pulumi SDK use list semantics. Introducing a map would require a transformation layer that adds complexity and potential for bugs.

4. **Order independence.** OKE does not assign meaning to placement config ordering — it distributes nodes across placements evenly regardless of order. A list communicates this correctly.

### How OKE Distributes Nodes

Given `size: 9` and 3 placement configs, OKE creates 3 nodes per placement. If `size: 10`, one placement gets 4 nodes and two get 3. The distribution algorithm is internal to OKE and not configurable. Users control which ADs/subnets are available; OKE controls the distribution within those placements.

## Why Preemptible Config Is Per-Placement, Not Pool-Wide

Unlike EKS where spot configuration is a pool-wide launch template setting, OKE configures preemptible instances per placement config. This is an OCI API design choice that the component preserves.

### The Benefit

A single node pool can have preemptible instances in AD-1 and on-demand instances in AD-2. This enables hybrid cost strategies within one pool — for example, running preemptible in ADs with surplus capacity and on-demand in ADs with tight capacity.

### Why Only `isPreserveBootVolume`

The `PreemptibleNodeConfig` message has a single field: `isPreserveBootVolume`. This appears minimal, but it reflects the OKE API accurately:

- **Preemption action is always TERMINATE.** OKE only supports one preemption action: terminate the instance. There is no "stop" or "hibernate" option. The IaC modules hardcode `type: "TERMINATE"` in the preemption action.
- **Boot volume preservation is the only decision.** When OCI reclaims a preemptible instance, the only user choice is whether to keep the boot volume for forensics or delete it to avoid orphaned storage costs.

A boolean field (`isPreemptible`) on the placement config was considered instead of a nested message. The nested message was chosen because:

1. It matches the provider API structure (a `preemptible_node_config` block with a nested `preemption_action`).
2. If OCI adds more preemption options in the future (e.g., stop/hibernate, grace period), the message extends naturally without restructuring the placement config.

## Why Deprecated Provider Features Are Excluded

The spec proto comment explicitly documents four intentionally omitted provider features:

### node_image_id / node_image_name

These were the original way to specify node OS images. They are superseded by `node_source_details`, which supports both the image OCID and boot volume size in a single block. The Terraform provider documentation marks these as deprecated. Including them would offer two ways to do the same thing, violating the principle of one obvious way.

### subnet_ids / quantity_per_subnet

These were the original way to specify node placement. They are superseded by `node_config_details` (with placement configs, NSGs, KMS, pod networking). The old model did not support per-AD subnet selection, fault domains, capacity reservations, or preemptible instances. Including them would create a confusing overlap with `nodeConfigDetails`.

## Pod Networking at the Node Pool Level

For clusters using VCN-native CNI (`oci_vcn_ip_native`), pod networking is configured at the **node pool** level, not the cluster level. This is different from EKS/GKE/AKS where pod networking is primarily a cluster concern.

### Why Node Pool Level?

OKE's VCN-native implementation allocates pod IPs from subnets at the node level. Each node reserves a set of VNICs from the pod subnet, and pods on that node receive IPs from those VNICs. The configuration knobs that control this are:

- **Pod subnets** (`podSubnetIds`): Where pod IPs come from. Different node pools can use different pod subnets, enabling network-level workload isolation.
- **Pod NSGs** (`podNsgIds`): Security rules applied to pod VNICs. Different node pools can have different pod security postures.
- **Max pods per node** (`maxPodsPerNode`): Determines how many VNICs (and therefore IPs) each node reserves. This is shape-dependent — a large shape can attach more VNICs than a small one.
- **CNI type** (`cniType`): Must match the cluster's CNI, but is declared at the node pool level because the pod network option details block requires it.

### Why This Is Useful

The per-node-pool pod networking model enables scenarios that are difficult with cluster-level pod networking:

1. **Workload isolation.** Sensitive workloads in one node pool get their own pod subnet and pod NSG, isolating them at the VCN level from workloads in other node pools.
2. **IP budget management.** A GPU node pool with 4 large nodes needs far fewer pod IPs than a general-purpose pool with 100 small nodes. Separate pod subnets prevent one pool from exhausting IPs needed by another.
3. **Security differentiation.** Pod NSGs can differ between pools — the batch processing pool might have broader egress rules than the production application pool.

### The `cniType` Redundancy

The `cniType` field in `podNetworkOptionDetails` must match the cluster's CNI configuration. This appears redundant (why declare it if it must match?), but it's required by the OKE API. The pod network option details block is keyed by CNI type in the provider API — the block structure differs between flannel and VCN-native. Setting it explicitly avoids ambiguity and matches the provider schema.

## Node Eviction Settings: Design Choices

The `NodeEvictionSettings` message exposes three OKE controls for graceful node lifecycle:

### evictionGraceDuration

An ISO 8601 duration string (e.g., `PT30M` for 30 minutes, `PT0M` for immediate deletion). This is OKE's equivalent of Kubernetes pod `terminationGracePeriodSeconds`, but applied at the node pool operation level — it controls how long OKE attempts to drain pods before giving up.

A `google.protobuf.Duration` type was considered but rejected because:

1. OKE's API accepts an ISO 8601 string, not seconds.
2. The IaC modules pass the string directly to the provider. A Duration type would require conversion logic.
3. ISO 8601 is more readable in YAML (`PT30M` vs `1800s`).

### Two Force Flags

`isForceActionAfterGraceDuration` and `isForceDeleteAfterGraceDuration` are separate because they control different behaviors:

- **Force action** means OKE proceeds with the node pool operation (marks the node for replacement, removes it from the pool's management) even if pods are still running.
- **Force delete** means OKE terminates the underlying compute instance even if the OS hasn't finished shutting down.

The typical production configuration sets force action to `true` (prevent stuck upgrades) and force delete to `false` (allow graceful OS shutdown). Setting both to `true` is appropriate for stateless batch workers where speed matters more than graceful shutdown.

Both fields use `optional bool` (proto3 wrapper) to distinguish between "not set" (use OKE default) and "explicitly false."

## Node Pool Cycling Details: Design Choices

### Why Strings for Surge and Unavailable

`maximumSurge` and `maximumUnavailable` are `string` fields that accept either an integer (`"1"`) or a percentage (`"25%"`). Using `string` instead of a `oneof { int32, float }` was deliberate:

1. **Matches the provider API.** The Terraform provider accepts strings for both fields. The Pulumi SDK does the same.
2. **Simpler YAML authoring.** Users write `maximumSurge: "25%"` or `maximumSurge: "1"`. A `oneof` would require a wrapper key (`maximumSurge: { percentage: 25 }` or `maximumSurge: { count: 1 }`), making the common case more verbose.
3. **Kubernetes precedent.** Kubernetes `Deployment.spec.strategy.rollingUpdate` uses the same pattern — `maxSurge` and `maxUnavailable` accept both integers and percentages as strings.

### Default Behavior

When cycling details are not set, OKE uses its default replacement behavior (which varies by operation type). When explicitly configured, the cycling details apply to all node pool operations — Kubernetes version upgrades, shape changes, and image changes.

## What's Deferred

Based on the 80/20 principle, the following features are not in the initial implementation:

- **Auto-Scaling Configuration** — OKE supports node pool autoscaling via the Kubernetes Cluster Autoscaler (integrated into the OKE control plane for enhanced clusters). Auto-scaling is configured via Kubernetes annotations and cluster add-ons, not via the node pool Terraform/Pulumi resource. The scaling policy is a cluster-level or add-on-level concern, not a node pool spec field.

- **Virtual Node Pools** — Enhanced clusters support virtual node pools where OCI manages the compute infrastructure entirely (no customer-managed instances). Virtual nodes use a different API surface (`oci_containerengine_virtual_node_pool`) and have distinct configuration — container runtime constraints, pod shape selection, and different networking. This is a separate component candidate, not a configuration option on the existing node pool.

- **Cordon and Drain Operations** — The spec manages the desired state of a node pool (size, shape, version, placement). Operational actions like cordoning individual nodes or draining specific pods are Kubernetes-level operations, not IaC concerns. They are performed via `kubectl` or the Kubernetes API, not via the node pool resource.

- **Defined Tags** — Like the cluster component, the node pool uses Planton freeform tags. OCI defined tags (namespace-scoped, schema-validated) require a tag namespace to be created first. Defined tag support can be added when the tag namespace pattern is established across OCI components.

- **Extended Metadata and Agent Config** — OKE nodes support agent configurations (monitoring, management, vulnerability scanning) and extended metadata beyond the basic `nodeMetadata` map. These are typically configured via instance configurations or launch templates in more advanced OKE setups. They can be added if demand arises.

## Research Notes

### Node Shape VNIC Limits and maxPodsPerNode

For VCN-native CNI, `maxPodsPerNode` is constrained by the node shape's VNIC attachment limit. Each pod VNIC consumes one VNIC attachment slot. Common limits:

| Shape | Max VNICs | Practical maxPodsPerNode |
|-------|-----------|-------------------------|
| VM.Standard.E4.Flex (1 OCPU) | 2 | 31 |
| VM.Standard.E4.Flex (2+ OCPU) | 2-24 (scales with OCPUs) | 31-110+ |
| VM.Standard.A1.Flex | 2-24 (scales with OCPUs) | 31-110+ |
| VM.GPU.A10.1 | 24 | 110 |
| BM.Standard.E4.128 | 256 | 150 (OKE max) |

OKE caps `maxPodsPerNode` at 110 for VMs and 150 for bare metal, regardless of VNIC capacity. When `maxPodsPerNode` is not set, OKE uses a default based on the shape.

### Preemptible Capacity Behavior

OCI preemptible instances follow these patterns:

- Instances are created from spare capacity in the AD. If capacity is not available at creation time, the node pool may report nodes in a "creating" state indefinitely.
- Reclamation can happen at any time with a ~30-second warning (via instance metadata).
- OKE maintains the desired pool size — when a preemptible instance is reclaimed, OKE attempts to replace it. However, replacement depends on capacity availability.
- Spreading preemptible nodes across all ADs reduces correlated reclamation risk.

### Regional Subnet Behavior in Placement Configs

A regional subnet spans all ADs in a region. When using regional subnets, each placement config specifies a different AD but the same subnet OCID. OCI routes traffic to the correct AD based on the instance's AD placement. This simplifies subnet management but still requires one placement config entry per AD to control distribution.
