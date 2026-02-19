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
  description = "Alibaba Cloud ALB Load Balancer specification"
  type = object({
    region                 = string
    vpc_id                 = string
    load_balancer_name     = optional(string, "")
    address_type           = optional(string, "Internet")
    load_balancer_edition  = optional(string, "Standard")
    resource_group_id      = optional(string, "")
    tags                   = optional(map(string), {})

    zone_mappings = list(object({
      zone_id    = string
      vswitch_id = string
    }))

    access_log_config = optional(object({
      log_project = string
      log_store   = string
    }))

    server_groups = optional(list(object({
      name      = string
      protocol  = optional(string, "HTTP")
      scheduler = optional(string, "Wrr")
      health_check_config = object({
        health_check_enabled      = bool
        health_check_protocol     = optional(string, "HTTP")
        health_check_path         = optional(string, "")
        health_check_host         = optional(string, "")
        health_check_method       = optional(string, "HEAD")
        health_check_connect_port = optional(number, 0)
        health_check_interval     = optional(number, 2)
        health_check_timeout      = optional(number, 5)
        healthy_threshold         = optional(number, 3)
        unhealthy_threshold       = optional(number, 3)
        health_check_codes        = optional(list(string), [])
      })
      sticky_session_config = optional(object({
        sticky_session_enabled = bool
        sticky_session_type    = optional(string, "")
        cookie                 = optional(string, "")
        cookie_timeout         = optional(number, 1000)
      }))
    })), [])

    listeners = optional(list(object({
      listener_port                    = number
      listener_protocol                = string
      default_action_server_group_name = string
      listener_description             = optional(string, "")
      certificate_id                   = optional(string, "")
      security_policy_id               = optional(string, "")
      gzip_enabled                     = optional(bool, true)
      http2_enabled                    = optional(bool, true)
      idle_timeout                     = optional(number, 60)
      request_timeout                  = optional(number, 60)
    })), [])
  })

  validation {
    condition     = length(var.spec.zone_mappings) >= 2
    error_message = "At least 2 zone_mappings are required for ALB high availability."
  }

  validation {
    condition     = contains(["Internet", "Intranet"], var.spec.address_type)
    error_message = "address_type must be one of: Internet, Intranet."
  }

  validation {
    condition     = contains(["Basic", "Standard", "StandardWithWaf"], var.spec.load_balancer_edition)
    error_message = "load_balancer_edition must be one of: Basic, Standard, StandardWithWaf."
  }
}
