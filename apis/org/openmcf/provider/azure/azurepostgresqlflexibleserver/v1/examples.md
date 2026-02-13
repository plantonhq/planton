# AzurePostgresqlFlexibleServer Examples

## Minimal Public Server (Dev/Test)

Smallest possible configuration for development and testing. Uses burstable compute
with 32 GB storage and public access with Azure services allowed.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: dev-pg
spec:
  region: eastus
  resource_group: dev-rg
  name: myapp-dev-pg
  administrator_login: pgadmin
  administrator_password: DevP@ssw0rd123!
  sku_name: B_Standard_B1ms
  storage_mb: 32768
  databases:
    - name: myapp
  firewall_rules:
    - name: allow-azure-services
      start_ip_address: "0.0.0.0"
      end_ip_address: "0.0.0.0"
```

## Production Server with VNet Integration

General-purpose compute with private VNet access, automatic storage growth,
extended backup retention, and a dedicated application database.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: prod-pg
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-rg
  name: myapp-prod-pg
  administrator_login: pgadmin
  administrator_password: Pr0dS3cur3P@ss!
  version: "16"
  sku_name: GP_Standard_D4s_v3
  storage_mb: 524288
  auto_grow_enabled: true
  delegated_subnet_id: /subscriptions/sub/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pg-subnet
  private_dns_zone_id: /subscriptions/sub/resourceGroups/prod-rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com
  backup_retention_days: 35
  databases:
    - name: myapp
    - name: reporting
  high_availability:
    mode: ZoneRedundant
    standby_availability_zone: "2"
  zone: "1"
```

## Production with Infra Chart valueFrom References

Uses StringValueOrRef to reference outputs from other resources in an infra chart.
This is the pattern used in the database-stack chart.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: prod-pg
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: myapp-prod-pg
  administrator_login: pgadmin
  administrator_password:
    value_from:
      kind: RandomPassword
      name: pg-admin-password
      field_path: status.outputs.result
  version: "16"
  sku_name: GP_Standard_D8s_v3
  storage_mb: 1048576
  auto_grow_enabled: true
  delegated_subnet_id:
    value_from:
      kind: AzureSubnet
      name: pg-subnet
      field_path: status.outputs.subnet_id
  private_dns_zone_id:
    value_from:
      kind: AzurePrivateDnsZone
      name: pg-dns-zone
      field_path: status.outputs.zone_id
  backup_retention_days: 35
  geo_redundant_backup_enabled: true
  databases:
    - name: app
    - name: analytics
  high_availability:
    mode: ZoneRedundant
```

## Database-Stack Pattern

A complete database-stack infra chart pattern showing PostgreSQL with its
supporting networking resources.

```yaml
# 1. Dedicated subnet for PostgreSQL (with delegation)
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: pg-subnet
spec:
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  vnet_id:
    value_from:
      kind: AzureVpc
      name: prod-vnet
      field_path: status.outputs.vnet_id
  name: pg-subnet
  address_prefix: "10.0.4.0/24"
  delegation:
    name: pg-delegation
    service_delegation_name: Microsoft.DBforPostgreSQL/flexibleServers
    actions:
      - Microsoft.Network/virtualNetworks/subnets/join/action
---
# 2. Private DNS zone for PostgreSQL
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: pg-dns
spec:
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: privatelink.postgres.database.azure.com
  vnet_id:
    value_from:
      kind: AzureVpc
      name: prod-vnet
      field_path: status.outputs.vnet_id
---
# 3. PostgreSQL Flexible Server
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: prod-pg
spec:
  region: westeurope
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: myapp-prod-pg
  administrator_login: pgadmin
  administrator_password:
    value_from:
      kind: RandomPassword
      name: pg-password
      field_path: status.outputs.result
  sku_name: GP_Standard_D4s_v3
  storage_mb: 262144
  auto_grow_enabled: true
  delegated_subnet_id:
    value_from:
      kind: AzureSubnet
      name: pg-subnet
      field_path: status.outputs.subnet_id
  private_dns_zone_id:
    value_from:
      kind: AzurePrivateDnsZone
      name: pg-dns
      field_path: status.outputs.zone_id
  backup_retention_days: 14
  databases:
    - name: myapp
  high_availability:
    mode: ZoneRedundant
```

## SameZone HA with Custom Database Settings

Uses SameZone HA mode for faster failover and custom database charset/collation.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: fast-ha-pg
spec:
  region: eastus
  resource_group: prod-rg
  name: analytics-pg
  administrator_login: analyst
  administrator_password: An@lyt1csP@ss!
  version: "17"
  sku_name: MO_Standard_E4s_v3
  storage_mb: 1048576
  auto_grow_enabled: true
  databases:
    - name: analytics
    - name: legacy
      charset: SQL_ASCII
      collation: C
  high_availability:
    mode: SameZone
  backup_retention_days: 14
  firewall_rules:
    - name: allow-office
      start_ip_address: "203.0.113.0"
      end_ip_address: "203.0.113.255"
    - name: allow-vpn
      start_ip_address: "198.51.100.10"
      end_ip_address: "198.51.100.10"
```

## Minimal with Geo-Redundant Backup

Enterprise server with cross-region backup replication for disaster recovery.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: dr-pg
  org: enterprise
  env: production
spec:
  region: eastus
  resource_group: dr-rg
  name: enterprise-pg
  administrator_login: pgadmin
  administrator_password: Ent3rpr1s3P@ss!
  sku_name: GP_Standard_D8s_v3
  storage_mb: 2097152
  auto_grow_enabled: true
  backup_retention_days: 35
  geo_redundant_backup_enabled: true
  databases:
    - name: core
  high_availability:
    mode: ZoneRedundant
    standby_availability_zone: "3"
  zone: "1"
```
