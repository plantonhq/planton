variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud NLB Load Balancer specification"
  type = object({
    region             = string
    vpc_id             = string
    load_balancer_name = optional(string, "")
    address_type       = optional(string, "Internet")
    resource_group_id  = optional(string, "")
    cross_zone_enabled = optional(bool, true)
    tags               = optional(map(string), {})

    zone_mappings = list(object({
      zone_id       = string
      vswitch_id    = string
      allocation_id = optional(string, "")
    }))

    server_groups = optional(list(object({
      name                       = string
      protocol                   = optional(string, "TCP")
      scheduler                  = optional(string, "Wrr")
      connection_drain_enabled   = optional(bool, false)
      connection_drain_timeout   = optional(number, 10)
      preserve_client_ip_enabled = optional(bool, true)
      health_check = object({
        health_check_enabled         = bool
        health_check_type            = optional(string, "TCP")
        health_check_connect_port    = optional(number, 0)
        health_check_connect_timeout = optional(number, 5)
        health_check_interval        = optional(number, 10)
        healthy_threshold            = optional(number, 2)
        unhealthy_threshold          = optional(number, 2)
        health_check_url             = optional(string, "")
        health_check_domain          = optional(string, "")
        http_check_method            = optional(string, "GET")
        health_check_http_codes      = optional(list(string), [])
      })
    })), [])

    listeners = optional(list(object({
      listener_port          = number
      listener_protocol      = string
      server_group_name      = string
      listener_description   = optional(string, "")
      idle_timeout           = optional(number, 900)
      proxy_protocol_enabled = optional(bool, false)
      certificate_ids        = optional(list(string), [])
      security_policy_id     = optional(string, "")
      ca_certificate_ids     = optional(list(string), [])
      ca_enabled             = optional(bool, false)
    })), [])
  })

  validation {
    condition     = length(var.spec.zone_mappings) >= 2
    error_message = "At least 2 zone_mappings are required for NLB high availability."
  }

  validation {
    condition     = contains(["Internet", "Intranet"], var.spec.address_type)
    error_message = "address_type must be one of: Internet, Intranet."
  }
}
