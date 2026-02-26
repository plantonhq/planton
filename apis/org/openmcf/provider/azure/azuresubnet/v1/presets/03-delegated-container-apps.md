# Delegated Container Apps Subnet

This preset creates a subnet delegated to Azure Container App Environments (`Microsoft.App/environments`). The /21 prefix provides 2,048 IP addresses — the minimum recommended for Container Apps environments. Container Apps requires a dedicated delegated subnet with sufficient IP space for per-revision IP allocation. This subnet is consumed by the `AzureContainerAppEnvironment` resource's `infrastructureSubnetId` field.

## When to Use

- VNet-integrated Azure Container App Environments (both public and internal load balancer modes)
- Microservices architectures deploying multiple Container Apps in a shared environment
- Enterprise environments requiring VNet isolation for containerized workloads
- Environments expecting scale-out to 10+ concurrent revisions

## Key Configuration Choices

- **Address prefix /21** (`addressPrefix: 10.0.4.0/21`) -- 2,048 IPs. Container Apps allocates IPs per active revision instance. A /23 (512 IPs) works for small deployments, but /21 is recommended for production with scale-out headroom. Minimum is /23
- **Delegation** (`serviceName: Microsoft.App/environments`) -- Required delegation. Prevents other resources from being placed in this subnet. The subnet becomes exclusive to Container App Environments
- **No service endpoints** -- Container Apps environments don't use service endpoints. They access other Azure services via managed identity or Private Endpoints in separate subnets
- **CIDR offset** -- The example uses 10.0.4.0/21 (IPs 10.0.4.0–10.0.11.255). Adjust based on your VNet's address space and existing subnet allocations

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<vnet-resource-id>` | ARM resource ID of the VNet | `AzureVpc` status outputs |

## Related Presets

- **01-general-purpose** -- General subnet with service endpoints (for application workloads)
- **02-delegated-postgresql** -- Subnet delegated to PostgreSQL Flexible Server
