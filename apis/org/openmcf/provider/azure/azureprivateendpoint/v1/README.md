# AzurePrivateEndpoint

## Overview

**AzurePrivateEndpoint** creates an Azure Private Endpoint with an optional Private DNS Zone Group, enabling private connectivity to Azure PaaS services (PostgreSQL, MySQL, Key Vault, Storage, etc.) over a private IP address within your Virtual Network.

Azure Private Endpoints enable secure, private access to Azure services without exposing traffic to the public internet. Traffic stays entirely on the Microsoft backbone network, providing data exfiltration protection and simplified network architecture.

This component bundles two Azure resources per DD03 (Composite Bundling Rules):
1. **Private Endpoint** -- The network interface that allocates a private IP from your subnet
2. **DNS Zone Group** (optional) -- Automatically registers the private IP as an A-record in a private DNS zone for seamless DNS resolution

## When to Use

- You need private connectivity to Azure PaaS services (PostgreSQL Flexible Server, MySQL Flexible Server, Redis, Key Vault, Storage, etc.)
- You're building the **database-stack** infra chart and need private endpoints for each database instance
- You want to eliminate public internet exposure for sensitive data services
- You need data exfiltration protection by restricting access to specific sub-resources

## Key Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | Azure region (must match subnet region) |
| `resource_group` | StringValueOrRef | Yes | Resource group for the private endpoint |
| `name` | string | Yes | Endpoint name (1-80 characters, alphanumeric, underscores, periods, hyphens) |
| `subnet_id` | StringValueOrRef | Yes | Subnet for private IP allocation |
| `private_connection_resource_id` | StringValueOrRef | Yes | Target resource ARM ID (polymorphic: PostgreSQL, MySQL, Key Vault, Storage, etc.) |
| `subresource_names` | repeated string | No | Sub-resource group IDs (e.g., "postgresqlServer", "vault", "blob") |
| `private_dns_zone_id` | StringValueOrRef | No | DNS zone for A-record registration |

## Outputs

| Output | Description |
|--------|-------------|
| `private_endpoint_id` | Azure Resource Manager ID of the private endpoint |
| `private_ip_address` | Private IP address allocated from the subnet |
| `network_interface_id` | Azure Resource Manager ID of the network interface |

## Azure Resources Created

- `azurerm_private_endpoint` -- The private endpoint itself
- `azurerm_private_dns_zone_group` (conditional) -- DNS zone group for A-record registration (only if `private_dns_zone_id` is provided)

## Key Behaviors

- **Auto-approved connection** -- Connections are automatically approved (`is_manual_connection` is hardcoded to `false`). Manual connections require approval workflows and are excluded per 80/20 scoping.
- **Conditional DNS zone group** -- The DNS zone group is only created when `private_dns_zone_id` is provided. This enables flexible DNS management patterns.
- **Auto-derived names** -- Private service connection name and DNS zone group name are auto-derived from `metadata.name` in IaC modules, following established patterns.
- **Dynamic IP allocation** -- Private IP is dynamically allocated from the subnet. Static IP assignment is excluded per 80/20 scoping.

## Common Sub-Resource Names

| Azure Service | Sub-Resource Name |
|--------------|-------------------|
| PostgreSQL Flexible Server | `postgresqlServer` |
| MySQL Flexible Server | `mysqlServer` |
| Azure SQL Database | `sqlServer` |
| Azure Key Vault | `vault` |
| Azure Blob Storage | `blob` |
| Azure Table Storage | `table` |
| Azure Queue Storage | `queue` |
| Azure File Storage | `file` |
| Azure Cosmos DB (SQL API) | `Sql` |
| Azure Cosmos DB (Mongo API) | `MongoDB` |
| Azure Cache for Redis | `redisCache` |
| Azure Container Registry | `registry` |

## Dependencies

- **AzureSubnet** -- The subnet where the private IP is allocated
- **AzurePrivateDnsZone** (optional) -- The DNS zone for A-record registration
- **Target Resource** -- The Azure PaaS service being connected (PostgreSQL, MySQL, Key Vault, etc.)

## Referenced By

- **Leaf resource** -- Currently, no other OpenMCF resources reference AzurePrivateEndpoint outputs. It serves as the final networking component in the infra chart DAG.

## Infra Chart Usage

- **database-stack** -- Creates private endpoints for each database instance (PostgreSQL, MySQL, MSSQL, Redis) and wires them to corresponding private DNS zones for seamless DNS resolution

## Known Limitations

- **No static IP assignment** -- Private IPs are dynamically allocated from the subnet. Static IP assignment is excluded per 80/20 scoping.
- **No manual connections** -- All connections are auto-approved. Manual approval workflows are excluded per 80/20 scoping.
