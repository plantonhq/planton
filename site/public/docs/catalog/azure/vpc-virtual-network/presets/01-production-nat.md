---
title: "Production VNet with NAT Gateway"
description: "This preset creates an Azure Virtual Network with a /16 address space, a /18 nodes subnet, and a NAT Gateway for outbound internet connectivity. This is the standard production configuration for..."
type: "preset"
rank: "01"
presetSlug: "01-production-nat"
componentSlug: "vpc-virtual-network"
componentTitle: "VPC (Virtual Network)"
provider: "azure"
icon: "package"
order: 1
---

# Production VNet with NAT Gateway

This preset creates an Azure Virtual Network with a /16 address space, a /18 nodes subnet, and a NAT Gateway for outbound internet connectivity. This is the standard production configuration for AKS-based workloads, providing 16,379 usable IPs in the nodes subnet and reliable outbound connectivity through the NAT Gateway.

## When to Use

- Production environments running AKS clusters with Azure CNI
- Workloads that need outbound internet access with a predictable source IP (via NAT Gateway)
- Standard hub-and-spoke or standalone VNet topologies

## Key Configuration Choices

- **Address space** (`addressSpaceCidr: 10.0.0.0/16`) -- 65,536 addresses; large enough for multiple subnets across environments
- **Nodes subnet** (`nodesSubnetCidr: 10.0.0.0/18`) -- 16,379 usable IPs; sufficient for large AKS clusters with Azure CNI overlay
- **NAT Gateway enabled** (`isNatGatewayEnabled: true`) -- Provides SNAT for outbound internet access with a static public IP, avoiding Azure's default outbound rules

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., `eastus`, `westeurope`) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |

## Related Presets

- **02-development** -- Use instead for development/test environments that don't need a NAT Gateway
