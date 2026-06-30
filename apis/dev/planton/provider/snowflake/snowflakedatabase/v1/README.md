# SnowflakeDatabase

Snowflake database resource for creating and managing databases in Snowflake. Provides comprehensive configuration options for Time Travel, data retention, Iceberg tables, task management, and performance optimization.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)
- **name**: Database identifier (must be unique for your account). Avoid these characters: `|`, `.`, `(`, `)`, `"`
- **comment**: Human-readable description of the database purpose
- **data_retention_time_in_days**: Number of days for which Time Travel actions (CLONE and UNDROP) can be performed. Also sets default retention for all schemas in the database.
- **is_transient**: If true, creates a transient database. Transient databases don't have a Fail-safe period, reducing storage costs but eliminating Fail-safe data protection.
- **drop_public_schema_on_creation**: If true, drops the public schema when creating the database (cannot be changed after creation).

### Advanced Fields (20% Use Case)

**Iceberg Configuration**
- **catalog**: Default catalog for Iceberg tables
- **external_volume**: Default external volume for Iceberg tables
- **storage_serialization_policy**: Encoding and compression strategy for Iceberg tables:
  - `COMPATIBLE`: Ensures interoperability with third-party compute engines
  - `OPTIMIZED`: Best performance within Snowflake
- **replace_invalid_characters**: Replace invalid UTF-8 characters with � in query results for Iceberg tables

**Collation and Encoding**
- **default_ddl_collation**: Default collation for all schemas and tables (can be overridden at schema/table level)
- **quoted_identifiers_ignore_case**: If true, case of quoted identifiers is ignored

**Logging and Monitoring**
- **log_level**: Severity level for ingestion into active event table
  - Options: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `OFF`
- **trace_level**: Controls trace event ingestion into event table
  - Options: `ALWAYS`, `ON_EVENT`, `OFF`
- **enable_console_output**: Enable stdout/stderr logging for anonymous stored procedures

**Data Retention and Streams**
- **max_data_extension_time_in_days**: Maximum days Snowflake can extend data retention to prevent streams from becoming stale

**Task Configuration**
- **suspend_task_after_num_failures**: Auto-suspend task after this many consecutive failures (0 = disabled)
- **task_auto_retry_attempts**: Maximum automatic retries for user tasks
- **user_task**: User task configuration object:
  - `managed_initial_warehouse_size`: Initial warehouse size for managed warehouses
  - `minimum_trigger_interval_in_seconds`: Minimum time between triggered task executions
  - `timeout_ms`: Task execution timeout in milliseconds

## Stack Outputs

- **id**: The provider-assigned unique ID for this managed resource
- **name**: The fully-qualified name of the created database
- **owner**: The owner role of the database
- **created_on**: Timestamp when the database was created
- **is_transient**: Boolean indicating if the database is transient
- **data_retention_time_in_days**: Configured data retention time in days

## How It Works

Planton provisions Snowflake databases via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (api.proto, spec.proto) and stack execution is orchestrated by the platform using the SnowflakeDatabaseStackInput (includes Snowflake credentials and IaC metadata).

### Time Travel and Data Retention

Snowflake's Time Travel feature allows you to access historical data that has been changed or deleted:
- **Time Travel**: Query, clone, and restore data from specific points in the past
- **Retention Period**: Controlled by `data_retention_time_in_days`
- **Standard**: Up to 1 day (default for most editions)
- **Enterprise**: Up to 90 days
- **Use Cases**: Recover from accidental deletions, analyze historical data, reproduce past results

### Transient vs. Permanent Databases

| Aspect | Permanent Database | Transient Database |
|--------|-------------------|-------------------|
| **Time Travel** | Yes (configured retention) | Yes (configured retention) |
| **Fail-safe** | Yes (7 days) | No |
| **Storage Cost** | Higher (includes Fail-safe) | Lower (no Fail-safe overhead) |
| **Data Protection** | Maximum | Reduced |
| **Best For** | Production, critical data | Staging, temporary data |

**Recommendation**: Use transient databases only for non-critical data where cost savings outweigh data protection needs.

### Iceberg Table Support

Snowflake supports Apache Iceberg tables with external catalogs:
- **Catalog**: Specify default catalog for Iceberg table metadata
- **External Volume**: Define storage location for Iceberg data files
- **Serialization Policy**:
  - `COMPATIBLE`: Use when sharing data with other compute engines (Spark, Trino, etc.)
  - `OPTIMIZED`: Use for Snowflake-only workloads (better performance)

### Task Management

Configure default task behavior at the database level:
- **Auto-Suspend**: Automatically suspend tasks after repeated failures
- **Auto-Retry**: Automatically retry failed tasks
- **Managed Warehouses**: Let Snowflake manage warehouse sizing for serverless tasks
- **Trigger Intervals**: Control minimum time between task executions

## Multi-Environment Best Practice

Use separate databases for each environment:
- `analytics_dev` → Development
- `analytics_staging` → Staging
- `analytics_prod` → Production

Each database provides isolation for:
- Data retention policies
- Access control and ownership
- Cost allocation and monitoring
- Task scheduling and execution

## Common Use Cases

### Standard Production Database
Permanent database with 30-day Time Travel retention for production workloads.

### Transient Staging Database
Cost-optimized transient database for staging environments where Fail-safe isn't required.

### Iceberg Data Lake Database
Database configured for Iceberg tables with external storage and compatible serialization.

### High-Compliance Database
Maximum data retention (90 days) with comprehensive logging and monitoring enabled.

### Task-Intensive Database
Configured with task auto-retry, managed warehouses, and failure handling for ETL workloads.

## Cost Optimization

### Storage Costs
- **Transient Databases**: Save on Fail-safe storage costs (~50% reduction for inactive data)
- **Data Retention**: Lower retention periods reduce storage costs but limit recovery options
- **Time Travel**: Balance between data protection and storage costs

### Compute Costs
- **Task Configuration**: Tune warehouse sizing and retry attempts to minimize wasted compute
- **Logging**: Use appropriate log levels to avoid excessive log ingestion costs

### Best Practices
- Use transient databases for non-critical, reproducible data
- Set data retention periods based on actual recovery needs (not maximums)
- Monitor storage growth and implement data lifecycle policies
- Use appropriate log levels (avoid TRACE/DEBUG in production)

## References

- Snowflake Database Documentation: https://docs.snowflake.com/en/sql-reference/sql/create-database
- Time Travel: https://docs.snowflake.com/en/user-guide/data-time-travel
- Fail-safe: https://docs.snowflake.com/en/user-guide/data-failsafe
- Iceberg Tables: https://docs.snowflake.com/en/user-guide/tables-iceberg
- Task Management: https://docs.snowflake.com/en/user-guide/tasks-intro
- Terraform Provider: https://registry.terraform.io/providers/Snowflake-Labs/snowflake/latest
- Pulumi Provider: https://www.pulumi.com/registry/packages/snowflake/
