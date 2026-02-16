variable "spec" {
  description = "GcpPubSubSubscription spec"
  type = object({
    project_id        = object({ value = string })
    subscription_name = string
    topic             = object({ value = string })

    ack_deadline_seconds       = optional(number, 0)
    message_retention_duration = optional(string, "")
    retain_acked_messages      = optional(bool, false)
    filter                     = optional(string, "")
    enable_message_ordering    = optional(bool, false)
    enable_exactly_once_delivery = optional(bool, false)

    expiration_policy = optional(object({
      ttl = string
    }), null)

    dead_letter_policy = optional(object({
      dead_letter_topic   = optional(object({ value = string }), null)
      max_delivery_attempts = optional(number, 0)
    }), null)

    retry_policy = optional(object({
      minimum_backoff = optional(string, "")
      maximum_backoff = optional(string, "")
    }), null)

    push_config = optional(object({
      push_endpoint = string
      attributes    = optional(map(string), {})
      oidc_token = optional(object({
        service_account_email = string
        audience              = optional(string, "")
      }), null)
      no_wrapper = optional(object({
        write_metadata = bool
      }), null)
    }), null)

    bigquery_config = optional(object({
      table                 = string
      use_topic_schema      = optional(bool, false)
      use_table_schema      = optional(bool, false)
      drop_unknown_fields   = optional(bool, false)
      write_metadata        = optional(bool, false)
      service_account_email = optional(string, "")
    }), null)

    cloud_storage_config = optional(object({
      bucket                   = object({ value = string })
      filename_prefix          = optional(string, "")
      filename_suffix          = optional(string, "")
      filename_datetime_format = optional(string, "")
      max_bytes                = optional(number, 0)
      max_duration             = optional(string, "")
      max_messages             = optional(number, 0)
      avro_config = optional(object({
        use_topic_schema = optional(bool, false)
        write_metadata   = optional(bool, false)
      }), null)
      service_account_email = optional(string, "")
    }), null)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key_base64 = optional(string, "")
  })
  default = { service_account_key_base64 = "" }
}
