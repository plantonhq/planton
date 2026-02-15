---
title: "PostgreSQL Delegated Subnet"
description: "This preset creates an Azure Subnet delegated to PostgreSQL Flexible Server. Delegation grants the PostgreSQL service permission to inject server instances directly into the subnet, enabling..."
type: "preset"
rank: "02"
presetSlug: "02-delegated-postgresql"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "azure"
icon: "package"
order: 2
---

# PostgreSQL Delegated Subnet

This preset creates an Azure Subnet delegated to PostgreSQL Flexible Server. Delegation grants the PostgreSQL service permission to inject server instances directly into the subnet, enabling VNet-integrated database access without public endpoints. This is a required prerequisite for VNet-injected PostgreSQL Flexible Server deployments.

## When to Use

- Creating a VNet-integrated PostgreSQL Flexible Server (`AzurePostgresqlFlexibleServer` with `delegatedSubnetId`)
- Private database access from within the VNet without public endpoints
- Compliance requirements mandating that database traffic stays within the private network

## Key Configuration Choices

- **Address prefix** (`addressPrefix: 10.0.2.0/28`) -- /28 provides 11 usable IPs (Azure reserves 5). Sufficient for most single-server deployments; use /27 or /26 for HA with multiple replicas
- **Delegation** (`delegation.serviceName: Microsoft.DBforPostgreSQL/flexibleServers`) -- Required for VNet injection. A delegated subnet cannot be shared with other resource types
- **No service endpoints** -- Not needed when the database is directly injected into the subnet

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Resource group containing the VNet | Azure portal or `AzureResourceGroup` status outputs |
| `<vnet-resource-id>` | Full ARM resource ID of the parent VNet | Azure portal or `AzureVpc` status outputs |

## Related Presets

- **01-general-purpose** -- Use instead for workloads that don't require service delegation (VMs, load balancers, private endpoints)
