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
  description = "AzureMonitorDataCollectionRule specification"
  type = object({
    resource_group = string
    name           = string
    region         = string

    # Omit for the default rule kind (all platforms). Once set, changing
    # it forces a new rule.
    kind = optional(string)

    description = optional(string, "")

    # Resolved ARM id of the Data Collection Endpoint, when the rule
    # ingests through one (required for custom streams).
    data_collection_endpoint_id = optional(string)

    # Managed identity: type is the spec enum's value name
    # (SYSTEM_ASSIGNED / USER_ASSIGNED); identity_ids carry resolved
    # ARM ids.
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))

    data_sources = optional(object({
      syslogs = optional(list(object({
        name           = string
        facility_names = list(string)
        log_levels     = list(string)
        streams        = list(string)
      })), [])
      performance_counters = optional(list(object({
        name                          = string
        sampling_frequency_in_seconds = number
        counter_specifiers            = list(string)
        streams                       = list(string)
      })), [])
      windows_event_logs = optional(list(object({
        name           = string
        x_path_queries = list(string)
        streams        = list(string)
      })), [])
      extensions = optional(list(object({
        name               = string
        extension_name     = string
        extension_json     = optional(string, "")
        input_data_sources = optional(list(string), [])
        streams            = list(string)
      })), [])
      iis_logs = optional(list(object({
        name            = string
        log_directories = optional(list(string), [])
        streams         = list(string)
      })), [])
      log_files = optional(list(object({
        name          = string
        file_patterns = list(string)
        format        = string
        settings = optional(object({
          text = object({
            record_start_timestamp_format = string
          })
        }))
        streams = list(string)
      })), [])
      prometheus_forwarders = optional(list(object({
        name = string
        label_include_filters = optional(list(object({
          label = string
          value = string
        })), [])
        streams = list(string)
      })), [])
      windows_firewall_logs = optional(list(object({
        name    = string
        streams = list(string)
      })), [])
      platform_telemetries = optional(list(object({
        name    = string
        streams = list(string)
      })), [])
      data_import = optional(object({
        event_hub_data_source = object({
          name           = string
          stream         = string
          consumer_group = optional(string, "")
        })
      }))
    }))

    destinations = object({
      log_analytics = optional(list(object({
        name                  = string
        workspace_resource_id = string
      })), [])
      azure_monitor_metrics = optional(object({
        name = string
      }))
      event_hub = optional(object({
        name         = string
        event_hub_id = string
      }))
      event_hub_direct = optional(object({
        name         = string
        event_hub_id = string
      }))
      monitor_accounts = optional(list(object({
        name               = string
        monitor_account_id = string
      })), [])
      storage_blobs = optional(list(object({
        name               = string
        container_name     = string
        storage_account_id = string
      })), [])
      storage_blob_directs = optional(list(object({
        name               = string
        container_name     = string
        storage_account_id = string
      })), [])
      storage_table_directs = optional(list(object({
        name               = string
        table_name         = string
        storage_account_id = string
      })), [])
    })

    data_flows = list(object({
      streams            = list(string)
      destinations       = list(string)
      built_in_transform = optional(string, "")
      output_stream      = optional(string, "")
      transform_kql      = optional(string, "")
    }))

    stream_declarations = optional(list(object({
      stream_name = string
      columns = list(object({
        name = string
        type = string
      }))
    })), [])

    tags = optional(map(string), {})
  })
}
