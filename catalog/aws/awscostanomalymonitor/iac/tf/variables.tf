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
  description = "AwsCostAnomalyMonitor specification"
  type = object({
    region                = string
    monitor_name          = string
    monitor_type          = optional(string, "")
    monitor_dimension     = optional(string, "")
    monitor_specification = optional(any)
    subscriptions = optional(list(object({
      name      = string
      frequency = optional(string, "")
      subscribers = list(object({
        address = string
        type    = optional(string, "")
      }))
      threshold_expression = optional(object({
        dimension = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = list(string)
        }))
        tag = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        cost_category = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        and = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        or = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        not = optional(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        }))
      }))
    })), [])
  })
}