# Standard Production OKE Cluster

This preset creates an Enhanced OKE cluster with VCN-native pod networking and a public Kubernetes API endpoint protected by a Network Security Group. It configures the service load balancer subnet so that Kubernetes `Service type: LoadBalancer` resources work out of the box, and sets explicit pod/service CIDRs to avoid overlap with VCN addressing. This is the standard starting point for the vast majority of production OKE deployments.

## When to Use

- Production Kubernetes workloads on OCI where pods need VCN-level network identity (NSGs on pods, native network policies)
- Teams adopting OKE for the first time who want a secure, fully-featured cluster without unnecessary complexity
- Any deployment where the Kubernetes API needs to be reachable from CI/CD pipelines or developer machines outside the VCN (public endpoint with NSG restriction)
- Clusters that will serve traffic via OCI Load Balancers created by Kubernetes Service resources

## Key Configuration Choices

- **Enhanced cluster type** (`type: enhanced_cluster`) -- Enables workload identity (pods can authenticate to OCI APIs without static credentials), cluster add-on lifecycle management, and virtual node pool support. Enhanced is the recommended type for all production clusters. Use `basic_cluster` only when cost or simplicity is the primary concern.
- **VCN-native CNI** (`cniType: oci_vcn_ip_native`) -- Each pod receives a real VCN IP address from a pod subnet, making pods first-class VCN citizens. This enables Network Security Groups on pods, Kubernetes NetworkPolicy enforcement via OCI-native mechanisms, and direct pod-to-service communication without NAT. Requires a dedicated pod subnet with sufficient IP space. The alternative `flannel_overlay` uses an overlay network that is simpler but less capable.
- **Public API endpoint with NSG** (`endpointConfig.isPublicIpEnabled: true`, `endpointConfig.nsgIds`) -- Assigns a public IP to the Kubernetes API server, enabling `kubectl` access from outside the VCN. The NSG should restrict ingress to trusted CIDR blocks (corporate VPN ranges, CI/CD runner IPs). If your security posture requires zero public exposure, use preset 02-private-cluster instead.
- **Service load balancer subnet** (`options.serviceLbSubnetIds`) -- Tells OKE where to place load balancers created by Kubernetes `Service type: LoadBalancer` resources. Without this, Service-based load balancers fail to provision. Use a public subnet for internet-facing services or a private subnet for internal services.
- **Explicit pod and service CIDRs** (`options.kubernetesNetworkConfig.podsCidr: 10.244.0.0/16`, `servicesCidr: 10.96.0.0/16`) -- The Kubernetes defaults. Setting them explicitly avoids surprises and documents the addressing scheme. These CIDRs must not overlap with the VCN CIDR or each other. If your VCN uses 10.244.0.0/16 or 10.96.0.0/16, change these values.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the cluster will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<vcn-ocid>` | OCID of the VCN hosting the cluster | OCI Console > Networking > VCNs, or `OciVcn` status outputs |
| `<kubernetes-version>` | Kubernetes version for the control plane (e.g., `v1.30.1`) | `oci ce cluster-options list --cluster-option-id all` or OCI Console > Developer Services > Kubernetes Clusters > Create |
| `<api-endpoint-subnet-ocid>` | OCID of the regional subnet hosting the API server endpoint | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<api-endpoint-nsg-ocid>` | OCID of the NSG controlling access to the API server endpoint | OCI Console > Networking > VCNs > Network Security Groups, or `OciNetworkSecurityGroup` status outputs |
| `<service-lb-subnet-ocid>` | OCID of the subnet where Kubernetes Service load balancers will be placed | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |

## Related Presets

- **02-private-cluster** -- Use instead when the Kubernetes API must not be reachable from the public internet (regulated industries, enterprise security policies)
- **03-development** -- Use instead for dev/test clusters where simplicity and fast setup outweigh production hardening
