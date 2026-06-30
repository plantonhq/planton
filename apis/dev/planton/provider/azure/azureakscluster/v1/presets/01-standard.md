# Standard Production AKS Cluster

This preset deploys a production-ready AKS cluster with a public API endpoint, Azure CNI Overlay networking, a 3-zone system node pool, a general-purpose user node pool with autoscaling, and all recommended addons enabled. This is the standard configuration for most production Kubernetes workloads on Azure.

## When to Use

- Production Kubernetes clusters that need a public API endpoint with optional IP restrictions
- Standard web, API, and microservice workloads running on Azure CNI Overlay
- Teams that want Azure-managed addons (Container Insights, Key Vault CSI, Azure Policy, Workload Identity) out of the box
- Clusters requiring 99.95% uptime SLA with availability zone distribution

## Key Configuration Choices

- **Standard SKU** (`controlPlaneSku: STANDARD`) -- Financially-backed 99.95% uptime SLA with availability zones; costs ~$73/month
- **Azure CNI Overlay** (`networkPlugin: AZURE_CNI`, `networkPluginMode: OVERLAY`) -- Pods get private CIDR IPs (10.244.0.0/16), avoiding VNet IP exhaustion while retaining full VNet integration
- **Public endpoint** (`privateClusterEnabled: false`) -- API server is publicly accessible; restrict with `authorizedIpRanges` if needed
- **System node pool** (`vmSize: Standard_D4s_v5`, 3-5 nodes, 3 zones) -- Dedicated to system components (CoreDNS, metrics-server); isolated from application workloads
- **User node pool** (`vmSize: Standard_D8s_v5`, 2-10 nodes, 3 zones) -- General-purpose pool for application workloads with autoscaling
- **All addons enabled** -- Container Insights for observability, Key Vault CSI for secrets, Azure Policy for governance, Workload Identity for credential-free auth

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<nodes-subnet-id>` | ARM resource ID of the subnet for cluster nodes | Azure portal or `AzureVpc` status outputs (`nodesSubnetId`) |
| `<log-analytics-workspace-id>` | ARM resource ID of the Log Analytics workspace | Azure portal or `AzureLogAnalyticsWorkspace` status outputs |

## Related Presets

- **02-private** -- Use instead when the API server must not be publicly accessible (VPN/ExpressRoute-only access)
