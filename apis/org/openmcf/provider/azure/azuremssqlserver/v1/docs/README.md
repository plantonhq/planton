# AzureMssqlServer: Research & Design Documentation

## Deployment Landscape

### Azure SQL Database vs SQL Server on VMs

Azure provides two paths for running SQL Server workloads:

1. **Azure SQL Database** (PaaS) -- Fully managed, automatic patching, built-in HA,
   elastic scaling. This is what `AzureMssqlServer` provisions.
2. **SQL Server on Azure VMs** (IaaS) -- Full OS access, lift-and-shift from
   on-premises, required for features like SQL Server Agent or cross-database
   queries across servers.

Azure SQL Database is the recommended default for new cloud-native workloads.
SQL Server on VMs is for legacy compatibility scenarios that require OS-level
access.

### Logical Server + Database Architecture

Azure SQL Database uses a unique **logical server + database** model that differs
fundamentally from PostgreSQL and MySQL Flexible Servers:

| Aspect | Azure SQL | PostgreSQL/MySQL Flexible Server |
|--------|-----------|--------------------------------|
| **Server model** | Logical container (no compute) | Physical server (has compute) |
| **Compute** | Per-database (sku_name on each DB) | Per-server (sku_name on server) |
| **Storage** | Per-database (max_size_gb on each DB) | Per-server (storage_mb on server) |
| **HA** | Per-database (zone_redundant) | Per-server (high_availability message) |
| **VNet integration** | Via Private Endpoint only | Via subnet delegation |
| **Billing** | Per-database | Per-server |
| **Authentication block** | No auth block on server | PostgreSQL has Authentication block |

This architectural difference means:
- The `AzureMssqlDatabase` message is **much richer** than PG/MySQL database messages
- The `AzureMssqlServerSpec` is **simpler** than PG/MySQL specs (no compute, storage, HA)
- Private networking uses a different mechanism (Private Endpoint vs. VNet delegation)

### Compute Tiers

Azure SQL Database offers three pricing models:

**DTU-based** (simpler, bundled):
- Basic (~5 DTU, 2 GB max) -- dev/test only
- Standard S0-S12 (10-3000 DTU, 250 GB-1 TB) -- most production workloads
- Premium P1-P15 (125-4000 DTU, 500 GB-4 TB) -- high-performance with In-Memory OLTP

**vCore-based** (flexible, independent compute/storage):
- General Purpose (GP_Gen5_*) -- remote storage, cost-effective for most workloads
- Business Critical (BC_Gen5_*) -- local SSD, built-in read replica, lowest latency
- Hyperscale (HS_Gen5_*) -- up to 100 TB, fast backup/restore, named replicas

**Serverless** (auto-pause):
- GP_S_Gen5_* -- auto-pauses after idle period, pay per second of compute
- Ideal for intermittent workloads, dev/test, and single-database applications

### Azure Hybrid Benefit

Customers with existing SQL Server licenses (Enterprise or Standard Edition with
active Software Assurance) can apply Azure Hybrid Benefit via the `license_type`
field set to `"BasePrice"`. This reduces compute costs by up to 55% for vCore-based
tiers (General Purpose, Business Critical, Hyperscale).

### Connection Policies

Azure SQL supports three connection policies that affect how TCP connections are
established:

- **Default**: Uses Redirect for connections originating from within Azure, and
  Proxy for connections from outside Azure. Recommended starting point.
- **Redirect**: After initial handshake through the gateway on port 1433, subsequent
  packets go directly to the database node. Lower latency, higher throughput.
  Requires firewall to allow ports 11000-11999.
- **Proxy**: All traffic routes through the Azure SQL gateway on port 1433.
  Simpler firewall requirements but higher latency.

For production workloads connecting from Azure VMs, AKS, or Container Apps,
**Redirect** provides measurably better performance.

## 80/20 Scoping Rationale

### Included in v1

| Feature | Rationale |
|---------|-----------|
| SQL authentication | Universal auth method, works everywhere |
| Databases with per-DB SKU | Core MSSQL architecture |
| Firewall rules | Essential for public access |
| Connection policy | Meaningful performance impact |
| License type | Up to 55% cost savings |
| Zone redundancy (per-DB) | Production availability requirement |
| Backup redundancy | Disaster recovery foundation |
| Min TLS version | Security baseline |
| Public network access toggle | Required for private-only deployments |

### Deferred to v2

| Feature | Rationale |
|---------|-----------|
| Azure AD authentication | Enterprise-advanced, configurable post-deployment |
| Elastic pools | Separate resource kind pattern (shared compute) |
| Failover groups | Cross-region HA, complex lifecycle |
| Transparent Data Encryption (CMK) | Compliance-niche, requires managed identity |
| Long-term retention policies | Backup retention niche, Azure defaults are reasonable |
| Threat detection policies | Security monitoring, can enable via portal |
| Database copy/restore | Advanced create modes with complex dependencies |
| Serverless auto-pause config | Serverless-specific (min_capacity, auto_pause_delay) |
| Read replicas / read scale | Advanced scaling pattern |
| Ledger / confidential computing | Niche compliance features |
| Maintenance window | Cannot be set on initial create in some cases |

### Deliberately Hardcoded

| Setting | Value | Rationale |
|---------|-------|-----------|
| TDE (Transparent Data Encryption) | Microsoft-managed | Enabled by default on all Azure SQL databases; CMK is compliance-niche |

## Provider Research

### Terraform Provider (azurerm)

Resources used:
- `azurerm_mssql_server` -- The logical server container
- `azurerm_mssql_database` -- Each database with its own SKU/storage
- `azurerm_mssql_firewall_rule` -- IP-based access rules

Key findings from provider source analysis:
- Server `version` only allows `"2.0"` or `"12.0"` -- no intermediate values
- `minimum_tls_version` in provider v5.0+ only allows `"1.2"` (1.0/1.1 deprecated)
- Database `transparent_data_encryption_enabled` defaults to `true` and must be
  `true` for non-DW SKUs
- Firewall rules and databases both reference the server via `server_id`
- `connection_policy` defaults to `"Default"` in the provider
- `fully_qualified_domain_name` is a computed attribute on the server

### Pulumi Azure Classic (v6)

Package: `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql`
- `mssql.NewServer` -- Server constructor
- `mssql.NewDatabase` -- Database constructor (uses `ServerId`)
- `mssql.NewFirewallRule` -- Firewall rule constructor (uses `ServerId`)

Server outputs include `FullyQualifiedDomainName` for connection strings.

## Design Decisions

### Why No VNet Delegation Fields

PostgreSQL and MySQL Flexible Servers support `delegated_subnet_id` -- the server
runs inside a delegated subnet with automatic private DNS. Azure SQL does not
support this model. Private connectivity for Azure SQL is exclusively via
Azure Private Endpoint, which is a separate resource in OpenMCF.

This means:
- No `delegated_subnet_id` field
- No `private_dns_zone_id` field
- No automatic derivation of `public_network_access_enabled` from subnet presence
- The explicit `public_network_access_enabled` boolean is required

### Why Databases Are Rich Objects

In PostgreSQL/MySQL, a database is just a name + charset + collation. In Azure SQL,
a database is where the compute and storage live. Omitting `sku_name` from the
database spec would mean every database inherits... nothing (the server has no SKU).
The `AzureMssqlDatabase` message must be richer to match Azure's architecture.

### Why connection_policy Is Included

Connection policy has measurable performance impact (Redirect vs Proxy can show
latency differences of 10-50ms per connection). For enterprise workloads with
thousands of connections, this adds up. It's a simple string field with clear
values -- worth the minimal complexity.

### Why license_type Is Included

Azure Hybrid Benefit saves up to 55% on compute costs. For a database costing
$1000/month on LicenseIncluded, setting `BasePrice` saves $550/month. This is too
impactful to defer to v2 for enterprises with existing SQL Server licenses.
