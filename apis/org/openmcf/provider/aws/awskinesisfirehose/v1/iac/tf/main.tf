# ---------------------------------------------------------------------------
# Kinesis Data Firehose delivery stream
# ---------------------------------------------------------------------------

resource "aws_kinesis_firehose_delivery_stream" "this" {
  name        = local.delivery_stream_name
  destination = local.destination_type
  tags        = local.tags

  # ---------------------------------------------------------------------------
  # Source configuration (optional — Direct PUT when absent)
  # ---------------------------------------------------------------------------

  dynamic "kinesis_source_configuration" {
    for_each = local.has_kinesis_source ? [var.spec.kinesis_stream_source] : []
    iterator = src
    content {
      kinesis_stream_arn = try(src.value.stream_arn.value, src.value.stream_arn)
      role_arn           = try(src.value.role_arn.value, src.value.role_arn)
    }
  }

  # ---------------------------------------------------------------------------
  # Server-side encryption (Direct PUT only)
  # ---------------------------------------------------------------------------

  dynamic "server_side_encryption" {
    for_each = local.sse_enabled ? [1] : []
    content {
      enabled  = true
      key_type = local.sse_key_type
      key_arn  = local.sse_kms_key_arn
    }
  }

  # ===========================================================================
  # Extended S3 destination
  # ===========================================================================

  dynamic "extended_s3_configuration" {
    for_each = local.destination_type == "extended_s3" ? [var.spec.extended_s3] : []
    iterator = dest
    content {
      # Required fields
      bucket_arn = try(dest.value.bucket_arn.value, dest.value.bucket_arn)
      role_arn   = try(dest.value.role_arn.value, dest.value.role_arn)

      # S3 delivery options
      prefix              = try(dest.value.prefix, null) != "" ? dest.value.prefix : null
      error_output_prefix = try(dest.value.error_output_prefix, null) != "" ? dest.value.error_output_prefix : null
      compression_format  = try(dest.value.compression_format, null) != "" ? dest.value.compression_format : null
      kms_key_arn         = try(dest.value.kms_key_arn.value, null)
      buffering_interval  = try(dest.value.buffering.interval_in_seconds, null) > 0 ? dest.value.buffering.interval_in_seconds : null
      buffering_size      = try(dest.value.buffering.size_in_mbs, null) > 0 ? dest.value.buffering.size_in_mbs : null
      custom_time_zone    = try(dest.value.custom_time_zone, null) != "" ? dest.value.custom_time_zone : null
      file_extension      = try(dest.value.file_extension, null) != "" ? dest.value.file_extension : null

      # S3 backup mode
      s3_backup_mode = try(dest.value.s3_backup_mode, null) != "" ? dest.value.s3_backup_mode : null

      # --- S3 backup configuration (source record backup) ---

      dynamic "s3_backup_configuration" {
        for_each = try(dest.value.s3_backup, null) != null ? [dest.value.s3_backup] : []
        iterator = bkp
        content {
          bucket_arn          = try(bkp.value.bucket_arn.value, bkp.value.bucket_arn)
          role_arn            = try(bkp.value.role_arn.value, bkp.value.role_arn)
          prefix              = try(bkp.value.prefix, null) != "" ? bkp.value.prefix : null
          error_output_prefix = try(bkp.value.error_output_prefix, null) != "" ? bkp.value.error_output_prefix : null
          compression_format  = try(bkp.value.compression_format, null) != "" ? bkp.value.compression_format : null
          kms_key_arn         = try(bkp.value.kms_key_arn.value, null)
          buffering_interval  = try(bkp.value.buffering.interval_in_seconds, null) > 0 ? bkp.value.buffering.interval_in_seconds : null
          buffering_size      = try(bkp.value.buffering.size_in_mbs, null) > 0 ? bkp.value.buffering.size_in_mbs : null
        }
      }

      # --- Processing configuration (Lambda) ---

      dynamic "processing_configuration" {
        for_each = try(dest.value.processing.enabled, false) ? [dest.value.processing] : []
        iterator = proc
        content {
          enabled = true
          processors {
            type = "Lambda"

            # LambdaArn (always required when processing is enabled)
            parameters {
              parameter_name  = "LambdaArn"
              parameter_value = try(proc.value.lambda_arn.value, proc.value.lambda_arn)
            }

            # BufferSizeInMBs (optional, 1-3 MiB)
            dynamic "parameters" {
              for_each = try(proc.value.buffer_size_in_mbs, 0) > 0 ? [proc.value.buffer_size_in_mbs] : []
              content {
                parameter_name  = "BufferSizeInMBs"
                parameter_value = tostring(parameters.value)
              }
            }

            # BufferIntervalInSeconds (optional, 60-900s)
            dynamic "parameters" {
              for_each = try(proc.value.buffer_interval_in_seconds, 0) > 0 ? [proc.value.buffer_interval_in_seconds] : []
              content {
                parameter_name  = "BufferIntervalInSeconds"
                parameter_value = tostring(parameters.value)
              }
            }

            # NumberOfRetries (optional, 0-300)
            dynamic "parameters" {
              for_each = try(proc.value.number_of_retries, 0) > 0 ? [proc.value.number_of_retries] : []
              content {
                parameter_name  = "NumberOfRetries"
                parameter_value = tostring(parameters.value)
              }
            }
          }
        }
      }

      # --- CloudWatch logging ---

      dynamic "cloudwatch_logging_options" {
        for_each = try(dest.value.logging.enabled, false) ? [dest.value.logging] : []
        iterator = log
        content {
          enabled         = true
          log_group_name  = log.value.log_group_name
          log_stream_name = log.value.log_stream_name
        }
      }

      # --- Dynamic partitioning (ForceNew) ---

      dynamic "dynamic_partitioning_configuration" {
        for_each = try(dest.value.dynamic_partitioning.enabled, false) ? [dest.value.dynamic_partitioning] : []
        iterator = dp
        content {
          enabled        = true
          retry_duration = try(dp.value.retry_duration_in_seconds, null) > 0 ? dp.value.retry_duration_in_seconds : null
        }
      }

      # --- Data format conversion (Parquet/ORC via Glue catalog) ---

      dynamic "data_format_conversion_configuration" {
        for_each = try(dest.value.data_format_conversion.enabled, false) ? [dest.value.data_format_conversion] : []
        iterator = dfc
        content {
          enabled = true

          # Input format (deserializer) — defaults to OpenX JSON
          input_format_configuration {
            deserializer {
              dynamic "open_x_json_ser_de" {
                for_each = coalesce(try(dfc.value.input_format, null), "OPENX_JSON") != "HIVE_JSON" ? [1] : []
                content {}
              }
              dynamic "hive_json_ser_de" {
                for_each = try(dfc.value.input_format, "") == "HIVE_JSON" ? [1] : []
                content {}
              }
            }
          }

          # Output format (serializer) — PARQUET or ORC
          output_format_configuration {
            serializer {
              dynamic "parquet_ser_de" {
                for_each = try(dfc.value.output_format, "PARQUET") != "ORC" ? [1] : []
                content {
                  compression = try(dfc.value.parquet_compression, null) != "" ? dfc.value.parquet_compression : null
                }
              }
              dynamic "orc_ser_de" {
                for_each = try(dfc.value.output_format, "") == "ORC" ? [1] : []
                content {
                  compression = try(dfc.value.orc_compression, null) != "" ? dfc.value.orc_compression : null
                }
              }
            }
          }

          # Glue Data Catalog schema
          schema_configuration {
            database_name = dfc.value.schema.database_name
            table_name    = dfc.value.schema.table_name
            role_arn      = try(dfc.value.schema.role_arn.value, dfc.value.schema.role_arn)
            catalog_id    = try(dfc.value.schema.catalog_id, null) != "" ? dfc.value.schema.catalog_id : null
            region        = try(dfc.value.schema.region, null) != "" ? dfc.value.schema.region : null
            version_id    = try(dfc.value.schema.version_id, null) != "" ? dfc.value.schema.version_id : null
          }
        }
      }
    }
  }

  # ===========================================================================
  # OpenSearch destination
  # ===========================================================================

  dynamic "opensearch_configuration" {
    for_each = local.destination_type == "opensearch" ? [var.spec.opensearch] : []
    iterator = dest
    content {
      # Target — exactly one of domain_arn or cluster_endpoint
      domain_arn       = try(dest.value.domain_arn.value, null)
      cluster_endpoint = try(dest.value.cluster_endpoint, null) != "" ? dest.value.cluster_endpoint : null

      # Indexing configuration
      index_name            = dest.value.index_name
      role_arn              = try(dest.value.role_arn.value, dest.value.role_arn)
      index_rotation_period = try(dest.value.index_rotation_period, null) != "" ? dest.value.index_rotation_period : null
      type_name             = try(dest.value.type_name, null) != "" ? dest.value.type_name : null

      # Delivery configuration
      buffering_interval = try(dest.value.buffering.interval_in_seconds, null) > 0 ? dest.value.buffering.interval_in_seconds : null
      buffering_size     = try(dest.value.buffering.size_in_mbs, null) > 0 ? dest.value.buffering.size_in_mbs : null
      retry_duration     = try(dest.value.retry_duration_in_seconds, null) > 0 ? dest.value.retry_duration_in_seconds : null

      # S3 backup mode
      s3_backup_mode = try(dest.value.s3_backup_mode, null) != "" ? dest.value.s3_backup_mode : null

      # --- S3 config (required — backs up failed/all documents) ---

      s3_configuration {
        bucket_arn          = try(dest.value.s3_config.bucket_arn.value, dest.value.s3_config.bucket_arn)
        role_arn            = try(dest.value.s3_config.role_arn.value, dest.value.s3_config.role_arn)
        prefix              = try(dest.value.s3_config.prefix, null) != "" ? dest.value.s3_config.prefix : null
        error_output_prefix = try(dest.value.s3_config.error_output_prefix, null) != "" ? dest.value.s3_config.error_output_prefix : null
        compression_format  = try(dest.value.s3_config.compression_format, null) != "" ? dest.value.s3_config.compression_format : null
        kms_key_arn         = try(dest.value.s3_config.kms_key_arn.value, null)
        buffering_interval  = try(dest.value.s3_config.buffering.interval_in_seconds, null) > 0 ? dest.value.s3_config.buffering.interval_in_seconds : null
        buffering_size      = try(dest.value.s3_config.buffering.size_in_mbs, null) > 0 ? dest.value.s3_config.buffering.size_in_mbs : null
      }

      # --- Processing configuration (Lambda) ---

      dynamic "processing_configuration" {
        for_each = try(dest.value.processing.enabled, false) ? [dest.value.processing] : []
        iterator = proc
        content {
          enabled = true
          processors {
            type = "Lambda"

            parameters {
              parameter_name  = "LambdaArn"
              parameter_value = try(proc.value.lambda_arn.value, proc.value.lambda_arn)
            }

            dynamic "parameters" {
              for_each = try(proc.value.buffer_size_in_mbs, 0) > 0 ? [proc.value.buffer_size_in_mbs] : []
              content {
                parameter_name  = "BufferSizeInMBs"
                parameter_value = tostring(parameters.value)
              }
            }

            dynamic "parameters" {
              for_each = try(proc.value.buffer_interval_in_seconds, 0) > 0 ? [proc.value.buffer_interval_in_seconds] : []
              content {
                parameter_name  = "BufferIntervalInSeconds"
                parameter_value = tostring(parameters.value)
              }
            }

            dynamic "parameters" {
              for_each = try(proc.value.number_of_retries, 0) > 0 ? [proc.value.number_of_retries] : []
              content {
                parameter_name  = "NumberOfRetries"
                parameter_value = tostring(parameters.value)
              }
            }
          }
        }
      }

      # --- CloudWatch logging ---

      dynamic "cloudwatch_logging_options" {
        for_each = try(dest.value.logging.enabled, false) ? [dest.value.logging] : []
        iterator = log
        content {
          enabled         = true
          log_group_name  = log.value.log_group_name
          log_stream_name = log.value.log_stream_name
        }
      }

      # --- VPC config (for VPC-deployed OpenSearch domains, ForceNew) ---

      dynamic "vpc_config" {
        for_each = try(dest.value.vpc_config, null) != null ? [dest.value.vpc_config] : []
        iterator = vpc
        content {
          role_arn           = try(vpc.value.role_arn.value, vpc.value.role_arn)
          subnet_ids         = [for s in try(vpc.value.subnet_ids, []) : try(s.value, s)]
          security_group_ids = [for s in try(vpc.value.security_group_ids, []) : try(s.value, s)]
        }
      }
    }
  }

  # ===========================================================================
  # HTTP Endpoint destination
  # ===========================================================================

  dynamic "http_endpoint_configuration" {
    for_each = local.destination_type == "http_endpoint" ? [var.spec.http_endpoint] : []
    iterator = dest
    content {
      # Endpoint configuration
      url        = dest.value.url
      name       = try(dest.value.name, null) != "" ? dest.value.name : null
      access_key = try(dest.value.access_key, null) != "" ? dest.value.access_key : null
      role_arn   = try(dest.value.role_arn.value, null)

      # Delivery configuration
      buffering_interval = try(dest.value.buffering.interval_in_seconds, null) > 0 ? dest.value.buffering.interval_in_seconds : null
      buffering_size     = try(dest.value.buffering.size_in_mbs, null) > 0 ? dest.value.buffering.size_in_mbs : null
      retry_duration     = try(dest.value.retry_duration_in_seconds, null) > 0 ? dest.value.retry_duration_in_seconds : null

      # S3 backup mode
      s3_backup_mode = try(dest.value.s3_backup_mode, null) != "" ? dest.value.s3_backup_mode : null

      # --- S3 config (required — backs up failed/all records) ---

      s3_configuration {
        bucket_arn          = try(dest.value.s3_config.bucket_arn.value, dest.value.s3_config.bucket_arn)
        role_arn            = try(dest.value.s3_config.role_arn.value, dest.value.s3_config.role_arn)
        prefix              = try(dest.value.s3_config.prefix, null) != "" ? dest.value.s3_config.prefix : null
        error_output_prefix = try(dest.value.s3_config.error_output_prefix, null) != "" ? dest.value.s3_config.error_output_prefix : null
        compression_format  = try(dest.value.s3_config.compression_format, null) != "" ? dest.value.s3_config.compression_format : null
        kms_key_arn         = try(dest.value.s3_config.kms_key_arn.value, null)
        buffering_interval  = try(dest.value.s3_config.buffering.interval_in_seconds, null) > 0 ? dest.value.s3_config.buffering.interval_in_seconds : null
        buffering_size      = try(dest.value.s3_config.buffering.size_in_mbs, null) > 0 ? dest.value.s3_config.buffering.size_in_mbs : null
      }

      # --- Processing configuration (Lambda) ---

      dynamic "processing_configuration" {
        for_each = try(dest.value.processing.enabled, false) ? [dest.value.processing] : []
        iterator = proc
        content {
          enabled = true
          processors {
            type = "Lambda"

            parameters {
              parameter_name  = "LambdaArn"
              parameter_value = try(proc.value.lambda_arn.value, proc.value.lambda_arn)
            }

            dynamic "parameters" {
              for_each = try(proc.value.buffer_size_in_mbs, 0) > 0 ? [proc.value.buffer_size_in_mbs] : []
              content {
                parameter_name  = "BufferSizeInMBs"
                parameter_value = tostring(parameters.value)
              }
            }

            dynamic "parameters" {
              for_each = try(proc.value.buffer_interval_in_seconds, 0) > 0 ? [proc.value.buffer_interval_in_seconds] : []
              content {
                parameter_name  = "BufferIntervalInSeconds"
                parameter_value = tostring(parameters.value)
              }
            }

            dynamic "parameters" {
              for_each = try(proc.value.number_of_retries, 0) > 0 ? [proc.value.number_of_retries] : []
              content {
                parameter_name  = "NumberOfRetries"
                parameter_value = tostring(parameters.value)
              }
            }
          }
        }
      }

      # --- CloudWatch logging ---

      dynamic "cloudwatch_logging_options" {
        for_each = try(dest.value.logging.enabled, false) ? [dest.value.logging] : []
        iterator = log
        content {
          enabled         = true
          log_group_name  = log.value.log_group_name
          log_stream_name = log.value.log_stream_name
        }
      }

      # --- Request configuration ---

      dynamic "request_configuration" {
        for_each = try(dest.value.request_config, null) != null ? [dest.value.request_config] : []
        iterator = rc
        content {
          content_encoding = try(rc.value.content_encoding, null) != "" ? rc.value.content_encoding : null

          dynamic "common_attributes" {
            for_each = try(rc.value.common_attributes, [])
            content {
              name  = common_attributes.value.name
              value = common_attributes.value.value
            }
          }
        }
      }
    }
  }

  # ===========================================================================
  # Redshift destination
  # ===========================================================================

  dynamic "redshift_configuration" {
    for_each = local.destination_type == "redshift" ? [var.spec.redshift] : []
    iterator = dest
    content {
      # Redshift target
      cluster_jdbcurl    = dest.value.cluster_jdbcurl
      role_arn           = try(dest.value.role_arn.value, dest.value.role_arn)
      data_table_name    = dest.value.data_table_name
      data_table_columns = try(dest.value.data_table_columns, null) != "" ? dest.value.data_table_columns : null
      copy_options       = try(dest.value.copy_options, null) != "" ? dest.value.copy_options : null

      # Authentication
      username = try(dest.value.username, null) != "" ? dest.value.username : null
      password = try(dest.value.password.value, try(dest.value.password, null))

      # Delivery configuration
      retry_duration = try(dest.value.retry_duration_in_seconds, null) > 0 ? dest.value.retry_duration_in_seconds : null

      # S3 backup mode for source records
      s3_backup_mode = try(dest.value.s3_backup_mode, null) != "" ? dest.value.s3_backup_mode : null

      # --- S3 intermediate staging config (required for Redshift COPY) ---

      s3_configuration {
        bucket_arn          = try(dest.value.s3_config.bucket_arn.value, dest.value.s3_config.bucket_arn)
        role_arn            = try(dest.value.s3_config.role_arn.value, dest.value.s3_config.role_arn)
        prefix              = try(dest.value.s3_config.prefix, null) != "" ? dest.value.s3_config.prefix : null
        error_output_prefix = try(dest.value.s3_config.error_output_prefix, null) != "" ? dest.value.s3_config.error_output_prefix : null
        compression_format  = try(dest.value.s3_config.compression_format, null) != "" ? dest.value.s3_config.compression_format : null
        kms_key_arn         = try(dest.value.s3_config.kms_key_arn.value, null)
        buffering_interval  = try(dest.value.s3_config.buffering.interval_in_seconds, null) > 0 ? dest.value.s3_config.buffering.interval_in_seconds : null
        buffering_size      = try(dest.value.s3_config.buffering.size_in_mbs, null) > 0 ? dest.value.s3_config.buffering.size_in_mbs : null
      }

      # --- S3 backup configuration (source record backup) ---

      dynamic "s3_backup_configuration" {
        for_each = try(dest.value.s3_backup, null) != null ? [dest.value.s3_backup] : []
        iterator = bkp
        content {
          bucket_arn          = try(bkp.value.bucket_arn.value, bkp.value.bucket_arn)
          role_arn            = try(bkp.value.role_arn.value, bkp.value.role_arn)
          prefix              = try(bkp.value.prefix, null) != "" ? bkp.value.prefix : null
          error_output_prefix = try(bkp.value.error_output_prefix, null) != "" ? bkp.value.error_output_prefix : null
          compression_format  = try(bkp.value.compression_format, null) != "" ? bkp.value.compression_format : null
          kms_key_arn         = try(bkp.value.kms_key_arn.value, null)
          buffering_interval  = try(bkp.value.buffering.interval_in_seconds, null) > 0 ? bkp.value.buffering.interval_in_seconds : null
          buffering_size      = try(bkp.value.buffering.size_in_mbs, null) > 0 ? bkp.value.buffering.size_in_mbs : null
        }
      }

      # --- Processing configuration (Lambda) ---

      dynamic "processing_configuration" {
        for_each = try(dest.value.processing.enabled, false) ? [dest.value.processing] : []
        iterator = proc
        content {
          enabled = true
          processors {
            type = "Lambda"

            parameters {
              parameter_name  = "LambdaArn"
              parameter_value = try(proc.value.lambda_arn.value, proc.value.lambda_arn)
            }

            dynamic "parameters" {
              for_each = try(proc.value.buffer_size_in_mbs, 0) > 0 ? [proc.value.buffer_size_in_mbs] : []
              content {
                parameter_name  = "BufferSizeInMBs"
                parameter_value = tostring(parameters.value)
              }
            }

            dynamic "parameters" {
              for_each = try(proc.value.buffer_interval_in_seconds, 0) > 0 ? [proc.value.buffer_interval_in_seconds] : []
              content {
                parameter_name  = "BufferIntervalInSeconds"
                parameter_value = tostring(parameters.value)
              }
            }

            dynamic "parameters" {
              for_each = try(proc.value.number_of_retries, 0) > 0 ? [proc.value.number_of_retries] : []
              content {
                parameter_name  = "NumberOfRetries"
                parameter_value = tostring(parameters.value)
              }
            }
          }
        }
      }

      # --- CloudWatch logging ---

      dynamic "cloudwatch_logging_options" {
        for_each = try(dest.value.logging.enabled, false) ? [dest.value.logging] : []
        iterator = log
        content {
          enabled         = true
          log_group_name  = log.value.log_group_name
          log_stream_name = log.value.log_stream_name
        }
      }
    }
  }
}
