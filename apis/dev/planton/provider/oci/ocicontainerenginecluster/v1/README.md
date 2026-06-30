# Overview

The **OCI Container Engine Cluster API Resource** provides a consistent and standardized interface for deploying and managing Oracle Cloud Infrastructure Container Engine for Kubernetes (OKE) clusters. An OKE cluster is a fully managed Kubernetes control plane — OCI runs the API server, etcd, scheduler, and controller manager on hardened infrastructure, while worker nodes are provisioned separately via OciContainerEngineNodePool resources. This component wraps the `oci_containerengine_cluster` API surface with the standard Planton KRM pattern.

## Purpose

This API resource streamlines the deployment of OKE clusters by offering a unified interface that covers the full range of cluster configurations — from a minimal development cluster to a production-grade private cluster with OIDC, KMS encryption, and image verification. It enables users to:

- **Choose the Right Cluster Type**: Select between basic clusters (standard Kubernetes features, no per-cluster charge) and enhanced clusters (virtual node pools, workload identity, cluster add-on lifecycle management). The `type` enum makes this a single-field decision.
- **Select Pod Networking Model**: Configure VCN-native CNI (pods receive VCN IP addresses, enabling network policies and NSGs on pods) or flannel overlay (simpler, legacy) via the `cniType` enum. This is one of the most consequential cluster decisions and is immutable after creation.
- **Control API Endpoint Access**: Configure the Kubernetes API server to be private (no public IP, accessible only within the VCN or via VPN/peering) or public. Apply NSGs to the endpoint for fine-grained network access control.
- **Encrypt Secrets at Rest**: Specify a KMS key to encrypt all Kubernetes secrets stored in etcd. Standard OKE encryption uses Oracle-managed keys; `kmsKeyId` enables customer-managed keys for compliance.
- **Authenticate via External Identity Providers**: Configure OIDC token authentication for the API server, enabling `kubectl` and API access through enterprise identity providers (Okta, Azure AD, Auth0, Keycloak). Two modes are supported: inline field configuration and base64-encoded configuration file.
- **Enforce Image Signing**: Enable image policy verification to require that all container images deployed to the cluster are signed with approved KMS keys — a supply chain security control.
- **Tag Kubernetes-Created Resources**: Propagate freeform and defined tags to load balancers and persistent volumes created by Kubernetes, reducing tag drift and enabling cost tracking without per-workload tagging.
- **Compose with Other OCI Resources**: Reference OciCompartment, OciVcn, OciSubnet, and OciSecurityGroup outputs via `StringValueOrRef` for declarative, cross-resource dependency chains.

## Key Features

- **Consistent Interface**: Aligns with the Planton pattern for deploying cloud infrastructure across providers.
- **Two Cluster Types**: Basic (no per-cluster charge, standard Kubernetes features) and enhanced (virtual node pools, workload identity, add-on management). One-way upgrade from basic to enhanced.
- **Two CNI Options**: VCN-native IP allocation (recommended for production — pods get VCN IPs, enabling network policies and NSGs on pods) and flannel overlay (simpler, suitable for development).
- **Private and Public Endpoints**: Configurable API server endpoint with subnet placement, public/private IP control, and NSG association. Private endpoints keep the API server off the internet.
- **KMS Secrets Encryption**: Customer-managed KMS key for encrypting Kubernetes secrets at rest. Requires Kubernetes version >= v1.13.0.
- **OIDC Authentication**: Two configuration modes — inline (issuer URL, client ID, claims) and configuration file (base64-encoded Kubernetes OIDC Auth Config). Supports signing algorithm selection, required claims, and OIDC Discovery endpoint.
- **Image Policy Verification**: Require container images to be signed with specified KMS keys before deployment. Multiple signing keys supported.
- **Service Load Balancer Configuration**: Designate subnets for Kubernetes Service load balancers, apply NSGs to backends, and propagate freeform and defined tags.
- **Persistent Volume Tagging**: Propagate freeform and defined tags to block volumes created by PersistentVolumeClaim resources.
- **Dual-Stack Networking**: Support for IPv4-only and IPv4+IPv6 dual-stack clusters via the `ipFamilies` field.
- **Automatic Tagging**: Standard Planton freeform tags applied to the cluster (resource kind, resource ID, organization, environment, and user-defined labels from metadata).
- **Infra-Chart Composability**: Exports 5 stack outputs (`clusterId`, `kubernetesVersion`, `kubernetesEndpoint`, `privateEndpoint`, `publicEndpoint`) for downstream `StringValueOrRef` references — most importantly, `clusterId` is consumed by OciContainerEngineNodePool.

## How OKE Differs from Other Managed Kubernetes Services

Understanding these differences is essential when coming from EKS, GKE, or AKS:

| Aspect | OKE | EKS | GKE | AKS |
|--------|-----|-----|-----|-----|
| **Cluster types** | Basic (free) and Enhanced (per-cluster charge, adds workload identity, virtual nodes, add-on management) | Single tier (per-cluster charge) | Standard and Autopilot | Free tier and Standard tier |
| **Pod CNI** | VCN-native (pods get VCN IPs) or Flannel overlay | VPC CNI (pods get VPC IPs) by default | VPC-native (alias IPs) by default | Azure CNI or kubenet |
| **API endpoint model** | Subnet-based: endpoint lives in a specific subnet, public/private controlled by IP assignment and subnet type | Endpoint access config: public, private, or both | Public or private endpoint with authorized networks | Public or private with authorized IP ranges |
| **Compartment scoping** | Every cluster lives in a compartment (OCI's hierarchical isolation model) | Clusters live in an AWS account | Clusters live in a GCP project | Clusters live in an Azure resource group |
| **Secrets encryption** | KMS key on cluster resource (encrypts etcd secrets) | KMS envelope encryption via separate config | Application-layer encryption via KMS key | Azure Key Vault integration |
| **Image policy** | Built-in image signature verification via KMS keys | No built-in (use admission webhooks) | Binary Authorization (separate service) | Azure Policy with Gatekeeper |
| **Node pool model** | Separate resource (OciContainerEngineNodePool) | Separate resource (EKS Managed Node Group) | Separate resource (GKE Node Pool) | Separate resource (AKS Node Pool) |
| **Control plane cost** | Basic: free. Enhanced: per-cluster/hour charge | Per-cluster/hour charge | Standard: per-cluster/hour. Autopilot: per-pod | Free tier available |

Key distinctions for OCI newcomers:

- **VCN-Native vs Flannel is immutable.** Once you choose a CNI type, the cluster cannot be changed. VCN-native is strongly recommended for production because it enables NSGs and network policies on individual pods. Flannel places all pods behind the node's IP, limiting network-level isolation.
- **Endpoint subnet matters.** Unlike EKS where the endpoint is controlled by access configuration flags, OKE places the API server endpoint in a specific subnet. The subnet's route table and security lists directly affect who can reach the API server.
- **Basic to Enhanced is one-way.** You can upgrade a basic cluster to enhanced, but you cannot downgrade. Enhanced clusters incur an additional per-cluster charge.

## Critical Constraints

- **VCN Is Immutable**: Changing `vcnId` after creation forces cluster recreation. Plan your VCN topology before creating the cluster.
- **CNI Type Is Immutable**: Changing `cniType` after creation forces cluster recreation. Choose VCN-native or flannel at creation time.
- **Pod and Service CIDRs Are Immutable**: `kubernetesNetworkConfig.podsCidr` and `kubernetesNetworkConfig.servicesCidr` cannot be changed after creation. Ensure CIDRs don't overlap with the VCN CIDR, each other, or CIDRs of peered VCNs.
- **Compartment Change Forces Recreation**: Moving a cluster to a different compartment via `compartmentId` change forces cluster recreation.
- **KMS Key Change Forces Recreation**: Changing `kmsKeyId` after creation forces cluster recreation. The KMS key must be in the same region as the cluster.
- **Basic to Enhanced Is One-Way**: Upgrading `type` from `basic_cluster` to `enhanced_cluster` is supported. Downgrading is not. Enhanced clusters incur additional charges.
- **Kubernetes Version Lifecycle**: OKE supports a limited set of Kubernetes versions. Older versions are deprecated and eventually removed. Run `oci ce cluster-options list` to check available versions before deploying or upgrading.
- **Service LB Subnet CIDRs**: Service load balancer subnets must have enough available IPs for the expected number of Kubernetes Service LoadBalancers. Each service creates one load balancer in the designated subnet(s).
- **IP Families Are Immutable**: Changing `ipFamilies` after creation forces cluster recreation. Dual-stack (IPv4+IPv6) requires VCN-native CNI.

## Use Cases

- **Production Kubernetes Platform**: Enhanced cluster with VCN-native CNI, private API endpoint, KMS-encrypted secrets, dedicated service LB subnets, and NSG-secured endpoints. The standard pattern for enterprise workloads.
- **Development and Testing**: Basic cluster with minimal configuration. No per-cluster charge on basic clusters. Flannel CNI is acceptable when network policy enforcement is not needed.
- **Multi-Tenant Platforms**: Enhanced clusters with workload identity, VCN-native CNI (namespace-level NSG isolation), and PV/LB tagging for per-tenant cost tracking.
- **CI/CD Infrastructure**: Basic or enhanced clusters for running build pipelines. The `clusterId` output feeds into OciContainerEngineNodePool for attaching compute-optimized node pools.
- **Compliance-Sensitive Workloads**: Private API endpoint (no internet exposure), OIDC authentication (enterprise SSO), KMS secrets encryption (customer-managed keys), and image policy verification (signed images only).
- **Hybrid Identity**: OIDC authentication integrates the cluster with enterprise identity providers (Okta, Azure AD, Auth0) for `kubectl` access without managing OCI IAM users for every developer.

## Production Features

This resource provides complete support for production-grade OKE cluster deployments, including:

- **Private API Endpoints**: The API server lives in a user-specified subnet with no public IP, accessible only within the VCN or via VPN/peering/bastion. NSGs on the endpoint control which networks can reach the API server.
- **KMS Secrets Encryption**: Customer-managed KMS keys encrypt all Kubernetes secrets in etcd. Standard OKE uses Oracle-managed keys; `kmsKeyId` adds customer control for compliance.
- **OIDC Authentication**: External identity provider integration for the Kubernetes API server. Supports inline configuration (issuer, client, claims) and configuration file mode. OIDC Discovery endpoint can be enabled for external signing key discovery.
- **Image Policy Verification**: Container image signature verification using KMS keys. When enabled, unsigned or incorrectly signed images are rejected at admission. Multiple signing keys allow key rotation.
- **Service LB Network Control**: NSGs applied to service load balancer backends, plus freeform and defined tag propagation to load balancers created by Kubernetes.
- **PV Tag Propagation**: Freeform and defined tags propagated to block volumes created by PersistentVolumeClaim resources, enabling cost tracking and compliance without per-workload tagging.
- **Dual-Stack Networking**: IPv4+IPv6 dual-stack support for clusters and pods when using VCN-native CNI.
- **Freeform Tagging**: Standard Planton labels applied as OCI freeform tags for resource management, cost tracking, and compliance.
- **Infrastructure as Code**: Full Pulumi (Go) and Terraform (HCL) implementations producing identical outputs.
- **Infra-Chart Composability**: Designed to compose with OciCompartment, OciVcn, OciSubnet, OciSecurityGroup, and OciContainerEngineNodePool via `StringValueOrRef`.
