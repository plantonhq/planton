variable "metadata" {
  description = "Resource metadata (name, org, env, id, labels)"
  type = object({
    name = string
    org  = optional(string, "")
    env  = optional(string, "")
    id   = optional(string, "")
  })
}

variable "spec" {
  description = "AwsGlobalAccelerator spec"
  type = object({
    enabled           = optional(bool, true)
    ip_address_type   = optional(string, "IPV4")
    ip_addresses      = optional(list(string), [])
    flow_logs = optional(object({
      enabled   = optional(bool, false)
      s3_bucket = optional(string, "")
      s3_prefix = optional(string, "")
    }), null)
    listeners = list(object({
      name             = string
      protocol         = string
      client_affinity  = optional(string, "NONE")
      port_ranges = list(object({
        from_port = number
        to_port   = number
      }))
      endpoint_groups = list(object({
        name                           = string
        endpoint_group_region          = optional(string, "")
        health_check_port              = optional(number, null)
        health_check_protocol          = optional(string, "TCP")
        health_check_path              = optional(string, "")
        health_check_interval_seconds  = optional(number, 30)
        threshold_count                = optional(number, 3)
        traffic_dial_percentage        = optional(number, 100.0)
        endpoints = optional(list(object({
          endpoint_id                    = string
          weight                         = optional(number, 128)
          client_ip_preservation_enabled = optional(bool, false)
        })), [])
        port_overrides = optional(list(object({
          listener_port = number
          endpoint_port = number
        })), [])
      }))
    }))
  })
}

variable "provider_config" {
  description = "AWS provider configuration"
  type = object({
    region            = string
    access_key_id     = optional(string, "")
    secret_access_key = optional(string, "")
    session_token     = optional(string, "")
  })
}
