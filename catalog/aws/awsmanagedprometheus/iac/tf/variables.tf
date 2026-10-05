variable "metadata" {
  description = "Catalog object metadata"
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
  description = "AwsManagedPrometheus specification"
  type = object({
    region = string
    alias = optional(string, "")
    kms_key_arn = optional(string, "")
    logging = optional(object({
      log_group_arn = string
    }))
    configuration = optional(object({
      retention_period_in_days = optional(number)
      out_of_order_time_window_in_seconds = optional(number)
      rule_query_offset_in_seconds = optional(number)
      limits_per_label_set = optional(list(object({
        label_set = optional(map(string), {})
        max_series = optional(number, 0)
      })), [])
    }))
    alert_manager_definition = optional(string, "")
    rule_group_namespaces = optional(list(object({
      name = string
      data = string
    })), [])
    query_logging = optional(object({
      destinations = list(object({
        log_group_arn = string
        qsp_threshold = optional(number, 0)
      }))
    }))
    resource_policy = optional(any)
    anomaly_detectors = optional(list(object({
      alias = string
      query = string
      evaluation_interval_in_seconds = optional(number)
      labels = optional(map(string), {})
      sample_size = optional(number)
      shingle_size = optional(number)
      ignore_near_expected_from_above = optional(object({
        amount = optional(number)
        ratio = optional(number)
      }))
      ignore_near_expected_from_below = optional(object({
        amount = optional(number)
        ratio = optional(number)
      }))
      missing_data_action = optional(string, "")
    })), [])
  })
}