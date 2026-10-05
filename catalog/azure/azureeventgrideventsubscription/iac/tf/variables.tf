variable "metadata" {
  description = "Catalog object metadata"
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
  description = "AzureEventgridEventSubscription specification"
  type = object({
    # Addressing: exactly one of scope / system_topic_id (the spec's
    # CEL enforces it before the module runs).
    scope           = optional(string)
    system_topic_id = optional(string)

    name = string

    # Exactly one destination arm (spec-enforced).
    destination = object({
      azure_function = optional(object({
        function_id                       = string
        max_events_per_batch              = optional(number)
        preferred_batch_size_in_kilobytes = optional(number)
      }))
      eventhub_id          = optional(string)
      hybrid_connection_id = optional(string)
      service_bus_queue_id = optional(string)
      service_bus_topic_id = optional(string)
      storage_queue = optional(object({
        storage_account_id                    = string
        queue_name                            = string
        queue_message_time_to_live_in_seconds = optional(number)
      }))
      webhook = optional(object({
        url                               = string
        max_events_per_batch              = optional(number)
        preferred_batch_size_in_kilobytes = optional(number)
        active_directory_tenant_id        = optional(string, "")
        active_directory_app_id_or_uri    = optional(string, "")
      }))
    })

    delivery_identity = optional(object({
      type                   = string
      user_assigned_identity = optional(string, "")
    }))

    delivery_properties = optional(list(object({
      header_name  = string
      type         = string
      value        = optional(string, "")
      source_field = optional(string, "")
      secret       = optional(bool, false)
    })), [])

    dead_letter = optional(object({
      storage_account_id          = string
      storage_blob_container_name = string
    }))

    dead_letter_identity = optional(object({
      type                   = string
      user_assigned_identity = optional(string, "")
    }))

    # The envelope events are delivered in; the platform default
    # (EventGridSchema) mirrors Azure's own.
    event_delivery_schema = optional(string, "EventGridSchema")

    included_event_types = optional(list(string), [])

    subject_filter = optional(object({
      subject_begins_with = optional(string, "")
      subject_ends_with   = optional(string, "")
      case_sensitive      = optional(bool)
    }))

    advanced_filter = optional(object({
      # Proto3 zero values (false / 0) are omitted by the tfvars
      # pipeline -- the defaults below restore them, mirroring proto3
      # semantics exactly (absent IS the zero value).
      bool_equals = optional(list(object({
        key   = string
        value = optional(bool, false)
      })), [])
      number_greater_than = optional(list(object({
        key   = string
        value = optional(number, 0)
      })), [])
      number_greater_than_or_equals = optional(list(object({
        key   = string
        value = optional(number, 0)
      })), [])
      number_less_than = optional(list(object({
        key   = string
        value = optional(number, 0)
      })), [])
      number_less_than_or_equals = optional(list(object({
        key   = string
        value = optional(number, 0)
      })), [])
      number_in = optional(list(object({
        key    = string
        values = list(number)
      })), [])
      number_not_in = optional(list(object({
        key    = string
        values = list(number)
      })), [])
      number_in_range = optional(list(object({
        key = string
        ranges = list(object({
          from = optional(number, 0)
          to   = optional(number, 0)
        }))
      })), [])
      number_not_in_range = optional(list(object({
        key = string
        ranges = list(object({
          from = optional(number, 0)
          to   = optional(number, 0)
        }))
      })), [])
      string_begins_with = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      string_not_begins_with = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      string_ends_with = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      string_not_ends_with = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      string_contains = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      string_not_contains = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      string_in = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      string_not_in = optional(list(object({
        key    = string
        values = list(string)
      })), [])
      is_not_null = optional(list(object({
        key = string
      })), [])
      is_null_or_undefined = optional(list(object({
        key = string
      })), [])
    }))

    advanced_filtering_on_arrays_enabled = optional(bool, false)

    labels = optional(list(string), [])

    expiration_time_utc = optional(string, "")

    retry_policy = optional(object({
      max_delivery_attempts = number
      event_time_to_live    = number
    }))
  })
}
