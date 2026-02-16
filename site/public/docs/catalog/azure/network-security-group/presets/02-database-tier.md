---
title: "Database Tier NSG"
description: "This preset creates a Network Security Group for database subnets, allowing only PostgreSQL and MySQL traffic from within the Virtual Network and explicitly denying all internet inbound traffic. This..."
type: "preset"
rank: "02"
presetSlug: "02-database-tier"
componentSlug: "network-security-group"
componentTitle: "Network Security Group"
provider: "azure"
icon: "package"
order: 2
---

# Database Tier NSG

This preset creates a Network Security Group for database subnets, allowing only PostgreSQL and MySQL traffic from within the Virtual Network and explicitly denying all internet inbound traffic. This is the standard NSG for subnets hosting managed databases or self-hosted database servers.

## When to Use

- Subnets hosting PostgreSQL or MySQL servers (managed or self-hosted)
- Back-end data tier in a multi-tier architecture
- Any subnet that should be completely isolated from direct internet access

## Key Configuration Choices

- **PostgreSQL from VNet** (`priority: 100, destinationPortRange: "5432"`) -- Allows PostgreSQL connections from any VNet resource (application servers, AKS pods)
- **MySQL from VNet** (`priority: 110, destinationPortRange: "3306"`) -- Allows MySQL connections from any VNet resource. Remove this rule if only PostgreSQL is used
- **Deny internet** (`priority: 4000, access: Deny`) -- Explicit deny-all for internet traffic. While Azure's implicit rules already deny most inbound, this explicit rule ensures the intent is documented and visible in the NSG
- **Source: VirtualNetwork** -- Azure service tag matching all VNet address spaces, including peered VNets

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the associated subnet) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-nsg-name>` | Name for the NSG (unique within resource group) | Your naming convention |

## Related Presets

- **01-web-tier** -- Use for internet-facing subnets (allows HTTP/HTTPS inbound)
- **03-bastion** -- Use for bastion/jump-host subnets (allows SSH/RDP from trusted IPs only)
