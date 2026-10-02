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
  description = "GcpDatastreamStream specification"
  type = object({
    # The GCP project the stream lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Datastream region, e.g. "us-central1" -- the profiles' region.
    # Immutable.
    location = string

    # The stream's ID. Defaults to metadata.name. Immutable.
    stream_id = optional(string, "")

    # The name shown in the console. Defaults to metadata.name.
    display_name = optional(string, "")

    # Labels on the stream. The platform attribution labels are added on top
    # and win on a key conflict.
    labels = optional(map(string), {})

    # The source and how it is read.
    source_config = object({
      # The source profile -- a GcpDatastreamConnectionProfile reference (its
      # name output) or a literal
      # projects/{project}/locations/{location}/connectionProfiles/{id} in the
      # stream's location. Salesforce and Spanner profiles are made outside
      # the catalog and named literally. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_connection_profile = string

      # A MySQL source.
      mysql_source_config = optional(object({
        # What to replicate; omitted means every database the user can read.
        include_objects = optional(object({
          # The databases in the set -- at least one.
          mysql_databases = list(object({
            # The database's name.
            database = string

            # The tables to narrow to; empty selects every table.
            mysql_tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              mysql_columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's collation, e.g. "utf8mb4_general_ci".
                collation = optional(string, "")

                # The column's MySQL data type, e.g. "VARCHAR".
                data_type = optional(string, "")

                # Whether the column is nullable.
                nullable = optional(bool, false)

                # The column's position in the table.
                ordinal_position = optional(number, 0)

                # Whether the column is part of the primary key.
                primary_key = optional(bool, false)
              })), [])
            })), [])
          }))
        }))

        # What to skip.
        exclude_objects = optional(object({
          # The databases in the set -- at least one.
          mysql_databases = list(object({
            # The database's name.
            database = string

            # The tables to narrow to; empty selects every table.
            mysql_tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              mysql_columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's collation, e.g. "utf8mb4_general_ci".
                collation = optional(string, "")

                # The column's MySQL data type, e.g. "VARCHAR".
                data_type = optional(string, "")

                # Whether the column is nullable.
                nullable = optional(bool, false)

                # The column's position in the table.
                ordinal_position = optional(number, 0)

                # Whether the column is part of the primary key.
                primary_key = optional(bool, false)
              })), [])
            })), [])
          }))
        }))

        # How many tables backfill at once. Empty (or 0) uses Google's default.
        max_concurrent_backfill_tasks = optional(number, 0)

        # How many change-capture tasks run at once. Empty (or 0) uses Google's
        # default.
        max_concurrent_cdc_tasks = optional(number, 0)

        # How changes are read (Google's cdc_method union):
        #   "GTID"                -- global transaction IDs; survives a failover
        #                            to a replica (needs gtid_mode=ON)
        #   "BINARY_LOG_POSITION" -- binary log file and position
        # Empty leaves Google's choice (binary log position).
        cdc_method = optional(string, "")
      }))

      # A PostgreSQL source.
      postgresql_source_config = optional(object({
        # What to replicate; omitted means every table in the publication.
        include_objects = optional(object({
          # The schemas in the set -- at least one.
          postgresql_schemas = list(object({
            # The schema's name, e.g. "public".
            schema = string

            # The tables to narrow to; empty selects every table.
            postgresql_tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              postgresql_columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's PostgreSQL data type, e.g. "TEXT".
                data_type = optional(string, "")

                # Whether the column is nullable.
                nullable = optional(bool, false)

                # The column's position in the table.
                ordinal_position = optional(number, 0)

                # Whether the column is part of the primary key.
                primary_key = optional(bool, false)
              })), [])
            })), [])
          }))
        }))

        # What to skip.
        exclude_objects = optional(object({
          # The schemas in the set -- at least one.
          postgresql_schemas = list(object({
            # The schema's name, e.g. "public".
            schema = string

            # The tables to narrow to; empty selects every table.
            postgresql_tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              postgresql_columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's PostgreSQL data type, e.g. "TEXT".
                data_type = optional(string, "")

                # Whether the column is nullable.
                nullable = optional(bool, false)

                # The column's position in the table.
                ordinal_position = optional(number, 0)

                # Whether the column is part of the primary key.
                primary_key = optional(bool, false)
              })), [])
            })), [])
          }))
        }))

        # The logical replication slot Datastream consumes, created beforehand
        # with the pgoutput plugin. One slot per stream; an idle slot makes the
        # server retain WAL, so delete a stream's slot with the stream.
        replication_slot = string

        # The publication naming the tables the stream may read, created
        # beforehand (CREATE PUBLICATION ... FOR ALL TABLES or FOR TABLE ...).
        publication = string

        # How many tables backfill at once. Empty (or 0) uses Google's default.
        max_concurrent_backfill_tasks = optional(number, 0)
      }))

      # An Oracle source.
      oracle_source_config = optional(object({
        # What to replicate; omitted means every schema the user can read.
        include_objects = optional(object({
          # The schemas in the set -- at least one.
          oracle_schemas = list(object({
            # The schema's name.
            schema = string

            # The tables to narrow to; empty selects every table.
            oracle_tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              oracle_columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's Oracle data type, e.g. "VARCHAR2".
                data_type = optional(string, "")
              })), [])
            })), [])
          }))
        }))

        # What to skip.
        exclude_objects = optional(object({
          # The schemas in the set -- at least one.
          oracle_schemas = list(object({
            # The schema's name.
            schema = string

            # The tables to narrow to; empty selects every table.
            oracle_tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              oracle_columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's Oracle data type, e.g. "VARCHAR2".
                data_type = optional(string, "")
              })), [])
            })), [])
          }))
        }))

        # How many tables backfill at once. Empty (or 0) uses Google's default.
        max_concurrent_backfill_tasks = optional(number, 0)

        # How many change-capture tasks run at once. Empty (or 0) uses Google's
        # default.
        max_concurrent_cdc_tasks = optional(number, 0)

        # What happens to LOB values -- CLOB, BLOB, NCLOB (Google's
        # large_objects_handling union):
        #   "DROP"   -- the values are dropped; rows replicate without them
        #   "STREAM" -- the values are streamed with the row
        # Empty leaves Google's default.
        large_objects_handling = optional(string, "")
      }))

      # A SQL Server source.
      sql_server_source_config = optional(object({
        # What to replicate; omitted means every schema the user can read.
        include_objects = optional(object({
          # The schemas in the set -- at least one.
          schemas = list(object({
            # The schema's name, e.g. "dbo".
            schema = string

            # The tables to narrow to; empty selects every table.
            tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's SQL Server data type, e.g. "NVARCHAR".
                data_type = optional(string, "")
              })), [])
            })), [])
          }))
        }))

        # What to skip.
        exclude_objects = optional(object({
          # The schemas in the set -- at least one.
          schemas = list(object({
            # The schema's name, e.g. "dbo".
            schema = string

            # The tables to narrow to; empty selects every table.
            tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              columns = optional(list(object({
                # The column's name.
                column = optional(string, "")

                # The column's SQL Server data type, e.g. "NVARCHAR".
                data_type = optional(string, "")
              })), [])
            })), [])
          }))
        }))

        # How many tables backfill at once. Empty (or 0) uses Google's default.
        max_concurrent_backfill_tasks = optional(number, 0)

        # How many change-capture tasks run at once. Empty (or 0) uses Google's
        # default.
        max_concurrent_cdc_tasks = optional(number, 0)

        # How changes are read (Google's cdc_method union):
        #   "CHANGE_TABLES"    -- SQL Server's CDC change tables (CDC enabled on
        #                         the database and each table)
        #   "TRANSACTION_LOGS" -- the transaction log directly; lighter on the
        #                         source, needs log access and retention
        # Empty leaves Google's default.
        cdc_method = optional(string, "")
      }))

      # A MongoDB source.
      mongodb_source_config = optional(object({
        # What to replicate; omitted means every database the user can read.
        include_objects = optional(object({
          # The databases in the set (at least one inside a backfill exclusion).
          databases = optional(list(object({
            # The database's name (required inside a backfill exclusion).
            database = optional(string, "")

            # The collections to narrow to; empty selects every collection.
            collections = optional(list(object({
              # The collection's name (required inside a backfill exclusion).
              collection = optional(string, "")

              # The fields to narrow to; empty selects every field.
              fields = optional(list(object({
                # The field's name.
                field = optional(string, "")
              })), [])
            })), [])
          })), [])
        }))

        # What to skip.
        exclude_objects = optional(object({
          # The databases in the set (at least one inside a backfill exclusion).
          databases = optional(list(object({
            # The database's name (required inside a backfill exclusion).
            database = optional(string, "")

            # The collections to narrow to; empty selects every collection.
            collections = optional(list(object({
              # The collection's name (required inside a backfill exclusion).
              collection = optional(string, "")

              # The fields to narrow to; empty selects every field.
              fields = optional(list(object({
                # The field's name.
                field = optional(string, "")
              })), [])
            })), [])
          })), [])
        }))

        # How many collections backfill at once, 0-50. Empty (or 0) uses
        # Google's default.
        max_concurrent_backfill_tasks = optional(number, 0)
      }))

      # A Salesforce source.
      salesforce_source_config = optional(object({
        # What to replicate; omitted means every object the integration user can
        # read.
        include_objects = optional(object({
          # The objects in the set -- at least one.
          objects = list(object({
            # The object's API name, e.g. "Account".
            object_name = optional(string, "")

            # The fields to narrow to; empty selects every field.
            fields = optional(list(object({
              # The field's API name, e.g. "AnnualRevenue".
              name = optional(string, "")
            })), [])
          }))
        }))

        # What to skip.
        exclude_objects = optional(object({
          # The objects in the set -- at least one.
          objects = list(object({
            # The object's API name, e.g. "Account".
            object_name = optional(string, "")

            # The fields to narrow to; empty selects every field.
            fields = optional(list(object({
              # The field's API name, e.g. "AnnualRevenue".
              name = optional(string, "")
            })), [])
          }))
        }))

        # How often each object is polled for changes, as a duration in seconds
        # -- Google allows 5 minutes to 24 hours ("300s" to "86400s"). Shorter
        # means fresher data and more Salesforce API calls against the org's
        # daily limit.
        polling_interval = string
      }))

      # A Spanner source.
      spanner_source_config = optional(object({
        # What to replicate; omitted means every table the change stream
        # watches.
        include_objects = optional(object({
          # The schemas in the set -- at least one.
          schemas = list(object({
            # The schema's name.
            schema = string

            # The tables to narrow to; empty selects every table.
            tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              columns = optional(list(object({
                # The column's name (required inside a backfill exclusion).
                column = optional(string, "")
              })), [])
            })), [])
          }))
        }))

        # What to skip.
        exclude_objects = optional(object({
          # The schemas in the set -- at least one.
          schemas = list(object({
            # The schema's name.
            schema = string

            # The tables to narrow to; empty selects every table.
            tables = optional(list(object({
              # The table's name.
              table = string

              # The columns to narrow to; empty selects every column.
              columns = optional(list(object({
                # The column's name (required inside a backfill exclusion).
                column = optional(string, "")
              })), [])
            })), [])
          }))
        }))

        # The Spanner change stream Datastream reads, created beforehand (CREATE
        # CHANGE STREAM ... FOR ALL). Google requires it. Immutable.
        change_stream_name = string

        # Run the backfill's reads on Spanner Data Boost -- independent compute
        # that leaves the instance's serving capacity untouched (billed by
        # Spanner). Off by default.
        backfill_data_boost_enabled = optional(bool, false)

        # A fine-grained access control role the reads run as. Empty reads with
        # the service agent's database-level permissions.
        fgac_role = optional(string, "")

        # How many tables backfill at once. Empty (or 0) uses Google's default.
        max_concurrent_backfill_tasks = optional(number, 0)

        # How many change-capture tasks run at once. Empty (or 0) uses Google's
        # default.
        max_concurrent_cdc_tasks = optional(number, 0)

        # The priority of Datastream's Spanner reads: "LOW", "MEDIUM", or
        # "HIGH". Lower yields to the application's traffic. Empty leaves
        # Google's default.
        spanner_rpc_priority = optional(string, "")
      }))
    })

    # The destination and how data lands.
    destination_config = object({
      # The destination profile -- a GcpDatastreamConnectionProfile reference
      # (its name output) or a literal full profile name in the stream's
      # location. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination_connection_profile = string

      # Write into BigQuery (needs a bigquery_profile).
      bigquery_destination_config = optional(object({
        # Every table into one dataset.
        single_target_dataset = optional(object({
          # The dataset -- a GcpBigQueryDataset reference (its self_link output,
          # which the modules trim to Google's projects/{project}/datasets/{dataset}
          # form) or a literal in that form or {project}:{dataset}. Tables are
          # named {schema}_{table}.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          dataset_id = string
        }))

        # One dataset per source schema or database, created by Datastream.
        source_hierarchy_datasets = optional(object({
          # The template for the created datasets.
          dataset_template = object({
            # Where the datasets live, e.g. "US" or "us-central1".
            location = string

            # A prefix for every created dataset's name, joined with an underscore:
            # prefix "crm" and schema "public" make "crm_public".
            dataset_id_prefix = optional(string, "")

            # A Cloud KMS key the created datasets use by default (CMEK) -- a
            # GcpKmsKey reference or a literal key path, in the datasets' location.
            # Datastream's service agent needs cryptoKeyEncrypterDecrypter on it.
            # Immutable.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            kms_key_name = optional(string, "")
          })

          # The project the datasets are created in -- a literal project ID or a
          # GcpProject reference. Empty uses the stream's project.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          project_id = optional(string, "")
        }))

        # How stale a query may read (merge mode), as a duration, e.g. "900s".
        # Lower is fresher and costs more BigQuery compute. Changing it affects
        # only tables created afterwards. Empty leaves Google's default.
        data_freshness = optional(string, "")

        # How changes land (Google's write_mode union):
        #   "MERGE"       -- tables mirror the source's current state; changes
        #                    are merged by primary key (tables need one)
        #   "APPEND_ONLY" -- every change is appended as a row with its change
        #                    type; the full history, no merging
        # Empty leaves Google's default (merge). Immutable.
        write_mode = optional(string, "")

        # Write BigLake managed (Iceberg) tables instead of native tables.
        blmt_config = optional(object({
          # The bucket holding the table data -- a GcpGcsBucket reference (its
          # bucket_name output) or a literal.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          bucket = string

          # The folder inside the bucket.
          root_path = optional(string, "")

          # The BigQuery connection whose service account writes the bucket -- a
          # GcpBigQueryConnection reference (its name output, which the modules
          # convert to Google's {project}.{location}.{connection_id} form) or a
          # literal in the dotted form. Use a cloud_resource connection and grant
          # its service account storage access on the bucket.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          connection_name = string

          # The data file format. Google offers "PARQUET".
          file_format = string

          # The table format. Google offers "ICEBERG".
          table_format = string
        }))
      }))

      # Write files into Cloud Storage (needs a gcs_profile).
      gcs_destination_config = optional(object({
        # The folder under the profile's root path, e.g. "/orders".
        path = optional(string, "")

        # The longest a file stays open before a new one starts -- Google allows
        # "15s" to "60s". Empty leaves Google's default.
        file_rotation_interval = optional(string, "")

        # The largest a file grows, in MB, before a new one starts. Empty (or 0)
        # leaves Google's default.
        file_rotation_mb = optional(number, 0)

        # Write Avro files. Google's Avro block has no settings: true declares
        # it.
        avro_file_format = optional(bool, false)

        # Write JSON files.
        json_file_format = optional(object({
          # "NO_COMPRESSION" or "GZIP". Empty leaves Google's default.
          compression = optional(string, "")

          # Whether an Avro schema file is written beside the data files:
          # "NO_SCHEMA_FILE" or "AVRO_SCHEMA_FILE". Empty leaves Google's default.
          schema_file_format = optional(string, "")
        }))
      }))
    })

    # Backfill every included object, except what is listed.
    backfill_all = optional(object({
      # MySQL objects to skip.
      mysql_excluded_objects = optional(object({
        # The databases in the set -- at least one.
        mysql_databases = list(object({
          # The database's name.
          database = string

          # The tables to narrow to; empty selects every table.
          mysql_tables = optional(list(object({
            # The table's name.
            table = string

            # The columns to narrow to; empty selects every column.
            mysql_columns = optional(list(object({
              # The column's name.
              column = optional(string, "")

              # The column's collation, e.g. "utf8mb4_general_ci".
              collation = optional(string, "")

              # The column's MySQL data type, e.g. "VARCHAR".
              data_type = optional(string, "")

              # Whether the column is nullable.
              nullable = optional(bool, false)

              # The column's position in the table.
              ordinal_position = optional(number, 0)

              # Whether the column is part of the primary key.
              primary_key = optional(bool, false)
            })), [])
          })), [])
        }))
      }))

      # PostgreSQL objects to skip.
      postgresql_excluded_objects = optional(object({
        # The schemas in the set -- at least one.
        postgresql_schemas = list(object({
          # The schema's name, e.g. "public".
          schema = string

          # The tables to narrow to; empty selects every table.
          postgresql_tables = optional(list(object({
            # The table's name.
            table = string

            # The columns to narrow to; empty selects every column.
            postgresql_columns = optional(list(object({
              # The column's name.
              column = optional(string, "")

              # The column's PostgreSQL data type, e.g. "TEXT".
              data_type = optional(string, "")

              # Whether the column is nullable.
              nullable = optional(bool, false)

              # The column's position in the table.
              ordinal_position = optional(number, 0)

              # Whether the column is part of the primary key.
              primary_key = optional(bool, false)
            })), [])
          })), [])
        }))
      }))

      # Oracle objects to skip.
      oracle_excluded_objects = optional(object({
        # The schemas in the set -- at least one.
        oracle_schemas = list(object({
          # The schema's name.
          schema = string

          # The tables to narrow to; empty selects every table.
          oracle_tables = optional(list(object({
            # The table's name.
            table = string

            # The columns to narrow to; empty selects every column.
            oracle_columns = optional(list(object({
              # The column's name.
              column = optional(string, "")

              # The column's Oracle data type, e.g. "VARCHAR2".
              data_type = optional(string, "")
            })), [])
          })), [])
        }))
      }))

      # SQL Server objects to skip.
      sql_server_excluded_objects = optional(object({
        # The schemas in the set -- at least one.
        schemas = list(object({
          # The schema's name, e.g. "dbo".
          schema = string

          # The tables to narrow to; empty selects every table.
          tables = optional(list(object({
            # The table's name.
            table = string

            # The columns to narrow to; empty selects every column.
            columns = optional(list(object({
              # The column's name.
              column = optional(string, "")

              # The column's SQL Server data type, e.g. "NVARCHAR".
              data_type = optional(string, "")
            })), [])
          })), [])
        }))
      }))

      # MongoDB objects to skip. Here every database and collection needs its
      # name, and at least one database is listed.
      mongodb_excluded_objects = optional(object({
        # The databases in the set (at least one inside a backfill exclusion).
        databases = optional(list(object({
          # The database's name (required inside a backfill exclusion).
          database = optional(string, "")

          # The collections to narrow to; empty selects every collection.
          collections = optional(list(object({
            # The collection's name (required inside a backfill exclusion).
            collection = optional(string, "")

            # The fields to narrow to; empty selects every field.
            fields = optional(list(object({
              # The field's name.
              field = optional(string, "")
            })), [])
          })), [])
        })), [])
      }))

      # Salesforce objects to skip.
      salesforce_excluded_objects = optional(object({
        # The objects in the set -- at least one.
        objects = list(object({
          # The object's API name, e.g. "Account".
          object_name = optional(string, "")

          # The fields to narrow to; empty selects every field.
          fields = optional(list(object({
            # The field's API name, e.g. "AnnualRevenue".
            name = optional(string, "")
          })), [])
        }))
      }))

      # Spanner objects to skip. Here every column needs its name.
      spanner_excluded_objects = optional(object({
        # The schemas in the set -- at least one.
        schemas = list(object({
          # The schema's name.
          schema = string

          # The tables to narrow to; empty selects every table.
          tables = optional(list(object({
            # The table's name.
            table = string

            # The columns to narrow to; empty selects every column.
            columns = optional(list(object({
              # The column's name (required inside a backfill exclusion).
              column = optional(string, "")
            })), [])
          })), [])
        }))
      }))
    }))

    # Backfill nothing: only changes from the stream's start replicate.
    # Google's block has no settings: true declares it.
    backfill_none = optional(bool, false)

    # Per-object BigQuery table customizations.
    rule_sets = optional(list(object({
      # The source object the rules apply to.
      object_filter = object({
        # The object.
        source_object_identifier = optional(object({
          # A MySQL table.
          mysql_identifier = optional(object({
            # The database.
            database = string

            # The table.
            table = string
          }))

          # A PostgreSQL table.
          postgresql_identifier = optional(object({
            # The schema.
            schema = string

            # The table.
            table = string
          }))

          # An Oracle table.
          oracle_identifier = optional(object({
            # The schema.
            schema = string

            # The table.
            table = string
          }))

          # A SQL Server table.
          sql_server_identifier = optional(object({
            # The schema.
            schema = string

            # The table.
            table = string
          }))

          # A MongoDB collection.
          mongodb_identifier = optional(object({
            # The database.
            database = string

            # The collection.
            collection = string
          }))

          # A Salesforce object.
          salesforce_identifier = optional(object({
            # The object's API name.
            object_name = string
          }))

          # A Spanner table.
          spanner_identifier = optional(object({
            # The schema; empty is the database's default schema.
            schema = optional(string, "")

            # The table.
            table = string
          }))
        }))
      })

      # The rules -- at least one.
      customization_rules = list(object({
        # Partitioning for the matched tables.
        bigquery_partitioning = optional(object({
          # Partition by arrival time.
          ingestion_time_partition = optional(object({
            # "PARTITIONING_TIME_GRANULARITY_HOUR", "..._DAY", "..._MONTH", or
            # "..._YEAR". Empty leaves Google's default.
            partitioning_time_granularity = optional(string, "")
          }))

          # Partition by a date or timestamp column.
          time_unit_partition = optional(object({
            # The partitioning column.
            column = string

            # "PARTITIONING_TIME_GRANULARITY_HOUR", "..._DAY", "..._MONTH", or
            # "..._YEAR". Empty leaves Google's default.
            partitioning_time_granularity = optional(string, "")
          }))

          # Partition by integer ranges.
          integer_range_partition = optional(object({
            # The partitioning column.
            column = string

            # The first range's start (inclusive).
            start = optional(number, 0)

            # The last range's end (exclusive).
            end = optional(number, 0)

            # Each range's width.
            interval = optional(number, 0)
          }))

          # Refuse queries over the tables that do not filter on the partition --
          # protects against full-table scans.
          require_partition_filter = optional(bool, false)
        }))

        # Clustering for the matched tables.
        bigquery_clustering = optional(object({
          # The clustering columns, most selective first (up to four in
          # BigQuery).
          columns = list(string)
        }))
      }))
    })), [])

    # A Cloud KMS key encrypting the data Datastream holds in flight (CMEK)
    # -- a GcpKmsKey reference or a literal key path in the stream's region.
    # Datastream's service agent needs cryptoKeyEncrypterDecrypter on it.
    # Empty uses a Google-managed key. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    customer_managed_encryption_key = optional(string, "")

    # Whether the stream runs:
    #   "" / "NOT_STARTED" -- created but not started; nothing moves or bills
    #   "RUNNING"          -- backfill (if chosen) and change capture run
    #   "PAUSED"           -- stopped, keeping its position; RUNNING resumes
    # Google accepts NOT_STARTED or RUNNING at create, then RUNNING or PAUSED;
    # a started stream never returns to NOT_STARTED.
    desired_state = optional(string, "")

    # Create the stream without Google's validation (source reachable,
    # objects exist, destination writable). Problems then surface when it
    # starts. Immutable.
    create_without_validation = optional(bool, false)

    # What happens to the stream when this resource is destroyed:
    #   "" / "DELETE" -- deleted (data already written stays in the
    #                    destination)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and keeps running in GCP
    deletion_policy = optional(string, "")
  })
}
