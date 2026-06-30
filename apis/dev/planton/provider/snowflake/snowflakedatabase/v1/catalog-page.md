# Snowflake Database

Deploys a Snowflake database with configurable Time Travel retention, transient mode, Iceberg table defaults, and task execution parameters. All database-level settings are managed declaratively through the spec, including collation, logging, and user task configuration.

## What Gets Created

When you deploy a SnowflakeDatabase resource, Planton provisions:

- **Snowflake Database** — a `snowflake_database` resource with the specified name and all configured parameters including Time Travel retention, collation, and Iceberg settings
- **Snowflake Provider** — configured using explicit credentials from provider config or environment variables

## Prerequisites

- **Snowflake credentials** configured via environment variables (`SNOWFLAKE_ACCOUNT`, `SNOWFLAKE_USER`, `SNOWFLAKE_PASSWORD`) or Planton provider config
- **A Snowflake account** with permissions to create databases
- **An external volume** if configuring Iceberg table defaults via `externalVolume`

## Quick Start

Create a file `snowflake-database.yaml`:

```yaml
apiVersion: snowflake.planton.dev/v1
kind: SnowflakeDatabase
metadata:
  name: my-database
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.SnowflakeDatabase.my-database
spec:
  name: MY_DATABASE
```

Deploy:

```shell
planton apply -f snowflake-database.yaml
```

This creates a standard (non-transient) Snowflake database named `MY_DATABASE` with default settings.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Snowflake identifier for the database. Must be unique within the account. Avoid characters: `\|`, `.`, `(`, `)`, `"`. | — |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `catalog` | `string` | `""` | Default catalog for Iceberg tables. Sets the `CATALOG` database parameter. |
| `comment` | `string` | `""` | Comment attached to the database. |
| `dataRetentionTimeInDays` | `int` | `0` | Number of days for Time Travel actions (CLONE and UNDROP). Also sets the default retention for all schemas in the database. |
| `defaultDdlCollation` | `string` | `""` | Default collation specification for all schemas and tables. Can be overridden at schema or table level. |
| `dropPublicSchemaOnCreation` | `bool` | `false` | Drops the `PUBLIC` schema on database creation. Has no effect if changed after creation. |
| `enableConsoleOutput` | `bool` | `false` | Enables stdout/stderr fast path logging for anonymous stored procedures. |
| `externalVolume` | `string` | `""` | Default external volume for Iceberg tables. Sets the `EXTERNAL_VOLUME` database parameter. |
| `isTransient` | `bool` | `false` | Creates a transient database. Transient databases have no Fail-safe period, reducing storage costs but removing Fail-safe data recovery. |
| `logLevel` | `string` | `""` | Severity level for event table ingestion. Valid values: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`, `OFF`. |
| `maxDataExtensionTimeInDays` | `int` | `0` | Maximum days Snowflake can extend data retention to prevent streams from becoming stale. |
| `quotedIdentifiersIgnoreCase` | `bool` | `false` | When `true`, the case of quoted identifiers is ignored. |
| `replaceInvalidCharacters` | `bool` | `false` | Replaces invalid UTF-8 characters with the Unicode replacement character in Iceberg table query results. Only applies to tables using an external Iceberg catalog. |
| `storageSerializationPolicy` | `string` | `""` | Storage serialization policy for Iceberg tables using Snowflake as catalog. Valid values: `COMPATIBLE` (interoperable with third-party engines), `OPTIMIZED` (best performance within Snowflake). |
| `suspendTaskAfterNumFailures` | `int` | `0` | Consecutive task failures before automatic suspension. `0` disables auto-suspending. |
| `taskAutoRetryAttempts` | `int` | `0` | Maximum automatic retry attempts for user tasks. |
| `traceLevel` | `string` | `""` | Controls trace event ingestion into the event table. Valid values: `ALWAYS`, `ON_EVENT`, `OFF`. |
| `userTask.managedInitialWarehouseSize` | `string` | `""` | Initial warehouse size for managed warehouses when no execution history exists. |
| `userTask.minimumTriggerIntervalInSeconds` | `int` | `0` | Minimum seconds between triggered task executions. |
| `userTask.timeoutMs` | `int` | `0` | User task execution timeout in milliseconds. |

## Examples

### Transient Database with Short Retention

A cost-optimized transient database with 1-day Time Travel retention, suitable for staging or ephemeral workloads:

```yaml
apiVersion: snowflake.planton.dev/v1
kind: SnowflakeDatabase
metadata:
  name: staging-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.SnowflakeDatabase.staging-db
spec:
  name: STAGING_DB
  isTransient: true
  dataRetentionTimeInDays: 1
  comment: "Staging environment - transient, short retention"
```

### Production Database with Extended Retention

A production database with 14-day Time Travel retention, extended data retention for stream staleness prevention, and debug-level logging:

```yaml
apiVersion: snowflake.planton.dev/v1
kind: SnowflakeDatabase
metadata:
  name: prod-analytics
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.SnowflakeDatabase.prod-analytics
spec:
  name: PROD_ANALYTICS
  dataRetentionTimeInDays: 14
  maxDataExtensionTimeInDays: 28
  logLevel: DEBUG
  traceLevel: ON_EVENT
  defaultDdlCollation: "en-ci"
  comment: "Production analytics database"
```

### Full-Featured Database with Iceberg and Task Configuration

A database configured for Iceberg table workloads with task execution settings, suitable for data lake integration:

```yaml
apiVersion: snowflake.planton.dev/v1
kind: SnowflakeDatabase
metadata:
  name: data-lake-db
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.SnowflakeDatabase.data-lake-db
spec:
  name: DATA_LAKE_DB
  dataRetentionTimeInDays: 7
  maxDataExtensionTimeInDays: 14
  catalog: "my_iceberg_catalog"
  externalVolume: "my_external_volume"
  storageSerializationPolicy: COMPATIBLE
  replaceInvalidCharacters: true
  logLevel: INFO
  traceLevel: ON_EVENT
  suspendTaskAfterNumFailures: 5
  taskAutoRetryAttempts: 3
  userTask:
    managedInitialWarehouseSize: "XSMALL"
    minimumTriggerIntervalInSeconds: 60
    timeoutMs: 3600000
  comment: "Data lake integration database with Iceberg support"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Provider-assigned unique ID for the database resource |
| `name` | `string` | Fully-qualified name of the created database |
| `owner` | `string` | Owner role of the database |
| `created_on` | `string` | Timestamp when the database was created |
| `is_transient` | `string` | Whether the database is transient (`"true"` or `"false"`) |
| `data_retention_time_in_days` | `string` | Configured data retention time in days |

## Related Components

No other Planton components have direct foreign key references to SnowflakeDatabase. This component is typically deployed as a standalone resource within a Snowflake account.
