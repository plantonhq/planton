variable "spec" {
  description = "GcpPubSubTopic spec"
  type = object({
    project_id = object({ value = string })
    topic_name = string

    kms_key_name               = optional(object({ value = string }), null)
    message_retention_duration = optional(string, "")

    message_storage_policy = optional(object({
      allowed_persistence_regions = list(string)
      enforce_in_transit          = optional(bool, false)
    }), null)

    schema_settings = optional(object({
      schema   = string
      encoding = optional(string, "")
    }), null)

    ingestion_data_source_settings = optional(object({
      aws_kinesis = optional(object({
        stream_arn          = string
        consumer_arn        = string
        aws_role_arn        = string
        gcp_service_account = string
      }), null)

      aws_msk = optional(object({
        cluster_arn         = string
        topic               = string
        aws_role_arn        = string
        gcp_service_account = string
      }), null)

      azure_event_hubs = optional(object({
        resource_group      = optional(string, "")
        namespace           = optional(string, "")
        event_hub           = optional(string, "")
        client_id           = optional(string, "")
        tenant_id           = optional(string, "")
        subscription_id     = optional(string, "")
        gcp_service_account = optional(string, "")
      }), null)

      cloud_storage = optional(object({
        bucket = object({ value = string })
        match_glob                = optional(string, "")
        minimum_object_create_time = optional(string, "")
        avro_format               = optional(bool, false)
        pubsub_avro_format        = optional(bool, false)
        text_format = optional(object({
          delimiter = optional(string, "")
        }), null)
      }), null)

      confluent_cloud = optional(object({
        bootstrap_server    = string
        topic               = string
        identity_pool_id    = string
        gcp_service_account = string
        cluster_id          = optional(string, "")
      }), null)

      platform_logs_settings = optional(object({
        severity = optional(string, "")
      }), null)
    }), null)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = { service_account_key = "" }
}
