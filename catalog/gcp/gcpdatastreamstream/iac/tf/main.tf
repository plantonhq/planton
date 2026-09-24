# The stream. Exactly one source arm, one destination arm, and one backfill
# mode are declared (the spec's rules). Empty optional strings, zero counts,
# and false flags are sent as null so Google's defaults apply; the object
# lists mirror the provider's blocks level for level.
resource "google_datastream_stream" "this" {
  project      = local.project_id
  location     = var.spec.location
  stream_id    = local.stream_id
  display_name = local.display_name

  create_without_validation       = var.spec.create_without_validation
  customer_managed_encryption_key = local.customer_managed_encryption_key
  desired_state                   = local.desired_state

  source_config {
    source_connection_profile = local.source.source_connection_profile

    dynamic "mysql_source_config" {
      for_each = local.source.mysql_source_config != null ? [local.source.mysql_source_config] : []
      content {
        max_concurrent_backfill_tasks = mysql_source_config.value.max_concurrent_backfill_tasks > 0 ? mysql_source_config.value.max_concurrent_backfill_tasks : null
        max_concurrent_cdc_tasks      = mysql_source_config.value.max_concurrent_cdc_tasks > 0 ? mysql_source_config.value.max_concurrent_cdc_tasks : null

        # The spec's cdc_method selects which of Google's two empty blocks
        # is sent.
        dynamic "gtid" {
          for_each = mysql_source_config.value.cdc_method == "GTID" ? [true] : []
          content {}
        }

        dynamic "binary_log_position" {
          for_each = mysql_source_config.value.cdc_method == "BINARY_LOG_POSITION" ? [true] : []
          content {}
        }

        dynamic "include_objects" {
          for_each = mysql_source_config.value.include_objects != null ? [mysql_source_config.value.include_objects] : []
          content {
            dynamic "mysql_databases" {
              for_each = include_objects.value.mysql_databases
              content {
                database = mysql_databases.value.database

                dynamic "mysql_tables" {
                  for_each = mysql_databases.value.mysql_tables
                  content {
                    table = mysql_tables.value.table

                    dynamic "mysql_columns" {
                      for_each = mysql_tables.value.mysql_columns
                      content {
                        column           = mysql_columns.value.column != "" ? mysql_columns.value.column : null
                        collation        = mysql_columns.value.collation != "" ? mysql_columns.value.collation : null
                        data_type        = mysql_columns.value.data_type != "" ? mysql_columns.value.data_type : null
                        nullable         = mysql_columns.value.nullable ? true : null
                        primary_key      = mysql_columns.value.primary_key ? true : null
                        ordinal_position = mysql_columns.value.ordinal_position > 0 ? mysql_columns.value.ordinal_position : null
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "exclude_objects" {
          for_each = mysql_source_config.value.exclude_objects != null ? [mysql_source_config.value.exclude_objects] : []
          content {
            dynamic "mysql_databases" {
              for_each = exclude_objects.value.mysql_databases
              content {
                database = mysql_databases.value.database

                dynamic "mysql_tables" {
                  for_each = mysql_databases.value.mysql_tables
                  content {
                    table = mysql_tables.value.table

                    dynamic "mysql_columns" {
                      for_each = mysql_tables.value.mysql_columns
                      content {
                        column           = mysql_columns.value.column != "" ? mysql_columns.value.column : null
                        collation        = mysql_columns.value.collation != "" ? mysql_columns.value.collation : null
                        data_type        = mysql_columns.value.data_type != "" ? mysql_columns.value.data_type : null
                        nullable         = mysql_columns.value.nullable ? true : null
                        primary_key      = mysql_columns.value.primary_key ? true : null
                        ordinal_position = mysql_columns.value.ordinal_position > 0 ? mysql_columns.value.ordinal_position : null
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    dynamic "postgresql_source_config" {
      for_each = local.source.postgresql_source_config != null ? [local.source.postgresql_source_config] : []
      content {
        replication_slot              = postgresql_source_config.value.replication_slot
        publication                   = postgresql_source_config.value.publication
        max_concurrent_backfill_tasks = postgresql_source_config.value.max_concurrent_backfill_tasks > 0 ? postgresql_source_config.value.max_concurrent_backfill_tasks : null

        dynamic "include_objects" {
          for_each = postgresql_source_config.value.include_objects != null ? [postgresql_source_config.value.include_objects] : []
          content {
            dynamic "postgresql_schemas" {
              for_each = include_objects.value.postgresql_schemas
              content {
                schema = postgresql_schemas.value.schema

                dynamic "postgresql_tables" {
                  for_each = postgresql_schemas.value.postgresql_tables
                  content {
                    table = postgresql_tables.value.table

                    dynamic "postgresql_columns" {
                      for_each = postgresql_tables.value.postgresql_columns
                      content {
                        column           = postgresql_columns.value.column != "" ? postgresql_columns.value.column : null
                        data_type        = postgresql_columns.value.data_type != "" ? postgresql_columns.value.data_type : null
                        nullable         = postgresql_columns.value.nullable ? true : null
                        primary_key      = postgresql_columns.value.primary_key ? true : null
                        ordinal_position = postgresql_columns.value.ordinal_position > 0 ? postgresql_columns.value.ordinal_position : null
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "exclude_objects" {
          for_each = postgresql_source_config.value.exclude_objects != null ? [postgresql_source_config.value.exclude_objects] : []
          content {
            dynamic "postgresql_schemas" {
              for_each = exclude_objects.value.postgresql_schemas
              content {
                schema = postgresql_schemas.value.schema

                dynamic "postgresql_tables" {
                  for_each = postgresql_schemas.value.postgresql_tables
                  content {
                    table = postgresql_tables.value.table

                    dynamic "postgresql_columns" {
                      for_each = postgresql_tables.value.postgresql_columns
                      content {
                        column           = postgresql_columns.value.column != "" ? postgresql_columns.value.column : null
                        data_type        = postgresql_columns.value.data_type != "" ? postgresql_columns.value.data_type : null
                        nullable         = postgresql_columns.value.nullable ? true : null
                        primary_key      = postgresql_columns.value.primary_key ? true : null
                        ordinal_position = postgresql_columns.value.ordinal_position > 0 ? postgresql_columns.value.ordinal_position : null
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    dynamic "oracle_source_config" {
      for_each = local.source.oracle_source_config != null ? [local.source.oracle_source_config] : []
      content {
        max_concurrent_backfill_tasks = oracle_source_config.value.max_concurrent_backfill_tasks > 0 ? oracle_source_config.value.max_concurrent_backfill_tasks : null
        max_concurrent_cdc_tasks      = oracle_source_config.value.max_concurrent_cdc_tasks > 0 ? oracle_source_config.value.max_concurrent_cdc_tasks : null

        # The spec's large_objects_handling selects which of Google's two
        # empty blocks is sent.
        dynamic "drop_large_objects" {
          for_each = oracle_source_config.value.large_objects_handling == "DROP" ? [true] : []
          content {}
        }

        dynamic "stream_large_objects" {
          for_each = oracle_source_config.value.large_objects_handling == "STREAM" ? [true] : []
          content {}
        }

        dynamic "include_objects" {
          for_each = oracle_source_config.value.include_objects != null ? [oracle_source_config.value.include_objects] : []
          content {
            dynamic "oracle_schemas" {
              for_each = include_objects.value.oracle_schemas
              content {
                schema = oracle_schemas.value.schema

                dynamic "oracle_tables" {
                  for_each = oracle_schemas.value.oracle_tables
                  content {
                    table = oracle_tables.value.table

                    dynamic "oracle_columns" {
                      for_each = oracle_tables.value.oracle_columns
                      content {
                        column    = oracle_columns.value.column != "" ? oracle_columns.value.column : null
                        data_type = oracle_columns.value.data_type != "" ? oracle_columns.value.data_type : null
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "exclude_objects" {
          for_each = oracle_source_config.value.exclude_objects != null ? [oracle_source_config.value.exclude_objects] : []
          content {
            dynamic "oracle_schemas" {
              for_each = exclude_objects.value.oracle_schemas
              content {
                schema = oracle_schemas.value.schema

                dynamic "oracle_tables" {
                  for_each = oracle_schemas.value.oracle_tables
                  content {
                    table = oracle_tables.value.table

                    dynamic "oracle_columns" {
                      for_each = oracle_tables.value.oracle_columns
                      content {
                        column    = oracle_columns.value.column != "" ? oracle_columns.value.column : null
                        data_type = oracle_columns.value.data_type != "" ? oracle_columns.value.data_type : null
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    dynamic "sql_server_source_config" {
      for_each = local.source.sql_server_source_config != null ? [local.source.sql_server_source_config] : []
      content {
        max_concurrent_backfill_tasks = sql_server_source_config.value.max_concurrent_backfill_tasks > 0 ? sql_server_source_config.value.max_concurrent_backfill_tasks : null
        max_concurrent_cdc_tasks      = sql_server_source_config.value.max_concurrent_cdc_tasks > 0 ? sql_server_source_config.value.max_concurrent_cdc_tasks : null

        # The spec's cdc_method selects which of Google's two empty blocks
        # is sent.
        dynamic "change_tables" {
          for_each = sql_server_source_config.value.cdc_method == "CHANGE_TABLES" ? [true] : []
          content {}
        }

        dynamic "transaction_logs" {
          for_each = sql_server_source_config.value.cdc_method == "TRANSACTION_LOGS" ? [true] : []
          content {}
        }

        dynamic "include_objects" {
          for_each = sql_server_source_config.value.include_objects != null ? [sql_server_source_config.value.include_objects] : []
          content {
            dynamic "schemas" {
              for_each = include_objects.value.schemas
              content {
                schema = schemas.value.schema

                dynamic "tables" {
                  for_each = schemas.value.tables
                  content {
                    table = tables.value.table

                    dynamic "columns" {
                      for_each = tables.value.columns
                      content {
                        column    = columns.value.column != "" ? columns.value.column : null
                        data_type = columns.value.data_type != "" ? columns.value.data_type : null
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "exclude_objects" {
          for_each = sql_server_source_config.value.exclude_objects != null ? [sql_server_source_config.value.exclude_objects] : []
          content {
            dynamic "schemas" {
              for_each = exclude_objects.value.schemas
              content {
                schema = schemas.value.schema

                dynamic "tables" {
                  for_each = schemas.value.tables
                  content {
                    table = tables.value.table

                    dynamic "columns" {
                      for_each = tables.value.columns
                      content {
                        column    = columns.value.column != "" ? columns.value.column : null
                        data_type = columns.value.data_type != "" ? columns.value.data_type : null
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    dynamic "mongodb_source_config" {
      for_each = local.source.mongodb_source_config != null ? [local.source.mongodb_source_config] : []
      content {
        max_concurrent_backfill_tasks = mongodb_source_config.value.max_concurrent_backfill_tasks > 0 ? mongodb_source_config.value.max_concurrent_backfill_tasks : null

        dynamic "include_objects" {
          for_each = mongodb_source_config.value.include_objects != null ? [mongodb_source_config.value.include_objects] : []
          content {
            dynamic "databases" {
              for_each = include_objects.value.databases
              content {
                database = databases.value.database != "" ? databases.value.database : null

                dynamic "collections" {
                  for_each = databases.value.collections
                  content {
                    collection = collections.value.collection != "" ? collections.value.collection : null

                    dynamic "fields" {
                      for_each = collections.value.fields
                      content {
                        field = fields.value.field != "" ? fields.value.field : null
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "exclude_objects" {
          for_each = mongodb_source_config.value.exclude_objects != null ? [mongodb_source_config.value.exclude_objects] : []
          content {
            dynamic "databases" {
              for_each = exclude_objects.value.databases
              content {
                database = databases.value.database != "" ? databases.value.database : null

                dynamic "collections" {
                  for_each = databases.value.collections
                  content {
                    collection = collections.value.collection != "" ? collections.value.collection : null

                    dynamic "fields" {
                      for_each = collections.value.fields
                      content {
                        field = fields.value.field != "" ? fields.value.field : null
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    dynamic "salesforce_source_config" {
      for_each = local.source.salesforce_source_config != null ? [local.source.salesforce_source_config] : []
      content {
        polling_interval = salesforce_source_config.value.polling_interval

        dynamic "include_objects" {
          for_each = salesforce_source_config.value.include_objects != null ? [salesforce_source_config.value.include_objects] : []
          content {
            dynamic "objects" {
              for_each = include_objects.value.objects
              content {
                object_name = objects.value.object_name != "" ? objects.value.object_name : null

                dynamic "fields" {
                  for_each = objects.value.fields
                  content {
                    name = fields.value.name != "" ? fields.value.name : null
                  }
                }
              }
            }
          }
        }

        dynamic "exclude_objects" {
          for_each = salesforce_source_config.value.exclude_objects != null ? [salesforce_source_config.value.exclude_objects] : []
          content {
            dynamic "objects" {
              for_each = exclude_objects.value.objects
              content {
                object_name = objects.value.object_name != "" ? objects.value.object_name : null

                dynamic "fields" {
                  for_each = objects.value.fields
                  content {
                    name = fields.value.name != "" ? fields.value.name : null
                  }
                }
              }
            }
          }
        }
      }
    }

    dynamic "spanner_source_config" {
      for_each = local.source.spanner_source_config != null ? [local.source.spanner_source_config] : []
      content {
        change_stream_name            = spanner_source_config.value.change_stream_name
        backfill_data_boost_enabled   = spanner_source_config.value.backfill_data_boost_enabled ? true : null
        fgac_role                     = spanner_source_config.value.fgac_role != "" ? spanner_source_config.value.fgac_role : null
        max_concurrent_backfill_tasks = spanner_source_config.value.max_concurrent_backfill_tasks > 0 ? spanner_source_config.value.max_concurrent_backfill_tasks : null
        max_concurrent_cdc_tasks      = spanner_source_config.value.max_concurrent_cdc_tasks > 0 ? spanner_source_config.value.max_concurrent_cdc_tasks : null
        spanner_rpc_priority          = spanner_source_config.value.spanner_rpc_priority != "" ? spanner_source_config.value.spanner_rpc_priority : null

        dynamic "include_objects" {
          for_each = spanner_source_config.value.include_objects != null ? [spanner_source_config.value.include_objects] : []
          content {
            dynamic "schemas" {
              for_each = include_objects.value.schemas
              content {
                schema = schemas.value.schema

                dynamic "tables" {
                  for_each = schemas.value.tables
                  content {
                    table = tables.value.table

                    dynamic "columns" {
                      for_each = tables.value.columns
                      content {
                        column = columns.value.column != "" ? columns.value.column : null
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "exclude_objects" {
          for_each = spanner_source_config.value.exclude_objects != null ? [spanner_source_config.value.exclude_objects] : []
          content {
            dynamic "schemas" {
              for_each = exclude_objects.value.schemas
              content {
                schema = schemas.value.schema

                dynamic "tables" {
                  for_each = schemas.value.tables
                  content {
                    table = tables.value.table

                    dynamic "columns" {
                      for_each = tables.value.columns
                      content {
                        column = columns.value.column != "" ? columns.value.column : null
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  destination_config {
    destination_connection_profile = local.destination.destination_connection_profile

    dynamic "bigquery_destination_config" {
      for_each = local.bigquery != null ? [local.bigquery] : []
      content {
        data_freshness = bigquery_destination_config.value.data_freshness != "" ? bigquery_destination_config.value.data_freshness : null

        dynamic "single_target_dataset" {
          for_each = local.single_target_dataset_id != null ? [local.single_target_dataset_id] : []
          content {
            dataset_id = single_target_dataset.value
          }
        }

        dynamic "source_hierarchy_datasets" {
          for_each = bigquery_destination_config.value.source_hierarchy_datasets != null ? [bigquery_destination_config.value.source_hierarchy_datasets] : []
          content {
            project_id = source_hierarchy_datasets.value.project_id != "" ? source_hierarchy_datasets.value.project_id : null

            dataset_template {
              location          = source_hierarchy_datasets.value.dataset_template.location
              dataset_id_prefix = source_hierarchy_datasets.value.dataset_template.dataset_id_prefix != "" ? source_hierarchy_datasets.value.dataset_template.dataset_id_prefix : null
              kms_key_name      = source_hierarchy_datasets.value.dataset_template.kms_key_name != "" ? source_hierarchy_datasets.value.dataset_template.kms_key_name : null
            }
          }
        }

        # The spec's write_mode selects which of Google's two empty blocks
        # is sent.
        dynamic "merge" {
          for_each = bigquery_destination_config.value.write_mode == "MERGE" ? [true] : []
          content {}
        }

        dynamic "append_only" {
          for_each = bigquery_destination_config.value.write_mode == "APPEND_ONLY" ? [true] : []
          content {}
        }

        dynamic "blmt_config" {
          for_each = bigquery_destination_config.value.blmt_config != null ? [bigquery_destination_config.value.blmt_config] : []
          content {
            bucket          = blmt_config.value.bucket
            root_path       = blmt_config.value.root_path != "" ? blmt_config.value.root_path : null
            connection_name = local.blmt_connection_name
            file_format     = blmt_config.value.file_format
            table_format    = blmt_config.value.table_format
          }
        }
      }
    }

    dynamic "gcs_destination_config" {
      for_each = local.gcs != null ? [local.gcs] : []
      content {
        path                   = gcs_destination_config.value.path != "" ? gcs_destination_config.value.path : null
        file_rotation_interval = gcs_destination_config.value.file_rotation_interval != "" ? gcs_destination_config.value.file_rotation_interval : null
        file_rotation_mb       = gcs_destination_config.value.file_rotation_mb > 0 ? gcs_destination_config.value.file_rotation_mb : null

        # Google's Avro block has no settings; the spec's bool emits the
        # empty marker block.
        dynamic "avro_file_format" {
          for_each = gcs_destination_config.value.avro_file_format ? [true] : []
          content {}
        }

        dynamic "json_file_format" {
          for_each = gcs_destination_config.value.json_file_format != null ? [gcs_destination_config.value.json_file_format] : []
          content {
            compression        = json_file_format.value.compression != "" ? json_file_format.value.compression : null
            schema_file_format = json_file_format.value.schema_file_format != "" ? json_file_format.value.schema_file_format : null
          }
        }
      }
    }
  }

  dynamic "backfill_all" {
    for_each = var.spec.backfill_all != null ? [var.spec.backfill_all] : []
    content {
      dynamic "mysql_excluded_objects" {
        for_each = backfill_all.value.mysql_excluded_objects != null ? [backfill_all.value.mysql_excluded_objects] : []
        content {
          dynamic "mysql_databases" {
            for_each = mysql_excluded_objects.value.mysql_databases
            content {
              database = mysql_databases.value.database

              dynamic "mysql_tables" {
                for_each = mysql_databases.value.mysql_tables
                content {
                  table = mysql_tables.value.table

                  dynamic "mysql_columns" {
                    for_each = mysql_tables.value.mysql_columns
                    content {
                      column           = mysql_columns.value.column != "" ? mysql_columns.value.column : null
                      collation        = mysql_columns.value.collation != "" ? mysql_columns.value.collation : null
                      data_type        = mysql_columns.value.data_type != "" ? mysql_columns.value.data_type : null
                      nullable         = mysql_columns.value.nullable ? true : null
                      primary_key      = mysql_columns.value.primary_key ? true : null
                      ordinal_position = mysql_columns.value.ordinal_position > 0 ? mysql_columns.value.ordinal_position : null
                    }
                  }
                }
              }
            }
          }
        }
      }

      dynamic "postgresql_excluded_objects" {
        for_each = backfill_all.value.postgresql_excluded_objects != null ? [backfill_all.value.postgresql_excluded_objects] : []
        content {
          dynamic "postgresql_schemas" {
            for_each = postgresql_excluded_objects.value.postgresql_schemas
            content {
              schema = postgresql_schemas.value.schema

              dynamic "postgresql_tables" {
                for_each = postgresql_schemas.value.postgresql_tables
                content {
                  table = postgresql_tables.value.table

                  dynamic "postgresql_columns" {
                    for_each = postgresql_tables.value.postgresql_columns
                    content {
                      column           = postgresql_columns.value.column != "" ? postgresql_columns.value.column : null
                      data_type        = postgresql_columns.value.data_type != "" ? postgresql_columns.value.data_type : null
                      nullable         = postgresql_columns.value.nullable ? true : null
                      primary_key      = postgresql_columns.value.primary_key ? true : null
                      ordinal_position = postgresql_columns.value.ordinal_position > 0 ? postgresql_columns.value.ordinal_position : null
                    }
                  }
                }
              }
            }
          }
        }
      }

      dynamic "oracle_excluded_objects" {
        for_each = backfill_all.value.oracle_excluded_objects != null ? [backfill_all.value.oracle_excluded_objects] : []
        content {
          dynamic "oracle_schemas" {
            for_each = oracle_excluded_objects.value.oracle_schemas
            content {
              schema = oracle_schemas.value.schema

              dynamic "oracle_tables" {
                for_each = oracle_schemas.value.oracle_tables
                content {
                  table = oracle_tables.value.table

                  dynamic "oracle_columns" {
                    for_each = oracle_tables.value.oracle_columns
                    content {
                      column    = oracle_columns.value.column != "" ? oracle_columns.value.column : null
                      data_type = oracle_columns.value.data_type != "" ? oracle_columns.value.data_type : null
                    }
                  }
                }
              }
            }
          }
        }
      }

      dynamic "sql_server_excluded_objects" {
        for_each = backfill_all.value.sql_server_excluded_objects != null ? [backfill_all.value.sql_server_excluded_objects] : []
        content {
          dynamic "schemas" {
            for_each = sql_server_excluded_objects.value.schemas
            content {
              schema = schemas.value.schema

              dynamic "tables" {
                for_each = schemas.value.tables
                content {
                  table = tables.value.table

                  dynamic "columns" {
                    for_each = tables.value.columns
                    content {
                      column    = columns.value.column != "" ? columns.value.column : null
                      data_type = columns.value.data_type != "" ? columns.value.data_type : null
                    }
                  }
                }
              }
            }
          }
        }
      }

      dynamic "mongodb_excluded_objects" {
        for_each = backfill_all.value.mongodb_excluded_objects != null ? [backfill_all.value.mongodb_excluded_objects] : []
        content {
          dynamic "databases" {
            for_each = mongodb_excluded_objects.value.databases
            content {
              database = databases.value.database != "" ? databases.value.database : null

              dynamic "collections" {
                for_each = databases.value.collections
                content {
                  collection = collections.value.collection != "" ? collections.value.collection : null

                  dynamic "fields" {
                    for_each = collections.value.fields
                    content {
                      field = fields.value.field != "" ? fields.value.field : null
                    }
                  }
                }
              }
            }
          }
        }
      }

      dynamic "salesforce_excluded_objects" {
        for_each = backfill_all.value.salesforce_excluded_objects != null ? [backfill_all.value.salesforce_excluded_objects] : []
        content {
          dynamic "objects" {
            for_each = salesforce_excluded_objects.value.objects
            content {
              object_name = objects.value.object_name != "" ? objects.value.object_name : null

              dynamic "fields" {
                for_each = objects.value.fields
                content {
                  name = fields.value.name != "" ? fields.value.name : null
                }
              }
            }
          }
        }
      }

      dynamic "spanner_excluded_objects" {
        for_each = backfill_all.value.spanner_excluded_objects != null ? [backfill_all.value.spanner_excluded_objects] : []
        content {
          dynamic "schemas" {
            for_each = spanner_excluded_objects.value.schemas
            content {
              schema = schemas.value.schema

              dynamic "tables" {
                for_each = schemas.value.tables
                content {
                  table = tables.value.table

                  dynamic "columns" {
                    for_each = tables.value.columns
                    content {
                      column = columns.value.column != "" ? columns.value.column : null
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  # Google's backfill_none block has no settings; the spec's bool emits the
  # empty marker block.
  dynamic "backfill_none" {
    for_each = var.spec.backfill_none ? [true] : []
    content {}
  }

  dynamic "rule_sets" {
    for_each = var.spec.rule_sets
    content {
      object_filter {
        dynamic "source_object_identifier" {
          for_each = rule_sets.value.object_filter.source_object_identifier != null ? [rule_sets.value.object_filter.source_object_identifier] : []
          content {
            dynamic "mysql_identifier" {
              for_each = source_object_identifier.value.mysql_identifier != null ? [source_object_identifier.value.mysql_identifier] : []
              content {
                database = mysql_identifier.value.database
                table    = mysql_identifier.value.table
              }
            }

            dynamic "postgresql_identifier" {
              for_each = source_object_identifier.value.postgresql_identifier != null ? [source_object_identifier.value.postgresql_identifier] : []
              content {
                schema = postgresql_identifier.value.schema
                table  = postgresql_identifier.value.table
              }
            }

            dynamic "oracle_identifier" {
              for_each = source_object_identifier.value.oracle_identifier != null ? [source_object_identifier.value.oracle_identifier] : []
              content {
                schema = oracle_identifier.value.schema
                table  = oracle_identifier.value.table
              }
            }

            dynamic "sql_server_identifier" {
              for_each = source_object_identifier.value.sql_server_identifier != null ? [source_object_identifier.value.sql_server_identifier] : []
              content {
                schema = sql_server_identifier.value.schema
                table  = sql_server_identifier.value.table
              }
            }

            dynamic "mongodb_identifier" {
              for_each = source_object_identifier.value.mongodb_identifier != null ? [source_object_identifier.value.mongodb_identifier] : []
              content {
                database   = mongodb_identifier.value.database
                collection = mongodb_identifier.value.collection
              }
            }

            dynamic "salesforce_identifier" {
              for_each = source_object_identifier.value.salesforce_identifier != null ? [source_object_identifier.value.salesforce_identifier] : []
              content {
                object_name = salesforce_identifier.value.object_name
              }
            }

            dynamic "spanner_identifier" {
              for_each = source_object_identifier.value.spanner_identifier != null ? [source_object_identifier.value.spanner_identifier] : []
              content {
                schema = spanner_identifier.value.schema != "" ? spanner_identifier.value.schema : null
                table  = spanner_identifier.value.table
              }
            }
          }
        }
      }

      dynamic "customization_rules" {
        for_each = rule_sets.value.customization_rules
        content {
          dynamic "bigquery_clustering" {
            for_each = customization_rules.value.bigquery_clustering != null ? [customization_rules.value.bigquery_clustering] : []
            content {
              columns = bigquery_clustering.value.columns
            }
          }

          dynamic "bigquery_partitioning" {
            for_each = customization_rules.value.bigquery_partitioning != null ? [customization_rules.value.bigquery_partitioning] : []
            content {
              require_partition_filter = bigquery_partitioning.value.require_partition_filter ? true : null

              dynamic "ingestion_time_partition" {
                for_each = bigquery_partitioning.value.ingestion_time_partition != null ? [bigquery_partitioning.value.ingestion_time_partition] : []
                content {
                  partitioning_time_granularity = ingestion_time_partition.value.partitioning_time_granularity != "" ? ingestion_time_partition.value.partitioning_time_granularity : null
                }
              }

              dynamic "time_unit_partition" {
                for_each = bigquery_partitioning.value.time_unit_partition != null ? [bigquery_partitioning.value.time_unit_partition] : []
                content {
                  column                        = time_unit_partition.value.column
                  partitioning_time_granularity = time_unit_partition.value.partitioning_time_granularity != "" ? time_unit_partition.value.partitioning_time_granularity : null
                }
              }

              dynamic "integer_range_partition" {
                for_each = bigquery_partitioning.value.integer_range_partition != null ? [bigquery_partitioning.value.integer_range_partition] : []
                content {
                  column   = integer_range_partition.value.column
                  start    = integer_range_partition.value.start
                  end      = integer_range_partition.value.end
                  interval = integer_range_partition.value.interval
                }
              }
            }
          }
        }
      }
    }
  }

  labels = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy
}
