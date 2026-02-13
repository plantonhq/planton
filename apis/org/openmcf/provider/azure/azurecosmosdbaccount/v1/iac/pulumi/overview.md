# AzureCosmosdbAccount Pulumi Module Architecture

## Resource Graph

```
cosmosdb.Account
├── [kind = GlobalDocumentDB]
│   ├── cosmosdb.SqlDatabase (per spec.sql_databases)
│   │   └── cosmosdb.SqlContainer (per database.containers)
│   │       └── Uses partition_key_paths, throughput/autoscale
│   └── (no mongo resources)
│
└── [kind = MongoDB]
    ├── cosmosdb.MongoDatabase (per spec.mongo_databases)
    │   └── cosmosdb.MongoCollection (per database.collections)
    │       └── Uses shard_key, indexes, throughput/autoscale
    └── (no SQL resources)
```

## Kind-Based Branching Logic

The module branches on `spec.kind`:

- **GlobalDocumentDB** (default): Creates `SqlDatabase` and `SqlContainer`
  resources. Each container requires `partition_key_paths`. Throughput can be
  at database level (shared) or container level (dedicated).

- **MongoDB**: Creates `MongoDatabase` and `MongoCollection` resources. Each
  collection requires `shard_key`. Optional `indexes` blocks define compound or
  unique indexes. The `EnableMongo` capability is automatically added when kind
  is MongoDB.

## Resource Dependencies

1. Account must exist before any database or container/collection.
2. SQL/Mongo databases must exist before their containers/collections.
3. Throughput (or autoscale) must be set at creation; switching between them
   requires destroy/recreate.

## Outputs

| Output | Source | Notes |
|--------|--------|-------|
| account_id | Account.Id | For AzurePrivateEndpoint private_connection_resource_id |
| endpoint | Account.Endpoint | SDK connectivity |
| primary_key | Account.PrimaryKey | Key-based auth |
| primary_connection_string | Account.PrimarySqlConnectionString | SQL API |
| primary_mongodb_connection_string | Account.PrimaryMongodbConnectionString | MongoDB API only |
| database_ids | Merge of SqlDatabase/MongoDatabase IDs | Map by database name |
