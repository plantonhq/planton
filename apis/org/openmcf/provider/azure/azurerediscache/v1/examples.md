# AzureRedisCache Examples

## Minimal Standard Cache (Dev/Test)

Smallest practical configuration for development. Standard tier with 1 GB cache
and Azure services access.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureRedisCache
metadata:
  name: dev-redis
spec:
  region: eastus
  resource_group: dev-rg
  name: myapp-dev-redis
  capacity: 1
  firewall_rules:
    - name: allow_azure_services
      start_ip: "0.0.0.0"
      end_ip: "0.0.0.0"
```

## Production Standard Cache

Standard tier with 6 GB, allkeys-lru eviction for a pure cache workload,
and weekly maintenance window on Saturday at 3 AM UTC.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureRedisCache
metadata:
  name: prod-redis
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group: prod-rg
  name: myapp-prod-redis
  sku_name: Standard
  capacity: 3
  maxmemory_policy: allkeys-lru
  public_network_access_enabled: false
  patch_schedules:
    - day_of_week: Saturday
      start_hour_utc: 3
  firewall_rules:
    - name: allow_office
      start_ip: "203.0.113.0"
      end_ip: "203.0.113.255"
    - name: allow_vpn
      start_ip: "198.51.100.10"
      end_ip: "198.51.100.10"
```

## Premium Cache with VNet Injection

Premium tier with VNet injection for network isolation. The cache is deployed
into a dedicated subnet with private IP addressing.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureRedisCache
metadata:
  name: secure-redis
  org: enterprise
  env: production
spec:
  region: eastus
  resource_group: prod-rg
  name: enterprise-redis
  sku_name: Premium
  capacity: 2
  subnet_id: /subscriptions/sub/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/redis-subnet
  zones:
    - "1"
    - "2"
  maxmemory_policy: allkeys-lru
  public_network_access_enabled: false
  patch_schedules:
    - day_of_week: Saturday
      start_hour_utc: 2
    - day_of_week: Sunday
      start_hour_utc: 2
```

## Premium Cache with Redis Cluster Sharding

Premium tier with 3 shards for higher throughput and larger data sets.
Total memory = 13 GB * (1 + 3) = 52 GB across 4 primary nodes.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureRedisCache
metadata:
  name: cluster-redis
  org: platform
  env: production
spec:
  region: westus2
  resource_group: platform-rg
  name: platform-cluster-redis
  sku_name: Premium
  capacity: 2
  shard_count: 3
  zones:
    - "1"
    - "2"
    - "3"
  maxmemory_policy: volatile-lfu
  patch_schedules:
    - day_of_week: Weekend
      start_hour_utc: 4
```

## Infra Chart valueFrom Pattern

Uses StringValueOrRef to reference outputs from other resources in an infra chart.
This is the pattern used when Redis is an optional component in database-stack,
container-apps-environment, or web-app-environment charts.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureRedisCache
metadata:
  name: app-redis
  org: mycompany
  env: production
spec:
  region: westeurope
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: myapp-redis
  sku_name: Standard
  capacity: 2
  maxmemory_policy: allkeys-lru
  public_network_access_enabled: false
```

## Database-Stack Pattern

A database-stack infra chart pattern showing Redis as an optional cache layer
alongside PostgreSQL, with shared networking resources.

```yaml
# 1. Redis cache (optional cache layer)
apiVersion: azure.openmcf.org/v1
kind: AzureRedisCache
metadata:
  name: app-redis
spec:
  region: westeurope
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: myapp-redis
  sku_name: Standard
  capacity: 2
  maxmemory_policy: allkeys-lru
  firewall_rules:
    - name: allow_azure
      start_ip: "0.0.0.0"
      end_ip: "0.0.0.0"
---
# 2. PostgreSQL Flexible Server (primary database)
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: app-pg
spec:
  region: westeurope
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: myapp-pg
  administrator_login: pgadmin
  administrator_password:
    value_from:
      kind: RandomPassword
      name: pg-password
      field_path: status.outputs.result
  sku_name: GP_Standard_D2s_v3
  storage_mb: 32768
  databases:
    - name: myapp
```

## Basic Cache for Testing

Minimal Basic tier cache for CI/CD pipelines and integration tests.
No SLA, no replication, lowest cost.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureRedisCache
metadata:
  name: test-redis
spec:
  region: eastus
  resource_group: test-rg
  name: ci-test-redis
  sku_name: Basic
  capacity: 0
```
