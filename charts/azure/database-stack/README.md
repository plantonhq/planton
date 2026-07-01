# Azure Database Stack

Production-ready managed databases on private Azure networks.

## What This Chart Deploys

| Resource | Kind | Condition |
|----------|------|-----------|
| Resource Group | `AzureResourceGroup` | Always |
| Virtual Network | `AzureVpc` | Always |
| PostgreSQL Subnet | `AzureSubnet` | `create_postgres` |
| PostgreSQL DNS Zone | `AzurePrivateDnsZone` | `create_postgres` |
| PostgreSQL Server | `AzurePostgresqlFlexibleServer` | `create_postgres` |
| MySQL Subnet | `AzureSubnet` | `create_mysql` |
| MySQL DNS Zone | `AzurePrivateDnsZone` | `create_mysql` |
| MySQL Server | `AzureMysqlFlexibleServer` | `create_mysql` |
| Private Endpoint Subnet | `AzureSubnet` | `create_mssql` |
| MSSQL DNS Zone | `AzurePrivateDnsZone` | `create_mssql` |
| MSSQL Server | `AzureMssqlServer` | `create_mssql` |
| MSSQL Private Endpoint | `AzurePrivateEndpoint` | `create_mssql` |
| Redis Cache | `AzureRedisCache` | `create_redis` |

## Network Architecture

PostgreSQL and MySQL Flexible Servers use **VNet-delegated subnets** for private
access. Each database type gets its own dedicated subnet with the appropriate
Azure service delegation. Public network access is automatically disabled when a
delegated subnet is provided.

MSSQL uses **Private Endpoints** for private connectivity because Azure SQL
Database does not support VNet-delegated subnets. A separate PE subnet hosts the
private endpoint, and a Private DNS zone enables FQDN resolution to the private IP.

Redis deploys in public access mode with TLS 1.2 enforced. For VNet-integrated
Redis (Premium SKU), configure the subnet reference after deployment.

## Default CIDR Allocation

| Subnet | CIDR | Purpose |
|--------|------|---------|
| Default (VPC internal) | 10.1.0.0/24 | Reserved, created by AzureVpc |
| PostgreSQL | 10.1.1.0/24 | Delegated to PostgreSQL Flexible Server |
| MySQL | 10.1.2.0/24 | Delegated to MySQL Flexible Server |
| Private Endpoint | 10.1.3.0/24 | Hosts MSSQL Private Endpoint |

## Parameters

### Foundation

| Parameter | Description | Default |
|-----------|-------------|---------|
| `region` | Azure region | `eastus` |
| `resource_group_name` | Resource group name suffix | `db-stack-rg` |
| `vnet_cidr` | VNet address space | `10.1.0.0/16` |
| `default_subnet_cidr` | Default subnet CIDR | `10.1.0.0/24` |

### PostgreSQL (default on)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_postgres` | Enable PostgreSQL | `true` |
| `postgres_name` | Server name (globally unique) | `my-postgres` |
| `postgres_subnet_cidr` | Delegated subnet CIDR | `10.1.1.0/24` |
| `postgres_sku_name` | Compute SKU | `B_Standard_B1ms` |
| `postgres_version` | PostgreSQL version | `16` |
| `postgres_storage_mb` | Storage in MB | `32768` |
| `postgres_admin_login` | Admin login | `pgadmin` |
| `postgres_admin_password` | Admin password | (empty) |
| `postgres_database_name` | Application database | `appdb` |

### MySQL (optional)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_mysql` | Enable MySQL | `false` |
| `mysql_name` | Server name (globally unique) | `my-mysql` |
| `mysql_subnet_cidr` | Delegated subnet CIDR | `10.1.2.0/24` |
| `mysql_sku_name` | Compute SKU | `B_Standard_B1ms` |
| `mysql_version` | MySQL version | `8.0.21` |
| `mysql_storage_size_gb` | Storage in GB | `20` |
| `mysql_admin_login` | Admin login | `mysqladmin` |
| `mysql_admin_password` | Admin password | (empty) |
| `mysql_database_name` | Application database | `appdb` |

### MSSQL (optional)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_mssql` | Enable MSSQL | `false` |
| `mssql_name` | Server name (globally unique) | `my-mssql` |
| `mssql_pe_subnet_cidr` | PE subnet CIDR | `10.1.3.0/24` |
| `mssql_version` | SQL Server version | `12.0` |
| `mssql_admin_login` | Admin login | `sqladmin` |
| `mssql_admin_password` | Admin password | (empty) |
| `mssql_database_name` | Application database | `appdb` |
| `mssql_database_sku` | Database SKU | `S0` |

### Redis (optional)

| Parameter | Description | Default |
|-----------|-------------|---------|
| `create_redis` | Enable Redis | `false` |
| `redis_name` | Cache name | `my-redis` |
| `redis_sku_name` | SKU (Basic/Standard/Premium) | `Standard` |
| `redis_capacity` | Cache size tier | `1` |
| `redis_version` | Redis version | `6` |

## Example

Deploy PostgreSQL with a Redis cache:

```yaml
params:
  region: westus2
  resource_group_name: myapp-db-rg
  postgres_name: myapp-pg
  postgres_admin_password: "S3cur3P@ssw0rd!"
  postgres_database_name: myapp
  postgres_sku_name: GP_Standard_D2s_v3
  postgres_storage_mb: "131072"
  create_redis: true
  redis_name: myapp-cache
```
