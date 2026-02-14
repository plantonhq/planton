# Development VNet

This preset creates an Azure Virtual Network with a /16 address space and a smaller /20 nodes subnet without a NAT Gateway. This is a cost-effective configuration for development and testing environments where outbound connectivity uses Azure's default rules and a large nodes subnet is not required.

## When to Use

- Development and testing environments
- Small-scale workloads that don't need predictable outbound IPs
- Cost-sensitive environments where NAT Gateway charges are not justified

## Key Configuration Choices

- **Address space** (`addressSpaceCidr: 10.0.0.0/16`) -- Same /16 as production for consistency; unused space costs nothing
- **Nodes subnet** (`nodesSubnetCidr: 10.0.0.0/20`) -- 4,091 usable IPs; sufficient for development clusters
- **NAT Gateway disabled** (`isNatGatewayEnabled: false`) -- Saves approximately $32/month plus data processing charges by using Azure default outbound access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |

## Related Presets

- **01-production-nat** -- Use instead for production environments that need a NAT Gateway for reliable outbound connectivity
