# AzureCosmosdbAccount Examples

## 1. Minimal SQL API

Single region, one database, one container with partition key. Uses provisioned throughput.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: my-cosmos-sql
spec:
  region: eastus
  resource_group: my-rg
  name: my-cosmos-sql
  kind: GlobalDocumentDB
  geo_locations:
    - location: eastus
      failover_priority: 0
  sql_databases:
    - name: mydb
      throughput: 400
      containers:
        - name: users
          partition_key_paths:
            - /userId
          throughput: 400
```

## 2. Minimal MongoDB API

Single region, one database, one collection with shard key.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: my-cosmos-mongo
spec:
  region: eastus
  resource_group: my-rg
  name: my-cosmos-mongo
  kind: MongoDB
  mongo_server_version: "4.2"
  geo_locations:
    - location: eastus
      failover_priority: 0
  mongo_databases:
    - name: mydb
      throughput: 400
      collections:
        - name: products
          shard_key: category
          throughput: 400
```

## 3. Multi-Region with Automatic Failover

Two regions, zone redundant, Session consistency. Automatic failover promotes West Europe if East US fails.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: prod-cosmos
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-rg
  name: prod-cosmos-account
  kind: GlobalDocumentDB
  consistency_policy:
    consistency_level: Session
  automatic_failover_enabled: true
  geo_locations:
    - location: eastus
      failover_priority: 0
      zone_redundant: true
    - location: westeurope
      failover_priority: 1
      zone_redundant: true
  sql_databases:
    - name: app
      autoscale_max_throughput: 4000
      containers:
        - name: orders
          partition_key_paths:
            - /tenantId
          autoscale_max_throughput: 1000
```

## 4. Serverless Mode

EnableServerless capability. No throughput config — pay per request.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: dev-cosmos-serverless
spec:
  region: eastus
  resource_group: dev-rg
  name: dev-cosmos-serverless
  kind: GlobalDocumentDB
  capabilities:
    - EnableServerless
  geo_locations:
    - location: eastus
      failover_priority: 0
  sql_databases:
    - name: devdb
      containers:
        - name: events
          partition_key_paths:
            - /eventId
```

## 5. BoundedStaleness Consistency

Reads lag behind writes by at most 300 seconds or 100000 versions.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: analytics-cosmos
spec:
  region: eastus
  resource_group: analytics-rg
  name: analytics-cosmos
  kind: GlobalDocumentDB
  consistency_policy:
    consistency_level: BoundedStaleness
    max_interval_in_seconds: 300
    max_staleness_prefix: 100000
  geo_locations:
    - location: eastus
      failover_priority: 0
  sql_databases:
    - name: analytics
      throughput: 1000
      containers:
        - name: metrics
          partition_key_paths:
            - /region
```

## 6. Private Access with VNet Rules

Restrict access to traffic from a specific subnet. Subnet must have `Microsoft.AzureCosmosDB` service endpoint.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: private-cosmos
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-rg
  name: private-cosmos-account
  kind: GlobalDocumentDB
  public_network_access_enabled: true
  is_virtual_network_filter_enabled: true
  virtual_network_rules:
    - subnet_id:
        value_from:
          kind: AzureSubnet
          name: app-subnet
          field_path: status.outputs.subnet_id
  geo_locations:
    - location: eastus
      failover_priority: 0
  sql_databases:
    - name: app
      throughput: 400
      containers:
        - name: data
          partition_key_paths:
            - /tenantId
```

## 7. Production SQL API

Multi-region, autoscale, multiple databases/containers, Continuous backup, zone redundant.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: prod-cosmos-full
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group: prod-rg
  name: prod-cosmos-full
  kind: GlobalDocumentDB
  consistency_policy:
    consistency_level: Session
  automatic_failover_enabled: true
  geo_locations:
    - location: eastus
      failover_priority: 0
      zone_redundant: true
    - location: westeurope
      failover_priority: 1
      zone_redundant: true
  backup:
    type: Continuous
    tier: Continuous7Days
  sql_databases:
    - name: core
      autoscale_max_throughput: 10000
      containers:
        - name: users
          partition_key_paths:
            - /tenantId
          autoscale_max_throughput: 4000
        - name: sessions
          partition_key_paths:
            - /userId
          autoscale_max_throughput: 1000
    - name: analytics
      autoscale_max_throughput: 4000
      containers:
        - name: events
          partition_key_paths:
            - /date
          default_ttl: 2592000
  ip_range_filter:
    - "0.0.0.0"
```

## 8. Infra-Chart valueFrom Pattern

Resource group from AzureResourceGroup output. Composes in a database-stack chart.

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureCosmosdbAccount
metadata:
  name: chart-cosmos
  org: mycompany
  env: production
spec:
  region: eastus
  resource_group:
    value_from:
      kind: AzureResourceGroup
      name: shared-rg
      field_path: status.outputs.resource_group_name
  name: chart-cosmos-account
  kind: GlobalDocumentDB
  geo_locations:
    - location: eastus
      failover_priority: 0
  sql_databases:
    - name: app
      throughput: 400
      containers:
        - name: items
          partition_key_paths:
            - /id
```
