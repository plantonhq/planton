variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsS3Bucket specification"
  type = object({
    region = string
    force_destroy = optional(bool, false)
    object_lock_enabled = optional(bool, false)
    bucket_namespace = optional(string, "")
    versioning_status = optional(string, "")
    encryption = optional(object({
      sse_algorithm = optional(string, "")
      kms_key_id = optional(string, "")
      bucket_key_enabled = optional(bool, false)
      blocked_encryption_types = optional(list(string), [])
    }))
    public_access_block = optional(object({
      block_public_acls = optional(bool)
      block_public_policy = optional(bool)
      ignore_public_acls = optional(bool)
      restrict_public_buckets = optional(bool)
    }))
    object_ownership = optional(string, "")
    acl = optional(string, "")
    policy = optional(any)
    transition_default_minimum_object_size = optional(string, "")
    lifecycle_rules = optional(list(object({
      id = string
      status = optional(string, "")
      filter = optional(object({
        prefix = optional(string, "")
        tags = optional(map(string), {})
        object_size_greater_than = optional(number, 0)
        object_size_less_than = optional(number, 0)
      }))
      transitions = optional(list(object({
        days = optional(number)
        date = optional(string, "")
        storage_class = string
      })), [])
      expiration = optional(object({
        days = optional(number, 0)
        date = optional(string, "")
        expired_object_delete_marker = optional(bool, false)
      }))
      noncurrent_version_transitions = optional(list(object({
        noncurrent_days = optional(number, 0)
        newer_noncurrent_versions = optional(number, 0)
        storage_class = string
      })), [])
      noncurrent_version_expiration = optional(object({
        noncurrent_days = optional(number, 0)
        newer_noncurrent_versions = optional(number, 0)
      }))
      abort_incomplete_multipart_upload_days = optional(number, 0)
    })), [])
    replication = optional(object({
      role_arn = string
      rules = list(object({
        id = string
        priority = optional(number, 0)
        status = optional(string, "")
        filter = optional(object({
          prefix = optional(string, "")
          tags = optional(map(string), {})
        }))
        destination = object({
          bucket_arn = string
          account = optional(string, "")
          storage_class = optional(string, "")
          change_replica_ownership_to_destination = optional(bool, false)
          replica_kms_key_id = optional(string, "")
          metrics_enabled = optional(bool, false)
          replication_time_control_enabled = optional(bool, false)
        })
        delete_marker_replication = optional(bool, false)
        existing_object_replication = optional(bool, false)
        replicate_replica_modifications = optional(bool, false)
        replicate_sse_kms_encrypted_objects = optional(bool, false)
      }))
    }))
    website = optional(object({
      index_document_suffix = optional(string, "")
      error_document_key = optional(string, "")
      redirect_all_requests_to = optional(object({
        host_name = string
        protocol = optional(string, "")
      }))
      routing_rules = optional(list(object({
        condition = optional(object({
          http_error_code_returned_equals = optional(string, "")
          key_prefix_equals = optional(string, "")
        }))
        redirect = object({
          host_name = optional(string, "")
          http_redirect_code = optional(string, "")
          protocol = optional(string, "")
          replace_key_prefix_with = optional(string, "")
          replace_key_with = optional(string, "")
        })
      })), [])
    }))
    logging = optional(object({
      target_bucket = string
      target_prefix = optional(string, "")
      partitioned_prefix_date_source = optional(string, "")
    }))
    cors_rules = optional(list(object({
      id = optional(string, "")
      allowed_methods = list(string)
      allowed_origins = list(string)
      allowed_headers = optional(list(string), [])
      expose_headers = optional(list(string), [])
      max_age_seconds = optional(number, 0)
    })), [])
    notification = optional(object({
      eventbridge = optional(bool, false)
      lambda_functions = optional(list(object({
        lambda_function_arn = string
        events = list(string)
        filter_prefix = optional(string, "")
        filter_suffix = optional(string, "")
      })), [])
      queues = optional(list(object({
        queue_arn = string
        events = list(string)
        filter_prefix = optional(string, "")
        filter_suffix = optional(string, "")
      })), [])
      topics = optional(list(object({
        topic_arn = string
        events = list(string)
        filter_prefix = optional(string, "")
        filter_suffix = optional(string, "")
      })), [])
    }))
    object_lock_default_retention = optional(object({
      mode = string
      days = optional(number, 0)
      years = optional(number, 0)
    }))
    acceleration_status = optional(string, "")
    request_payer = optional(string, "")
    intelligent_tiering_configurations = optional(list(object({
      name = string
      status = optional(string, "")
      filter_prefix = optional(string, "")
      filter_tags = optional(map(string), {})
      tiers = list(object({
        access_tier = string
        days = optional(number, 0)
      }))
    })), [])
    abac_status = optional(string, "")
    analytics_configurations = optional(list(object({
      name = string
      filter_prefix = optional(string, "")
      filter_tags = optional(map(string), {})
      export = optional(object({
        bucket_arn = string
        bucket_account_id = optional(string, "")
        prefix = optional(string, "")
      }))
    })), [])
    inventory_configurations = optional(list(object({
      name = string
      disabled = optional(bool, false)
      included_object_versions = string
      frequency = string
      optional_fields = optional(list(string), [])
      filter_prefix = optional(string, "")
      destination = object({
        bucket_arn = string
        format = string
        account_id = optional(string, "")
        prefix = optional(string, "")
        sse_kms_key_id = optional(string, "")
        sse_s3 = optional(bool, false)
      })
    })), [])
    metrics_configurations = optional(list(object({
      name = string
      filter_access_point_arn = optional(string, "")
      filter_prefix = optional(string, "")
      filter_tags = optional(map(string), {})
    })), [])
    metadata_configuration = optional(object({
      inventory_table_enabled = optional(bool, false)
      inventory_table_encryption = optional(object({
        sse_algorithm = string
        kms_key_arn = optional(string, "")
      }))
      journal_record_expiration = object({
        enabled = optional(bool, false)
        days = optional(number, 0)
      })
      journal_table_encryption = optional(object({
        sse_algorithm = string
        kms_key_arn = optional(string, "")
      }))
    }))
  })
}
