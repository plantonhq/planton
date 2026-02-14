# General-Purpose Subnet

This preset creates a general-purpose Azure Subnet with a /24 CIDR block (254 usable IPs) and common service endpoints for Storage, Key Vault, and SQL. This is the standard subnet configuration for workloads like virtual machines, private endpoints, internal load balancers, and application gateways.

## When to Use

- General workloads: VMs, internal load balancers, application gateways
- Subnets that need secure access to Azure PaaS services via service endpoints
- Private endpoint subnets (default network policies allow this)

## Key Configuration Choices

- **Address prefix** (`addressPrefix: 10.0.1.0/24`) -- 254 usable IPs; standard sizing for most workloads. Adjust for larger deployments
- **Service endpoints** -- `Microsoft.Storage`, `Microsoft.KeyVault`, `Microsoft.Sql` provide optimized, private-network-only access to these Azure services
- **No delegation** -- Keeps the subnet flexible for multiple resource types
- **Default network policies** -- Private endpoint policies disabled (Azure default), allowing private endpoints to receive traffic without NSG interference

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Resource group containing the VNet | Azure portal or `AzureResourceGroup` status outputs |
| `<vnet-resource-id>` | Full ARM resource ID of the parent VNet | Azure portal or `AzureVpc` status outputs |

## Related Presets

- **02-delegated-postgresql** -- Use instead when creating a subnet dedicated to PostgreSQL Flexible Server
