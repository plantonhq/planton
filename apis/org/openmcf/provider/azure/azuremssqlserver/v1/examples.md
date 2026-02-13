# AzureMssqlServer Examples

## Minimal Server with One Database

The simplest setup: a SQL Server with a single Standard-tier database and Azure
service access enabled.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: my-sql-server
spec:
  region: eastus
  resource_group: my-rg
  name: my-sql-server
  administrator_login: sqladmin
  administrator_password: P@ssw0rd1234!
  databases:
    - name: myapp
      sku_name: S0
  firewall_rules:
    - name: allow-azure-services
      start_ip_address: "0.0.0.0"
      end_ip_address: "0.0.0.0"
```

## Production Server with Multiple Databases

A production setup with multiple databases at different performance tiers,
zone redundancy, and Azure Hybrid Benefit for cost savings.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: prod-sql
  org: mycompany
  env: production
spec:
  region: westus2
  resource_group: prod-rg
  name: prod-sql-server
  administrator_login: sqladmin
  administrator_password: SuperSecureP@ss123!
  connection_policy: Redirect
  databases:
    - name: app-primary
      sku_name: GP_Gen5_4
      max_size_gb: 500
      license_type: BasePrice
      zone_redundant: true
      storage_account_type: GeoZone
    - name: reporting
      sku_name: GP_Gen5_2
      max_size_gb: 250
      license_type: BasePrice
    - name: staging
      sku_name: S0
      max_size_gb: 50
  firewall_rules:
    - name: office-vpn
      start_ip_address: "203.0.113.0"
      end_ip_address: "203.0.113.255"
    - name: ci-cd-runner
      start_ip_address: "198.51.100.10"
      end_ip_address: "198.51.100.10"
```

## Private-Only Server (No Public Access)

A server accessible only through Azure Private Endpoint. Public network
access is disabled -- firewall rules have no effect.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: private-sql
spec:
  region: westeurope
  resource_group: private-rg
  name: private-sql-server
  administrator_login: sqladmin
  administrator_password: PrivateP@ss456!
  public_network_access_enabled: false
  minimum_tls_version: "1.2"
  databases:
    - name: app
      sku_name: BC_Gen5_2
      max_size_gb: 100
      zone_redundant: true
      storage_account_type: Zone
```

Use an `AzurePrivateEndpoint` with `subresource_names: ["sqlServer"]` to
connect to this server from your VNet.

## Business Critical with Hybrid Benefit

A high-performance Business Critical tier database with Azure Hybrid Benefit
(55% cost savings with existing SQL Server license + Software Assurance).

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: bc-sql
spec:
  region: eastus2
  resource_group: enterprise-rg
  name: bc-sql-server
  administrator_login: sqladmin
  administrator_password: EntP@ss789!
  connection_policy: Redirect
  databases:
    - name: core-erp
      sku_name: BC_Gen5_8
      max_size_gb: 1024
      license_type: BasePrice
      zone_redundant: true
      collation: SQL_Latin1_General_CP1_CI_AS
      storage_account_type: GeoZone
  firewall_rules:
    - name: allow-azure-services
      start_ip_address: "0.0.0.0"
      end_ip_address: "0.0.0.0"
```

## Infra Chart Reference (valueFrom)

When used within a database-stack infra chart, resource_group and password
come from other resources in the chart via `valueFrom` references.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: chart-sql
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: shared-rg
      fieldPath: status.outputs.resource_group_name
  name: chart-sql-server
  administrator_login: sqladmin
  administrator_password:
    valueFrom:
      kind: AzureKeyVault
      name: db-secrets
      fieldPath: status.outputs.secret_value
  connection_policy: Redirect
  databases:
    - name: app
      sku_name: GP_Gen5_2
      max_size_gb: 100
      license_type: BasePrice
  firewall_rules:
    - name: allow-azure-services
      start_ip_address: "0.0.0.0"
      end_ip_address: "0.0.0.0"
```

## Database-Stack Pattern

A complete database-stack with SQL Server behind a Private Endpoint
and Private DNS Zone for VNet-internal name resolution.

```yaml
# 1. Private DNS Zone for SQL Server
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: sql-dns
spec:
  name: privatelink.database.windows.net
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: db-rg
      fieldPath: status.outputs.resource_group_name
  vnet_id:
    valueFrom:
      kind: AzureVpc
      name: main-vnet
      fieldPath: status.outputs.vnet_id
---
# 2. SQL Server (private only)
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: db-sql
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: db-rg
      fieldPath: status.outputs.resource_group_name
  name: db-sql-server
  administrator_login: sqladmin
  administrator_password:
    valueFrom:
      kind: AzureKeyVault
      name: db-secrets
      fieldPath: status.outputs.secret_value
  public_network_access_enabled: false
  connection_policy: Redirect
  databases:
    - name: app
      sku_name: GP_Gen5_4
      max_size_gb: 500
      license_type: BasePrice
---
# 3. Private Endpoint for SQL Server
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: sql-pe
spec:
  region: eastus
  resource_group:
    valueFrom:
      kind: AzureResourceGroup
      name: db-rg
      fieldPath: status.outputs.resource_group_name
  name: sql-private-endpoint
  subnet_id:
    valueFrom:
      kind: AzureSubnet
      name: pe-subnet
      fieldPath: status.outputs.subnet_id
  private_connection_resource_id:
    valueFrom:
      kind: AzureMssqlServer
      name: db-sql
      fieldPath: status.outputs.server_id
  subresource_names:
    - sqlServer
  private_dns_zone_id:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: sql-dns
      fieldPath: status.outputs.zone_id
```
