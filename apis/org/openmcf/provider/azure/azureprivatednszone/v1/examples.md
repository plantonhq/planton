# AzurePrivateDnsZone Examples

## Minimal: PostgreSQL Private Link Zone

The most common use case -- a privatelink zone for PostgreSQL Flexible Server with a VNet link.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: pg-private-dns
spec:
  resource_group: my-resource-group
  name: privatelink.postgres.database.azure.com
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet
```

## MySQL Private Link Zone

A privatelink zone for MySQL Flexible Server.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: mysql-private-dns
spec:
  resource_group: my-resource-group
  name: privatelink.mysql.database.azure.com
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet
```

## Key Vault Private Link Zone

Enable private connectivity to Azure Key Vault.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: kv-private-dns
  org: mycompany
  env: production
spec:
  resource_group: prod-network-rg
  name: privatelink.vaultcore.azure.net
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet
```

## Custom Internal DNS with Auto-Registration

An internal DNS zone with auto-registration enabled for VM hostname resolution.
VMs created in the linked VNet will automatically get A-records in this zone.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: internal-dns
  org: mycompany
  env: development
spec:
  resource_group: dev-rg
  name: contoso.internal
  vnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/dev-rg/providers/Microsoft.Network/virtualNetworks/dev-vnet
  registration_enabled: true
```

## Infra Chart Reference: Database Stack with StringValueOrRef

In an infra chart, all references use `valueFrom` to wire resources together.
This creates a PostgreSQL privatelink zone linked to the VNet from the same chart.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: pg-private-dns
  org: mycompany
  env: production
spec:
  resource_group:
    valueFrom:
      name: prod-rg
  name: privatelink.postgres.database.azure.com
  vnet_id:
    valueFrom:
      name: prod-vpc
```

## Production: Multiple Privatelink Zones (Database Stack Pattern)

In a database-stack infra chart, you typically create one privatelink zone per database type.
Each zone is a separate AzurePrivateDnsZone instance.

### PostgreSQL Zone

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: pg-dns
  org: acmecorp
  env: production
spec:
  resource_group:
    valueFrom:
      name: prod-network-rg
  name: privatelink.postgres.database.azure.com
  vnet_id:
    valueFrom:
      name: prod-vpc
```

### Redis Zone

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateDnsZone
metadata:
  name: redis-dns
  org: acmecorp
  env: production
spec:
  resource_group:
    valueFrom:
      name: prod-network-rg
  name: privatelink.redis.cache.windows.net
  vnet_id:
    valueFrom:
      name: prod-vpc
```
