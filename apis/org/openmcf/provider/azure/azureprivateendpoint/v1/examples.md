# AzurePrivateEndpoint Examples

## Minimal: PostgreSQL Private Endpoint

The most common use case -- a private endpoint for PostgreSQL Flexible Server with literal values.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: pg-pe
spec:
  region: eastus
  resource_group: my-resource-group
  name: pg-private-endpoint
  subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet
  private_connection_resource_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-postgresql
  subresource_names:
    - postgresqlServer
```

## Key Vault Private Endpoint

A private endpoint for Azure Key Vault with org/env metadata.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: kv-pe
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: kv-private-endpoint
  subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pe-subnet
  private_connection_resource_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.KeyVault/vaults/my-keyvault
  subresource_names:
    - vault
```

## Storage Blob Private Endpoint

A private endpoint for Azure Blob Storage, demonstrating a different service type.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: storage-blob-pe
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: storage-blob-private-endpoint
  subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pe-subnet
  private_connection_resource_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Storage/storageAccounts/mystorageaccount
  subresource_names:
    - blob
```

## Private Endpoint with DNS Zone Group

A complete setup with DNS zone group for automatic A-record registration.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: pg-pe-with-dns
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-network-rg
  name: pg-private-endpoint
  subnet_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pe-subnet
  private_connection_resource_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/prod-postgresql
  subresource_names:
    - postgresqlServer
  private_dns_zone_id: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com
```

## Infra Chart Reference: Database Stack with StringValueOrRef

In an infra chart, all references use `valueFrom` to wire resources together. This creates a PostgreSQL private endpoint wired to the subnet and DNS zone from the same chart.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: pg-pe
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      name: prod-rg
  name: pg-private-endpoint
  subnet_id:
    valueFrom:
      name: pe-subnet
  private_connection_resource_id:
    valueFrom:
      name: prod-postgresql
  subresource_names:
    - postgresqlServer
  private_dns_zone_id:
    valueFrom:
      name: pg-dns
```

## Production: Multiple Private Endpoints

In a database-stack infra chart, you typically create one private endpoint per database instance. Each endpoint is a separate AzurePrivateEndpoint instance.

### PostgreSQL Private Endpoint

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: pg-pe
  org: acmecorp
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      name: prod-network-rg
  name: pg-private-endpoint
  subnet_id:
    valueFrom:
      name: pe-subnet
  private_connection_resource_id:
    valueFrom:
      name: prod-postgresql
  subresource_names:
    - postgresqlServer
  private_dns_zone_id:
    valueFrom:
      name: pg-dns
```

### Redis Private Endpoint

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePrivateEndpoint
metadata:
  name: redis-pe
  org: acmecorp
  env: production
spec:
  region: eastus
  resource_group:
    valueFrom:
      name: prod-network-rg
  name: redis-private-endpoint
  subnet_id:
    valueFrom:
      name: pe-subnet
  private_connection_resource_id:
    valueFrom:
      name: prod-redis
  subresource_names:
    - redisCache
  private_dns_zone_id:
    valueFrom:
      name: redis-dns
```
