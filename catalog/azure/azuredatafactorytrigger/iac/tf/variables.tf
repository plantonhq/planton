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
  description = "AzureDataFactoryTrigger specification"
  type = object({
    data_factory_id = string
    name            = string
    description     = optional(string, "")
    annotations     = optional(list(string), [])
    activated       = optional(bool)
    schedule = optional(object({
      frequency  = optional(string)
      interval   = optional(number)
      start_time = optional(string, "")
      end_time   = optional(string, "")
      time_zone  = optional(string, "")
      recurrence_schedule = optional(object({
        days_of_month = optional(list(number), [])
        days_of_week  = optional(list(string), [])
        hours         = optional(list(number), [])
        minutes       = optional(list(number), [])
        monthly = optional(list(object({
          weekday = string
          week    = optional(number)
        })), [])
      }))
      pipelines = list(object({
        name       = string
        parameters = optional(map(string), {})
      }))
    }))
    tumbling_window = optional(object({
      frequency       = string
      interval        = optional(number, 0)
      start_time      = string
      end_time        = optional(string, "")
      delay           = optional(string, "")
      max_concurrency = optional(number)
      retry = optional(object({
        count    = optional(number, 0)
        interval = optional(number)
      }))
      dependencies = optional(list(object({
        trigger_name = optional(string, "")
        offset       = optional(string, "")
        size         = optional(string, "")
      })), [])
      additional_properties = optional(map(string), {})
      pipeline = object({
        name       = string
        parameters = optional(map(string), {})
      })
    }))
    blob_event = optional(object({
      storage_account_id    = string
      events                = list(string)
      blob_path_begins_with = optional(string, "")
      blob_path_ends_with   = optional(string, "")
      ignore_empty_blobs    = optional(bool)
      additional_properties = optional(map(string), {})
      pipelines = list(object({
        name       = string
        parameters = optional(map(string), {})
      }))
    }))
    custom_event = optional(object({
      eventgrid_topic_id    = string
      events                = list(string)
      subject_begins_with   = optional(string, "")
      subject_ends_with     = optional(string, "")
      additional_properties = optional(map(string), {})
      pipelines = list(object({
        name       = string
        parameters = optional(map(string), {})
      }))
    }))
  })
}