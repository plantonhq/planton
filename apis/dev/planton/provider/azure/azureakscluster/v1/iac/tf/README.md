# Azure AKS Cluster Terraform Module

## Overview

This Terraform module provisions an Azure Kubernetes Service (AKS) cluster with production-ready defaults following Azure best practices. The module creates the AKS cluster along with its required resource group and configures essential features like managed identity, networking, and optional monitoring integration.

## Features

- **Production-Ready Configuration**: Default settings follow Azure AKS best practices
- **Managed Identity**: Uses System-Assigned Managed Identity for secure, secret-less authentication
- **Network Integration**: Supports both Azure CNI and Kubenet network plugins
- **VNet Integration**: Deploys nodes into your existing VNet subnet
- **Azure AD RBAC**: Optional Azure Active Directory integration for cluster access control
- **Private Cluster Support**: Can deploy as a private cluster with no public endpoint
- **Authorized IP Ranges**: Restrict API server access to specific CIDR blocks
- **Monitoring Integration**: Optional Azure Monitor Container Insights integration
- **Auto-scaling**: System node pool with auto-scaling enabled (3-5 nodes)

## Prerequisites

Before using this module, ensure you have:

1. **Azure CLI** installed and authenticated
2. **Terraform** >= 1.5 installed
3. An existing **Azure VNet and Subnet** for the AKS nodes
4. Appropriate **Azure permissions** to create resources

## Usage

### Basic Example

```hcl
module "aks_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "my-aks-cluster"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    vnet_subnet_id = "/subscriptions/sub-id/resourceGroups/rg-network/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/aks-subnet"
  }
}
```

### Production Example with All Features

```hcl
module "aks_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "prod-app-aks"
    env  = "production"
    org  = "mycompany"
    labels = {
      team    = "platform"
      project = "infrastructure"
    }
  }

  spec = {
    region              = "eastus"
    vnet_subnet_id      = "/subscriptions/.../subnets/aks-subnet"
    kubernetes_version  = "1.30"
    network_plugin      = "AZURE_CNI"
    
    # Security
    private_cluster_enabled = false
    authorized_ip_ranges = [
      "203.0.113.0/24",  # Office network
      "198.51.100.0/24"  # CI/CD agents
    ]
    
    # Azure AD RBAC
    disable_azure_ad_rbac = false
    
    # Monitoring
    log_analytics_workspace_id = "/subscriptions/.../workspaces/my-workspace"
  }
}
```

### Development/Test Example

```hcl
module "aks_cluster" {
  source = "./path/to/module"

  metadata = {
    name = "dev-aks"
    env  = "development"
  }

  spec = {
    region         = "eastus"
    vnet_subnet_id = "/subscriptions/.../subnets/dev-subnet"
    
    # Use defaults for simpler dev setup
    kubernetes_version = "1.30"
  }
}
```

## Inputs

### Metadata Object

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `name` | Name of the AKS cluster | `string` | Yes |
| `id` | Optional resource ID | `string` | No |
| `org` | Organization name | `string` | No |
| `env` | Environment (e.g., production, staging, dev) | `string` | No |
| `labels` | Key-value labels/tags | `map(string)` | No |

### Spec Object

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `region` | Azure region (e.g., "eastus") | `string` | - | Yes |
| `vnet_subnet_id` | Azure Resource ID of the VNet subnet for nodes | `string` | - | Yes |
| `network_plugin` | Network plugin: "AZURE_CNI" or "KUBENET" | `string` | `"AZURE_CNI"` | No |
| `kubernetes_version` | Kubernetes version | `string` | `"1.30"` | No |
| `private_cluster_enabled` | Deploy as private cluster | `bool` | `false` | No |
| `authorized_ip_ranges` | CIDR blocks allowed to access API server | `list(string)` | `[]` | No |
| `disable_azure_ad_rbac` | Disable Azure AD RBAC integration | `bool` | `false` | No |
| `log_analytics_workspace_id` | Log Analytics Workspace ID for monitoring | `string` | `""` | No |

## Outputs

| Name | Description |
|------|-------------|
| `api_server_endpoint` | The FQDN of the AKS API server |
| `cluster_resource_id` | The Azure Resource ID of the cluster |
| `cluster_kubeconfig` | Base64-encoded kubeconfig (sensitive) |
| `managed_identity_principal_id` | Principal ID of the cluster's managed identity |
| `resource_group_name` | Name of the resource group |
| `cluster_name` | Name of the cluster |
| `node_resource_group` | Auto-generated resource group for cluster resources |

## Architecture

### Resources Created

1. **Resource Group**: Container for all AKS-related resources
2. **AKS Cluster**: Managed Kubernetes cluster with:
   - System-Assigned Managed Identity
   - System node pool (3 nodes, auto-scaling to 5)
   - Network integration with your VNet
   - Optional Azure AD RBAC
   - Optional monitoring integration

### Network Configuration

The module uses the following network configuration:
- **Network Plugin**: Azure CNI (default) or Kubenet
- **Load Balancer SKU**: Standard
- **Service CIDR**: 10.0.0.0/16
- **DNS Service IP**: 10.0.0.10

Ensure your VNet subnet does not overlap with the service CIDR.

### Node Pool Configuration

**System Node Pool (Default)**:
- Name: `system`
- VM Size: `Standard_D2s_v3`
- Node Count: 3 (auto-scales between 3-5)
- Labels: `node-role=system`
- Purpose: Critical system workloads (CoreDNS, metrics-server, etc.)

## Security Considerations

### Managed Identity

The cluster uses **System-Assigned Managed Identity** instead of Service Principals. This provides:
- No secrets to manage or rotate
- Automatic identity lifecycle management
- Improved security posture

### API Server Access

Configure `authorized_ip_ranges` to restrict API server access:

```hcl
spec = {
  authorized_ip_ranges = [
    "203.0.113.0/24"  # Office network
  ]
}
```

For maximum security, use `private_cluster_enabled = true` to eliminate public endpoint entirely.

### Azure AD RBAC

By default, Azure AD RBAC is enabled (`disable_azure_ad_rbac = false`), providing:
- Centralized identity management
- Role-based access control using Azure AD groups
- No need for cluster-admin kubeconfig distribution

## Monitoring

To enable Azure Monitor Container Insights, provide a Log Analytics Workspace ID:

```hcl
spec = {
  log_analytics_workspace_id = "/subscriptions/.../workspaces/my-workspace"
}
```

This enables:
- Container logs (stdout/stderr)
- Performance metrics
- Kubernetes event monitoring
- Integration with Azure Monitor dashboards

## Known Limitations

1. **Single System Node Pool**: Module creates one system node pool. Additional user node pools should be added separately.
2. **Network CIDR**: Service CIDR (10.0.0.0/16) is fixed. Ensure no overlap with your VNet.
3. **Node Pool Zones**: Availability zone configuration is not exposed. Nodes deploy to Azure-selected zones.

## Terraform Version

- **Required**: >= 1.5
- **Azure Provider**: ~> 3.0

## Validation

After applying, verify cluster status:

```bash
# Get credentials
az aks get-credentials --resource-group rg-<cluster-name> --name <cluster-name>

# Check nodes
kubectl get nodes

# Check system pods
kubectl get pods -n kube-system
```

## Cleanup

To destroy the cluster:

```bash
terraform destroy
```

**Note**: This will delete the cluster and its resource group. Ensure you have backups if needed.

## Support

For issues or questions:
- Review the [Azure AKS Documentation](https://docs.microsoft.com/en-us/azure/aks/)
- Check the [Terraform AzureRM Provider Docs](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster)
- Open an issue in the repository

## License

This module is licensed under the MIT License.

