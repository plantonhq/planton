# Snowflake Database Terraform Module

## Overview

This Terraform module provisions and manages Snowflake databases with comprehensive configuration support for all Snowflake database parameters. The module is designed to work with Planton's standardized infrastructure patterns and provides cost-conscious defaults for database creation.

## Key Features

### Database Configuration

- **Complete Parameter Support**: All Snowflake database parameters from the `SnowflakeDatabaseSpec` are supported, including:
  - **Data Retention Policies**: Configure Time Travel and Fail-safe retention periods
  - **Cost Optimization**: Support for transient databases to reduce storage costs
  - **Iceberg Integration**: Catalog and external volume configuration for Iceberg tables
  - **Task Management**: User task configuration for automated workflows
  - **Logging and Tracing**: Configurable log levels and trace settings
  - **Collation and Serialization**: Default collation and storage serialization policies

- **Provider Configuration**: Secure authentication using Snowflake credentials (account, region, username, password)

- **Flexible Naming**: Database naming aligned with organizational conventions

### Cost Governance

The module emphasizes cost-conscious defaults:

- **Transient Database Support**: The `is_transient` flag eliminates the 7-day Fail-safe period, significantly reducing storage costs for development environments, CI/CD pipelines, and recreatable data.

- **Data Retention Control**: Configurable `data_retention_time_in_days` to balance recovery needs with storage costs.

- **Smart Defaults**: Optional parameters only apply values when explicitly configured, avoiding unnecessary cost implications.

## Prerequisites

- Terraform >= 1.0
- Snowflake account with appropriate permissions
- Snowflake credentials (account identifier, region, username, password)

## Provider Configuration

The module uses the Snowflake Terraform provider from Snowflake Labs:

```hcl
terraform {
  required_providers {
    snowflake = {
      source = "Snowflake-Labs/snowflake"
    }
  }
}
```

Credentials are passed via the `snowflake_credential` variable.

## Usage

### Basic Example

```hcl
module "snowflake_database" {
  source = "./iac/tf"

  metadata = {
    name = "my-database"
    id   = "db-prod-001"
    org  = "engineering"
    env  = "production"
    labels = {
      team    = "data-platform"
      project = "analytics"
    }
  }

  spec = {
    name                        = "ANALYTICS_PROD"
    comment                     = "Production analytics database"
    is_transient                = false
    data_retention_time_in_days = 7

    # User task configuration
    user_task = {
      managed_initial_warehouse_size = "MEDIUM"
      minimum_trigger_interval_in_seconds = 60
      timeout_ms = 3600000
    }

    # Other optional parameters
    catalog                     = ""
    default_ddl_collation       = ""
    drop_public_schema_on_creation = false
    enable_console_output       = true
    external_volume             = ""
    log_level                   = "INFO"
    max_data_extension_time_in_days = 0
    quoted_identifiers_ignore_case = false
    replace_invalid_characters  = false
    storage_serialization_policy = ""
    suspend_task_after_num_failures = 10
    task_auto_retry_attempts    = 3
    trace_level                 = "ON_EVENT"
  }

  snowflake_credential = {
    account  = "xy12345.us-east-1"
    region   = "us-east-1"
    username = "terraform_service_account"
    password = var.snowflake_password  # Pass via variable, not hardcoded
  }
}
```

### Development/CI Database (Cost-Optimized)

```hcl
module "dev_database" {
  source = "./iac/tf"

  metadata = {
    name = "dev-database"
    env  = "development"
  }

  spec = {
    name                        = "ANALYTICS_DEV"
    comment                     = "Development database - transient for cost savings"
    is_transient                = true  # Eliminates Fail-safe storage costs
    data_retention_time_in_days = 1     # Minimal Time Travel retention

    # Minimal configuration for dev
    user_task = {
      managed_initial_warehouse_size = "XSMALL"
      minimum_trigger_interval_in_seconds = 0
      timeout_ms = 0
    }

    catalog                     = ""
    default_ddl_collation       = ""
    drop_public_schema_on_creation = false
    enable_console_output       = true
    external_volume             = ""
    log_level                   = ""
    max_data_extension_time_in_days = 0
    quoted_identifiers_ignore_case = false
    replace_invalid_characters  = false
    storage_serialization_policy = ""
    suspend_task_after_num_failures = 0
    task_auto_retry_attempts    = 0
    trace_level                 = ""
  }

  snowflake_credential = {
    account  = "xy12345.us-east-1"
    region   = "us-east-1"
    username = "terraform_service_account"
    password = var.snowflake_password
  }
}
```

### Iceberg-Enabled Database

```hcl
module "iceberg_database" {
  source = "./iac/tf"

  metadata = {
    name = "iceberg-lakehouse"
    env  = "production"
  }

  spec = {
    name                        = "LAKEHOUSE_PROD"
    comment                     = "Iceberg-enabled lakehouse database"
    is_transient                = false
    data_retention_time_in_days = 7

    # Iceberg-specific configuration
    catalog                     = "iceberg_catalog"
    external_volume             = "s3_external_volume"
    storage_serialization_policy = "OPTIMIZED"

    user_task = {
      managed_initial_warehouse_size = ""
      minimum_trigger_interval_in_seconds = 0
      timeout_ms = 0
    }

    default_ddl_collation       = ""
    drop_public_schema_on_creation = false
    enable_console_output       = false
    log_level                   = ""
    max_data_extension_time_in_days = 0
    quoted_identifiers_ignore_case = false
    replace_invalid_characters  = false
    suspend_task_after_num_failures = 0
    task_auto_retry_attempts    = 0
    trace_level                 = ""
  }

  snowflake_credential = {
    account  = "xy12345.us-east-1"
    region   = "us-east-1"
    username = "terraform_service_account"
    password = var.snowflake_password
  }
}
```

## Input Variables

### metadata

Resource metadata for identification and organization:

- `name` (string, required): Resource name
- `id` (string, optional): Unique identifier
- `org` (string, optional): Organization name
- `env` (string, optional): Environment (dev, staging, prod)
- `labels` (map(string), optional): Key-value labels
- `tags` (list(string), optional): Resource tags
- `version` (object, optional): Version tracking

### spec

Snowflake database specification with all configuration parameters:

- `name` (string, required): Database name (must be unique in Snowflake account)
- `catalog` (string): Default catalog for Iceberg tables
- `comment` (string): Database description
- `data_retention_time_in_days` (number): Time Travel retention period (0-90 days)
- `default_ddl_collation` (string): Default collation specification
- `drop_public_schema_on_creation` (bool): Whether to drop public schema on creation
- `enable_console_output` (bool): Enable stdout/stderr logging for stored procedures
- `external_volume` (string): Default external volume for Iceberg tables
- `is_transient` (bool): Create as transient database (eliminates Fail-safe)
- `log_level` (string): Log message severity level (TRACE, DEBUG, INFO, WARN, ERROR, FATAL, OFF)
- `max_data_extension_time_in_days` (number): Maximum data extension days for streams
- `quoted_identifiers_ignore_case` (bool): Ignore case for quoted identifiers
- `replace_invalid_characters` (bool): Replace invalid UTF-8 characters
- `storage_serialization_policy` (string): Storage policy (COMPATIBLE, OPTIMIZED)
- `suspend_task_after_num_failures` (number): Auto-suspend threshold
- `task_auto_retry_attempts` (number): Automatic retry attempts for tasks
- `trace_level` (string): Trace event ingestion level (ALWAYS, ON_EVENT, OFF)
- `user_task` (object): User task configuration
  - `managed_initial_warehouse_size` (string): Initial warehouse size
  - `minimum_trigger_interval_in_seconds` (number): Minimum trigger interval
  - `timeout_ms` (number): Task execution timeout

### snowflake_credential

Snowflake authentication credentials:

- `account` (string, required): Snowflake account identifier
- `region` (string, required): Snowflake region
- `username` (string, required): Authentication username
- `password` (string, required): Authentication password (use secrets management)

## Outputs

### id

The Snowflake database resource ID.

### name

The database name as created in Snowflake.

### is_transient

Boolean indicating if the database is transient (cost indicator).

### data_retention_time_in_days

Number of days configured for Time Travel retention (cost indicator).

### bootstrap_endpoint, crn, rest_endpoint

These outputs are placeholders for proto compatibility and are not applicable to Snowflake databases. They return empty strings.

## Cost Considerations

### Critical Parameters for Cost Control

1. **`is_transient`**: Set to `true` for development, testing, CI/CD, and staging environments. This eliminates the 7-day Fail-safe period, reducing storage costs by up to 50% for recreatable data.

2. **`data_retention_time_in_days`**: Balance recovery needs with storage costs:
   - Production: 7-90 days (typical: 7 days)
   - Development: 1 day (minimal retention)
   - CI/CD: 0 days (no retention needed)

3. **Parameter Inheritance**: Remember that database-level settings cascade to all schemas and tables. Set cost-conscious defaults at the database level.

### Storage Cost Formula

For permanent databases:
```
Storage Cost = Active Data + Time Travel Data + Fail-safe Data (7 days)
```

For transient databases:
```
Storage Cost = Active Data + Time Travel Data (no Fail-safe)
```

## Best Practices

1. **Use Transient for Ephemeral Data**: Development, testing, CI/CD, and staging databases should typically be transient.

2. **Secure Credentials**: Never hardcode passwords. Use Terraform variables, environment variables, or secrets management tools.

3. **Database-per-Environment**: Follow the industry pattern of separate databases for dev, test, and prod environments.

4. **Minimize Data Retention**: Set `data_retention_time_in_days` to the minimum required for your use case.

5. **Tagging and Labeling**: Use metadata labels for cost allocation and resource organization.

6. **Zero-Copy Cloning**: For creating development environments from production, consider using Snowflake's `CREATE DATABASE ... CLONE` feature (outside Terraform) which is instant and cost-effective.

## Validation

After deployment, verify the database configuration:

```sql
-- Connect to Snowflake and verify database
SHOW DATABASES LIKE 'YOUR_DATABASE_NAME';

-- Check database properties
SHOW PARAMETERS FOR DATABASE YOUR_DATABASE_NAME;

-- Verify transient status and retention
SELECT 
  database_name,
  is_transient,
  retention_time,
  comment
FROM information_schema.databases
WHERE database_name = 'YOUR_DATABASE_NAME';
```

## Troubleshooting

### Authentication Issues

If provider authentication fails:
- Verify account identifier format (include region if required)
- Check username and password are correct
- Ensure user has `CREATE DATABASE` privilege

### Permission Errors

```
Error: Insufficient privileges to operate on database
```

Solution: Grant appropriate role:
```sql
GRANT CREATE DATABASE ON ACCOUNT TO ROLE terraform_role;
```

### Parameter Validation Errors

Some parameters have specific valid values (e.g., `log_level`, `trace_level`). Refer to [Snowflake documentation](https://docs.snowflake.com/en/sql-reference/sql/create-database) for valid options.

## References

- [Snowflake CREATE DATABASE Documentation](https://docs.snowflake.com/en/sql-reference/sql/create-database)
- [Snowflake Time Travel Documentation](https://docs.snowflake.com/en/user-guide/data-time-travel)
- [Snowflake Terraform Provider](https://registry.terraform.io/providers/Snowflake-Labs/snowflake/latest/docs/resources/database)
- [Planton Architecture](../../../../../architecture/deployment-component.md)

## License

This module is part of Planton and follows the project's licensing terms.

## Support

For issues, questions, or contributions, please refer to the [main repository](https://github.com/plantonhq/planton).




