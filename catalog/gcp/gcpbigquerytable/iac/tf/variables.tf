variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpBigQueryTable specification"
  type = object({
    # The GCP project the table is created in. Accepts a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The dataset containing the table. Accepts a literal dataset ID or a
    # reference to a GcpBigQueryDataset resource. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    dataset_id = string

    # Unique identifier for the table within the dataset. Immutable.
    # Letters, numbers, and underscores; maximum 1024 characters.
    # Example: "events_raw", "revenue_summary"
    table_id = string

    # User-friendly display name for the table.
    friendly_name = optional(string, "")

    # Description of the table's contents or purpose.
    description = optional(string, "")

    # Labels applied to the table for cost attribution and organization.
    # Merged with Planton's platform labels (which win on key conflicts).
    labels = optional(map(string), {})

    # Resource Manager tags bound to the table, as
    # "tagKeys-namespaced-name" -> "tagValue short name" pairs. Unlike
    # labels, tags participate in IAM conditions and organization policy.
    resource_tags = optional(map(string), {})

    # Table schema as a JSON array of field definitions, e.g.
    # [{"name":"id","type":"INT64","mode":"REQUIRED"},
    #  {"name":"payload","type":"JSON"}]. Columns are add-only in place:
    # removing or retyping a column recreates the table. Omit for views
    # (defined by their query) and autodetected external tables.
    schema = optional(string, "")

    # Time-based partitioning. Mutually exclusive with range_partitioning.
    # The partitioning field is immutable.
    time_partitioning = optional(object({
      # Partition granularity: DAY, HOUR, MONTH, or YEAR. DAY is the common
      # default for event data; HOUR suits very high-volume streams.
      type = string

      # The DATE, TIMESTAMP, or DATETIME column to partition on. Immutable.
      # If omitted, the table is ingestion-time partitioned (the pseudo-columns
      # _PARTITIONTIME / _PARTITIONDATE carry the partition key).
      field = optional(string, "")

      # Number of milliseconds to keep each partition before it is dropped.
      # If not set (0), partitions never expire (the dataset's
      # default_partition_expiration_ms still applies to new tables).
      expiration_ms = optional(number, 0)
    }))

    # Integer-range partitioning. Mutually exclusive with time_partitioning.
    # The partitioning field is immutable.
    range_partitioning = optional(object({
      # The INTEGER column to partition on. Immutable.
      field = string

      # The partition key space (start / end / interval).
      range = object({
        # Start of range partitioning, inclusive.
        start = optional(number, 0)

        # End of range partitioning, exclusive. Must be greater than start.
        end = optional(number, 0)

        # Width of each partition interval. Must be at least 1.
        interval = optional(number, 0)
      })
    }))

    # Up to four columns to cluster by, in precedence order. Queries
    # filtering on the leading clustering columns scan less data.
    clustering = optional(list(string), [])

    # Require every query against the table to carry a partition filter
    # predicate — the cost guard for large partitioned tables.
    require_partition_filter = optional(bool, false)

    # Time when the table expires and is deleted, in milliseconds since
    # epoch. If not set (0), the table never expires (the dataset's
    # default_table_expiration_ms applied at creation still governs).
    expiration_time = optional(number, 0)

    # Maximum staleness tolerated when reading a BigLake table with metadata
    # caching (SQL interval string, e.g. "0-0 0 4:0:0" for 4 hours).
    max_staleness = optional(string, "")

    # Cloud KMS key encrypting the table (CMEK). Immutable — changing the
    # key recreates the table. Overrides the dataset's default key. The
    # BigQuery service agent must hold cryptoKeyEncrypterDecrypter on it.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Makes the table a logical view. Mutually exclusive with
    # materialized_view and external_data_configuration.
    view = optional(object({
      # The SQL query defining the view. Referenced tables must live in the
      # same location as the view's dataset.
      query = string

      # Whether the query uses BigQuery's legacy SQL dialect. Defaults to
      # false (GoogleSQL) — both engines always send the value explicitly, so
      # the BigQuery API's own legacy-SQL-by-default behavior for views never
      # silently applies.
      use_legacy_sql = optional(bool, false)
    }))

    # Makes the table a materialized view. Mutually exclusive with view and
    # external_data_configuration.
    materialized_view = optional(object({
      # The SQL query defining the materialized view. Immutable — changing the
      # query recreates the table.
      query = string

      # Whether BigQuery automatically refreshes the results when base tables
      # change. When omitted, GCP's default (enabled) applies; set false to
      # refresh only manually.
      enable_refresh = optional(bool)

      # Maximum staleness between refreshes, in milliseconds. If not set (0),
      # GCP's default of 1800000 (30 minutes) applies.
      refresh_interval_ms = optional(number, 0)

      # Allow a query shape BigQuery cannot refresh incrementally (full
      # re-computation on refresh). Immutable. Default false.
      allow_non_incremental_definition = optional(bool, false)
    }))

    # Makes the table external (data stays in GCS/Sheets/Bigtable/...).
    # Mutually exclusive with view and materialized_view.
    external_data_configuration = optional(object({
      # Let BigQuery infer the schema and options from the source. Both
      # engines always send this value explicitly (the API requires the field).
      autodetect = optional(bool, false)

      # Fully qualified source URIs, e.g. gs://bucket/path/*. Wildcards are
      # allowed after the bucket name (not for Bigtable or Google Drive).
      source_uris = list(string)

      # Format of the data: CSV, GOOGLE_SHEETS, NEWLINE_DELIMITED_JSON, AVRO,
      # ICEBERG, DATASTORE_BACKUP, PARQUET, ORC, BIGTABLE, or DELTA_LAKE.
      # Mutually exclusive with object_metadata (object tables).
      source_format = optional(string, "")

      # Set to SIMPLE to create an OBJECT table over unstructured GCS objects
      # (requires connection_id; mutually exclusive with source_format).
      object_metadata = optional(string, "")

      # Compression of the source data: NONE (default) or GZIP.
      compression = optional(string, "")

      # Explicit schema for the external data as a JSON array of field
      # definitions. Immutable. Mutually exclusive with autodetect in
      # practice; most formats are self-describing.
      schema = optional(string, "")

      # Tolerate rows with values that do not match the schema (extra columns
      # ignored, missing values become NULL).
      ignore_unknown_values = optional(bool, false)

      # Maximum number of bad records to tolerate before failing the query.
      max_bad_records = optional(number, 0)

      # Connection (projects/{p}/locations/{l}/connections/{c}) whose
      # credential reads the source — this is what upgrades a plain external
      # table to a BigLake table. Plain string until a connection kind exists.
      connection_id = optional(string, "")

      # Reference file providing the table schema for AVRO/PARQUET/ORC.
      reference_file_schema_uri = optional(string, "")

      # Metadata caching for BigLake tables: AUTOMATIC (refresh within
      # max_staleness) or MANUAL (on-demand refresh). Requires connection_id.
      metadata_cache_mode = optional(string, "")

      # How source URIs are interpreted: FILE_SYSTEM_MATCH (default; expand
      # via object listing) or NEW_LINE_DELIMITED_MANIFEST (URIs point to
      # manifest files, one data URI per line).
      file_set_spec_type = optional(string, "")

      # Parse extension for JSON data: GEOJSON (newline-delimited GeoJSON).
      json_extension = optional(string, "")

      # Format-specific parsing options — set the block matching source_format.
      csv_options = optional(object({
        # The value used to quote data sections. When omitted, the API default
        # double-quote (") applies; set an explicit empty string for unquoted
        # data. Presence-tracked because the empty string is meaningful.
        quote = optional(string)

        # Accept rows missing trailing optional columns (missing values become
        # NULL).
        allow_jagged_rows = optional(bool, false)

        # Allow quoted data sections containing newlines.
        allow_quoted_newlines = optional(bool, false)

        # Character encoding of the data: UTF-8 (default) or ISO-8859-1.
        encoding = optional(string, "")

        # Field separator. Defaults to comma.
        field_delimiter = optional(string, "")

        # Number of header rows to skip.
        skip_leading_rows = optional(number, 0)

        # How source columns map onto the table schema: POSITION (by ordering,
        # the safe choice for stable extracts) or NAME (reads the header row
        # and reorders to match schema field names — tolerates column
        # reshuffles, requires a header).
        source_column_match = optional(string, "")
      }))

      json_options = optional(object({
        # Character encoding of the data. Defaults to UTF-8.
        encoding = optional(string, "")
      }))
      google_sheets_options = optional(object({
        # Sheet range to query, e.g. "sheet1!A1:B20". When omitted the first
        # sheet is used.
        range = optional(string, "")

        # Number of header rows to skip.
        skip_leading_rows = optional(number, 0)
      }))
      hive_partitioning_options = optional(object({
        # Partition-key inference mode: AUTO (infer types), STRINGS (all keys as
        # strings), or CUSTOM (types encoded in source_uri_prefix).
        mode = optional(string, "")

        # Require a partition filter predicate in every query against the table.
        require_partition_filter = optional(bool, false)

        # Common prefix of all source URIs before partition-key encoding begins,
        # e.g. gs://bucket/path (or gs://bucket/path/{dt:DATE} for CUSTOM mode).
        source_uri_prefix = optional(string, "")
      }))
      avro_options = optional(object({
        # Interpret Avro logical types (timestamp-micros, decimal, ...) as their
        # corresponding BigQuery types instead of raw primitives. Both engines
        # always send this value explicitly.
        use_avro_logical_types = optional(bool, false)
      }))
      parquet_options = optional(object({
        # Infer Parquet ENUM logical type as STRING instead of BYTES.
        enum_as_string = optional(bool, false)

        # Infer Parquet LIST logical type as the list's element type instead of
        # a repeated record wrapper.
        enable_list_inference = optional(bool, false)
      }))
      bigtable_options = optional(object({
        # Skip unspecified column families instead of failing.
        ignore_unspecified_column_families = optional(bool, false)

        # Read the row key as a STRING instead of BYTES.
        read_rowkey_as_string = optional(bool, false)

        # Expose unlisted column families as a JSON-typed column each.
        output_column_families_as_json = optional(bool, false)

        # Column families to expose, with per-family and per-column typing.
        column_families = optional(list(object({
          # Identifier of the column family.
          family_id = optional(string, "")

          # Default type for values in this family (overridable per column).
          type = optional(string, "")

          # Default encoding for values in this family (overridable per column).
          encoding = optional(string, "")

          # Default only-latest-version posture for this family.
          only_read_latest = optional(bool, false)

          # Columns of the family to expose individually.
          columns = optional(list(object({
            # Qualifier of the column, base64-encoded (for qualifiers that are not
            # valid UTF-8). Exactly one of qualifier_encoded or qualifier_string.
            qualifier_encoded = optional(string, "")

            # Qualifier of the column as a UTF-8 string.
            qualifier_string = optional(string, "")

            # Field name to use in the table instead of the qualifier.
            field_name = optional(string, "")

            # Type to convert the value to: BYTES (default), STRING, INTEGER, FLOAT,
            # BOOLEAN, JSON.
            type = optional(string, "")

            # Encoding of the values: TEXT or BINARY (default).
            encoding = optional(string, "")

            # Expose only the latest cell version instead of all versions.
            only_read_latest = optional(bool, false)
          })), [])
        })), [])
      }))

      # Types decimal source values may convert to. The API picks the FIRST
      # type (in its fixed NUMERIC → BIGNUMERIC → STRING precedence, not this
      # list's order) that is listed here and fits the value's precision and
      # scale.
      decimal_target_types = optional(list(string), [])
    }))

    # Unenforced primary/foreign keys for the optimizer and lineage tools.
    table_constraints = optional(object({
      # Primary key of the table.
      primary_key = optional(object({
        # Columns composing the primary key.
        columns = list(string)
      }))

      # Foreign keys of the table.
      foreign_keys = optional(list(object({
        # Optional name of the constraint.
        name = optional(string, "")

        # The table this key references.
        referenced_table = object({
          # Project of the referenced table.
          project_id = string

          # Dataset of the referenced table.
          dataset_id = string

          # The referenced table. Accepts a literal table ID or a reference to a
          # GcpBigQueryTable resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          table_id = string
        })

        # The referencing/referenced column pair.
        column_references = object({
          # The column in this table.
          referencing_column = string

          # The column in the referenced table.
          referenced_column = string
        })
      })), [])
    }))

    # Makes the table a replica of a source materialized view. Immutable.
    table_replication_info = optional(object({
      # Project of the source materialized view.
      source_project_id = string

      # Dataset of the source materialized view.
      source_dataset_id = string

      # The source materialized view.
      source_table_id = string

      # Replication interval in milliseconds. If not set (0), GCP's default of
      # 300000 (5 minutes) applies.
      replication_interval_ms = optional(number, 0)
    }))

    # BigLake managed-table storage (Iceberg files in your GCS bucket).
    # Immutable.
    biglake_configuration = optional(object({
      # Connection (projects/{p}/locations/{l}/connections/{c}) whose
      # credential accesses the storage. Plain string until a connection kind
      # exists.
      connection_id = string

      # Fully qualified storage location prefix, e.g. gs://bucket/path.
      storage_uri = string

      # Open-source file format of the data. Currently PARQUET.
      file_format = string

      # Open-source table format managing the metadata. Currently ICEBERG.
      table_format = string
    }))

    # Marks the schema as expressed in a foreign type system (e.g. HIVE).
    # Immutable.
    schema_foreign_type_info = optional(object({
      # The type system of the schema, e.g. HIVE. Immutable.
      type_system = string
    }))

    # Hive-metastore-compatible metadata for open-source engines.
    external_catalog_table_options = optional(object({
      # Hive-table-style key/value parameters (the whole map is limited to
      # 30 KB by the API).
      parameters = optional(map(string), {})

      # Connection whose credential accesses the storage.
      connection_id = optional(string, "")

      # Physical storage description of the table.
      storage_descriptor = optional(object({
        # Physical location of the table, e.g. gs://bucket/path.
        location_uri = optional(string, "")

        # Fully qualified Hive input format class name.
        input_format = optional(string, "")

        # Fully qualified Hive output format class name.
        output_format = optional(string, "")

        # Serializer/deserializer information.
        serde_info = optional(object({
          # Optional name of the SerDe.
          name = optional(string, "")

          # Fully qualified SerDe class name, e.g.
          # org.apache.hadoop.hive.ql.io.orc.OrcSerde.
          serialization_library = string

          # SerDe key/value parameters.
          parameters = optional(map(string), {})
        }))
      }))
    }))

    # Prevents the table from being destroyed while true. Defaults to true:
    # a destroy fails until this is set to false — the guard for tables
    # holding real data. Set false for disposable/dev tables.
    deletion_protection = optional(bool)

    # What destroying this resource does to the table. Two levers guard a
    # table: this one decides whether removal is attempted at all, and
    # deletion_protection (checked second) blocks an attempted delete.
    # ABANDON therefore bypasses deletion_protection — the table leaves
    # management untouched — while DELETE still requires
    # deletion_protection=false to actually destroy data:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the table (and its data) is deleted, if
    #                deletion_protection allows it
    #   "PREVENT" -- destroy FAILS before deletion_protection is even
    #                consulted
    #   "ABANDON" -- the table is removed from management and keeps serving
    #                queries in GCP
    deletion_policy = optional(string, "")

    # Hide diffs from columns BigQuery adds on its own (e.g. columns
    # materialized by flexible-schema features) so they are not fought over
    # on every apply. Terraform-plan-level behavior; no API field.
    ignore_auto_generated_schema = optional(bool, false)

    # Schema sub-fields treated as non-authoritative per column — the
    # provider stops reconciling them. The provider currently supports
    # exactly "dataPolicies" (column-level data-policy attachments managed
    # outside this spec); other strings are accepted by the provider but
    # are inert.
    ignore_schema_changes = optional(list(string), [])

    # How much table metadata the provider requests when reading the table
    # back: BASIC (cheapest), STORAGE_STATS (the API default), or FULL.
    # A read-tuning knob for very large fleets; no effect on the table
    # itself.
    table_metadata_view = optional(string, "")
  })
}
