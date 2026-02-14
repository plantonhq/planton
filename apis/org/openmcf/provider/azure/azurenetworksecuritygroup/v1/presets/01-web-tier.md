# Web Tier NSG

This preset creates a Network Security Group for web-facing subnets, allowing inbound HTTP and HTTPS traffic from the internet. This is the standard NSG for subnets hosting load balancers, application gateways, or directly internet-exposed web servers.

## When to Use

- Subnets hosting Azure Application Gateway or Azure Load Balancer with public IPs
- Web application subnets that need inbound internet traffic on ports 80 and 443
- Front-end tier in a multi-tier architecture

## Key Configuration Choices

- **HTTPS inbound** (`priority: 100, destinationPortRange: "443"`) -- Primary web traffic; highest priority
- **HTTP inbound** (`priority: 110, destinationPortRange: "80"`) -- For HTTP-to-HTTPS redirect at the load balancer or app level
- **Source: Internet** (`sourceAddressPrefix: Internet`) -- Azure service tag allowing all internet traffic
- **Azure defaults apply** -- VNet-to-VNet traffic and Azure Load Balancer probes are allowed by Azure's implicit rules (priorities 65000+)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the associated subnet) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-nsg-name>` | Name for the NSG (unique within resource group) | Your naming convention |

## Related Presets

- **02-database-tier** -- Use for subnets hosting databases (allows only VNet-internal traffic)
- **03-bastion** -- Use for bastion/jump-host subnets (allows SSH/RDP from trusted IPs only)
