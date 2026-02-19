# AlicloudRdsInstance Research Documentation

## Provider Resource Analysis

### alicloud_db_instance (Terraform) / rds.Instance (Pulumi)

The RDS instance is the core resource. Key findings from provider analysis:

- **Multi-engine**: A single resource type handles MySQL, PostgreSQL, SQL Server, MariaDB, and PPAS via the `engine` parameter (per DD02)
- **Category architecture**: The `category` field controls deployment topology (Basic, HighAvailability, AlwaysOn, Finance, cluster), replacing the simplified `high_availability` boolean from the initial spec
- **Storage types**: `db_instance_storage_type` selects the disk tier (local_ssd, cloud_ssd, cloud_essd, cloud_essd2, cloud_essd3)
- **Connection strings**: The instance exposes an intranet `connection_string` and `port` as computed outputs; public endpoints require a separate `alicloud_db_connection` resource

### alicloud_db_database (Terraform) / rds.Database (Pulumi)

- `data_base_name` is the primary field (ForceNew)
- `character_set` defaults vary by engine; the IaC module applies engine-appropriate defaults
- Character set is ForceNew -- changing it requires recreating the database

### alicloud_rds_account (Terraform) / rds.RdsAccount (Pulumi)

- `account_name` and `account_type` are ForceNew
- `account_type` supports "Normal", "Super", "Sysadmin" (Sysadmin is SQL Server only, excluded from spec for simplicity)
- Password complexity enforced by the provider API

### alicloud_db_account_privilege (Terraform) / rds.AccountPrivilege (Pulumi)

- Links an account to database(s) with one privilege level
- Privilege levels: ReadOnly, ReadWrite, DDLOnly, DMLOnly, DBOwner
- DBOwner is SQL Server specific
- The `privilege` field is ForceNew

## Design Rationale

### No Public Connection Endpoint

The `alicloud_db_connection` resource creates a public internet endpoint for the RDS instance. This is intentionally excluded from the component because:

1. Exposing databases to the internet is a security anti-pattern
2. The instance already provides an intranet connection string for VPC-internal access
3. Production workloads almost always connect via VPC, not the public internet
4. Users who need public access can manage the `alicloud_db_connection` resource separately

### Structured Account Privileges

The initial spec design used `repeated string database_privileges` which was unstructured. The provider models privileges as a separate resource (`alicloud_db_account_privilege`) with explicit fields. The structured `AlicloudRdsAccountPrivilege` message provides a clean YAML experience while accurately mapping to the provider's resource model.

### Category over Boolean HA

The initial spec used `bool high_availability` which oversimplifies the provider's `category` field. The actual deployment architectures (Basic, HighAvailability, AlwaysOn, Finance, cluster) have meaningful differences in node count, failover behavior, and cost that users need to control explicitly.

### Fields Excluded from Spec

| Field | Reason |
| --- | --- |
| `babelfish_config` | Very niche PostgreSQL feature |
| `pg_hba_conf` | Too operational; belongs in parameter tuning |
| `sql_collector_*` | Monitoring detail, not infrastructure config |
| `serverless_config` | Different billing paradigm; could be separate component |
| `storage_auto_scale` | Operational scaling policy, not declarative config |
| `cold_data_enabled` | Niche MySQL-only feature |
| `bursting_enabled` | Performance tuning detail |
| `optimized_writes` | MySQL-specific optimization |
| `recovery_model` | SQL Server-specific |
| `pg_bouncer_enabled` | PostgreSQL connection pooling tuning |
| `tcp_connection_type` | Network tuning detail |

### Composite Bundling (DD07)

RDS Instance + Databases + Accounts + Privileges are bundled because:
- An RDS instance without databases has no schemas for applications to use
- Accounts are required for any client to connect
- Privileges define the access model -- without them, Normal accounts have no access
- Together they form a complete, functional database deployment unit
