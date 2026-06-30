# Overview

The **OCI Container Engine Node Pool API Resource** provides a consistent and standardized interface for deploying and managing Oracle Cloud Infrastructure Container Engine for Kubernetes (OKE) node pools. A node pool is a set of worker nodes within an OKE cluster that share the same configuration — compute shape, OS image, Kubernetes version, and placement rules. Multiple node pools can be attached to a single cluster, enabling heterogeneous workloads (GPU nodes for ML, ARM nodes for cost optimization, preemptible nodes for batch jobs). This component wraps the `oci_containerengine_node_pool` API surface with the standard Planton KRM pattern.

## Purpose

This API resource streamlines the deployment of OKE node pools by offering a unified interface that covers the full range of worker node configurations — from a minimal development pool to a production-grade multi-AD pool with VCN-native pod networking, KMS encryption, and rolling upgrade strategies. It enables users to:

- **Right-Size Compute with Flex Shapes**: Configure OCPUs and memory independently on flex shapes (E4.Flex, A1.Flex) via `nodeShapeConfig`. This allows precise resource allocation — a 2-OCPU / 32 GB node for memory-intensive workloads, or a 4-OCPU / 16 GB node for CPU-bound services — without being locked to fixed shape ratios.
- **Distribute Nodes Across Availability Domains**: Define placement configs for each AD, each with its own subnet, fault domain constraints, and optional capacity reservations. OKE distributes the pool's nodes as evenly as possible across the configured placements.
- **Use Preemptible (Spot) Instances**: Enable preemptible nodes on a per-placement basis for fault-tolerant and batch workloads. Preemptible instances cost significantly less but can be reclaimed by OCI when capacity is needed.
- **Configure VCN-Native Pod Networking**: For clusters using `oci_vcn_ip_native` CNI, configure pod subnets, pod NSGs, and max pods per node at the node pool level. Each pod receives a VCN IP address, enabling network-level isolation and observability.
- **Encrypt Boot Volumes at Rest and In Transit**: Specify a KMS key for boot volume encryption and enable in-transit encryption for paravirtualized attachments — meeting compliance requirements without per-node configuration.
- **Control Rolling Upgrades**: Configure node pool cycling details (maximum surge, maximum unavailable) and eviction settings (grace duration, force actions) to control how OKE replaces nodes during Kubernetes version upgrades or shape changes.
- **Label Nodes for Scheduling**: Apply Kubernetes labels to nodes at pool creation time, enabling workload scheduling via `nodeSelector` and affinity rules without post-creation label management.
- **Compose with Other OCI Resources**: Reference OciContainerEngineCluster, OciCompartment, OciSubnet, and OciSecurityGroup outputs via `StringValueOrRef` for declarative, cross-resource dependency chains.

## Key Features

- **Consistent Interface**: Aligns with the Planton pattern for deploying cloud infrastructure across providers.
- **Flex Shape Support**: Independent OCPU and memory configuration via `nodeShapeConfig` for E4.Flex, A1.Flex, and other flex shapes. Fixed shapes (E4.1, GPU.A10.1) work without shape config.
- **Multi-AD Placement**: Repeated placement configs distribute nodes across availability domains for high availability. Each placement specifies its AD, subnet, fault domains, and optional capacity reservation.
- **Preemptible Instances**: Per-placement preemptible configuration for cost-sensitive, fault-tolerant workloads. The preemption action is always TERMINATE; the only configurable option is whether to preserve the boot volume.
- **VCN-Native Pod Networking**: Pod subnet allocation, pod NSGs, and max pods per node configuration. Required for clusters using `oci_vcn_ip_native` CNI, where each pod gets a VCN IP address.
- **Boot Volume Encryption**: KMS key for encrypting boot volumes at rest (`kmsKeyId`). In-transit encryption for paravirtualized volume attachments (`isPvEncryptionInTransitEnabled`).
- **Custom OS Image**: Specify a platform or custom image OCID and boot volume size via `nodeSourceDetails`. When omitted, OKE uses the default Oracle Linux image for the cluster's Kubernetes version.
- **Node Labels and Metadata**: Kubernetes labels applied at node join time (`initialNodeLabels`). Instance metadata key/value pairs for cloud-init user data (`nodeMetadata`).
- **Eviction Settings**: Configurable grace duration (ISO 8601), force action, and force delete behavior during node pool operations (scale-down, upgrades, shape changes).
- **Rolling Upgrade Strategy**: Node pool cycling with configurable maximum surge and maximum unavailable counts or percentages. Controls how aggressively nodes are replaced during upgrades.
- **Automatic Tagging**: Standard Planton freeform tags applied to the node pool and its node config (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Independent Version Management**: Kubernetes version can be set independently of the cluster version, enabling rolling worker node upgrades without control plane changes.
- **Infra-Chart Composability**: Exports 2 stack outputs (`nodePoolId`, `kubernetesVersion`) for downstream `StringValueOrRef` references. Consumes `clusterId` from OciContainerEngineCluster.

## How OKE Node Pools Differ from Other Providers

Understanding these differences is essential when coming from EKS, GKE, or AKS:

| Aspect | OKE Node Pool | EKS Managed Node Group | GKE Node Pool | AKS Agent Pool |
|--------|--------------|----------------------|---------------|----------------|
| **Compute model** | OCI shapes (fixed and flex with independent OCPU/memory) | EC2 instance types (fixed vCPU/memory ratios) | GCE machine types (predefined and custom) | Azure VM sizes (fixed vCPU/memory ratios) |
| **Placement** | Availability domains + fault domains + subnets (per-placement config) | Subnets across AZs (single config) | Zones (list of zones) | Availability zones (list of zones) |
| **Preemptible/spot** | Per-placement config (not pool-wide) | Launch template spot config (pool-wide) | Per-pool preemptible flag | Per-pool spot config |
| **Pod networking** | VCN-native (pods get VCN IPs, configured at node pool level) or flannel | VPC CNI (pods get VPC IPs, configured at cluster level) | VPC-native alias IPs (configured at cluster level) | Azure CNI (configured at cluster level) |
| **Capacity reservation** | Per-placement config | Capacity reservation group on launch template | Reservation affinity on node pool | Capacity reservation group |
| **Version management** | Independent `kubernetesVersion` per node pool (inherits cluster version when omitted) | Independent version per node group | Independent version per node pool | Independent version per agent pool |
| **Upgrade strategy** | Node pool cycling (max surge/unavailable) + eviction settings | Update config (max unavailable) | Surge upgrade (max surge/unavailable) | Max surge |
| **Node tagging** | Freeform tags on node pool and node config | Tags on EC2 instances via launch template | Labels and metadata on GCE instances | Tags on Azure VMs |

Key distinctions for OCI newcomers:

- **Per-Placement Preemptible.** Unlike EKS where spot is a pool-wide launch template setting, OKE configures preemptible instances per placement config. A single node pool can have preemptible instances in AD-1 and on-demand instances in AD-2. This is unusual and enables hybrid cost strategies within one pool.
- **Flex Shapes Are the Norm.** OCI flex shapes (E4.Flex, A1.Flex) allow independent OCPU and memory scaling. This is more granular than AWS/GCP/Azure instance types where vCPU and memory are fixed ratios. `nodeShapeConfig` is practically required for any flex shape — without it, OCPUs and memory default to the shape's minimum.
- **Pod Networking at Node Pool Level.** For VCN-native CNI clusters, pod subnet and pod NSG configuration happens at the node pool level via `podNetworkOptionDetails`, not at the cluster level. This means different node pools can use different pod subnets — useful for isolating workloads at the network level.
- **Fault Domains Are OCI-Specific.** Each OCI availability domain contains 3 fault domains (independent power, cooling, networking). The `faultDomains` field constrains which fault domains receive nodes, enabling anti-affinity at the infrastructure level for applications that need it.

## Critical Constraints

- **Cluster Is Immutable**: Changing `clusterId` after creation forces node pool recreation. A node pool cannot be moved between clusters.
- **Compartment Change Forces Recreation**: Moving a node pool to a different compartment via `compartmentId` change forces node pool recreation.
- **Node Shape Change Requires Cycling**: Changing `nodeShape` or `nodeShapeConfig` triggers node replacement via the cycling strategy. Nodes are not updated in place — new nodes are created with the new shape and old nodes are drained and deleted.
- **Version Must Be Compatible**: The node pool `kubernetesVersion` must be within one minor version of the cluster's control plane version. OKE enforces this at the API level.
- **Placement Configs and Subnets**: Each placement config's subnet must be in the specified availability domain (or be a regional subnet). For VCN-native CNI, pod subnets must have enough available IPs for `maxPodsPerNode * nodesInThatAD`.
- **Preemptible Termination**: Preemptible instances can be terminated at any time by OCI. Workloads must be fault-tolerant. OKE does not automatically replace preemptible nodes that are reclaimed — the pool's desired size is maintained, but replacement timing depends on capacity availability.
- **CNI Type Must Match Cluster**: The `podNetworkOptionDetails.cniType` must match the cluster's CNI configuration. A node pool cannot use VCN-native CNI on a flannel cluster or vice versa.
- **Image Lifecycle**: Node images are tied to Kubernetes versions. When upgrading `kubernetesVersion`, the image may also need updating if using a custom `nodeSourceDetails.imageId`. Default images (when `nodeSourceDetails` is omitted) are automatically matched to the Kubernetes version.

## Use Cases

- **General-Purpose Production Pool**: E4.Flex nodes with 4 OCPUs and 64 GB RAM distributed across 3 ADs with VCN-native CNI, KMS encryption, and rolling upgrade settings. The standard pattern for running stateless microservices.
- **GPU/ML Acceleration**: GPU.A10.1 or GPU.A10.2 nodes with Kubernetes labels for `nodeSelector`-based scheduling. Training and inference workloads are directed to GPU nodes while other services run on general-purpose pools.
- **Batch Processing with Preemptible Nodes**: Preemptible instances in multiple ADs for fault-tolerant batch jobs. Significant cost savings (60-90% discount) with the tradeoff of potential instance reclamation. Kubernetes `PodDisruptionBudget` and job retry logic handle preemption gracefully.
- **ARM Cost Optimization**: A1.Flex (Ampere Altra) nodes for cost-optimized workloads that support ARM64 architecture. Approximately 50% cost reduction compared to x86 flex shapes at equivalent performance for compatible workloads.
- **Multi-Tier Architecture**: Multiple node pools per cluster — a small pool for system services (monitoring, ingress), a larger pool for application workloads, and a GPU pool for ML inference. Kubernetes labels and taints route workloads to the appropriate pool.
- **Compliance-Sensitive Workloads**: KMS-encrypted boot volumes, in-transit encryption, specific fault domain placement for data sovereignty, and capacity reservations for guaranteed compute availability.
- **Kubernetes Version Rolling Upgrades**: Pin `kubernetesVersion` on the node pool independently of the cluster. Upgrade the control plane first, then roll out worker node upgrades pool by pool with controlled surge and eviction settings.

## Production Features

This resource provides complete support for production-grade OKE node pool deployments, including:

- **Multi-AD High Availability**: Placement configs across all availability domains in a region. OKE distributes nodes evenly, and Kubernetes scheduler respects topology spread constraints for pod-level HA.
- **VCN-Native Pod Networking**: Pod subnets, pod NSGs, and max pods per node — the full VCN-native configuration surface. Different node pools can use different pod subnets for workload isolation at the network layer.
- **Encryption at Rest and In Transit**: KMS key for boot volume encryption ensures data at rest is protected with customer-managed keys. In-transit encryption protects data between the instance and its paravirtualized volumes.
- **Rolling Upgrade Strategy**: Node pool cycling with `maximumSurge` and `maximumUnavailable` controls how aggressively nodes are replaced. Combined with eviction settings (grace duration, force actions), this enables zero-downtime upgrades for production workloads.
- **Capacity Reservations**: Per-placement capacity reservation guarantees that compute capacity is available when nodes need to be created or replaced. Critical for production workloads where spot capacity is not acceptable.
- **Preemptible Cost Optimization**: Per-placement preemptible configuration enables mixed on-demand/preemptible pools. The `isPreserveBootVolume` option allows boot volume forensics when preemptible instances are terminated.
- **Freeform Tagging**: Standard Planton labels applied as OCI freeform tags on both the node pool resource and its node config for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Infra-Chart Composability**: Designed to compose with OciContainerEngineCluster (upstream dependency), OciCompartment, OciSubnet, and OciSecurityGroup via `StringValueOrRef`.
