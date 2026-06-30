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
  description = "Azure Application Gateway specification"
  type = object({
    region         = string
    resource_group = string
    name           = string
    subnet_id      = string
    public_ip_id   = string
    sku            = string

    capacity  = optional(number, 2)
    autoscale = optional(object({
      min_capacity = number
      max_capacity = optional(number)
    }))

    backend_address_pools = list(object({
      name         = string
      fqdns        = optional(list(string), [])
      ip_addresses = optional(list(string), [])
    }))

    backend_http_settings = list(object({
      name                                = string
      port                                = number
      protocol                            = string
      cookie_based_affinity               = optional(string, "Disabled")
      request_timeout                     = optional(number, 30)
      probe_name                          = optional(string)
      host_name                           = optional(string)
      pick_host_name_from_backend_address = optional(bool, false)
    }))

    http_listeners = list(object({
      name                 = string
      port                 = number
      protocol             = string
      host_name            = optional(string)
      ssl_certificate_name = optional(string)
    }))

    request_routing_rules = list(object({
      name                        = string
      http_listener_name          = string
      backend_address_pool_name   = string
      backend_http_settings_name  = string
      priority                    = number
    }))

    probes = optional(list(object({
      name                = string
      protocol            = string
      path                = string
      host                = optional(string)
      interval            = optional(number, 30)
      timeout             = optional(number, 30)
      unhealthy_threshold = optional(number, 3)
    })), [])

    ssl_certificates = optional(list(object({
      name                 = string
      key_vault_secret_id  = string
    })), [])

    identity_ids = optional(list(string), [])

    waf_enabled  = optional(bool, false)
    waf_mode     = optional(string, "Prevention")
    enable_http2 = optional(bool, false)
  })
}
