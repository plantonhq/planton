---
title: "Development OKE Cluster"
description: "This preset creates a minimal OKE cluster optimized for development, testing, and experimentation. It uses the Basic cluster type with flannel overlay networking to minimize setup complexity and..."
type: "preset"
rank: "03"
presetSlug: "03-development"
componentSlug: "container-engine-cluster"
componentTitle: "Container Engine Cluster"
provider: "oci"
icon: "package"
order: 3
---

# Development OKE Cluster

This preset creates a minimal OKE cluster optimized for development, testing, and experimentation. It uses the Basic cluster type with flannel overlay networking to minimize setup complexity and cost. The public API endpoint allows immediate `kubectl` access without VPN or Bastion configuration. No service load balancer subnet, KMS encryption, or image policy is configured -- these are production concerns that add friction to development workflows.

## When to Use

- Local development and experimentation with OKE before committing to a production architecture
- CI/CD environments that spin up ephemeral clusters for integration testing
- Learning and training environments where fast cluster creation matters more than security hardening
- Budget-constrained teams that want the simplest viable Kubernetes cluster on OCI

## Key Configuration Choices

- **Basic cluster type** (`type: basic_cluster`) -- Omits enhanced features (workload identity, add-on management, virtual node pools) that are unnecessary for development. Basic clusters have no additional cost beyond the underlying compute. Upgrade to `enhanced_cluster` when graduating to production.
- **Flannel overlay CNI** (`cniType: flannel_overlay`) -- Uses a software overlay network for pod-to-pod communication instead of allocating VCN IPs to each pod. This eliminates the need for a dedicated pod subnet with large IP space, making VCN design simpler for dev environments. The tradeoff is no NSG enforcement on individual pods and no OCI-native network policy support. For development, this is an acceptable simplification.
- **Public API endpoint without NSG** (`endpointConfig.isPublicIpEnabled: true`, no `nsgIds`) -- Enables immediate `kubectl` access from any network. The endpoint is protected by Kubernetes RBAC and the kubeconfig token, not by network-level restrictions. This is intentionally open for developer convenience. For shared dev clusters, add an NSG to restrict access to office/VPN CIDR blocks.
- **No service LB subnet** -- Omitted because dev clusters typically use `kubectl port-forward` or `NodePort` services rather than provisioning OCI Load Balancers. If you need `Service type: LoadBalancer`, add `options.serviceLbSubnetIds` with a subnet OCID.
- **No KMS encryption or image policy** -- Intentionally omitted. Dev clusters rarely need customer-managed encryption or image signing enforcement. Adding these increases provisioning time and creates dependencies on Vault and key management infrastructure.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the cluster will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN hosting the cluster | OCI Console > Networking > VCNs, or `OciVcn` status outputs |
| `<kubernetes-version>` | Kubernetes version for the control plane (e.g., `v1.30.1`) | `oci ce cluster-options list --cluster-option-id all` or OCI Console > Developer Services > Kubernetes Clusters > Create |
| `<api-endpoint-subnet-ocid>` | OCID of a public subnet hosting the API server endpoint | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |

## Related Presets

- **01-standard-production** -- Use instead for production workloads with enhanced features, VCN-native CNI, and service load balancer support
- **02-private-cluster** -- Use instead for regulated environments requiring private API endpoints and KMS encryption
