---
title: "Cosmos DB Account"
description: "Cosmos DB Account deployment documentation"
icon: "package"
order: 100
componentName: "azurecosmosdbaccount"
---

# Azure Cosmos DB Account

Deploys an Azure Cosmos DB account supporting both SQL/NoSQL and MongoDB APIs, with configurable consistency levels, global distribution across multiple regions, automatic failover, throughput provisioning (fixed or autoscale), backup policies, VNet rules, and IP-based firewall. The component bundles the account with its databases and containers/collections as a single deployable unit.

## What Gets Created

When you deploy an AzureCosmosdbAccount resource, OpenMCF provisions:

- **Cosmos DB Account** -- a `cosmosdb.Account` resource in the specified region and resource group, configured with the chosen API kind, consistency policy, geo-locations, capabilities, and network access rules
- **SQL Databases** -- a `cosmosdb.SqlDatabase` for each entry in `sqlDatabases` (when `kind` is `GlobalDocumentDB`), with optional shared throughput
- **SQL Containers** -- a `cosmosdb.SqlContainer` for each container within a SQL database, configured with partition key, optional dedicated throughput, and TTL
- **MongoDB Databases** -- a `cosmosdb.MongoDatabase` for each entry in `mongoDatabases` (when `kind` is `MongoDB`), with optional shared throughput
- **MongoDB Collections** -- a `cosmosdb.MongoCollection` for each collection within a MongoDB database, configured with shard key, optional throughput, TTL, and indexes
- **Azure Tags** -- resource metadata tags applied to the account for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the account will be created (can reference an AzureResourceGroup resource)
- **A globally unique account name** -- the name becomes the endpoint `https://{name}.documents.azure.com:443/`
- **Partition key design** -- determine the partition key (SQL) or shard key (MongoDB) for each container/collection before deployment; this cannot be changed after creation

## Quick Start

Create a file `cosmosdb.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: my-cosmos
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureCosmosdbAccount.my-cosmos
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-cosmos-db
  geoLocations:
    - location: eastus
      failoverPriority: 0
  sqlDatabases:
    - name: myapp
      containers:
        - name: items
          partitionKeyPaths:
            - /tenantId
```

Deploy:

```shell
openmcf apply -f cosmosdb.yaml
```

This creates a GlobalDocumentDB (SQL API) Cosmos DB account with Session consistency, a single `myapp` database, and an `items` container partitioned by `/tenantId`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the account (e.g., `eastus`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Globally unique account name. Becomes the endpoint `https://{name}.documents.azure.com:443/`. **ForceNew**. | Required, 3-50 characters, pattern `^[-a-z0-9]{3,50}$` |
| `geoLocations` | `list` | Geographic regions for the account. At least one required. First entry with `failoverPriority: 0` is the primary write region. | Minimum 1 item |
| `geoLocations[].location` | `string` | Azure region name (e.g., `eastus`). | Required |
| `geoLocations[].failoverPriority` | `int` | Failover priority (0 = primary write region). Must be unique and contiguous. | Required, >= 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kind` | `string` | `"GlobalDocumentDB"` | API kind. `GlobalDocumentDB` (SQL/NoSQL API) or `MongoDB` (MongoDB wire protocol). **ForceNew**. |
| `consistencyPolicy.consistencyLevel` | `string` | `"Session"` | Default consistency. Values: `Strong`, `BoundedStaleness`, `Session`, `ConsistentPrefix`, `Eventual`. |
| `consistencyPolicy.maxIntervalInSeconds` | `int` | `5` | Max staleness interval (BoundedStaleness only). Range: 5-86400. |
| `consistencyPolicy.maxStalenessPrefix` | `int` | `100` | Max staleness prefix (BoundedStaleness only). Minimum: 10. |
| `geoLocations[].zoneRedundant` | `bool` | `false` | Enable availability zone redundancy for this region. |
| `capabilities` | `string[]` | `[]` | Account capabilities (e.g., `EnableServerless`, `EnableAggregationPipeline`). `EnableMongo` is auto-added for MongoDB kind. |
| `freeTierEnabled` | `bool` | `false` | Enable free tier (1000 RU/s + 25 GB, one per subscription). **ForceNew**. |
| `automaticFailoverEnabled` | `bool` | `false` | Auto-promote next region on failure. Recommended for multi-region. |
| `multipleWriteLocationsEnabled` | `bool` | `false` | Enable active-active multi-region writes. Requires `automaticFailoverEnabled`. |
| `publicNetworkAccessEnabled` | `bool` | `true` | Allow public internet access. |
| `isVirtualNetworkFilterEnabled` | `bool` | `false` | Restrict access to allowed VNet subnets only. |
| `virtualNetworkRules` | `list` | `[]` | Subnet rules. Each has `subnetId` (StringValueOrRef to AzureSubnet). |
| `ipRangeFilter` | `string[]` | `[]` | Allowed CIDR ranges or IPs. Use `0.0.0.0` for all Azure services. |
| `backup.type` | `string` | -- | Backup type: `Periodic` or `Continuous`. |
| `backup.intervalInMinutes` | `int` | `240` | Periodic backup interval (60-1440). |
| `backup.retentionInHours` | `int` | `8` | Periodic backup retention (8-720). |
| `backup.storageRedundancy` | `string` | `"Geo"` | Periodic backup storage: `Geo`, `Local`, `Zone`. |
| `backup.tier` | `string` | -- | Continuous backup tier: `Continuous7Days`, `Continuous30Days`. |
| `mongoServerVersion` | `string` | -- | MongoDB wire protocol version (MongoDB kind only). Values: `3.6`, `4.0`, `4.2`, `5.0`, `6.0`, `7.0`. |
| `sqlDatabases` | `list` | `[]` | SQL API databases with containers (GlobalDocumentDB kind). |
| `mongoDatabases` | `list` | `[]` | MongoDB API databases with collections (MongoDB kind). |

**SQL database/container fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sqlDatabases[].name` | `string` | -- | Database name (required, 1-255 chars) |
| `sqlDatabases[].throughput` | `int` | -- | Shared provisioned RU/s (min 400). Mutually exclusive with `autoscaleMaxThroughput`. |
| `sqlDatabases[].autoscaleMaxThroughput` | `int` | -- | Autoscale max RU/s (min 1000). |
| `containers[].name` | `string` | -- | Container name (required) |
| `containers[].partitionKeyPaths` | `string[]` | -- | Partition key paths (required, e.g., `["/tenantId"]`) |
| `containers[].partitionKeyKind` | `string` | `"Hash"` | `Hash` (single key) or `MultiHash` (hierarchical). |
| `containers[].throughput` | `int` | -- | Dedicated RU/s (min 400). |
| `containers[].autoscaleMaxThroughput` | `int` | -- | Autoscale max RU/s (min 1000). |
| `containers[].defaultTtl` | `int` | -- | Document TTL in seconds (-1 = enabled without default). |

**MongoDB database/collection fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mongoDatabases[].name` | `string` | -- | Database name (required) |
| `mongoDatabases[].throughput` | `int` | -- | Shared provisioned RU/s (min 400). |
| `mongoDatabases[].autoscaleMaxThroughput` | `int` | -- | Autoscale max RU/s (min 1000). |
| `collections[].name` | `string` | -- | Collection name (required) |
| `collections[].shardKey` | `string` | -- | Shard key field (required, e.g., `tenantId`) |
| `collections[].throughput` | `int` | -- | Dedicated RU/s (min 400). |
| `collections[].defaultTtlSeconds` | `int` | -- | Document TTL in seconds. |
| `collections[].indexes` | `list` | `[]` | Indexes with `keys` (string[]) and optional `unique` (bool). |

## Examples

### SQL API with Autoscale

A SQL API account with autoscale throughput and multiple containers:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: app-cosmos
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureCosmosdbAccount.app-cosmos
spec:
  region: eastus
  resourceGroup: prod-rg
  name: app-cosmos-db
  automaticFailoverEnabled: true
  backup:
    type: Continuous
    tier: Continuous7Days
  geoLocations:
    - location: eastus
      failoverPriority: 0
      zoneRedundant: true
  sqlDatabases:
    - name: appdata
      autoscaleMaxThroughput: 4000
      containers:
        - name: users
          partitionKeyPaths:
            - /tenantId
          defaultTtl: -1
        - name: sessions
          partitionKeyPaths:
            - /userId
          defaultTtl: 86400
```

### MongoDB API Account

A MongoDB account with MongoDB 7.0 wire protocol, a sharded collection, and indexes:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: mongo-cosmos
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureCosmosdbAccount.mongo-cosmos
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: mongo-cosmos-db
  kind: MongoDB
  mongoServerVersion: "7.0"
  automaticFailoverEnabled: true
  geoLocations:
    - location: westeurope
      failoverPriority: 0
  mongoDatabases:
    - name: catalog
      autoscaleMaxThroughput: 4000
      collections:
        - name: products
          shardKey: categoryId
          indexes:
            - keys:
                - name
              unique: false
            - keys:
                - sku
              unique: true
        - name: reviews
          shardKey: productId
          defaultTtlSeconds: 7776000
```

### Multi-Region with Strong Consistency

A globally distributed SQL API account with BoundedStaleness consistency, zone-redundant regions, and continuous backup:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: global-cosmos
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureCosmosdbAccount.global-cosmos
spec:
  region: eastus
  resourceGroup: prod-rg
  name: global-cosmos-db
  automaticFailoverEnabled: true
  consistencyPolicy:
    consistencyLevel: BoundedStaleness
    maxIntervalInSeconds: 300
    maxStalenessPrefix: 100000
  backup:
    type: Continuous
    tier: Continuous30Days
  geoLocations:
    - location: eastus
      failoverPriority: 0
      zoneRedundant: true
    - location: westeurope
      failoverPriority: 1
      zoneRedundant: true
    - location: southeastasia
      failoverPriority: 2
  sqlDatabases:
    - name: orders
      containers:
        - name: transactions
          partitionKeyPaths:
            - /regionId
            - /customerId
          partitionKeyKind: MultiHash
          autoscaleMaxThroughput: 10000
```

### Using Foreign Key References

Reference OpenMCF-managed resources for the resource group and VNet rules:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: ref-cosmos
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureCosmosdbAccount.ref-cosmos
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-cosmos-db
  isVirtualNetworkFilterEnabled: true
  virtualNetworkRules:
    - subnetId:
        valueFrom:
          kind: AzureSubnet
          name: app-subnet
          field: status.outputs.subnet_id
  geoLocations:
    - location: eastus
      failoverPriority: 0
  sqlDatabases:
    - name: mydb
      containers:
        - name: items
          partitionKeyPaths:
            - /id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `account_id` | `string` | Azure Resource Manager ID of the Cosmos DB account. Referenced by AzurePrivateEndpoint for private connectivity. |
| `account_name` | `string` | Name of the Cosmos DB account |
| `endpoint` | `string` | Document endpoint URI (e.g., `https://{name}.documents.azure.com:443/`) |
| `primary_key` | `string` | Primary access key for authentication (sensitive) |
| `primary_connection_string` | `string` | SQL API connection string (sensitive). Always populated. |
| `primary_mongodb_connection_string` | `string` | MongoDB API connection string (sensitive). Only populated when `kind` is `MongoDB`. |
| `database_ids` | `map<string, string>` | Map of database names to their Azure Resource Manager IDs |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/resource-group) -- provides the resource group for account placement
- [AzureSubnet](/docs/catalog/azure/subnet) -- provides subnets for VNet access rules
- [AzurePrivateEndpoint](/docs/catalog/azure/private-endpoint) -- establishes private connectivity to the account
- [AzureKeyVault](/docs/catalog/azure/key-vault) -- stores the primary key or connection strings as secrets
