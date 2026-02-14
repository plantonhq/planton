# Private AKS Cluster

This preset deploys a private AKS cluster with no public API server endpoint. The Kubernetes API is accessible only from within the VNet or via peered networks (VPN, ExpressRoute). All other configuration is identical to the standard preset: Azure CNI Overlay, 3-zone system and user node pools, and all recommended addons.

## When to Use

- Regulated or security-sensitive environments that prohibit public Kubernetes API endpoints
- Clusters accessed exclusively via VPN, ExpressRoute, or Azure Bastion
- Compliance requirements mandating private-only control plane access
- Enterprise environments with strict network perimeter policies

## Key Configuration Choices

- **Private cluster** (`privateClusterEnabled: true`) -- API server has no public IP; accessible only from within the VNet or peered networks
- **Standard SKU** (`controlPlaneSku: STANDARD`) -- Financially-backed 99.95% uptime SLA with availability zones
- **Azure CNI Overlay** (`networkPlugin: AZURE_CNI`, `networkPluginMode: OVERLAY`) -- Private pod CIDR avoids VNet IP exhaustion
- **System node pool** (`vmSize: Standard_D4s_v5`, 3-5 nodes, 3 zones) -- Dedicated system components with HA
- **User node pool** (`vmSize: Standard_D8s_v5`, 2-10 nodes, 3 zones) -- General-purpose autoscaling pool
- **All addons enabled** -- Container Insights, Key Vault CSI, Azure Policy, Workload Identity

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<nodes-subnet-id>` | ARM resource ID of the subnet for cluster nodes | Azure portal or `AzureVpc` status outputs (`nodesSubnetId`) |
| `<log-analytics-workspace-id>` | ARM resource ID of the Log Analytics workspace | Azure portal or `AzureLogAnalyticsWorkspace` status outputs |

## Related Presets

- **01-standard** -- Use instead when a public API endpoint is acceptable (simpler access for developers)
