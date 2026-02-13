# AzureMysqlFlexibleServer Examples

## Minimal Public Server (Dev/Test)

Smallest possible configuration for development and testing. Uses burstable compute
with 20 GB storage, MySQL 8.0.21, and public access with Azure services allowed.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: dev-mysql
spec:
  region: eastus
  resource_group: dev-rg
  name: myapp-dev-mysql
  administrator_login: mysqladmin
  administrator_password: DevP@ssw0rd123!
  sku_name: B_Standard_B1ms
  storage_size_gb: 20
  databases:
    - name: myapp
  firewall_rules:
    - name: allow-azure-services
      start_ip_address: "0.0.0.0"
      end_ip_address: "0.0.0.0"
```

## VNet-Integrated Server

General-purpose compute with private VNet access using a delegated subnet and
private DNS zone. Public access is automatically disabled when `delegated_subnet_id`
is set.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: vnet-mysql
  org: mycompany
  env: staging
spec:
  region: westeurope
  resource_group: staging-rg
  name: myapp-staging-mysql
  administrator_login: mysqladmin
  administrator_password: St@g1ngS3cur3!
  version: "8.0.21"
  sku_name: GP_Standard_D2ds_v4
  storage_size_gb: 128
  delegated_subnet_id: /subscriptions/sub/resourceGroups/staging-rg/providers/Microsoft.Network/virtualNetworks/staging-vnet/subnets/mysql-subnet
  private_dns_zone_id: /subscriptions/sub/resourceGroups/staging-rg/providers/Microsoft.Network/privateDnsZones/privatelink.mysql.database.azure.com
  backup_retention_days: 14
  databases:
    - name: myapp
```

## HA Production Server

Production-grade configuration with zone-redundant high availability, 512 GB storage,
multiple databases, and extended backup retention.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: prod-mysql
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-rg
  name: myapp-prod-mysql
  administrator_login: mysqladmin
  administrator_password: Pr0dS3cur3P@ss!
  version: "8.0.21"
  sku_name: GP_Standard_D4ds_v4
  storage_size_gb: 512
  auto_grow_enabled: true
  delegated_subnet_id: /subscriptions/sub/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/mysql-subnet
  private_dns_zone_id: /subscriptions/sub/resourceGroups/prod-rg/providers/Microsoft.Network/privateDnsZones/privatelink.mysql.database.azure.com
  backup_retention_days: 35
  databases:
    - name: myapp
    - name: reporting
      charset: utf8mb4
      collation: utf8mb4_0900_ai_ci
  high_availability:
    mode: ZoneRedundant
    standby_availability_zone: "2"
  zone: "1"
```

## Infra Chart valueFrom Pattern

Uses StringValueOrRef to reference outputs from other resources in an infra chart.
This is the pattern used in the database-stack chart where resource group, subnet,
DNS zone, and password are all sourced from sibling resources.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: prod-mysql
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: myapp-prod-mysql
  administrator_login: mysqladmin
  administrator_password:
    value_from:
      kind: RandomPassword
      name: mysql-admin-password
      field_path: status.outputs.result
  version: "8.0.21"
  sku_name: GP_Standard_D8ds_v4
  storage_size_gb: 1024
  auto_grow_enabled: true
  delegated_subnet_id:
    value_from:
      kind: AzureSubnet
      name: mysql-subnet
      field_path: status.outputs.subnet_id
  private_dns_zone_id:
    value_from:
      kind: AzurePrivateDnsZone
      name: mysql-dns-zone
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

A complete database-stack infra chart pattern showing MySQL with its
supporting networking resources. This demonstrates how `AzureMysqlFlexibleServer`
fits alongside `AzureSubnet` and `AzurePrivateDnsZone` in a composed stack.

```yaml
# 1. Dedicated subnet for MySQL (with delegation)
apiVersion: azure.openmcf.org/v1
kind: AzureSubnet
metadata:
  name: mysql-subnet
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
  name: mysql-subnet
  address_prefix: "10.0.5.0/24"
  delegation:
    name: mysql-delegation
    service_delegation_name: Microsoft.DBforMySQL/flexibleServers
    actions:
      - Microsoft.Network/virtualNetworks/subnets/join/action
---
# 2. Private DNS zone for MySQL
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: mysql-dns
spec:
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: privatelink.mysql.database.azure.com
  vnet_id:
    value_from:
      kind: AzureVpc
      name: prod-vnet
      field_path: status.outputs.vnet_id
---
# 3. MySQL Flexible Server
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: prod-mysql
spec:
  region: westeurope
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: myapp-prod-mysql
  administrator_login: mysqladmin
  administrator_password:
    value_from:
      kind: RandomPassword
      name: mysql-password
      field_path: status.outputs.result
  sku_name: GP_Standard_D4ds_v4
  storage_size_gb: 256
  auto_grow_enabled: true
  delegated_subnet_id:
    value_from:
      kind: AzureSubnet
      name: mysql-subnet
      field_path: status.outputs.subnet_id
  private_dns_zone_id:
    value_from:
      kind: AzurePrivateDnsZone
      name: mysql-dns
      field_path: status.outputs.zone_id
  backup_retention_days: 14
  databases:
    - name: myapp
  high_availability:
    mode: ZoneRedundant
```

## Geo-Redundant Enterprise Server

Enterprise-grade Memory Optimized server with cross-region backup replication,
latest MySQL 8.4, and zone-redundant HA for maximum resilience.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: enterprise-mysql
  org: enterprise
  env: production
spec:
  region: eastus
  resource_group: dr-rg
  name: enterprise-mysql
  administrator_login: mysqladmin
  administrator_password: Ent3rpr1s3P@ss!
  version: "8.4"
  sku_name: MO_Standard_E4ds_v4
  storage_size_gb: 2048
  auto_grow_enabled: true
  backup_retention_days: 35
  geo_redundant_backup_enabled: true
  databases:
    - name: core
      charset: utf8mb4
      collation: utf8mb4_0900_ai_ci
    - name: audit
  high_availability:
    mode: ZoneRedundant
    standby_availability_zone: "3"
  zone: "1"
  firewall_rules:
    - name: allow-office
      start_ip_address: "203.0.113.0"
      end_ip_address: "203.0.113.255"
```
