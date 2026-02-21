# AlicloudPolardbCluster Pulumi Module -- Architecture Overview

## Resource Graph

```
alicloud.Provider (region)
  └── polardb.Cluster (main cluster)
        ├── polardb.Database (per databases[] entry)
        ├── polardb.Account (per accounts[] entry)
        │     └── polardb.AccountPrivilege (per privileges[] entry)
        └── Stack Outputs: cluster_id, connection_string, port, database_ids
```

## Execution Flow

1. **Provider** -- configured with the region from spec
2. **Cluster** -- created with engine, node class, node count, and all optional configurations
3. **Databases** -- created as children of the cluster, with engine-appropriate charset defaults
4. **Accounts** -- created as children of the cluster, after databases exist
5. **Privileges** -- created as children of accounts, linking accounts to databases
6. **Outputs** -- cluster ID, primary connection string, port, and database ID map exported

## Key Design Points

- All resources use `pulumi.Parent()` to establish clear ownership hierarchy
- Database charset defaults are engine-aware (utf8 for MySQL, UTF8 for PG/Oracle)
- Optional spec fields use nil-checking helpers to avoid passing zero values
- Tags are computed from metadata (name, org, env, id) plus user-specified tags
