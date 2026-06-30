variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Front Door Profile specification"
  type = object({
    resource_group           = string
    name                     = string
    sku                      = optional(string, "Standard_AzureFrontDoor")
    response_timeout_seconds = optional(number, 120)

    endpoints = optional(list(object({
      name    = string
      enabled = optional(bool, true)
    })), [])

    origin_groups = optional(list(object({
      name                     = string
      session_affinity_enabled = optional(bool, true)
      load_balancing = optional(object({
        sample_size                        = optional(number, 4)
        successful_samples_required        = optional(number, 3)
        additional_latency_in_milliseconds = optional(number, 50)
      }), {})
      health_probe = optional(object({
        protocol            = string
        path                = optional(string, "/")
        request_type        = optional(string, "HEAD")
        interval_in_seconds = number
      }))
      origins = optional(list(object({
        name                           = string
        host_name                      = string
        certificate_name_check_enabled = optional(bool, true)
        origin_host_header             = optional(string)
        http_port                      = optional(number, 80)
        https_port                     = optional(number, 443)
        priority                       = optional(number, 1)
        weight                         = optional(number, 500)
        enabled                        = optional(bool, true)
        private_link = optional(object({
          location               = string
          private_link_target_id = string
          request_message        = optional(string, "Access request for CDN FrontDoor Private Link Origin")
          target_type            = optional(string)
        }))
      })), [])
    })), [])

    routes = optional(list(object({
      name                   = string
      endpoint_name          = string
      origin_group_name      = string
      patterns_to_match      = list(string)
      supported_protocols    = list(string)
      forwarding_protocol    = optional(string, "MatchRequest")
      https_redirect_enabled = optional(bool, true)
      link_to_default_domain = optional(bool, true)
      enabled                = optional(bool, true)
      cache = optional(object({
        query_string_caching_behavior = optional(string, "IgnoreQueryString")
        query_strings                 = optional(list(string), [])
        compression_enabled           = optional(bool, false)
        content_types_to_compress     = optional(list(string), [])
      }))
    })), [])
  })
}
