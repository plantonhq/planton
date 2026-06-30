variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name = string,
    id = optional(string),
    org = optional(string),
    env = optional(string),
    labels = optional(map(string)),
    tags = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}


variable "spec" {
  description = "spec"
  type = object({

    # The database parameter that specifies the default catalog to use for Iceberg tables
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#catalog_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters#catalog
    catalog = string

    # Specifies a comment for the database
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#comment_yaml
    comment = string

    # Specifies the number of days for which Time Travel actions (CLONE and UNDROP) can be performed on the database,
    # as well as specifying the default Time Travel retention time for all schemas created in the database.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#dataretentiontimeindays_yaml
    # https://docs.snowflake.com/en/user-guide/data-time-travel
    data_retention_time_in_days = number

    # Specifies a default collation specification for all schemas and tables added to the database.
    # It can be overridden on schema or table level.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#defaultddlcollation_yaml
    # https://docs.snowflake.com/en/sql-reference/collation#label-collation-specification
    default_ddl_collation = string

    # Specifies whether to drop public schema on creation or not. Modifying the parameter after database is
    # already created won't have any effect.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#droppublicschemaoncreation_yaml
    drop_public_schema_on_creation = bool

    # If true, enables stdout/stderr fast path logging for anonymous stored procedures.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#enableconsoleoutput_yaml
    enable_console_output = bool

    # The database parameter that specifies the default external volume to use for Iceberg tables
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#externalvolume_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters#external-volume
    external_volume = string

    # Specifies the database as transient. Transient databases do not have a Fail-safe period so they do not incur
    # additional storage costs once they leave Time Travel; however, this means they are also not protected by
    # Fail-safe in the event of a data loss.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#istransient_yaml
    is_transient = bool

    # Specifies the severity level of messages that should be ingested and made available in the active event table.
    # Valid options are: [TRACE DEBUG INFO WARN ERROR FATAL OFF]. Messages at the specified level (and at more severe levels) are ingested.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#loglevel_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters.html#label-log-level
    log_level = string

    # Object parameter that specifies the maximum number of days for which Snowflake can extend the data retention period
    # for tables in the database to prevent streams on the tables from becoming stale.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#maxdataextensiontimeindays_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters.html#label-max-data-extension-time-in-days
    max_data_extension_time_in_days = number

    # Specifies the identifier for the database; must be unique for your account. As a best practice for Database
    # Replication and Failover, it is recommended to give each secondary database the same name as its primary database.
    # This practice supports referencing fully-qualified objects (i.e. '\n\n.\n\n.\n\n') by other objects in the
    # same database, such as querying a fully-qualified table name in a view. If a secondary database has a
    # different name from the primary database, then these object references would break in the secondary database.
    # Due to technical limitations (read more here), avoid using the following characters: |, ., (, ), "
    name = string

    # If true, the case of quoted identifiers is ignored
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#quotedidentifiersignorecase_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters#quoted-identifiers-ignore-case
    quoted_identifiers_ignore_case = bool

    # Specifies whether to replace invalid UTF-8 characters with the Unicode replacement character (�) in query results
    # for an Iceberg table. You can only set this parameter for tables that use an external Iceberg catalog
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#replaceinvalidcharacters_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters#replace-invalid-characters
    replace_invalid_characters = bool

    # The storage serialization policy for Iceberg tables that use Snowflake as the catalog.
    # Valid options are: [COMPATIBLE OPTIMIZED]. COMPATIBLE: Snowflake performs encoding and compression of data
    # files that ensures interoperability with third-party compute engines. OPTIMIZED: Snowflake performs encoding and
    # compression of data files that ensures the best table performance within Snowflake.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#storageserializationpolicy_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters#storage-serialization-policy
    storage_serialization_policy = string

    # How many times a task must fail in a row before it is automatically suspended. 0 disables auto-suspending.
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#suspendtaskafternumfailures_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters#suspend-task-after-num-failures
    suspend_task_after_num_failures = number

    # Maximum automatic retries allowed for a user task
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#taskautoretryattempts_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters#task-auto-retry-attempts
    task_auto_retry_attempts = number

    # Controls how trace events are ingested into the event table. Valid options are: [ALWAYS ON*EVENT OFF]
    # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#tracelevel_yaml
    # https://docs.snowflake.com/en/sql-reference/parameters.html#label-trace-level
    trace_level = string

    # snowflake database user task
    user_task = object({

      # The initial size of warehouse to use for managed warehouses in the absence of history.
      # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#usertaskmanagedinitialwarehousesize_yaml
      # https://docs.snowflake.com/en/sql-reference/parameters#user-task-managed-initial-warehouse-size
      managed_initial_warehouse_size = string

      # Minimum amount of time between Triggered Task executions in seconds.
      # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#usertaskminimumtriggerintervalinseconds_yaml
      minimum_trigger_interval_in_seconds = number

      # User task execution timeout in milliseconds
      # https://www.pulumi.com/registry/packages/snowflake/api-docs/database/#usertasktimeoutms_yaml
      # https://docs.snowflake.com/en/sql-reference/parameters#user-task-timeout-ms
      timeout_ms = number
    })
  })
}