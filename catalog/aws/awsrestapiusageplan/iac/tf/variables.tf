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
  description = "AwsRestApiUsagePlan specification"
  type = object({
    region       = string
    description  = optional(string, "")
    product_code = optional(string, "")
    api_stages = optional(list(object({
      rest_api_id = string
      stage_name  = string
      method_throttles = optional(list(object({
        path        = string
        burst_limit = optional(number, 0)
        rate_limit  = optional(number, 0)
      })), [])
    })), [])
    quota = optional(object({
      limit  = optional(number, 0)
      period = optional(string, "")
      offset = optional(number, 0)
    }))
    throttle = optional(object({
      burst_limit = optional(number, 0)
      rate_limit  = optional(number, 0)
    }))
    api_keys = optional(list(object({
      name        = string
      description = optional(string, "")
      enabled     = optional(bool)
      customer_id = optional(string, "")
      value       = optional(string, "")
    })), [])
  })
}