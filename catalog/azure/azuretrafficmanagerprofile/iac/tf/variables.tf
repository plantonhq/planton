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
  description = "AzureTrafficManagerProfile specification"
  type = object({
    resource_group = string
    name           = string
    routing_method = string
    dns_config = object({
      relative_name = string
      ttl_seconds   = optional(number)
    })
    monitor_config = object({
      protocol                     = string
      port                         = number
      path                         = optional(string, "")
      interval_in_seconds          = optional(number)
      timeout_in_seconds           = optional(number)
      tolerated_number_of_failures = optional(number)
      expected_status_code_ranges  = optional(list(string), [])
      custom_headers = optional(list(object({
        name  = string
        value = string
      })), [])
    })
    enabled              = optional(bool)
    max_return           = optional(number)
    traffic_view_enabled = optional(bool, false)
    tags                 = optional(map(string), {})
  })
}
