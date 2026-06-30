# AzureAksCluster

Azure Kubernetes Service (AKS) cluster resource for deploying production-ready managed Kubernetes clusters on Azure. Provides the 80/20 configuration needed for most AKS deployments based on Microsoft's baseline architecture.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **region**: Azure region for the cluster (e.g., `eastus`, `westus2`, `westeurope`)
  - Required field that determines the physical location of the cluster
- **vnet_subnet_id**: Azure subnet resource ID for cluster nodes
  - Can reference an AzureVpc resource's `status.outputs.nodes_subnet_id`
  - Nodes will be deployed in this subnet
- **kubernetes_version**: Kubernetes version for control plane (e.g., `1.30`)
  - Recommended to pin explicitly to prevent unintended upgrades
  - Default: `1.30`
- **control_plane_sku**: Control plane tier
  - `STANDARD` (default): 99.95% SLA with AZs, production-ready (~$73/month)
  - `FREE`: No SLA, dev/test only
- **network_plugin**: Networking plugin for the cluster
  - `AZURE_CNI` (default): Full VNet integration, recommended
  - `KUBENET`: Deprecated (retiring March 2028), do not use
- **network_plugin_mode**: Network plugin mode (only for AZURE_CNI)
  - `OVERLAY` (default): Pods get IPs from private CIDR (10.244.0.0/16), solves VNet IP exhaustion
  - `DYNAMIC`: Pods get real VNet IPs dynamically from dedicated pod subnet
- **system_node_pool**: System node pool configuration (required)
  - `vm_size`: Azure VM SKU (e.g., `Standard_D4s_v5`)
  - `autoscaling`: Min/max node count
    - `min_count`: Minimum nodes (recommended: 3 for production HA)
    - `max_count`: Maximum nodes
  - `availability_zones`: List of zones (e.g., `["1", "2", "3"]`)
    - Production: 3 zones for 99.95% SLA
    - Dev/test: Single zone acceptable

### Advanced Fields (20% Use Case)

- **private_cluster_enabled**: Deploy as private cluster (no public API endpoint)
  - `true`: API server only accessible from VNet
  - `false` (default): Public endpoint (use `authorized_ip_ranges` to restrict)
- **authorized_ip_ranges**: CIDR blocks allowed to access API server
  - Only applies to public clusters
  - Empty list allows all (0.0.0.0/0)
- **disable_azure_ad_rbac**: Disable Azure AD integration
  - `false` (default): Azure AD RBAC enabled (recommended)
  - `true`: Disable Azure AD integration
- **user_node_pools**: User node pools for application workloads
  - Array of user pools with same config as system pool plus:
  - `name`: Pool name (lowercase alphanumeric, max 12 chars)
  - `spot_enabled`: Use Spot VMs for 30-90% cost savings
- **addons**: Cluster add-ons configuration
  - `enable_container_insights`: Azure Monitor Container Insights (requires `log_analytics_workspace_id`)
  - `enable_key_vault_csi_driver`: Mount secrets from Key Vault (default: true)
  - `enable_azure_policy`: Policy-based governance (default: true)
  - `enable_workload_identity`: Workload Identity for secret-less auth (default: true)
  - `log_analytics_workspace_id`: Log Analytics workspace resource ID
- **advanced_networking**: Custom networking CIDRs
  - `pod_cidr`: Pod CIDR for Overlay mode (default: 10.244.0.0/16)
  - `service_cidr`: Kubernetes service CIDR (default: 10.0.0.0/16)
  - `dns_service_ip`: DNS service IP within service CIDR (default: 10.0.0.10)
  - `custom_dns_servers`: Custom DNS servers for VNet

## Stack Outputs

- **api_server_endpoint**: Kubernetes API server endpoint URL
- **cluster_resource_id**: Azure Resource ID of the AKS cluster
- **cluster_kubeconfig**: Base64-encoded kubeconfig file contents
- **managed_identity_principal_id**: Azure AD principal ID of cluster's managed identity

## How It Works

Planton provisions AKS clusters via Pulumi or Terraform modules defined in this repository. The resource follows Microsoft's baseline architecture for production-ready AKS deployments:

1. **Control Plane**: Azure-managed Kubernetes control plane (etcd, API server, scheduler, controller manager)
2. **System Node Pool**: Runs critical cluster components (CoreDNS, metrics-server)
   - Automatically tainted to prevent application workloads
   - Minimum 3 nodes across 3 AZs for production HA
3. **User Node Pools**: Run application workloads
   - Optional but recommended for production (isolates apps from system components)
   - Support for Spot instances, autoscaling, and multi-zone deployment
4. **Networking**: Azure CNI with Overlay mode by default
   - Solves VNet IP exhaustion while maintaining full integration
   - Supports all network policy options
5. **Add-ons**: Managed add-ons for monitoring, secrets, policy, and identity

## Use Cases

### Production Kubernetes Platform
Deploy a production-ready AKS cluster with:
- Standard tier control plane (99.95% SLA)
- Multi-AZ system and user node pools
- Container Insights monitoring
- Workload Identity for secret-less authentication
- Azure Policy for governance

### Private AKS Cluster
Create a fully private AKS cluster with no public API endpoint:
- Private cluster enabled
- API server only accessible from VNet
- Requires VPN or bastion for kubectl access

### Development/Test Environment
Cost-optimized AKS cluster for dev/test:
- FREE tier control plane (no SLA)
- Single AZ deployment
- Smaller VM sizes
- Spot instances for user workloads

### Multi-Node Pool Architecture
Specialized node pools for different workload types:
- General-purpose pool (Standard_D8s_v5)
- Compute-optimized pool (Standard_F16s_v2)
- Memory-optimized pool (Standard_E8s_v5)
- Batch/background jobs pool (Spot instances)

## References

- Azure AKS Documentation: https://docs.microsoft.com/en-us/azure/aks/
- AKS Baseline Architecture: https://learn.microsoft.com/en-us/azure/architecture/reference-architectures/containers/aks/baseline-aks
- Azure CNI Overlay: https://learn.microsoft.com/en-us/azure/aks/azure-cni-overlay
- AKS Pricing: https://azure.microsoft.com/en-us/pricing/details/kubernetes-service/
- Terraform Provider: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster
- Pulumi Provider: https://www.pulumi.com/registry/packages/azure-native/api-docs/containerservice/managedcluster/

