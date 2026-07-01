# OKE Environment InfraChart

This chart provisions a **production-ready Kubernetes environment on Oracle Cloud Infrastructure**:

* Custom VCN with internet gateway, NAT gateway, and service gateway
* Public subnet (API endpoint, service load balancers) with internet routing
* Private subnet (worker nodes) with NAT routing
* Network security groups for API endpoint and worker nodes
* OKE cluster with configurable type (basic/enhanced) and CNI (VCN-native/flannel)
* Autoscaled managed node pool across one or two availability domains
* Optional DNS zone for domain management

## Resources Created

| Resource | Kind | Condition |
|----------|------|-----------|
| Virtual Cloud Network | `OciVcn` | Always |
| Public Subnet | `OciSubnet` | Always |
| Private Subnet | `OciSubnet` | Always |
| API Endpoint NSG | `OciSecurityGroup` | Always |
| Worker Node NSG | `OciSecurityGroup` | Always |
| OKE Cluster | `OciContainerEngineCluster` | Always |
| Node Pool | `OciContainerEngineNodePool` | Always |
| DNS Zone | `OciDnsZone` | `enable_dns` |

## Parameters

| Name | Description | Default |
|------|-------------|---------|
| `compartment_ocid` | OCI compartment OCID | — |
| `vcn_cidr` | VCN CIDR block | `10.0.0.0/16` |
| `public_subnet_cidr` | Public subnet CIDR | `10.0.0.0/24` |
| `private_subnet_cidr` | Private subnet CIDR | `10.0.1.0/24` |
| `cluster_name` | Cluster name | `oke-demo` |
| `kubernetes_version` | K8s version | `v1.30.1` |
| `cluster_type` | basic_cluster / enhanced_cluster | `enhanced_cluster` |
| `cni_type` | oci_vcn_ip_native / flannel_overlay | `oci_vcn_ip_native` |
| `is_public_endpoint` | Public API endpoint | `false` |
| `node_pool_name` | Node pool name | `default-pool` |
| `node_shape` | Compute shape | `VM.Standard.E4.Flex` |
| `node_ocpus` | OCPUs per node | `2` |
| `node_memory_gb` | Memory per node (GB) | `32` |
| `node_image_id` | Custom OS image OCID | (OKE default) |
| `node_pool_size` | Node count | `3` |
| `availability_domain_1` | Primary AD | `US-ASHBURN-AD-1` |
| `availability_domain_2` | Secondary AD (blank to skip) | `US-ASHBURN-AD-2` |
| `ssh_public_key` | SSH key for nodes | (disabled) |
| `enable_dns` | Create DNS zone | `false` |
| `domain_name` | DNS zone domain | `example.com` |

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  VCN (10.0.0.0/16)                                      │
│  ┌─────────────────────┐  ┌───────────────────────────┐ │
│  │  Public Subnet       │  │  Private Subnet           │ │
│  │  10.0.0.0/24         │  │  10.0.1.0/24              │ │
│  │                      │  │                           │ │
│  │  • API Endpoint      │  │  • Worker Nodes           │ │
│  │  • Service LBs       │  │  • Node Pool              │ │
│  │  → Internet GW       │  │  → NAT GW                 │ │
│  └─────────────────────┘  └───────────────────────────┘ │
│                                                          │
│  ┌─────────────────────┐  ┌───────────────────────────┐ │
│  │  API Endpoint NSG   │  │  Worker NSG               │ │
│  │  • TCP 6443 from VCN│  │  • All from VCN           │ │
│  │  • TCP 12250        │  │  • All egress             │ │
│  │  • All egress       │  │                           │ │
│  └─────────────────────┘  └───────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```
