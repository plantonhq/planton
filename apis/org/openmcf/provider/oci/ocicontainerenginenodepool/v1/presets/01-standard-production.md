# Standard Production OKE Node Pool

This preset creates a general-purpose OKE node pool with E4.Flex compute shapes, VCN-native pod networking, and zero-downtime rolling upgrade settings. It distributes 3 nodes across 3 availability domains for high availability and configures Network Security Groups on both node and pod VNICs. This is the standard starting point for production OKE workloads and pairs naturally with the `01-standard-production` OKE cluster preset.

## When to Use

- Production Kubernetes workloads requiring general-purpose compute (web services, APIs, microservices)
- Any OKE cluster using VCN-native pod networking (`oci_vcn_ip_native` CNI) that needs a worker node pool
- Teams deploying their first node pool alongside a standard production OKE cluster
- Workloads that need zero-downtime rolling upgrades during Kubernetes version changes or shape modifications

## Key Configuration Choices

- **VM.Standard.E4.Flex with 2 OCPUs / 32 GB** (`nodeShape`, `nodeShapeConfig`) -- AMD EPYC-based flex shape at the 16 GB/OCPU ratio, which is the standard general-purpose configuration for Kubernetes workers. 2 OCPUs provides enough headroom for typical pod density without over-provisioning. Scale OCPUs up for compute-heavy workloads or add more nodes for horizontal scaling.
- **3 nodes across 3 ADs** (`nodeConfigDetails.size: 3`, three `placementConfigs`) -- OKE distributes nodes evenly across placement configs. One node per AD ensures the pool survives a full AD failure. For regional subnets, all three placement configs reference the same subnet OCID. Increase `size` to 6 or 9 for larger pools while maintaining even AD distribution.
- **VCN-native pod networking** (`podNetworkOptionDetails.cniType: oci_vcn_ip_native`) -- Each pod gets a real VCN IP from the pod subnet, enabling NSG enforcement on individual pods and OCI-native network policy. Must match the cluster's CNI configuration. The `maxPodsPerNode: 31` reflects the VNIC attachment limit for 2-OCPU E4.Flex shapes -- OKE calculates this automatically but setting it explicitly documents the constraint.
- **Worker and pod NSGs** (`nsgIds`, `podNsgIds`) -- Separate NSGs for node VNICs and pod VNICs allow fine-grained network segmentation. The worker NSG controls SSH access and inter-node communication; the pod NSG controls pod-to-pod and pod-to-service traffic.
- **Zero-downtime rolling upgrades** (`nodePoolCyclingDetails.maximumSurge: "1"`, `maximumUnavailable: "0"`) -- During Kubernetes version upgrades or shape changes, OKE creates one new node before draining and removing an old one. This guarantees no reduction in cluster capacity during the rolling operation. The tradeoff is slower upgrades (one node at a time) and temporarily higher compute cost.
- **60-minute eviction grace with forced cleanup** (`nodeEvictionSettings`) -- OKE attempts graceful pod eviction for up to 60 minutes during node operations. If pods cannot be evicted within the grace period (stuck finalizers, PDB violations), the node is forcefully deleted to prevent indefinitely blocked upgrades.
- **Node label for scheduling** (`initialNodeLabels: pool=general-purpose`) -- Enables `nodeSelector` and affinity rules to target this pool specifically, which becomes important when multiple node pools coexist (e.g., general-purpose alongside GPU or ARM pools).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the node pool will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<oke-cluster-ocid>` | OCID of the OKE cluster to attach this node pool to | OCI Console > Developer Services > Kubernetes Clusters, or `OciContainerEngineCluster` status outputs |
| `<kubernetes-version>` | Kubernetes version for the worker nodes (e.g., `v1.30.1`) | Should match or be compatible with the cluster's Kubernetes version. `oci ce node-pool-options get --node-pool-option-id all` |
| `<availability-domain-1>` | First availability domain name (e.g., `Uocm:PHX-AD-1`) | `oci iam availability-domain list` or OCI Console > Compute > Instances > Create Instance |
| `<availability-domain-2>` | Second availability domain name (e.g., `Uocm:PHX-AD-2`) | Same as above |
| `<availability-domain-3>` | Third availability domain name (e.g., `Uocm:PHX-AD-3`) | Same as above |
| `<worker-subnet-ocid>` | OCID of the regional subnet for worker node VNICs | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<worker-nsg-ocid>` | OCID of the NSG applied to worker node VNICs | OCI Console > Networking > Network Security Groups, or `OciNetworkSecurityGroup` status outputs |
| `<pod-subnet-ocid>` | OCID of the subnet for pod IP allocation (VCN-native CNI) | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<pod-nsg-ocid>` | OCID of the NSG applied to pod VNICs | OCI Console > Networking > Network Security Groups, or `OciNetworkSecurityGroup` status outputs |

## Related Presets

- **02-hardened-encrypted** -- Use instead for regulated environments requiring KMS boot volume encryption, in-transit encryption, and no SSH access
- **03-preemptible-dev** -- Use instead for development clusters where cost optimization via preemptible (spot) instances is preferred over high availability
